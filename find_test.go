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
	projectDccDir   = filepath.Join(testDataDir, DefaultDccDir)
	projectRootDir  = filepath.Join(testDataDir, "root")
	projectChildDir = filepath.Join(projectRootDir, "child")
)

func makeTestDirs(t *testing.T) {
	mustMkdir := func(path string) {
		if err := os.MkdirAll(path, 0777); err != nil {
			t.Fatal(err)
		}
	}
	mustMkdir(projectRootDir)
	mustMkdir(projectChildDir)
	mustMkdir(projectDccDir)
}

func removeTestDirs(t *testing.T) {
	if err := os.RemoveAll("testdata"); err != nil {
		t.Fatal(err)
	}
}

func makeFile(t *testing.T, path string) {
	if err := ioutil.WriteFile(path, []byte(path), 0666); err != nil {
		t.Fatal(err)
	}
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

	cwd, err = filepath.Abs(cwd)
	if err != nil {
		t.Fatal(err)
	}

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
	singleFileTestCase(t, projectChildDir, testFilename, projectChildDir)
	singleFileTestCase(t, projectRootDir, testFilename, projectChildDir)
	singleFileTestCase(t, projectDccDir, testFilename, projectChildDir)
}

func Test_FindFile_PlatformSpecificSearches(t *testing.T) {
	setupTest(t)
	defer removeTestDirs(t)
	singleFileTestCase(t, projectChildDir, OsSpecificFilename(testFilename), projectChildDir)
	singleFileTestCase(t, projectDccDir, OsAndArchSpecificFilename(testFilename), projectChildDir)
}
