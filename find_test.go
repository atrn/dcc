package main

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"
)

func invalidateStatCache() {
	statCacheMutex.Lock()
	statCache = make(map[string]os.FileInfo)
	statCacheMutex.Unlock()
}

func runOneTestCase(t *testing.T, dirname, filename, cwd string) {
	invalidateStatCache()

	fullpath := filepath.Join(dirname, filename)
	abspath, err := filepath.Abs(fullpath)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(fullpath, []byte{}, 0666)
	if err != nil {
		t.Fatal(err)
	}

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
	}
}

func TestFind(t *testing.T) {
	// DebugFind = true

	removeTestData := func() {
		err := os.RemoveAll("testdata")
		if err != nil {
			os.Stderr.WriteString("ERROR: " + err.Error())
		}
	}

	defer removeTestData()
	removeTestData()

	err := os.MkdirAll("testdata/root/child", 0777)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Mkdir("testdata/root/"+DefaultDccDir, 0777)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Mkdir("testdata/"+DefaultDccDir, 0777)
	if err != nil {
		t.Fatal(err)
	}

	dirname1 := "testdata/root/child"
	dirname2 := path.Join("testdata/root", DefaultDccDir)
	dirname3 := path.Join("testdata", DefaultDccDir)

	filename := "testfile"
	runOneTestCase(t, dirname1, filename, "testdata/root/child")
	runOneTestCase(t, dirname2, filename, "testdata/root/child")
	runOneTestCase(t, dirname3, filename, "testdata/root/child")
	runOneTestCase(t, dirname1, filename+"."+runtime.GOOS, "testdata/root/child")
	runOneTestCase(t, dirname3, filename+"."+runtime.GOOS+"_"+runtime.GOARCH, "testdata/root/child")
}
