// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//

package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CurrentDirectory is the path of the process working directory at startup
//
var CurrentDirectory = MustGetwd()

// NewestOf returns the os.FileInfo time for the most recently
// modified file in the slice of file names.
//
// Each file is stat'd and its modification time used to determine
// if it is newer than any previous file. Any error doing this
// results in an non-nil error return and the most recently
// set most recent time.
//
func NewestOf(filenames []string) (time.Time, error) {
	var t time.Time
	info, err := Stat(filenames[0])
	if err != nil {
		return t, err
	}
	t = info.ModTime()
	for i := 1; i < len(filenames); i++ {
		if s, err := Stat(filenames[i]); err != nil {
			return t, err
		} else if s.ModTime().After(t) {
			t = s.ModTime()
		}
	}
	return t, nil
}

// IsNewer returns true if the first file is newer than the second.
//
func IsNewer(a os.FileInfo, b os.FileInfo) bool {
	switch {
	case a == nil:
		return false
	case b == nil:
		return true
	default:
		return a.ModTime().After(b.ModTime())
	}
}

// ObjectFilename returns the name of the object file for a given source file.
//
func ObjectFilename(path string, d string) string {
	if IsSourceFile(path) {
		stem := strings.TrimSuffix(path, filepath.Ext(path))
		return filepath.Clean(filepath.Join(d, stem+platform.ObjectFileSuffix))
	}
	if IsHeaderFile(path) {
		return filepath.Clean(filepath.Join(d, path+".gch")) // XXX platform
	}
	return filepath.Clean(filepath.Join(d, path+".o")) // something other than the input filename!
}

// DepsFilename returns the name of the dependencies file for a given object file.
//
func DepsFilename(path string) string {
	if DepsDir == "" {
		return path + ".d"
	}
	head, tail := filepath.Dir(path), filepath.Base(path)
	return filepath.Join(head, DepsDir, tail) + ".d"
}
