// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//

package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

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

var serializeCmdStart sync.Mutex

// Exec executes a command with the supplied arguments and directs its
// standard error output stream to the supplied io.Writer. The
// command's standard input is connected to /dev/null and the output
// stream connected to our standard output.
//
func Exec(path string, args []string, stderr io.Writer) error {
	if Debug {
		log.Println("EXEC:", path, strings.Join(args, " "))
	}
	cmd := exec.Command(path, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, os.Stdout, stderr

	// cmd.Run/cmd.Start are not safe to use concurrently (cmd.Run
	// sometimes fails in strange ways, only on Linux so far).
	// Calls to cmd.Start are serialized via a mutex.  I'll assume
	// Wait must be safe otherwise we can't use os/exe to run
	// multiple commands at the same time.
	//
	serializeCmdStart.Lock()
	err := cmd.Start()
	serializeCmdStart.Unlock()

	if err == nil {
		err = cmd.Wait()
	}
	return err
}
