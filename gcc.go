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
	"bytes"
	"io"
	"io/ioutil"
	"strings"
)

// GccStyleCompiler is-a CompilerDriver that uses gcc-style options to
// generate make-format dependencies.
//
type GccStyleCompiler struct {
	command string
}

// NewGccStyleCompiler returns a CompilerDriver using
// gcc-style options that will run the supplied command
// name. This is used to run gcc, clang and icc and any
// compiler with "gcc"" in its command name (e.g. typically
// named cross compilers).
//
func NewGccStyleCompiler(name string) Compiler {
	return &GccStyleCompiler{
		command: name,
	}
}

// Name returns the Compiler's name.
//
func (gcc *GccStyleCompiler) Name() string {
	return gcc.command
}

// Compile runs the compiler to compile a source code to object code.
//
func (gcc *GccStyleCompiler) Compile(source, object, deps string, options []string, w io.Writer) error {
	args := append([]string{}, options...)
	args = append(args, "-MD", "-MF", deps, "-c", source, "-o", object)
	return Exec(gcc.command, args, w)
}

// ReadDependencies reads make-style dependency specification from the named file
// and returns the names of the target, the dependent files and an error value,
// non-nil if the file failed to be parsed or opened.
//
func (gcc *GccStyleCompiler) ReadDependencies(path string) (string, []string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", nil, err
	}

	input := bufio.NewScanner(bytes.NewReader(data))
	input.Split(bufio.ScanWords)
	if !input.Scan() {
		return "", nil, ErrUnexpectedEOF
	}

	target := input.Text()
	if strings.HasSuffix(target, ":") {
		target = target[0 : len(target)-1]
	} else if !input.Scan() || input.Text() != ":" {
		return "", nil, ErrNoColon
	}

	var filenames = make([]string, 0, 1000)
	for input.Scan() {
		filename := input.Text()
		filename = strings.TrimSuffix(filename, "\\")
		if filename == "" {
			continue
		}
		if filename == ":" || strings.HasSuffix(filename, ":") {
			return "", nil, ErrMultipleTargets
		}
		filenames = append(filenames, filename)
	}

	return target, filenames, nil
}
