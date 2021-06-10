// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Assert that a predicate is true, panic if not.
//
func Assert(predicate bool, message string) {
	if !predicate {
		_, filename, line, ok := runtime.Caller(1)
		if ok {
			panic(fmt.Errorf("%s (%s:%d)", message, filename, line))
		} else {
			panic(errors.New(message))
		}
	}
}

// GetProgramName returns the command name used to run this program.
//
func GetProgramName() string {
	name := filepath.Base(os.Args[0])
	if runtime.GOOS == "windows" {
		// Sometimes you get uppercase paths on Windows.
		if strings.HasSuffix(name, ".EXE") {
			name = strings.TrimSuffix(name, ".EXE")
		} else {
			name = strings.TrimSuffix(name, ".exe")
		}
	}
	return name
}

// Mkdir makes a directory using os.MkdirAll to ensure
// any parent directories exist.
//
func Mkdir(path string) error {
	return os.MkdirAll(path, 0777)
}

// Getenv returns the value of an environment variable or a default
// value if the variable is not set in the environment.
//
func Getenv(key, value string) string {
	if s := os.Getenv(key); s != "" {
		return s
	}
	return value
}

// GetenvInt returns the value of an integer-valued environment
// variable or a default value if it is not set or fails to parse.
//
func GetenvInt(key string, value int) int {
	if s := os.Getenv(key); s != "" {
		n, err := strconv.Atoi(s)
		if err == nil {
			return n
		}
		log.Print("environment variable ", key, " has invalid value: ", err)
		return value
	}
	return value
}

// ConfigureLogger configures the global logger to makes its output
// more suitable for a command line tool. We turn off all date/time
// output and set the log message prefix based on the program name.
//
func ConfigureLogger(prefix string) {
	log.SetFlags(0)
	log.SetPrefix(prefix + ": ")
}

// MustGetwd returns PWD without requiring users to deal with that
// pesky error. Any errors are fatal.
//
func MustGetwd() string {
	s, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return s
}

// Remove any, single byte, delimiters from a possibly delimited
// string.
//
func RemoveDelimiters(s string, start, end byte) string {
	switch n := len(s); {
	case n < 2:
		return s
	case s[0] != start:
		return s
	case s[n-1] != end:
		return s
	default:
		return s[1 : n-1]
	}
}

//
func LowercaseFilenameExtension(path string) string {
	return strings.ToLower(filepath.Ext(path))
}
