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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type msvcCompiler struct {
}

func (cl *msvcCompiler) Name() string {
	return "cl"
}

func msvcScrapeShowIncludes(r io.Reader, deps io.Writer, nonDeps io.Writer, source string) {
	const prefix = "Note: including file:"
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, prefix) {
			filename := strings.TrimSpace(strings.TrimPrefix(line, prefix))
			fmt.Fprintln(deps, filename)
		} else if line != source {
			fmt.Fprintln(nonDeps, line)
		}
	}
}

func (cl *msvcCompiler) Compile(source, object, deps string, options []string, stderr io.Writer) error {
	r, w, err := os.Pipe()
	if err != nil {
		return err
	}
	defer r.Close()
	defer w.Close()

	depsFile, err := os.Create(deps)
	if err != nil {
		return err
	}

	fmt.Fprintln(depsFile, object)

	go msvcScrapeShowIncludes(r, depsFile, os.Stdout, filepath.Base(source))

	args := append([]string{}, options...)
	args = append(args, "/nologo", "/showIncludes", "/c", source, "/Fo"+object)
	cmd := exec.Command(cl.Name(), args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, w, stderr
	err = cmd.Run()
	err2 := depsFile.Close()
	if err != nil {
		os.Remove(deps)
		return err
	}
	if err2 != nil {
		os.Remove(deps)
		return err2
	}
	return err
}

func (cl *msvcCompiler) ReadDependencies(path string) (string, []string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", nil, err
	}

	input := bufio.NewScanner(bytes.NewReader(data))
	if !input.Scan() {
		return "", nil, ErrUnexpectedEOF
	}

	target := input.Text()

	var filenames = make([]string, 0, 1000)
	for input.Scan() {
		filename := input.Text()
		filenames = append(filenames, filename)
	}

	return target, filenames, nil
}

// NewMsvcCompiler returns a CompilerDriver using Microsoft's
// cl.exe C/C++ compiler.
func NewMsvcCompiler() Compiler {
	return &msvcCompiler{}
}

func (cl *msvcCompiler) DefineExecutableArgs(exeName string) []string {
	args := make([]string, 1)
	args[0] = fmt.Sprintf("/Fe%s", exeName)
	return args
}
