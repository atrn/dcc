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
)

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
	return cmd.Run()
}
