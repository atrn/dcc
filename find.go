// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

var DebugFind = false // enable for verbose debug output

// FindFile returns a path for a filename and a flag indicating if the
// file was actually found.
//
func FindFile(filename string) (string, bool) {
	ofilename := filename
	if DebugFind {
		log.Printf("DEBUG FIND: FindFile %q", ofilename)
	}
	path, _, exists, err := FindFileFromCwd(filename)
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
func FindFileFromCwd(filename string) (string, os.FileInfo, bool, error) {
	if DebugFind {
		log.Printf("DEBUG FIND: FindFileFromCwd %q", filename)
	}
	return FindFileFromDirectory(filename, MustGetwd())
}

// FindFileFromDirectory finds a file starting from the specified directory, search towards the root.
//
func FindFileFromDirectory(filename, dir string) (string, os.FileInfo, bool, error) {
	if DebugFind {
		log.Printf("DEBUG FIND: FindFileFromDirectory %q %q", filename, dir)
	}
	const root = string(filepath.Separator)
	paths := []string{dir}
	for dir != root {
		dir = filepath.Dir(dir)
		paths = append(paths, dir)
	}
	paths = append(paths, dir)
	return FindFileOnPath(paths, filename)
}

// FindFileOnPath finds a file along a search path.
//
func FindFileOnPath(paths []string, filename string) (string, os.FileInfo, bool, error) {
	if DebugFind {
		log.Printf("DEBUG FIND: FindFileOnPath %q %q", paths, filename)
	}
	for _, dir := range paths {
		if path, info, found, err := FindFileInDirectory(filename, dir); err != nil {
			return "", nil, false, err
		} else if found {
			return path, info, true, nil
		}
	}
	if DebugFind {
		log.Printf("DEBUG FIND: %q not found", filename)
	}
	return "", nil, false, nil
}

func FindFileInDirectory(filename string, dirname string) (string, os.FileInfo, bool, error) {
	filenameOs := fmt.Sprintf("%s.%s", filename, runtime.GOOS)
	filenameOsArch := fmt.Sprintf("%s_%s", filenameOs, runtime.GOARCH)

	try := func(dirname, filename string) (string, os.FileInfo, bool, error) {
		path := filepath.Join(dirname, filename)
		if DebugFind {
			log.Printf("DEBUG FIND: FindFileInDirectory trying %q", path)
		}
		if info, err := Stat(path); err == nil {
			if DebugFind {
				log.Printf("DEBUG FIND: FindFileInDirectory returning %q", path)
			}
			return path, info, true, nil
		} else if !os.IsNotExist(err) {
			if DebugFind {
				log.Printf("DEBUG FIND: FindFileInDirectory %q: %s", path, err.Error())
			}
			return path, nil, true, err
		}
		return "", nil, false, nil
	}

	tryAll := func() (string, os.FileInfo, bool, error) {
		if path, info, found, err := try(dirname, filenameOsArch); err != nil || found {
			return path, info, found, err
		}
		if path, info, found, err := try(dirname, filenameOs); err != nil || found {
			return path, info, found, err
		}
		return try(dirname, filename)
	}

	if path, info, found, err := tryAll(); err != nil || found {
		return path, info, found, err
	}

	dirname = filepath.Join(dirname, DccDir)

	if path, info, found, err := tryAll(); err != nil || found {
		return path, info, found, err
	}

	return "", nil, false, nil
}

// FindLibrary finds a library file on a search path, either static or dynamic.
//
func FindLibrary(paths []string, name string) (string, os.FileInfo, bool, error) {
	if path, info, found, err := FindFileOnPath(paths, platform.DynamicLibrary(name)); found || err != nil {
		return path, info, found, err
	}
	if path, info, found, err := FindFileOnPath(paths, platform.StaticLibrary(name)); found || err != nil {
		return path, info, found, err
	}
	return "", nil, false, nil
}
