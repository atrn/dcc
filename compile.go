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
	"strings"

	"github.com/atrn/par"
)

// CompileAll compiles all of the given sources using the supplied
// options and returns true if all were sucessfully compiled.
//
// CompileAll compiles the sources in parallel using, up to, NJOBS
// parallel compilations.
//
func CompileAll(sources []string, options *Options, objdir string) (ok bool) {
	ok = true
	// The standard error output of each compile is routed via an
	// OutputMux which ensures output is not interleaved. We don't
	// bother with standard output yet but probably should.
	//
	mux := NewOutputMux(os.Stderr)
	go mux.Run()
	defer mux.Stop()

	// Our process structure is a simple fan-out that feeds the
	// names of the source files to a number of "workers" for
	// compilation.
	//
	// Source file names are fed to a 'filenames' channel which
	// is read by N worker tasks. Each worker reads a filename
	// from the channel, compiles the source file and sends an
	// error value result for compilation on the errors channel.
	// An error 'collecter' reads and reports any non-nil errors.
	//
	// Synchronization is done by par.DO and par.FOR.
	//
	filenames := make(chan string, len(sources))
	errs := make(chan error, len(sources))

	par.DO(
		func() {
			for _, filename := range sources {
				filenames <- filename
			}
			close(filenames)
		},
		func() {
			par.FOR(0, NumJobs, func(int) {
				for filename := range filenames {
					ofile := ObjectFilename(filename, objdir)
					stderr := mux.NewWriter()
					errs <- Compile(filename, options, ofile, stderr, objdir)
				}
			})
			close(errs)
		},
		func() {
			for err := range errs {
				if err != nil {
					log.Print(err)
					ok = false
				}
			}
		},
	)

	return
}

// Compile a single source file, if required. Returns a non-nil error if compilation fails.
//
func Compile(filename string, options *Options, ofile string, stderr io.WriteCloser, objdir string) error {
	defer stderr.Close()
	if ofile == "" {
		ofile = ObjectFilename(filename, objdir)
	}
	ofileDir := filepath.Dir(ofile)
	if err := Mkdir(ofileDir); err != nil {
		return err
	}
	depsFilename := DepsFilename(ofile)
	depsFileDir := filepath.Dir(depsFilename)
	if depsFileDir != ofileDir {
		if err := Mkdir(depsFileDir); err != nil {
			return err
		}
	}
	if !IgnoreDependencies {
		sourceInfo, err := Stat(filename)
		if err != nil {
			return err
		}
		target, deps, err := ActualCompiler.ReadDependencies(depsFilename)
		if os.IsNotExist(err) {
			// ignore
		} else if err != nil {
			return err
		} else if filepath.Base(target) != filepath.Base(ofile) {
			log.Printf("WARNING: got dependency target %q for object file %q", target, ofile)
		}
		if err == nil {
			uptodate, err := IsUptoDate(ofile, deps, sourceInfo, options)
			if err != nil || uptodate {
				return err
			}
		}
	}

	// Compile the file.
	//

	ClearCachedStat(ofile) // it will change

	// Do we need to output a command line? We don't output
	// the raw command as we'll add options to it. So we
	// prepare something similar for the user.
	//
	if !Quiet {
		var displayed []string
		if Verbose {
		    displayed = append(displayed, ActualCompiler.Name())
		    displayed = append(displayed, options.Values...)
		    displayed = append(displayed, filename)
		    if objdir != "" {
			    displayed = append(displayed, "-o", ofile)
		    }
		} else {
		    displayed = append(displayed, ActualCompiler.Name())
		    displayed = append(displayed, filename)
		}
		fmt.Fprintln(os.Stderr, strings.Join(displayed, " "))
	}

	return ActualCompiler.Compile(filename, ofile, depsFilename, options.Values, stderr)
}

// IsUptoDate determines if a given target file is up to date with
// respect to the input files, and compiler options, that led any
// previous generation of the target file.
//
func IsUptoDate(target string, deps []string, sourceInfo os.FileInfo, options *Options) (bool, error) {
	result := func(current bool, err error, caption string) (bool, error) {
		if Debug {
			how := "up to"
			if !current {
				how = "out of"
			}
			log.Printf("deps: %q (%q) -> (%s date, %v) - %s", sourceInfo.Name(), target, how, err, caption)
		}
		return current, err
	}

	UpToDate := func() (bool, error) {
		return result(true, nil, "target up to date")
	}
	OutOfDate := func(caption string) (bool, error) {
		return result(false, nil, caption)
	}
	Error := func(err error, caption string) (bool, error) {
		return result(false, err, caption)
	}

	targetInfo, err := Stat(target)
	switch {
	case os.IsNotExist(err):
		return OutOfDate("target does not exist")
	case err != nil:
		return Error(err, "error checking target file existence")
	case IsNewer(sourceInfo, targetInfo):
		return OutOfDate("source is newer than target")
	case IsNewer(options.FileInfo(), targetInfo):
		return OutOfDate("compiler options file is newer than target")
	}

	for _, filename := range deps {
		switch info, err := Stat(filename); {
		case os.IsNotExist(err):
			return OutOfDate(fmt.Sprintf("%q dependent file does not exist", filename))
		case err != nil:
			return Error(err, fmt.Sprintf("%q error when checking dependent file", filename))
		case IsNewer(info, targetInfo):
			return OutOfDate(fmt.Sprintf("%q dependency is newer than target", filename))
		}
	}
	return UpToDate()
}
