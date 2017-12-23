// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//
package main

import (
	"io"
	"log"
	"strings"
)

// The Compiler interface defines the interface to a compiler
// that can generate dependency information.
//
// The interface exists to accomonodate the differences between
// gcc-style dependency generation (as done by gcc, clang and icc)
// and the approach taken with Microsoft C++ (parsing the output
// of its /showIncludes switch).
//
type Compiler interface {
	// Return a name for the compiler.
	//
	Name() string

	// Compile the source file named by source using the supplied options
	// and create an object file named object and a dependencies file
	// called deps.
	//
	Compile(source, object, deps string, options []string, w io.Writer) error

	// Read a compiler-generated depdencies file and return the dependent filenames.
	//
	ReadDependencies(path string) (string, []string, error)
}

// GetCompiler is a factory function to return a value that implements
// the Compiler interface.
//
func GetCompiler(name string) Compiler {
	switch name {
	case "cl", "cl.exe":
		return NewMsvcCompiler()
	case "cc", "c++", "gcc", "g++", "clang", "clang++", "icc", "icpc":
		return NewGccStyleCompiler(name)
	default:
		if strings.Contains(name, "gcc") || strings.Contains(name, "clang") {
			return NewGccStyleCompiler(name)
		}
		if strings.Contains(name, "g++") || strings.Contains(name, "clang++") {
			return NewGccStyleCompiler(name)
		}
	}
	log.Fatalf("%s: unsupported compiler", name)
	return nil
}
