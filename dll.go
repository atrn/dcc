// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//
package main

import (
	"os"
	"strings"
)

// Dll runs the compiler to link the given inputs and create the given
// shared/dynamic library target using the supplied options. Returns a
// corresponding error value. If the target exists and is newer than any
// of the inputs no linking occurs.
//
func Dll(target string, inputs []string, libs *Options, options *Options, otherFiles *Options, frameworks []string) error {
	createDll := func() error {
		inputs = append(inputs, otherFiles.Values...)
		return platform.CreateDLL(target, inputs, libs.Values, options.Values)
	}
	if IgnoreDependencies {
		return createDll()
	}
	targetInfo, err := Stat(target)
	if os.IsNotExist(err) {
		return createDll()
	}
	if err != nil {
		return err
	}
	if MoreRecentOf(options, libs).After(targetInfo.ModTime()) {
		return createDll()
	}
	newestInput, err := Newest(inputs)
	if err != nil {
		return err
	}
	if newestInput.After(targetInfo.ModTime()) {
		return createDll()
	}
	if len(otherFiles.Values) > 0 {
		newestOther, err := Newest(otherFiles.Values)
		if err != nil {
			return err
		}
		if newestOther.After(targetInfo.ModTime()) {
			return createDll()
		}
	}
	skipNext := false
	for _, name := range libs.Values {
		if skipNext {
			skipNext = false
			continue
		}
		if name == "-framework" {
			skipNext = true
			continue
		}
		if !strings.HasPrefix(name, "-l") {
			if libInfo, err := Stat(name); err != nil {
				return err
			} else if IsNewer(libInfo, targetInfo) {
				return createDll()
			}
		}
	}
	return nil
}
