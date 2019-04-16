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
	"strings"
	"sync"
)

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

	// cmd.Run/cmd.Start are not safe to use concurrently (cmd.Run
	// sometimes fails in strange ways, only on Linux so far).
	// Calls to cmd.Start are serialized via a mutex.  I'll assume
	// Wait must be safe otherwise we can't use os/exe to run
	// multiple commands at the same time.
	//
	serializeCmdStart.Lock()

	cmd := exec.Command(path, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, os.Stdout, stderr
	err := cmd.Start()

	serializeCmdStart.Unlock()

	if err == nil {
		err = cmd.Wait()
	}
	return err
}
