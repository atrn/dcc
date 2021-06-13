// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//

package main

import (
	"os"
	"path/filepath"
)

var platform = Platform{
	DefaultCC:         "cl",
	DefaultCXX:        "cl",
	ObjectFileSuffix:  ".obj",
	StaticLibSuffix:   ".lib",
	DynamicLibSuffix:  ".dll",
	DefaultExecutable: "program.exe",
	LibraryPaths:      filepath.SplitList(Getenv("LIB", "")),
	CreateLibrary:     WindowsCreateLibrary,
	CreateDLL:         WindowsCreateDLL,
	CreatePlugin:      WindowsCreateDLL,
}

// WindowsCreateLibrary creates a static library from the supplied object files
// using Microsoft's LIB.EXE
//
func WindowsCreateLibrary(filename string, objectFiles []string) error {
	args := append([]string{"/out:" + filename}, objectFiles...)
	return Exec("lib", args, os.Stderr)
}

// WindowsCreateDLL creates a dynamic library from the supplied object files
// and library files using Microsoft's LINK.EXE.
//
func WindowsCreateDLL(filename string, objectFiles []string, libraryFiles []string, linkerOptions []string, frameworks []string) error {
	args := append([]string{"/DLL", "/OUT:" + filename}, objectFiles...)
	args = append(args, linkerOptions...)
	args = append(args, libraryFiles...)
	return Exec("link", args, os.Stderr)
}
