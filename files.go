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
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	OsSuffix     = fmt.Sprintf(".%s", runtime.GOOS)
	ArchSuffix   = fmt.Sprintf(".%s", runtime.GOARCH)
	OsArchSuffix = fmt.Sprintf(".%s_%s", runtime.GOOS, runtime.GOARCH)
)

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

// FileIsNewer returns true if the first file is newer than the second
// as identified by their FileInfo's modification time.
//
func FileIsNewer(a os.FileInfo, b os.FileInfo) bool {
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
	dirname, basename := filepath.Split(path)
	parts := strings.Split(dirname, "/")
	first, n := 0, len(parts)
	if n > 0 {
		for first < n && parts[first] == ".." {
			first++
		}
		if first > 0 {
			if first == n {
				path = filepath.Join(".", basename)
			} else {
				path = filepath.Join(strings.Join(parts[first:], "/"), basename)
			}
		}
	}
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
	dirname, basename := filepath.Split(path)
	return filepath.Join(dirname, DepsDir, basename) + ".d"
}

// Basename returns base portion of a path, the filename.Base(),
// taking into account the possibility the path may have an OS
// and/or architecture specific suffix, which is removed.
//
func Basename(path string) string {
	b := filepath.Base(path)
	if strings.HasSuffix(b, OsArchSuffix) {
		return strings.TrimSuffix(b, OsArchSuffix)
	}
	if strings.HasSuffix(b, ArchSuffix) {
		return strings.TrimSuffix(b, ArchSuffix)
	}
	if strings.HasSuffix(b, OsSuffix) {
		return strings.TrimSuffix(b, OsSuffix)
	}
	return b
}

// Dirname returns the directory portion of a path taking into account
// the possibility the path may reside in the "dcc directory". I.e. it
// returns the filepath.Dir() of a path removing the final ".dcc" if
// it has one.
//
func Dirname(path string) string {
	d := filepath.Dir(path)
	if filepath.Base(d) == DccDir {
		d = filepath.Dir(d)
	}
	return d
}

// OsSpecificFilename returns the OS-specfic verson of a path
// (filename) by appending the OsSuffix to the path.
//
func OsSpecificFilename(path string) string {
	return path + OsSuffix
}

// ArchSpecificFilename returns the architecture-specfic verson of a
// path (filename) by appending the ArchSuffix to the path.
//
func ArchSpecificFilename(path string) string {
	return path + ArchSuffix
}

// OsAndArchSpecificFilename returns the architecture and OS-specfic
// verson of a path (filename) by appending the OsArchSuffix to the
// path.
//
func OsAndArchSpecificFilename(path string) string {
	return path + OsArchSuffix
}

// ParentDir returns the parent directory of a path taking into
// account the possibility the path is in a ".dcc" directory.
//
func ParentDir(path string) string {
	return filepath.Clean(filepath.Join(Dirname(path), ".."))
}
