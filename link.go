// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Link runs the compiler to link the given inputs and create the given
// target using the supplied options. Return a corresponding error
// value. If the target exists and is newer than any of the inputs no
// linking occurs.
//
func Link(target string, inputs []string, libs *Options, options *Options, otherFiles *Options, frameworks []string) error {
	if target == "" {
		target = platform.DefaultExecutable
	}
	link := func() error {
		var args []string
		args = append(args, options.Values...)
		args = append(args, inputs...)
		args = append(args, endash(libs.Values)...)
		args = append(args, frameworks...)
		args = append(args, "-o", target)
		if !Quiet {
			if Verbose {
				fmt.Fprintln(os.Stderr, ActualCompiler.Name(), strings.Join(args, " "))
			} else {
				fmt.Fprintln(os.Stderr, "ld", target)
			}
		}
		return Exec(ActualCompiler.Name(), args, os.Stderr)
	}
	if IgnoreDependencies {
		return link()
	}
	targetInfo, err := Stat(target)
	if os.IsNotExist(err) {
		return link()
	}
	if err != nil {
		return err
	}
	if MoreRecentOf(options, libs).After(targetInfo.ModTime()) {
		return link()
	}
	newestInput, err := NewestOf(inputs)
	if err != nil {
		return err
	}
	if newestInput.After(targetInfo.ModTime()) {
		return link()
	}
	if len(otherFiles.Values) > 0 {
		newest, err := NewestOf(otherFiles.Values)
		if err != nil {
			return err
		}
		if newest.After(targetInfo.ModTime()) {
			return link()
		}
	}
	for _, name := range libs.Values {
		if !strings.HasPrefix(name, "-l") {
			if libInfo, err := Stat(name); err != nil {
				return err
			} else if FileIsNewer(libInfo, targetInfo) {
				return link()
			}
		}
	}
	return nil
}

var standardPath map[string]struct{} = nil

func endash(values []string) (dashed []string) {
	if standardPath == nil {
		standardPath = make(map[string]struct{})
		for _, name := range platform.LibraryPaths {
			standardPath[name] = struct{}{}
		}
	}
	for _, name := range values {
		dir, base := filepath.Dir(name), filepath.Base(name)
		if _, found := standardPath[dir]; found {
			s := strings.TrimSuffix(base, platform.StaticLibSuffix)
			if s == base {
				s = strings.TrimSuffix(base, platform.DynamicLibSuffix)
			}
			t := strings.TrimPrefix(s, platform.StaticLibPrefix)
			if t == s {
				t = strings.TrimPrefix(s, platform.DynamicLibPrefix)
			}
			dashed = append(dashed, "-l" + t)
		} else {
			dashed = append(dashed, name)
		}
	}
	return dashed
}
