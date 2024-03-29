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
	"runtime"
)

var lib32 = []string{
	"/usr/local/lib",
	"/usr/lib32",
	"/usr/lib/gcc/x86_64-linux-gnu/8",
	"/usr/lib",
	"/lib",
}

var lib64 = []string{
	"/usr/local/lib64",
	"/usr/lib64",
	"/lib64",
	"/usr/lib/x86_64-linux-gnu",
	"/lib/x86_64-linux-gnu",
	"/usr/local/lib",
	"/usr/lib",
	"/lib",
}

func linuxDefaultLibraryPaths() []string {
	if runtime.GOARCH == "amd64" {
		return lib64
	}
	return lib32
}

func linuxSelectTarget(p *Platform, target string) error {
	switch target {
	case "-m32":
		p.LibraryPaths = lib32
		return nil
	case "-m64":
		p.LibraryPaths = lib64
		return nil
	default:
		return fmt.Errorf("%s: unhandled target", target)
	}
}

var platform = Platform{
	DefaultCC:         "cc",
	DefaultCXX:        "c++",
	ObjectFileSuffix:  ".o",
	DynamicLibPrefix:  "lib",
	DynamicLibSuffix:  ".so",
	PluginPrefix:      "lib",
	PluginSuffix:      ".so",
	StaticLibPrefix:   "lib",
	StaticLibSuffix:   ".a",
	DefaultExecutable: "a.out",
	LibraryPaths:      linuxDefaultLibraryPaths(),
	CreateLibrary:     ElfCreateLibrary,
	CreateDLL:         ElfCreateDLL,
	CreatePlugin:      ElfCreateDLL,
	SelectTarget:      linuxSelectTarget,
	IsRoot:		   UnixIsRoot,
}
