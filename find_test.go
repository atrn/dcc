package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func runOneTestCase(cwd string, fn func() error) error {
	opwd := MustGetwd()
	defer func() {
		os.Chdir(opwd)
	}()
	err := os.Chdir(cwd)
	if err == nil {
		err = fn()
	}
	return err
}

func TestFind(t *testing.T) {
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
		DebugFind = true
		_, found := FindFile(filename, PlatformSpecific)
		if !found {
			t.Fatalf("%q not found", filename)
		}
		return nil
	})
}
