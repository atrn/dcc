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
	DefaultDepsDir:    ".dcc.d",
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
			fmt.Fprintln(os.Stderr, cmd, strings.Join(args, " "))
		} else {
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
			fmt.Fprintln(os.Stderr, cmd, filename)
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
func MacosCreateDLL(filename string, objectFiles []string, libraryFiles []string, linkerOptions []string) error {
	args := []string{"-dynamic", "-o", filename}
	args = append(args, linkerOptions...)
	args = append(args, objectFiles...)
	args = append(args, libraryFiles...)
	return libtool(args)
}
