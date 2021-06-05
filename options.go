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
	"io"
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
	actualFilename, _ := FindFile(filename, PlatformSpecific)
	if Debug {
		log.Printf("OPTIONS: %q -> actual filename %q", filename, actualFilename)
	}
	file, err := os.Open(actualFilename)
	if err != nil {
		return false, err
	}
	defer file.Close()
	o.Path = actualFilename
	info, err := file.Stat()
	if err != nil {
		return true, err
	}
	o.mtime = info.ModTime()
	o.fileinfo = info
	return o.ReadFromReader(file, filename, filter)
}

// Read options from the given io.Reader.
//
func (o *Options) ReadFromReader(r io.Reader, filename string, filter func(string) string) (bool, error) {
	if filter == nil {
		filter = func(s string) string { return s }
	}
	var conditional Conditional
	input := bufio.NewScanner(r)
	lineNumber := 0
	for input.Scan() {
		line := input.Text()
		lineNumber++

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		evalCondition := func(invert bool) error {
			if len(fields) != 2 {
				return reportErrorInFile(filename, lineNumber, fmt.Sprintf("%s requires a single parameter", fields[0]))
			}
			val := os.Getenv(fields[1])
			state1, state2 := TrueConditionState, FalseConditionState
			if invert {
				state1, state2 = FalseConditionState, TrueConditionState
			}
			if val != "" {
				conditional.PushState(state1)
			} else {
				conditional.PushState(state2)
			}
			return nil
		}

		if fields[0] == "#error" {
			if !conditional.IsSkippingLines() {
				message := strings.Join(fields[1:], " ")
				if message == "" {
					message = "#error raised without message"
				}
				return false, reportErrorInFile(filename, lineNumber, message)
			}
			continue
		}

		if fields[0] == "#ifdef" {
			if conditional.IsSkippingLines() {
				conditional.PushState(conditional.CurrentState())
			} else if err := evalCondition(false); err != nil {
				return false, err
			} else {
				continue
			}
		}

		if fields[0] == "#ifndef" {
			if conditional.IsSkippingLines() {
				conditional.PushState(conditional.CurrentState())
			} else if err := evalCondition(true); err != nil {
				return false, err
			} else {
				continue
			}
		}

		if fields[0] == "#else" {
			if !conditional.IsActive() {
				return false, reportErrorInFile(filename, lineNumber, ErrNoCondition.Error())
			}
			conditional.ToggleState()
			continue
		}

		if fields[0] == "#endif" {
			if !conditional.IsActive() {
				return false, reportErrorInFile(filename, lineNumber, ErrNoCondition.Error())
			}
			if err := conditional.PopState(); err != nil {
				return false, err
			}
			continue
		}

		if conditional.IsSkippingLines() {
			continue
		}

		// #include <filename>
		if fields[0] == "#include" {
			if err := o.includeFile(filename, lineNumber, line, fields, filter); err != nil {
				return false, err
			}
			continue
		}

		// #inherit
		if fields[0] == "#inherit" {
			if err := o.inheritFile(filename, lineNumber, line, fields, filter); err != nil {
				return false, err
			}
			continue
		}

		// Any other line that starts with a '#' is a comment.
		if line[0] == '#' {
			continue
		}

		// Otherwise, treat fields (tokens) as options to be included.
		// Expand (interpolate) any variable references, filter and
		// collect any non-empty strings.
		for _, field := range fields {
			field = os.ExpandEnv(field)
			fields2 := strings.Fields(field)
			for _, field2 := range fields2 {
				if field2 = filter(field2); field2 != "" {
					o.Values = append(o.Values, field2)
				}
			}
		}
	}
	return true, nil
}

func removeDelimiters(s string, start, end byte) string {
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
		return removeDelimiters(filename, '"', '"')
	}
	if filename[0] == '<' {
		return removeDelimiters(filename, '<', '>')
	}
	return filename
}

func reportErrorInFile(filename string, lineNumber int, what string) error {
	return fmt.Errorf("error: %s:%d %s", filename, lineNumber, what)
}

func malformedLine(filename string, lineNumber int, what, line string) error {
	return reportErrorInFile(filename, lineNumber, fmt.Sprintf("malformed %s - %s", what, line))
}

func (o *Options) includeFile(parentFilename string, lineNumber int, line string, fields []string, filter func(string) string) error {
	if len(fields) != 2 {
		return malformedLine(parentFilename, lineNumber, "#include", line)
	}
	name := extractFilename(fields[1])
	path := filepath.Join(filepath.Dir(parentFilename), name)
	if Debug {
		log.Printf("DEBUG: %q include -> %q", parentFilename, path)
	}
	_, err := o.ReadFromFile(path, filter)
	return err
}

func (o *Options) inheritFile(parentFilename string, lineNumber int, line string, fields []string, filter func(string) string) error {
	if len(fields) != 1 {
		return malformedLine(parentFilename, lineNumber, "#inherit", line)
	}

	inheritedFilename := filepath.Base(parentFilename)
	if filepath.Dir(inheritedFilename) != "." {
		return reportErrorInFile(parentFilename, lineNumber, fmt.Sprintf("filename parameter to '#inherit' cannot contain path elements %q", line))
	}
	path, _, found, err := FindFileFromDirectory(
		inheritedFilename,
		filepath.Clean(filepath.Join(filepath.Dir(parentFilename), "..")),
		nil,
	)
	if err != nil {
		return err
	}
	if !found {
		return reportErrorInFile(parentFilename, lineNumber, fmt.Sprintf("#inherited file %q not found", inheritedFilename))
	}

	if Debug {
		log.Printf("DEBUG: %q #inherit -> %q", parentFilename, path)
	}

	ok, err := o.ReadFromFile(path, filter)
	if err != nil {
		return err
	}

	if !ok {
		return reportErrorInFile(parentFilename, lineNumber, fmt.Sprintf("error reading inherited file %q", path))
	}

	return nil
}

//
// MostRecentModTime returns the modification time of the most recently
// modified of the two Options.
//
func MostRecentModTime(a *Options, b *Options) time.Time {
	if a.mtime.After(b.mtime) {
		return a.mtime
	}
	return b.mtime
}
