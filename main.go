// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const (
	// The default "hidden" directory where options files reside.
	//
	DefaultDccDir = ".dcc"

	// The default "hidden" directory where compiler-generated
	// dependency files reside.
	//
	DefaultDepsDir = ".dcc.d"
)

var (
	// Myname is the program's invocation name. We use it to
	// prefix log messages and check it to see if we should
	// work as a C++ compiler driver.
	//
	Myname string

	// DccCurrentDirectory is the path of the process working directory at startup
	//
	DccCurrentDirectory = MustGetwd()

	// IgnoreDependencies has dcc not perform any dependency checking
	// and makes it assume everything is out of date, forcing everything
	// to be rebuilt.
	//
	// This is set by the --force command line option.
	//
	IgnoreDependencies bool

	// ActualCompiler is the Compiler (compiler.go) for the real
	// compiler executable. The Compiler type abstracts the
	// functions of the underlying compiler and lets dcc work with
	// different compilers, i.e. Microsoft's cl.exe requires
	// different handling to the gcc-style compilers.
	//
	ActualCompiler Compiler

	// DefaultNumJobs is the number of concurrent compilations
	// dcc will perform by default.  The "2 + numcpu" is the
	// same value used by the ninja build tool.
	//
	DefaultNumJobs = 2 + runtime.NumCPU()

	// NumJobs is the number of concurrent compilations dcc will
	// perform. By default this is two times the number of CPUs.
	// Why? Well basically because that's what works well for me.
	// The old rule-of-thumb used to be 1+NCPU but we do more I/O
	// in our builds these days (C++ especially) so there's more
	// I/O to overlap. Ninja uses 2+NCPU by default and dcc
	// used to do that but 2*NCPU works better for what I build
	// on my machines (not so modern C++, SSDs, 4/8 cores).
	//
	NumJobs = GetenvInt("NUMJOBS", DefaultNumJobs)

	// DccDir is the name of the directory where dcc-related files
	// are stored. If this directory does not exist the current
	// directory is used. DccDir defaults to ".dcc" allowing users
	// to "hide" the various build files by putting them in a "dot"
	// directory (which makes for cleaner source directories).
	//
	DccDir = Getenv("DCCDIR", DefaultDccDir)

	// DepsDir is the name of the directory where the per-object
	// file dependency files (.d files) are stored. This is usually
	// ".dcc.d" except on windows.
	//
	DepsDir = Getenv("DEPSDIR", DefaultDepsDir)

	// ObjsDir is the name of the directory where object files
	// are written. It is usually a relative path, relative to
	// the source file.
	//
	ObjsDir = "."

	// Quiet disables most messages when true. Usually dcc will
	// print, to stderr, the commands being executed.
	//
	Quiet = false

	// Verbose, when true, enables various messages from dcc.
	//
	Verbose = false

	// Debug, when true, enables debug output.
	//
	Debug = false

	// Debug "FindFile" operations when true.
	//
	DebugFind = false

	// CCFILE is the name of the "options file" that defines
	// the name of the C compiler.
	//
	CCFILE = Getenv("CCFILE", "CC")

	// CXXFILE is the name of the "options file" that defines
	// the name of the C++ compiler.
	//
	CXXFILE = Getenv("CXXFILE", "CXX")

	// CFLAGSFILE is the name of the "options file" that
	// defines the options passed to the C compiler.
	//
	CFLAGSFILE = Getenv("CFLAGSFILE", "CFLAGS")

	// CXXFLAGSFILE is the name of the "options file" that
	// defines the options passed to the C++ compiler.
	//
	CXXFLAGSFILE = Getenv("CXXFLAGSFILE", "CXXFLAGS")

	// LDFLAGSFILE is the name of the "options file" that defines
	// the options passed to the linker (aka "loader").
	//
	LDFLAGSFILE = Getenv("LDFLAGSFILE", "LDFLAGS")

	// LIBSFILE is the name of the "options file" that defines
	// the names of the libraries used when linking.
	//
	LIBSFILE = Getenv("LIBSFILE", "LIBS")
)

func main() {
	Myname = GetProgramName()
	ConfigureLogger(Myname)

	// Enable debug as early as possible.

	if s := os.Getenv("DCCDEBUG"); s != "" {
		if s == "find" {
			DebugFind = true
		} else if s == "all" {
			Debug = true
			DebugFind = true
		} else {
			Debug = true
		}
	} else {
		for i := 1; i < len(os.Args); i++ {
			if os.Args[i] == "--debug" {
				Debug = true
			}
			if os.Args[i] == "--debug-find" {
				DebugFind = true
			}
		}
	}

	if !Debug {
		defer CatchPanics()
	}

	runningMode := ModeNotSpecified
	outputPathname := ""
	dasho := ""
	dashm := ""
	writeCompileCommands := false
	appendCompileCommands := false

	cCompiler := makeCompilerOption(CCFILE, platform.DefaultCC)
	cppCompiler := makeCompilerOption(CXXFILE, platform.DefaultCXX)

	// We assume we're compiling C and define a function to switch
	// things to C++. We'll check for that next.
	//
	underlyingCompiler, optionsFilename := cCompiler, CFLAGSFILE
	cplusplus := func() {
		underlyingCompiler, optionsFilename = cppCompiler, CXXFLAGSFILE
	}

	// Rules for deciding if we're need to compile with C++:
	//
	// - our invocation name ends in "++"
	// - the --cpp option was supplied
	// - any of the input files are C++ source files
	//
	if strings.HasSuffix(Myname, "++") {
		cplusplus()
	} else {
		for i := 1; i < len(os.Args); i++ {
			if os.Args[i] == "--cpp" {
				cplusplus()
				break
			}
			if os.Args[i][0] != '-' && IsCPlusPlusFile(os.Args[i]) {
				cplusplus()
				break
			}
		}
	}

	// The state...
	//
	// Options can be read from files, other state variables are
	// set by interpreting command line options. The
	// sourceFileIndex is explained below, basically it defines
	// which files in inputFilenames are the names of source files.
	//
	var (
		err             error
		compilerOptions = new(Options)
		linkerOptions   = new(Options)
		libraryFiles    = new(Options)
		otherFiles      = new(Options)
		libraryDirs     = make([]string, 0)
		inputFilenames  = make([]string, 0)
		sourceFilenames = make([]string, 0)
		sourceFileIndex = make([]int, 0) // position of sourceFilenames[x] within inputFilenames
		frameworks      = make([]string, 0)
	)

	// The default set of libraryDirs comes from the platform's standard directories.
	//
	libraryDirs = append(libraryDirs, platform.LibraryPaths...)

	// Get compiler options from the options file.
	//
	if path, found := FindFile(optionsFilename); found {
		_, err := compilerOptions.ReadFromFile(path, nil)
		if err != nil {
			log.Fatal(err)
		}
	}

	// A -o in an options file is problematic, we keep that information elsewhere.
	// So we have to look for a -o and transfer it's value to our variable and
	// remove it from the options.
	//
	if index := compilerOptions.OptionIndex("-o"); index != -1 {
		if index == len(compilerOptions.Values)-1 {
			log.Fatalf("invalid -o option in compiler options file %q", optionsFilename)
		}
		dasho = compilerOptions.Values[index+1]
		compilerOptions.Values = append(compilerOptions.Values[0:index], compilerOptions.Values[index+2:]...)
	}

	// We use the compilerOptions modtime as a dependency but in
	// the unlikely case the compiler's modtime is newer we'll use
	// that instead - automatic rebuilds when the compiler
	// changes (v.unsafe - ref. Thompson's Reflections on Trust).
	//
	compilerOptions.SetModTime(MostRecentModTime(compilerOptions, underlyingCompiler))

	// Now we do the same for the linker options. We use the "LDFLAGS" file.
	//
	if path, found := FindFile(LDFLAGSFILE); found {
		_, err = linkerOptions.ReadFromFile(path, func(s string) string {
			// Collect directories named via -L in libraryDirs
			if strings.HasPrefix(s, "-L") {
				libraryDirs = append(libraryDirs, s[2:])
			}
			return s
		})
	}
	setLinkOpts := false
	if os.IsNotExist(err) {
		setLinkOpts = true
		// remember that we need to set the linker options
		// from the compiler options ... but only when
		// linking an executable.
		// linkerOptions.SetFrom(compilerOptions)
	} else if err != nil {
		log.Fatal(err)
	}

	// Libraries
	//
	if err = readLibs(LIBSFILE, libraryFiles, &libraryDirs, &frameworks); err != nil {
		log.Fatal(err)
	}
	if Debug {
		log.Printf("LIBS: %d libraries, %d dirs, %d frameworks", libraryFiles.Len(), len(libraryDirs), len(frameworks))
		log.Println("LIBS:", libraryFiles.String())
		log.Printf("LIBS: paths %v", libraryDirs)
		log.Printf("LIBS: frameworks %v", frameworks)
	}

	// Helper to set our running mode, once and once only.
	//
	setMode := func(arg string, mode RunningMode) {
		if runningMode != ModeNotSpecified {
			fmt.Fprintf(os.Stderr, "%s: a running mode has already been defined\n\n", arg)
			UsageError(os.Stderr, 1)
		}
		runningMode = mode
	}

	// Helper to take a filename supplied on the command line
	// and incorporate it in the appropriate place according
	// to its type.
	//
	collectInputFile := func(path string) {
		inputFilenames = append(inputFilenames, path)
		switch {
		case FileWillBeCompiled(path):
			sourceFilenames = append(sourceFilenames, path)
			sourceFileIndex = append(sourceFileIndex, len(inputFilenames)-1)

		case IsLibraryFile(path):
			libraryFiles.Append(path)

		default:
			otherFiles.Append(path)
		}
	}

	// Helper to issue a warning message re. the -o option being set more
	// that once and that we accept the first -o to set an output path
	// and ignore subsequent ones.
	//
	warnOutputPathAlreadySet := func(ignoredPath string) {
		log.Printf("warning: output pathname has already been set, ignoring this -o (%s)", ignoredPath)
	}

	// Windows is special... When using Microsoft's "cl.exe"
	// as the underlying compiler, on Windows, we need to support
	// its command line option syntax which allows for '/' as an
	// option specifier.
	//
	// Likewise darwin (MacOS) is special in that it has the
	// concept of frameworks.
	//
	windows := runtime.GOOS == "windows"
	macos := runtime.GOOS == "darwin"

	// Go over the command line and collect option and the names of
	// source files to be compiled and/or files to pass to the linker.
	//
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--help":
			UsageError(os.Stdout, 0)

		case windows && arg[0] == '/':
			// Compiler option for cl.exe
			//
			switch {
			case arg == "/c":
				setMode(arg, CompileSourceFiles)

			case strings.HasPrefix(arg, "/Fo"):
				dasho = arg[3:]

			case strings.HasPrefix(arg, "/Fe"):
				outputPathname = arg[3:]

			default:
				if _, err := Stat(arg); err != nil {
					compilerOptions.Append(arg)
					compilerOptions.Changed()
				} else {
					collectInputFile(arg)
				}
			}

		case arg[0] != '-':
			// Non-option arguments are filenames, possibly sources to compile or
			// files to pass to the linker - classical cc(1) behaviour.
			//
			collectInputFile(arg)

		case arg == "--write-compile-commands":
			writeCompileCommands = true

		case arg == "--append-compile-commands":
			appendCompileCommands = true

		case arg == "--cpp":
			break // ignore, handled above

		case arg == "--force":
			IgnoreDependencies = true

		case arg == "--quiet":
			Quiet = true

		case arg == "--verbose":
			Verbose = true

		case arg == "--debug":
			break // also handled above

		case arg == "--version":
			fmt.Print(versionNumber)
			os.Exit(0)

		case strings.HasPrefix(arg, "-j"):
			if arg == "-j" {
				NumJobs = runtime.NumCPU()
			} else if n, err := strconv.Atoi(arg[2:]); err != nil {
				log.Fatalf("%s: %s", arg, err)
			} else if n < 1 {
				log.Fatalf("%s: bad number of jobs", arg)
			} else {
				NumJobs = n
			}

		case arg == "--objdir":
			if i++; i < len(os.Args) {
				ObjsDir = os.Args[i]
			} else {
				log.Fatalf("%s: directory name required", arg)
			}

		case arg == "--exe":
			if i++; i < len(os.Args) {
				outputPathname = os.Args[i]
			} else {
				log.Fatalf("%s: program filename required", arg)
			}
			setMode(arg, CompileAndLink)

		case arg == "--lib":
			if i++; i < len(os.Args) {
				outputPathname = os.Args[i]
			} else {
				log.Fatalf("%s: library filename required", arg)
			}
			setMode(arg, CompileAndMakeLib)

		case arg == "--dll":
			if i++; i < len(os.Args) {
				outputPathname = os.Args[i]
			} else {
				log.Fatalf("%s: library filename required", arg)
			}
			setMode(arg, CompileAndMakeDLL)

		case arg == "--plugin":
			if i++; i < len(os.Args) {
				outputPathname = os.Args[i]
			} else {
				log.Fatalf("%s: library filename required", arg)
			}
			setMode(arg, CompileAndMakePlugin)

		case macos && arg == "-framework":
			libraryFiles.Append(arg)
			if i++; i < len(os.Args) {
				libraryFiles.Append(os.Args[i])
			}
			libraryFiles.Changed()

		case macos && (arg == "-macosx_version_min" || arg == "-macosx_version_max"):
			linkerOptions.Append(arg)
			if i++; i < len(os.Args) {
				linkerOptions.Append(os.Args[i])
			}
			linkerOptions.Changed()

		case macos && arg == "-bundle_loader":
			linkerOptions.Append(arg)
			if i++; i < len(os.Args) {
				linkerOptions.Append(os.Args[i])
			}
			linkerOptions.Changed()

		case arg == "-o":
			if i++; i < len(os.Args) {
				if outputPathname != "" {
					warnOutputPathAlreadySet(outputPathname)
				} else {
					dasho = os.Args[i]
				}
			} else {
				log.Fatal(arg + " pathname parameter required")
			}

		case strings.HasPrefix(arg, "-o"):
			if outputPathname != "" {
				warnOutputPathAlreadySet(outputPathname)
			} else {
				dasho = arg[2:]
			}

		case strings.HasPrefix(arg, "-L"):
			linkerOptions.Append(arg)
			linkerOptions.Changed()
			libraryDirs = append(libraryDirs, arg[2:])

		case strings.HasPrefix(arg, "-l"):
			libraryFiles.Prepend(arg)

		case arg == "-c":
			setMode(arg, CompileSourceFiles)

		case arg == "-m32":
			if dashm != "" && dashm != arg {
				log.Fatalf("conflicting options %s and %s", dashm, arg)
			}
			dashm = arg

		case arg == "-m64":
			if dashm != "" && dashm != arg {
				log.Fatalf("conflicting options %s and %s", dashm, arg)
			}
			dashm = arg

		default:
			compilerOptions.Append(arg)
			compilerOptions.Changed()
		}
	}

	// We have to at least have one filename to process. It doesn't
	// need to be a source file but we need something.
	//
	if len(inputFilenames) == 0 {
		UsageError(os.Stderr, 1)
	}

	// ----------------------------------------------------------------

	// If no mode was explicitly specified compile and link like cc(1).
	//
	if runningMode == ModeNotSpecified {
		runningMode = CompileAndLink
	}

	// Deal with any -m32/-m64 option
	//
	if dashm != "" {
		if platform.SelectTarget != nil {
			err = platform.SelectTarget(&platform, dashm)
		}
		if err != nil {
			log.Fatal(err)
		}
	}

	// Deal with any -o option.
	//
	// The cc(1) behaviour is:
	//
	// - without -c the -o's parameter names the executable
	// - -c -o<path> names the object file but is only permitted when compiling a single file
	//
	// Dcc should add:
	//
	// -c -o <path> permitted with multiple files if <path> names
	// a directory.
	//
	if dasho != "" && runningMode == CompileSourceFiles && len(sourceFilenames) > 1 {
		log.Fatal("-o <file> may not be supplied with -c and more one input file")
	}

	// outputPathname will be empty if no dcc-specific option that sets it
	// was used.  If a -o was supplied we propogate its value to outputPathname.
	//
	if outputPathname == "" && dasho != "" {
		outputPathname = dasho
		dasho = ""
	}

	// objdir is the object file directory.
	//
	objdir := dasho
	if objdir == "" {
		objdir = ObjsDir
	}

	// Next, replace any source file names with their object file
	// name in the inputFilenams slice. This is then the list of
	// files given to the linker, or librarian.  During this
	// replacement we replace any header file names with empty
	// strings as they are not used as inputs to the linker or
	// librarian. We do this so we don't invalidate the indices
	// stored in the sourceFileIndex slice during the loop. The
	// empty strings are removed once we've done this pass over
	// the names. And we only do all this if we need to since
	// pre-compiling headers is relatively rare.
	//
	removeEmptyNames := false
	for index, filename := range sourceFilenames {
		if IsSourceFile(filename) {
			inputFilenames[sourceFileIndex[index]] = ObjectFilename(filename, objdir)
		} else if IsHeaderFile(filename) {
			inputFilenames[sourceFileIndex[index]] = ""
			removeEmptyNames = true
		}
	}

	// NB. sourceFileIndex is no longer used/required and we're
	// possibly about to re-write th inputFilenames slice which
	// would invalidate the sourceFileIndex anyway so we nil it
	// to help detect now invalid accesses.
	//
	sourceFileIndex = nil

	// We typically don't do this. Empty strings only end up in
	// inputFilenames if a header file is being pre-compiled.
	//
	if removeEmptyNames {
		tmp := make([]string, 0, len(inputFilenames))
		for _, filename := range inputFilenames {
			if filename != "" {
				tmp = append(tmp, filename)
			}
		}
		inputFilenames = tmp
		// If there are no inputs remaining then we do not need to
		// link (or create a library).
		//
		if len(inputFilenames) == 0 {
			runningMode = CompileSourceFiles
		}
	}

	// Update any -l<name> library references in LibraryFiles with
	// the actual file path. We want to "stat" these files to determine
	// if they're newer than the executable/DLL that depends on them.
	//
	for index, name := range libraryFiles.Values {
		if strings.HasPrefix(name, "-l") {
			path, _, found, err := FindLibrary(libraryDirs, name[2:])
			switch {
			case err != nil:
				log.Fatal(err)
			case !found:
				log.Printf("warning: %q library not found on path %v", name, libraryDirs)
				// ... and we let the linker deal with it
			default:
				libraryFiles.Values[index] = path
				if Debug {
					log.Print("DEBUG LIB: '", name[2:], "' -> ", path)
				}
			}
		}
	}

	// Set the actual compiler we'll be using.
	//
	ActualCompiler = GetCompiler(underlyingCompiler.String())

	// Generate a compile_commands.json if requested.
	//
	if appendCompileCommands {
		if err := AppendCompileCommandsDotJson(filepath.Join(objdir, CompileCommandsFilename), sourceFilenames, compilerOptions, objdir); err != nil {
			log.Fatal(err)
		}
	} else if writeCompileCommands {
		if err := WriteCompileCommandsDotJson(filepath.Join(objdir, CompileCommandsFilename), sourceFilenames, compilerOptions, objdir); err != nil {
			log.Fatal(err)
		}
	}

	// And now we're ready to compile everything.
	//
	if !CompileAll(sourceFilenames, compilerOptions, objdir) {
		os.Exit(1)
	}

	// Then, if required, link an executable or DLL, or create a
	// static library.
	//
	switch runningMode {
	case CompileAndLink:
		if setLinkOpts && linkerOptions.Empty() {
			linkerOptions.SetFrom(compilerOptions)
		}
		err = Link(outputPathname, inputFilenames, libraryFiles, linkerOptions, otherFiles, frameworks)

	case CompileAndMakeDLL:
		err = Dll(outputPathname, inputFilenames, libraryFiles, linkerOptions, otherFiles, frameworks)

	case CompileAndMakePlugin:
		err = Plugin(outputPathname, inputFilenames, libraryFiles, linkerOptions, otherFiles, frameworks)

	case CompileAndMakeLib:
		err = Lib(outputPathname, inputFilenames)
	}

	// And that's it. Report any final error and exit.
	//
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	os.Exit(0)
}

// Helper function to create an Options (see options.go) that
// defines the underlying compiler, either CC and CXX depending
// upon mode.
//
// This function is given the name of the file and the default
// command name. If an environment variable with the same name
// as the file is set we use its value, or a default. Then we
// search for a, possibly platform and/or architecture
// specific, file with that name. If that file exists we use
// its contents.
//
// We need to do the above twice which is why its a function.
//
func makeCompilerOption(name, defcmd string) *Options {
	cmd := Getenv(name, defcmd)
	opts := new(Options)
	path, fileExists := FindFile(name)
	if fileExists {
		_, err := opts.ReadFromFile(path, nil)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		opts.Append(cmd)
	}
	if opts.Empty() {
		opts.Append(defcmd)
	}
	return opts
}

// UsageError outputs a program usage message to the given
// writer and exits the process with the given status.
//
func UsageError(w io.Writer, status int) {
	fmt.Fprintf(w, `Usage: %s [options] [compiler-options] filename...

Options, other than those listed below, are passed to the underlying
compiler. Any -c or -o and similar options are noted and used to control
linking and object file locations. Compilation is done in parallel using,
by default, as many jobs as there are CPUs and dependency files written
to a directory alongside the object files.

Options:
    --exe path	    Create executable program 'path'.
    --lib path	    Create static library 'path'.
    --dll path	    Create shared/dynamic library 'path'.
    --objdir path   Create object files in 'path'.
    -j[N]           Use 'N' compile jobs (note single dash, default is one per CPU).
    --cpp	    Compile source files as C++.
    --force         Ignore dependencies, always compile/link/lib.
    --clean         Remove dcc-maintained files.
    --quiet         Disable non-error messages.
    --verbose       Show more output.
    --debug         Enable debug messages.
    --version       Report dcc version and exit.
    --write-compile-commands
                    Output a compile_commands.json file.
    --append-compile-commands
                    Append to an existng compile_commands.json file
		    if it exists.

With anything else is passed to the underlying compiler.

Environment
    CC              C compiler (%s).
    CXX             C++ compiler (%s).
    DEPSDIR	    Name of .d file directory (%s).
    OBJDIR	    Name of .o file directory (%s).
    DCCDIR	    Name of the dcc-options directory (%s).
    NJOBS           Number of compile jobs (%d).

The following variables define the actual names used for
the options files (see "Files" below).

    CCFILE	    Name of the CC options file.
    CXXFILE	    Name of the CXX options files.
    CFLAGSFILE	    Name of the CFLAGS options file.
    CXXFLAGSFILE    Name of the CXXFLAGS options file.
    LDFLAGSFILE	    Name of the LDFLAGS options file.
    LIBSFILE	    Name of the LIBS options file.

Files
    CC              C compiler name.
    CXX             C++ compiler name.
    CFLAGS          Compiler options for C.
    CXXFLAGS        Compiler options for C++.
    LDFLAGS         Linker options.
    LIBS            Libraries and library paths.

Options files may be put in a directory named by the DCCDIR
environment variable, ".dcc" by default, or if that directory
does not exist, the current directory.

Options files are text files storing compiler and linker options.
The contents are turned into program options by removing #-style
line comments and using the whitespace separated words as individual
program arguments. Options files act as dependencies and invoke
rebuilds when changed.

Platform-specific file naming may be used to have different options
files for different platforms and/or architectures.  When opening a
file dcc appends extensions formed from the host OS and archtecture
to the file name and tries to use those files. The first extension
tried is of the form ".<os>_<arch>", followed by ".<os>" and finally
no extension.

<os> is something like "windows", "linux", "darwin", "freebsd".
<arch> is one of  "amd64", "386", "arm8le" , "s390x", etc...
The actual names come from the Go runtime.  This host is "%s_%s".

Environment variables may be used to define the names of the
different options files and select specific options.

`,
		Myname,
		platform.DefaultCC,
		platform.DefaultCXX,
		DefaultDepsDir,
		ObjsDir,
		DccDir,
		DefaultNumJobs,
		runtime.GOOS,
		runtime.GOARCH,
	)
	os.Exit(status)
}

// CatchPanics catches and reports panics. It is intended to be used
// at the top level of main to avoid printing unsightly stack traces.
//
func CatchPanics() {
	if x := recover(); x != nil {
		if err, ok := x.(error); ok {
			fmt.Fprintf(os.Stderr, "UNHANDLED ERROR: %s\n", err.Error())
		} else {
			fmt.Fprintf(os.Stderr, "PANIC: %v\n", x)
		}
		os.Exit(1)
	}
}

func readLibs(libsFile string, libraryFiles *Options, libraryDirs *[]string, frameworks *[]string) error {
	var frameworkDirs []string
	captureNext := false
	prevs := ""
	path, found := FindFile(libsFile)
	if !found {
		return nil
	}
	_, err := libraryFiles.ReadFromFile(path, func(s string) string {
		if Debug {
			log.Printf("LIBS: filter %q", s)
		}
		if captureNext {
			if prevs == "-L" {
				*libraryDirs = append(*libraryDirs, s)
			} else if prevs == "-F" {
				frameworkDirs = append(frameworkDirs, s)
			} else {
				*frameworks = append(*frameworks, s)
			}
			captureNext = false
			prevs = s
			return ""
		}
		if s == "-framework" {
			*frameworks = append(*frameworks, s)
			captureNext = true
			prevs = s
			return ""
		}
		if s == "-F" || s == "-L" {
			captureNext = true
			prevs = s
			return ""
		}
		if strings.HasPrefix(s, "-F") {
			frameworkDirs = append(frameworkDirs, s[2:])
			prevs = s
			return ""
		}
		if strings.HasPrefix(s, "-L") {
			*libraryDirs = append(*libraryDirs, s[2:])
			prevs = s
			return ""
		}
		if strings.HasPrefix(s, "-l") {
			if path, _, found, err := FindLibrary(*libraryDirs, s[2:]); err != nil {
				log.Printf("warning: failed to find %s: %s", *libraryDirs, err)
			} else if found {
				return path
			}
			prevs = s
		}
		return s
	})
	if err != nil {
		return err
	}
	for index := range frameworkDirs {
		frameworkDirs[index] = "-F" + frameworkDirs[index]
	}
	*frameworks = append(frameworkDirs, (*frameworks)...)
	return nil
}
