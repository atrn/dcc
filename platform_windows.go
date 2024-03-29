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
	"strings"
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
	IsRoot:            WindowsIsRoot,
}

// WindowsCreateLibrary creates a static library from the supplied object files
// using Microsoft's LIB.EXE
func WindowsCreateLibrary(filename string, objectFiles []string) error {
	args := append([]string{"/nologo", "/out:" + filename}, objectFiles...)
	return Exec("lib", args, os.Stderr)
}

// WindowsCreateDLL creates a dynamic library from the supplied object files
// and library files using Microsoft's LINK.EXE.
func WindowsCreateDLL(filename string, objectFiles []string, libraryFiles []string, linkerOptions []string, frameworks []string) error {
	args := append([]string{"/nologo", "/DLL", "/OUT:" + filename}, objectFiles...)
	args = append(args, linkerOptions...)
	args = append(args, libraryFiles...)
	return Exec("link", args, os.Stderr)
}

// WindowsIsRoot determines if a pathname represents a "root"
// directory.  A "root" is defined to be the top-level directory of a
// volume (aka drive).
func WindowsIsRoot(path string) bool {
	path = filepath.Clean(path)
	path = strings.TrimPrefix(path, filepath.VolumeName(path))
	return path == "\\"
}
