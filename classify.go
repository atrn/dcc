// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//

package main

import (
	"path/filepath"
)

var cppSet = StringSet{
	".cc":  {},
	".cpp": {},
	".cxx": {},
	".c++": {},
	".hh":  {},
	".hpp": {},
	".hxx": {},
	".h++": {},
}

var sourceFileSet = StringSet{
	".c":   {},
	".cc":  {},
	".cpp": {},
	".cxx": {},
	".c++": {},
	".m":   {},
	".mm":  {},
}

var headerFileSet = StringSet{
	".h":   {},
	".hh":  {},
	".hpp": {},
	".hxx": {},
	".h++": {},
}

// IsCPlusPlusFile returns true if the supplied pathname is that
// of a C++ source file.
//
func IsCPlusPlusFile(path string) bool {
	return cppSet.Contains(LowercaseFilenameExtension(path))
}

// IsSourceFile returns true if the supplied pathname is that
// of a C, C++ or Objective-C source file.
//
func IsSourceFile(path string) bool {
	return sourceFileSet.Contains(LowercaseFilenameExtension(path))
}

// IsHeaderFile returns true if the supplied pathname is that
// of a C/C++ header file.
//
func IsHeaderFile(path string) bool {
	return headerFileSet.Contains(LowercaseFilenameExtension(path))
}

// IsLibraryFile returns true if the supplied pathname is
// that of a library of some kind?
//
func IsLibraryFile(path string) bool {
	ext := filepath.Ext(path)
	if ext == platform.StaticLibSuffix {
		return true
	}
	if ext == platform.DynamicLibSuffix {
		return true
	}
	return false
}

// FileWillBeCompiled returns true if the supplied pathname is the
// name of a some source file that the compiler would compile.
// Header files are considered source files to accomodate pre-
// compiling header files.
//
func FileWillBeCompiled(path string) bool {
	return IsSourceFile(path) || IsHeaderFile(path)
}
