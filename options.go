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
// file and act as a dependency to the build.
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

// Len returns the number of values defined by the receiver.
//
func (o *Options) Len() int {
	return len(o.Values)
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

// Prepend inserts an option at the start of the set of options.
// Note, Prepend does NOT modify the mtime of the receiver.
//
func (o *Options) Prepend(option string) {
	o.Values = append([]string{option}, o.Values...)
}

// SetModTime sets the receiver's modification time to the supplied
// value.
//
func (o *Options) SetModTime(t time.Time) {
	o.mtime = t
}

func (o *Options) ModTime() time.Time {
	return o.mtime
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
	ofilename := filename
	filename, _ = FindFile(filename, PlatformSpecific)
	if Debug {
		log.Printf("OPTIONS: %q -> actual %q", ofilename, filename)
	}
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
		line := os.ExpandEnv(strings.TrimSpace(input.Text()))
		if strings.HasPrefix(line, "#include") {
			if err := o.includeFile(line, filename, filter); err != nil {
				return false, err
			}
			continue
		}
		if strings.HasPrefix(line, "#inherit") {
			if err := o.inheritFile(line, filename, filter); err != nil {
				return false, err
			}
			continue
		}
		if line == "" || line[0] == '#' {
			continue
		}
		fields := strings.Fields(line)
		for _, field := range fields {
			if field[0] == '$' {
				field = os.Getenv(field[1:])
				if field == "" {
					continue
				}
			}
			if field = filter(field); field != "" {
				o.Values = append(o.Values, field)
			}
		}
	}
	return true, nil
}

func undelim(s string, start, end byte) string {
	switch n := len(s) - 1; {
	case n < 1:
		return s
	case s[0] != start:
		return s
	case s[n] != end:
		return s
	default:
		return s[1:n]
	}
}

func extractFilename(filename string) string {
	if len(filename) < 2 {
		return filename
	}
	if filename[0] == '"' {
		return undelim(filename, '"', '"')
	}
	if filename[0] == '<' {
		return undelim(filename, '<', '>')
	}
	return filename
}

func getFilename(why, line, filename string) (string, error) {
	malformed := func() (string, error) {
		return "", fmt.Errorf("%q - malformed '%s' line", why, line)
	}
	dir := filepath.Dir(filename)
	filename = filepath.Base(filename) // CXXFLAGS, CFLAGS, etc...
	fields := strings.Fields(line)
	switch numFields := len(fields); {
	case numFields == 2:
		return filepath.Join(dir, extractFilename(fields[1])), nil
	case numFields == 1:
		if why == "#inherit" {
			return filename, nil
		}
		return malformed()
	default:
		return malformed()
	}
}

func (o *Options) includeFile(line string, filename string, filter func(string) string) error {
	path, err := getFilename("#include", line, filename)
	if err != nil {
		return err
	}
	if Debug {
		log.Printf("DEBUG: %q include -> %q", filename, path)
	}
	_, err = o.ReadFromFile(path, filter)
	return err
}

func (o *Options) inheritFile(line string, filename string, filter func(string) string) error {
	inheritedFilename, err := getFilename("#inherit", line, filename)
	if err != nil {
		return err
	}
	if filepath.Dir(inheritedFilename) != "." {
		return fmt.Errorf("%q: '#inherit' filename cannot contain path elements", inheritedFilename)
	}
	path, _, found, err := FindFileFromDirectory(
		inheritedFilename,
		filepath.Clean(filepath.Join(filepath.Dir(filename), "..")),
		nil,
	)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("%q: #inherited file not found", inheritedFilename)
	}

	if Debug {
		log.Printf("DEBUG: %q #inherit -> %q", filename, path)
	}

	ok, err := o.ReadFromFile(path, filter)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("%q: error reading file", path)
	}

	return nil
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
