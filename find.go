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

var DebugFind = false // enable for verbose debug output

// Given the base name for an options file this function finds the
// actual file, using the various rules for searching for and
// selecting files according to platform and returns the actual path
// of the file that should be read.
//
func FindFileWithName(name string) (string, bool) {
	return FindFile(filepath.Join(DccDir, name), nil)
}

// FindFile returns a path for a filename and a flag
// indicating if the file was actually found.
//
func FindFile(filename string, f func(string) string) (string, bool) {
	ofilename := filename
	if DebugFind {
		log.Printf("DEBUG FIND: FindFile %q", ofilename)
	}
	path, _, exists, err := FindFileFromCwd(filename, f)
	if err == nil && exists {
		if DebugFind {
			log.Printf("DEBUG FIND: FindFile %q -> %q", filename, path)
		}
		return path, true
	}
	if err != nil {
		log.Print(err)
	}
	if DebugFind {
		log.Printf("DEBUG FIND: FindFile %q -> %q", filename, filename)
	}
	return filename, false
}

// FindFileFromCwd finds a file starting from the current directory and searching towards the root.
//
func FindFileFromCwd(filename string, f func(string) string) (string, os.FileInfo, bool, error) {
	if DebugFind {
		log.Printf("DEBUG FIND: FindFileFromCwd %q", filename)
	}
	return FindFileFromDirectory(filename, CurrentDirectory, f)
}

// FindFileFromDirectory finds a file starting from the specified directory, search towards the root.
//
func FindFileFromDirectory(filename, dir string, f func(string) string) (string, os.FileInfo, bool, error) {
	if DebugFind {
		log.Printf("DEBUG FIND: FindFileFromDirectory %q %q", filename, dir)
	}
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
	logit := func(format string, args ...interface{}) {
		if DebugFind {
			log.Printf("DEBUG FIND: "+format, args...)
		}
	}
	logit("FindFileOnPath %q %q", paths, filename)
	for _, dir := range paths {
		path := filepath.Join(dir, filename)
		logit("trying %q", path)
		if f != nil {
			newpath := f(path)
			logit("transformed %q -> %q", path, newpath)
			path = newpath
		}
		if info, err := Stat(path); err == nil {
			logit("returning %q", path)
			return path, info, true, nil
		} else if !os.IsNotExist(err) {
			logit("%s", err.Error())
			return "", nil, false, err
		}

		path = filepath.Join(dir, DefaultDccDir, filename)
		logit("trying %q", path)
		if f != nil {
			newpath := f(path)
			logit("transformed %q -> %q", path, newpath)
			path = newpath
		}
		if info, err := Stat(path); err == nil {
			logit("returning %q", path)
			return path, info, true, nil
		} else if !os.IsNotExist(err) {
			logit("%s", err.Error())
			return "", nil, false, err
		}
	}
	logit("%q not found", filename)
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
