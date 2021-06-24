// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func invalidateStatCache() {
	statCacheMutex.Lock()
	statCache = make(map[string]os.FileInfo)
	statCacheMutex.Unlock()
}

const (
	testDataDir  = "testdata"
	testFilename = "testfile"
)

var (
	testProjectDccDir   = filepath.Join(testDataDir, DefaultDccDir)
	testProjectRootDir  = filepath.Join(testDataDir, "root")
	testProjectChildDir = filepath.Join(testProjectRootDir, "child")
)

func makeTestDirs(t *testing.T) {
	mustMkdir := func(path string) {
		if err := os.MkdirAll(path, 0777); err != nil {
			t.Fatal(err)
		}
	}
	mustMkdir(testProjectRootDir)
	mustMkdir(testProjectChildDir)
	mustMkdir(testProjectDccDir)
}

func removeTestDirs(t *testing.T) {
	if err := os.RemoveAll("testdata"); err != nil {
		t.Fatal(err)
	}
}

func makeFileWithContent(t *testing.T, path, content string) {
	if err := ioutil.WriteFile(path, []byte(content), 0666); err != nil {
		t.Fatal(err)
	}
}

func makeFile(t *testing.T, path string) {
	makeFileWithContent(t, path, path)
}

func checkFile(t *testing.T, path, content string) {
	if data, err := os.ReadFile(path); err != nil {
		t.Fatal(err)
	} else if s := string(data); s != content {
		t.Fatalf("%s: content (%q) does not match expected content (%q)", path, s, content)
	}
}

func setupTest(t *testing.T) {
	removeTestDirs(t)
	makeTestDirs(t)
}

func singleFileTestCase(t *testing.T, dirname, filename, cwd string) {
	invalidateStatCache()

	fullpath := filepath.Join(dirname, filename)
	abspath, err := filepath.Abs(fullpath)
	if err != nil {
		t.Fatal(err)
	}
	makeFile(t, fullpath)

	defer func(name string) {
		if err := os.Remove(name); err != nil {
			t.Fatal(err)
		}
	}(fullpath)

	defer func(dir string) {
		if err := os.Chdir(dir); err != nil {
			t.Fatal(err)
		}
	}(MustGetwd())

	if err = os.Chdir(cwd); err != nil {
		t.Fatal(err)
	} else if foundpath, found := FindFile(filename); !found {
		t.Fatalf("%q not found", filename)
	} else if foundpath != abspath {
		t.Fatalf("found %q which is not the expected %q", foundpath, fullpath)
	} else {
		checkFile(t, foundpath, fullpath)
	}
}

func Test_FindFile_DirectorySearching(t *testing.T) {
	setupTest(t)
	defer removeTestDirs(t)
	singleFileTestCase(t, testProjectChildDir, testFilename, testProjectChildDir)
	singleFileTestCase(t, testProjectRootDir, testFilename, testProjectChildDir)
	singleFileTestCase(t, testProjectDccDir, testFilename, testProjectChildDir)
}

func Test_FindFile_PlatformSpecificSearches(t *testing.T) {
	setupTest(t)
	defer removeTestDirs(t)
	singleFileTestCase(t, testProjectChildDir, OsSpecificFilename(testFilename), testProjectChildDir)
	singleFileTestCase(t, testProjectDccDir, OsAndArchSpecificFilename(testFilename), testProjectChildDir)
}
