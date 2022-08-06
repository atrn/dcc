// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//
//go:build windows

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMsvc(t *testing.T) {
	makeTestDirs(t)
	defer removeTestDirs(t)

	msvc := &msvcCompiler{}

	sourceFile := filepath.Join(testProjectRootDir, "main.cpp")

	makeFileWithContent(t, sourceFile, `// test source file
#include <iostream>
int main() { std::cout << "Hello\n"; }
`,
	)

	objectFile := ObjectFilename(filepath.Base(sourceFile), testProjectRootDir)
	depsFile := filepath.Join(testProjectRootDir, "main.d")
	options := []string{"/nologo", "/EHsc", "/W3", "/O2"}
	err := msvc.Compile(sourceFile, objectFile, depsFile, options, os.Stderr)
	if err != nil {
		t.Fatal(err)
	}
}
