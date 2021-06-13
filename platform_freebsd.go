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
	ObjectFileSuffix:  ".o",
	DynamicLibPrefix:  "lib",
	DynamicLibSuffix:  ".so",
	PluginPrefix:      "lib",
	PluginSuffix:      ".so",
	StaticLibPrefix:   "lib",
	StaticLibSuffix:   ".a",
	DefaultExecutable: "a.out",
	LibraryPaths:      []string{"/usr/lib", "/lib"},
	CreateLibrary:     ElfCreateLibrary,
	CreateDLL:         ElfCreateDLL,
	CreatePlugin:      ElfCreateDLL,
}
