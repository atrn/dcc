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
	"os"
	"strings"
)

var platform = Platform{
	DefaultCC:         "cc",
	DefaultCXX:        "c++",
	DefaultDepsDir:    DefaultDccDir + ".d",
	ObjectFileSuffix:  ".o",
	StaticLibPrefix:   "lib",
	StaticLibSuffix:   ".a",
	DynamicLibPrefix:  "lib",
	DynamicLibSuffix:  ".dylib",
	DefaultExecutable: "a.out",
	LibraryPaths:      []string{"/usr/lib"},
	CreateLibrary:     MacosCreateLibrary,
	CreateDLL:         MacosCreateDLL,
}

// Run libtool with the supplied arguments.
//
func libtool(args []string) error {
	const cmd = "libtool"
	if !Quiet {
		if Verbose {
			fmt.Fprintln(os.Stdout, cmd, strings.Join(args, " "))
		} else if !Quiet {
			filename := ""
			nargs := len(args)
			for index, arg := range args {
				if arg == "-o" {
					if index < nargs-1 {
						filename = args[index+1]
						break
					}
				}
			}
			fmt.Fprintln(os.Stdout, cmd, filename)
		}
	}
	return Exec(cmd, args, os.Stderr)
}

// MacosCreateLibrary creates a static library using the MacOS libtool
// program passing it the -static option.
//
func MacosCreateLibrary(filename string, objectFiles []string) error {
	args := []string{"-static", "-o", filename}
	args = append(args, objectFiles...)
	return libtool(args)
}

// MacosCreateDLL creates a dynamic library using the MacOS libtool
// program passing it the -dynamic option.
//
func MacosCreateDLL(filename string, objectFiles []string, libraryFiles []string, linkerOptions []string, frameworks []string) error {
	// The LDFLAGS and LIBS options files are shared between the compiler
	// and libtool.  However libtool doesn't support the compiler's -Wl
	// mechanism for passing options to the actual linker so we strip
	// any such prefix from options we send to libtool. libtool also
	// doesn't understand -rpath and its argument.
	//
	removeWl := func(args []string) (r []string) {
		const dashWl = "-Wl,"
		skipNext := false
		for _, arg := range args {
			if skipNext {
				skipNext = false
			} else {
				if strings.HasPrefix(arg, "-Wl,-rpath") {
					skipNext = true
				} else if strings.HasPrefix(arg, dashWl) {
					r = append(r, strings.TrimPrefix(arg, dashWl))
				} else {
					r = append(r, arg)
				}
			}
		}
		return
	}

	args := []string{"-dynamic", "-o", filename}
	args = append(args, removeWl(linkerOptions)...)
	args = append(args, objectFiles...)
	args = append(args, libraryFiles...)
	args = append(args, frameworks...)
	return libtool(args)
}
