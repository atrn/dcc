// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//

package main

var lib32 = []string{
	"/usr/local/lib",
	"/usr/lib32",
	"/usr/lib/gcc/x86_64-linux-gnu/8",
	"/usr/lib",
	"/lib",
}

var lib64 = []string{
	"/usr/local/lib",
	"/usr/lib",
	"/usr/lib/x86_64-linux-gnu",
	"/lib/x86_64-linux-gnu",
	"/lib",
}

func linuxDefaultLibraryPaths() []string {
	if runtime.ARCH == "amd64" {
		return lib64
	}
	return lib32
}

func linuxSelectTarget(target string) error {
	switch target {
	case "-m32":
		platform.LibraryPaths = lib32
		return nil
	case "-m64":
		platform.LibraryPaths = lib64
		return nil
	default:
		return fmt.Errof("%s: unhandled target", target)
	}
}

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
	LibraryPaths:      linuxDefaultLibraryPaths(),
	CreateLibrary:     ElfCreateLibrary,
	CreateDLL:         ElfCreateDLL,
	SelectTarget:      linuxSelectTarget,
}
