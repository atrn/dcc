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
	"io"
	"os"
	"os/exec"
)

var errNotImplemented = errors.New("MSVC support not yet implemented")

type msvcCompiler struct {
}

func (cl *msvcCompiler) Name() string {
	return "cl"
}

func (cl *msvcCompiler) Compile(source, object, deps string, options []string, w io.Writer) error {
	args := append([]string{}, options...)
	args = append(args, "/showIncludes", "/c", source, "/Fo", object)
	cmd := exec.Command(cl.Name(), args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return err
}

func (cl *msvcCompiler) ReadDependencies(path string) (string, []string, error) {
	return "", nil, errNotImplemented
}

// NewMsvcCompiler returns a CompilerDriver using Microsoft's
// cl.exe C/C++ compiler.
//
func NewMsvcCompiler() Compiler {
	return &msvcCompiler{}
}
