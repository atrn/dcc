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
	"runtime"
	"strings"
)

// The Platform type defines various platform-specific values and
// functions. There is a single global value of this type, called
// platform, defined by the platform_xxx.go file for the target
// system.
//
type Platform struct {
	DefaultCC         string
	DefaultCXX        string
	DefaultDepsDir    string
	ObjectFileSuffix  string
	StaticLibPrefix   string
	StaticLibSuffix   string
	DynamicLibPrefix  string
	DynamicLibSuffix  string
	DefaultExecutable string
	LibraryPaths      []string
	CreateLibrary     func(string, []string) error
	CreateDLL         func(string, []string, []string, []string) error
}

// StaticLibrary transforms a filename "stem" to the name of a
// static library on the host platform.
//
func (p *Platform) StaticLibrary(name string) string {
	return p.StaticLibPrefix + name + p.StaticLibSuffix
}

// DynamicLibrary transforms a filename "stem" to the name of
// a dynamic library on the host platform.
//
func (p *Platform) DynamicLibrary(name string) string {
	return p.DynamicLibPrefix + name + p.DynamicLibSuffix
}

// PlatformSpecific is a pathname filter function used with the
// MustFindFile function to transform pathnames to so-called platform-
// specific versions so they may be used in preference to other files.
//
func PlatformSpecific(path string) string {
	try := func(path string) (string, bool) {
		_, err := Stat(path)
		return path, err == nil
	}
	if p, ok := try(path + "." + runtime.GOOS + "_" + runtime.GOARCH); ok {
		return p
	}
	if p, ok := try(path + "." + runtime.GOOS); ok {
		return p
	}
	return path
}

// ElfCreateLibrary creates a static library file using the UNIX ar
// program.
//
func ElfCreateLibrary(filename string, objectFiles []string) error {
	args := append([]string{"rcv", filename}, objectFiles...)
	if !Quiet {
		fmt.Fprintln(os.Stderr, "ar", strings.Join(args, " "))
	}
	return Exec("ar", args, os.Stderr)
}

// ElfCreateDLL creates a dynamic library using the compiler, passing
// a -shared option.
//
func ElfCreateDLL(filename string, objectFiles []string, libraryFiles []string, linkerOptions []string) error {
	args := []string{"-shared", "-o", filename}
	args = append(args, objectFiles...)
	args = append(args, libraryFiles...)
	if !Quiet {
		fmt.Fprintln(os.Stderr, ActualCompiler.Name(), strings.Join(args, " "))
	}
	return Exec(ActualCompiler.Name(), args, os.Stderr)
}
