// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//

// +build !windows !darwin

package main

import (
	"fmt"
	"os"
	"strings"
)

// ElfCreateLibrary creates a static library file using the UNIX ar
// program.
//
func ElfCreateLibrary(filename string, objectFiles []string) error {
	args := append([]string{"rc", filename}, objectFiles...)
	if Verbose {
		fmt.Fprintln(os.Stdout, "ar", strings.Join(args, " "))
	} else if !Quiet {
		fmt.Fprintln(os.Stdout, "ar", filename)
	}
	return Exec("ar", args, os.Stderr)
}

// ElfCreateDLL creates a dynamic library using the compiler, passing
// a -shared option.
//
func ElfCreateDLL(filename string, objectFiles []string, libraryFiles []string, linkerOptions []string, frameworks []string) error {
	args := []string{"-shared", "-o", filename}
	args = append(args, linkerOptions...)
	args = append(args, objectFiles...)
	args = append(args, libraryFiles...)
	if Verbose {
		fmt.Fprintln(os.Stdout, ActualCompiler.Name(), strings.Join(args, " "))
	} else if !Quiet {
		fmt.Fprintln(os.Stdout, "ld", filename)
	}
	return Exec(ActualCompiler.Name(), args, os.Stderr)
}
