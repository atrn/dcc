// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//

package main

var platform = Platform{
	DefaultCC:         "cc",
	DefaultCXX:        "c++",
	DefaultDepsDir:    ".dcc.d",
	ObjectFileSuffix:  ".o",
	DynamicLibPrefix:  "lib",
	DynamicLibSuffix:  ".so",
	StaticLibPrefix:   "lib",
	StaticLibSuffix:   ".a",
	DefaultExecutable: "a.out",
	LibraryPaths:      []string{"/usr/local/lib64", "/usr/local/lib", "/usr/lib/x86_64-linux-gnu", "/usr/lib64", "/usr/lib", "/lib/x86_64-linux-gnu", "lib64", "/lib"},
	CreateLibrary:     ElfCreateLibrary,
	CreateDLL:         ElfCreateDLL,
}
