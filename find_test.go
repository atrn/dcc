package main

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func runOneTestCase(cwd string, fn func() error) error {
	ocwd := MustGetwd()
	defer func() {
		os.Chdir(ocwd)
		CurrentDirectory = ocwd
	}()
	cwd, err := filepath.Abs(cwd)
	if err != nil {
		return err
	}
	err = os.Chdir(cwd)
	if err == nil {
		CurrentDirectory = cwd
		err = fn()
	}
	return err
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

	filename := "test.file"
	fullpath := path.Join("testdata/root/", DefaultDccDir, filename)

	err = ioutil.WriteFile(fullpath, []byte{}, 0666)
	if err != nil {
		t.Fatal(err)
	}

	err = runOneTestCase("testdata/root/child", func() error {
		_, found := FindFile(filename, PlatformSpecific)
		if !found {
			t.Fatalf("%q not found", filename)
		}
		return nil
	})
}
