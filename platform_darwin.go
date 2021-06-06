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
	PluginPrefix:      "",
	PluginSuffix:      ".bundle",
	DefaultExecutable: "a.out",
	LibraryPaths:      []string{"/usr/lib"},
	CreateLibrary:     MacosCreateLibrary,
	CreateDLL:         MacosCreateDLL,
	CreatePlugin:      MacosCreatePlugin,
}

func logCommand(cmd string, args []string) {
	if Quiet {
		return
	}
	if Verbose {
		fmt.Fprintln(os.Stdout, cmd, strings.Join(args, " "))
		return
	}
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

// MacosCreateLibrary creates a static library using libtool.
//
func MacosCreateLibrary(filename string, objectFiles []string) error {
	const libtool = "libtool"
	args := []string{"-static", "-o", filename}
	args = append(args, objectFiles...)
	logCommand(libtool, args)
	return Exec(libtool, args, os.Stderr)
}

// MacosCreateDLL creates a dynamic library using the underlying
// compiler and its -shared option.
//
func MacosCreateDLL(filename string, objectFiles []string, libraryFiles []string, linkerOptions []string, frameworks []string) error {
	args := []string{"-shared", "-o", filename}
	args = append(args, linkerOptions...)
	args = append(args, objectFiles...)
	args = append(args, libraryFiles...)
	args = append(args, frameworks...)
	logCommand(ActualCompiler.Name(), args)
	return Exec(ActualCompiler.Name(), args, os.Stderr)
}

// MacosCreatePlugin creates a bundle
//
func MacosCreatePlugin(filename string, objectFiles []string, libraryFiles []string, linkerOptions []string, frameworks []string) error {
	args := []string{"-bundle", "-o", filename}
	args = append(args, linkerOptions...)
	args = append(args, objectFiles...)
	args = append(args, libraryFiles...)
	args = append(args, frameworks...)
	logCommand(ActualCompiler.Name(), args)
	return Exec(ActualCompiler.Name(), args, os.Stderr)
}
