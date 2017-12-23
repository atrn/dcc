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
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var (
	// Myname is the program's invocation name. We use it to
	// prefix log messages and check it to see if we should
	// work as a C++ compiler driver.
	//
	Myname string

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

	// NumJobs is the number of concurrent compilations dcc will
	// perform. By default this is the number of available CPUs.
	//
	NumJobs int

	// DccDir is the name of the directory where dcc-related files
	// are stored. If this directory does not exist the current
	// is used in its place.
	//
	DccDir string

	// DepsDir is the name of the directory where the per-object
	// file dependency files (.d files) are stored.
	//
	DepsDir string

	// Quiet disables most messages when true. Usually dcc will
	// print, to stderr, the commands being executed.
	//
	Quiet = false

	// Debug enables debugging features if true.
	//
	Debug = false
)

func main() {
	Myname = GetProgramName()
	ConfigureLogger(Myname)

	// Enable debug as early as possible.

	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--debug" {
			Debug = true
		}
	}
	if !Debug {
		defer CatchPanics()
	}

	NumJobs = GetenvInt("NJOBS", runtime.NumCPU())
	DepsDir = Getenv("DCCDEPS", platform.DefaultDepsDir)
	DccDir = Getenv("DCCDIR", ".dcc")

	_, err := Stat(DccDir)
	if os.IsNotExist(err) {
		DccDir = "."
	} else if err != nil {
		log.Fatal(err)
	}

	runningMode := ModeNotSpecified
	outputPathname := ""
	dasho := ""

	cCompiler := makeCompilerOption("CC", platform.DefaultCC)
	cppCompiler := makeCompilerOption("CXX", platform.DefaultCXX)

	// We assume we're compiling C and define a function to switch
	// things to C++. Which we check for next.
	//
	underlyingCompiler, optionsFilename := cCompiler, "CFLAGS"
	cplusplus := func() {
		underlyingCompiler, optionsFilename = cppCompiler, "CXXFLAGS"
	}

	// Rules for deciding if we're need to compile with C++:
	//
	// - our invocation name ends in "++"
	// - the --cpp option was supplied
	// - any input file names are C++ source files
	//
	if strings.HasSuffix(Myname, "++") {
		cplusplus()
	} else {
		for i := 1; i < len(os.Args); i++ {
			if os.Args[i] == "--cpp" {
				cplusplus()
				break
			}
			if os.Args[i][0] != '-' && IsCPlusPlus(os.Args[i]) {
				cplusplus()
				break
			}
		}
	}

	var (
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
	_, err = compilerOptions.ReadFromFile(filepath.Join(DccDir, optionsFilename), nil)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	// A -o in an options file is problematic, we keep that information elsewhere.
	// So we have to look for a -o and transfer it's value to our variable and
	// remove it from the options.
	//
	if index := compilerOptions.FindFile("-o"); index != -1 {
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
	compilerOptions.SetModTime(MoreRecentOf(compilerOptions, underlyingCompiler))

	// Now we do the same for the linker options. We use the file "LDFLAGS".
	//
	_, err = linkerOptions.ReadFromFile(filepath.Join(DccDir, "LDFLAGS"), func(s string) string {
		// Collect directories named via -L in libraryDirs
		if strings.HasPrefix(s, "-L") {
			libraryDirs = append(libraryDirs, s[2:])
		}
		return s
	})
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

	if err = ReadLibs(libraryFiles, libraryDirs, frameworks); err != nil {
		log.Fatal(err)
	}

	// Helper to set our running mode and set it once.
	//
	setMode := func(arg string, mode RunningMode) {
		if runningMode != ModeNotSpecified {
			fmt.Fprintf(os.Stderr, "%s: a running mode has already been defined\n\n", arg)
			UsageError(1)
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

	// Windows is special... When using Microsoft's "cl.exe"
	// as the underlying compiler, on Windows, we need to support
	// its command line option syntax which allows for '/' as an
	// option specifier.
	//
	windows := runtime.GOOS == "windows"

	// Go over the command line and collect option and the names of
	// source files to be compiled and/or files to pass to the linker.
	//
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--help":
			UsageError(0)

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
				} else {
					collectInputFile(arg)
				}
			}

		case arg[0] != '-':
			// Non-option arguments are filenames, possibly sources to compile or
			// files to pass to the linker - classical cc(1) behaviour.
			//
			collectInputFile(arg)

		case arg == "--clean":
			setMode(arg, CleanupOutputFiles)

		case arg == "--cpp":
			break // ignore, handled above

		case arg == "--force":
			IgnoreDependencies = true

		case arg == "--quiet":
			Quiet = true

		case arg == "--debug":
			break // also handled above

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

		case runtime.GOOS == "darwin" && arg == "-framework":
			libraryFiles.Append(arg)
			if i++; i < len(os.Args) {
				libraryFiles.Append(os.Args[i])
			}

		case runtime.GOOS == "darwin" && (arg == "-macosx_version_min" || arg == "-macosx_version_max"):
			linkerOptions.Append(arg)
			if i++; i < len(os.Args) {
				linkerOptions.Append(os.Args[i])
			}

		case arg == "-o":
			if i++; i < len(os.Args) {
				if outputPathname != "" {
					log.Print("warning: output pathname has already been set, ignoring -o")
				} else {
					dasho = os.Args[i]
				}
			} else {
				log.Fatal(arg + " pathname parameter required")
			}

		case strings.HasPrefix(arg, "-o"):
			if outputPathname != "" {
				log.Print("warning: output pathname has already been set, ignoring -o")
			} else {
				dasho = arg[2:]
			}

		case strings.HasPrefix(arg, "-L"):
			linkerOptions.Append(arg)
			libraryDirs = append(libraryDirs, arg[2:])

		case strings.HasPrefix(arg, "-l"):
			libraryFiles.Append(arg)

		case arg == "-c":
			setMode(arg, CompileSourceFiles)

		default:
			compilerOptions.Append(arg)
		}
	}

	// We have to at least have one filename to process. It doesn't
	// need to be a source file but we need something.
	//
	if len(inputFilenames) == 0 {
		UsageError(1)
	}

	// ----------------------------------------------------------------

	// Cleaning overrides everything. We don't check anything and
	// just remove the entire depsdir then quit.
	//
	// TODO: only delete what we create
	//
	if runningMode == CleanupOutputFiles {
		exitStatus := 0
		os.Exit(exitStatus)
	}

	// If no mode was explicitly specified compile and link like cc(1).
	//
	if runningMode == ModeNotSpecified {
		runningMode = CompileAndLink
	}

	// Deal with any -o option.
	//
	// The cc(1) behaviour is:
	//
	// - without -c the -o's parameter names the executable
	// - -c -o<path> names the object file but is only permitted
	//   when compiling a single file
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

	//
	objdir := dasho
	if objdir == "" {
		objdir = "."
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
			path, _, found, err := FindLib(libraryDirs, name[2:])
			switch {
			case err != nil:
				log.Fatal(err)
			case !found:
				log.Printf("warning: %q library not found (%v)", name, libraryDirs)
				// ... and we let the linker deal with it
			default:
				libraryFiles.Values[index] = path
				if Debug {
					log.Print(" LIB: ", name[2:], " -> ", path)
				}
			}
		}
	}

	// Set the actual compiler we'll be using.
	//
	ActualCompiler = GetCompiler(underlyingCompiler.String())

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
// defines the underling compiler, either CC and CXX depending
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
	fileExists, err := opts.ReadFromFile(filepath.Join(DccDir, name), nil)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}
	if !fileExists {
		opts.Append(cmd)
	} else if opts.Empty() {
		opts.Append(defcmd)
	}
	return opts
}

// UsageError outputs a program usage message and exits the process.
//
func UsageError(status int) {
	fmt.Fprintf(os.Stderr, `Usage: %s [options] filename...

Options, other than those listed below, are passed to the underlying
compiler. Any -c or -o and similar options are noted and used to control
linking and object file locations. Compilation is done in parallel using,
by default, as many jobs as there are CPUs and dependency files written
to a .dcc directory alongside the object files.

Non-compiler Options:
    --exe path	Create executable program 'path'.
    --lib path	Create static library 'path'.
    --dll path	Create shared/dynamic library 'path'.
    -j[N]       Use 'N' compile jobs (note single dash, default is one per CPU).
    --cpp	Compile source files as C++.
    --force     Ignore dependencies, always compile/link/lib.
    --clean     Remove dcc-maintained files.
    --quiet     Disable non-error messages.
    --debug     Enable debug messages.

Environment
    CC          C compiler (%s).
    CXX         C++ compiler (%s).
    DCCDEPS	Name of .d file directory (%s).
    DCCDIR	Name of the dcc-options directory (%s).
    NJOBS       Number of compile job (%d).

Files
    CC          C compiler name.
    CXX         C++ compiler name.
    CFLAGS      Compiler options for C.
    CXXFLAGS    Compiler options for C++.
    LDFLAGS     Linker options.
    LIBS        Libraries and library paths.

Platform-specific filenames may be used which override
the generic files. A platform-specific filename has an
filename extenion of the form '.<os>' or '.<os>_<arch>'
where <os> is one of 'windows', 'darwin', 'linux', 'freebsd'
and <arch> the processor architecture, 'i386', 'amd64',
'arm32', 'arm64', etc... (the actual strings being defined
by the Go programming language implementation).
`,
		Myname,
		platform.DefaultCC,
		platform.DefaultCXX,
		platform.DefaultDepsDir,
		DccDir,
		runtime.NumCPU(),
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

// ReadLibs reads a "LIBS" file
//
func ReadLibs(libraryFiles *Options, libraryDirs []string, frameworks []string) error {
	captureNext := false
	_, err := libraryFiles.ReadFromFile(filepath.Join(DccDir, "LIBS"), func(s string) string {
		if captureNext {
			frameworks = append(frameworks, s)
			captureNext = false
			return ""
		}
		if s == "-framework" {
			frameworks = append(frameworks, s)
			captureNext = true
			return ""
		}
		if strings.HasPrefix(s, "-L") {
			libraryDirs = append(libraryDirs, s[2:])
			return ""
		}
		if strings.HasPrefix(s, "-l") {
			if path, _, found, err := FindLib(libraryDirs, s[2:]); err != nil {
				log.Fatal(err) // FIXME - don't fatal here
			} else if found {
				return path
			}
		}
		return s
	})
	if err != nil && os.IsNotExist(err) {
		err = nil
	}
	return err
}
