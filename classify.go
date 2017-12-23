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
	"strings"
)

var cppSet = map[string]struct{}{
	".cc":  struct{}{},
	".cpp": struct{}{},
	".cxx": struct{}{},
	".c++": struct{}{},
	".hh":  struct{}{},
	".hpp": struct{}{},
	".hxx": struct{}{},
	".h++": struct{}{},
}

var sourceFileSet = map[string]struct{}{
	".c":   struct{}{},
	".cc":  struct{}{},
	".cpp": struct{}{},
	".cxx": struct{}{},
	".c++": struct{}{},
	".m":   struct{}{},
	".mm":  struct{}{},
}

var headerFileSet = map[string]struct{}{
	".h":   struct{}{},
	".hh":  struct{}{},
	".hpp": struct{}{},
	".hxx": struct{}{},
	".h++": struct{}{},
}

func is(path string, set map[string]struct{}) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, found := set[ext]
	return found
}

// IsCPlusPlus returns true if the supplied pathname is that
// of a C++ source file.
//
func IsCPlusPlus(path string) bool {
	return is(path, cppSet)
}

// IsSourceFile returns true if the supplied pathname is that
// of a C, C++ or Objective-C source file.
//
func IsSourceFile(path string) bool {
	return is(path, sourceFileSet)
}

// IsHeaderFile returns true if the supplied pathname is that
// of a C/C++ header file.
//
func IsHeaderFile(path string) bool {
	return is(path, headerFileSet)
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
