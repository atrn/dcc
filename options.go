// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Options represents a series of words and is used to represent
// compiler and linker options. Options are intended to be read from a
// file.
//
// An Options has a slice of strings, the option "values".
//
type Options struct {
	Values   []string    // option values
	Path     string      // associated options file path
	fileinfo os.FileInfo // options file info
	mtime    time.Time   // options file modtime, mutable
}

// FileInfo returns the receiver's os.FileInfo set when the
// receiver was successfully read from a file.
//
func (o *Options) FileInfo() os.FileInfo {
	return o.fileinfo
}

// Empty returns true if the Options has no values
//
func (o *Options) Empty() bool {
	return len(o.Values) == 0
}

// String returns a string representation of the receiver.  This
// returns a space separated list of the options' Values.
//
func (o *Options) String() string {
	return strings.Join(o.Values, " ")
}

// Append appends an option to the set of options.  Note, Append does
// NOT modify the mtime of the receiver.
//
func (o *Options) Append(option string) {
	o.Values = append(o.Values, option)
}

// SetModTime sets the receiver's modification time to the supplied
// value.
//
func (o *Options) SetModTime(t time.Time) {
	o.mtime = t
}

// SetFrom copies options from another Options leaving
// the receiver's Path unmodified.
//
func (o *Options) SetFrom(other *Options) {
	o.Values = make([]string, len(other.Values))
	copy(o.Values, other.Values)
	o.mtime = other.mtime
	o.fileinfo = other.fileinfo
}

// FindFile locates a specific option and returns its index within the
// receiver's Values.  If no option is found -1 is returned.
//
func (o *Options) FindFile(s string) int {
	for i := 0; i < len(o.Values); i++ {
		if o.Values[i] == s {
			return i
		}
	}
	return -1
}

// ReadFromFile reads options from a text file.
//
// Options files are word-based with each non-blank, non-comment line
// being split into space-separated fields.
//
// An optional 'filter' function can be supplied which is applied to
// each option word before it is added to the receiver's Values slice.
//
// Returns true if everything worked, false if the file could not be
// read for some reason.
//
func (o *Options) ReadFromFile(filename string, filter func(string) string) (bool, error) {
	filename, _ = FindFile(filename, PlatformSpecific)
	if filter == nil {
		filter = func(s string) string { return s }
	}
	file, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer file.Close()
	o.Path = filename
	info, err := file.Stat()
	if err != nil {
		return true, err
	}
	o.mtime = info.ModTime()
	o.fileinfo = info
	input := bufio.NewScanner(file)
	for input.Scan() {
		line := strings.TrimSpace(input.Text())
		if line == "#include" {
			path, _, found, err := FindFileFromDirectory(filepath.Base(filename), filepath.Dir(CurrentDirectory), nil)
			if err != nil {
				return false, err
			}
			if !found {
				return false, fmt.Errorf("no file to '#include' in file %q", filename)
			}
			if Debug {
				log.Println(filename, "#include ->", path)
			}
			ok, err := o.ReadFromFile(path, filter)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, fmt.Errorf("error reading '#include' file %q", path)
			}
			continue
		}
		if line == "" || line[0] == '#' {
			continue
		}
		fields := strings.Fields(line)
		for _, field := range fields {
			if field = filter(field); field != "" {
				o.Values = append(o.Values, field)
			}
		}
	}
	return true, nil
}

//
// MoreRecentOf returns the most recent modtime of two options.
//
func MoreRecentOf(a *Options, b *Options) time.Time {
	if a.mtime.After(b.mtime) {
		return a.mtime
	}
	return b.mtime
}
