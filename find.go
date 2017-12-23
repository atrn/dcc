// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//
package main

import (
	"log"
	"os"
	"path/filepath"
)

// FindFile returns a path for a filename and a flag
// indicating if the file was actually found.
//
func FindFile(filename string, f func(string) string) (string, bool) {
	path, _, exists, err := FindFileFromCwd(filename, f)
	if err == nil && exists {
		return path, true
	}
	if err != nil {
		log.Print(err)
	}
	return filename, false
}

// FindFileFromCwd finds a file starting from the current directory and searching towards the root.
//
func FindFileFromCwd(filename string, f func(string) string) (string, os.FileInfo, bool, error) {
	return FindFileFromDirectory(filename, CurrentDirectory, f)
}

// FindFileFromDirectory finds a file starting from the specified directory, search towards the root.
//
func FindFileFromDirectory(filename, dir string, f func(string) string) (string, os.FileInfo, bool, error) {
	const root = string(filepath.Separator)
	paths := []string{dir}
	for dir != root {
		dir = filepath.Dir(dir)
		paths = append(paths, dir)
	}
	return FindFileOnPath(paths, filename, f)
}

// FindFileOnPath finds a file along a search path.
//
func FindFileOnPath(paths []string, filename string, f func(string) string) (string, os.FileInfo, bool, error) {
	for _, dir := range paths {
		path := filepath.Join(dir, filename)
		if f != nil {
			path = f(path)
		}
		if info, err := Stat(path); err == nil {
			return path, info, true, nil
		} else if !os.IsNotExist(err) {
			return "", nil, false, err
		} // else, ignore non-existent file
	}
	return "", nil, false, nil
}

// FindLib finds a library file on a search path, either static or dynamic.
//
func FindLib(paths []string, name string) (string, os.FileInfo, bool, error) {
	if path, info, found, err := FindFileOnPath(paths, platform.DynamicLibrary(name), nil); found || err != nil {
		return path, info, found, err
	}
	if path, info, found, err := FindFileOnPath(paths, platform.StaticLibrary(name), nil); found || err != nil {
		return path, info, found, err
	}
	return "", nil, false, nil
}
