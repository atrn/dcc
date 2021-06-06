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
// corresponding error value. If the target exists and is newer than
// any of the inputs no linking occurs.
//
func dllOrPlugin(target string, inputs []string, libs *Options, options *Options, otherFiles *Options, frameworks []string, create func(string, []string, []string, []string, []string) error) error {
	createDllOrPlugin := func() error {
		inputs = append(inputs, otherFiles.Values...)
		return create(target, inputs, libs.Values, options.Values, frameworks)
	}
	if IgnoreDependencies {
		return createDllOrPlugin()
	}
	targetInfo, err := Stat(target)
	if os.IsNotExist(err) {
		return createDllOrPlugin()
	}
	if err != nil {
		return err
	}
	if MostRecentModTime(options, libs).After(targetInfo.ModTime()) {
		return createDllOrPlugin()
	}
	newestInput, err := NewestOf(inputs)
	if err != nil {
		return err
	}
	if newestInput.After(targetInfo.ModTime()) {
		return createDllOrPlugin()
	}
	if len(otherFiles.Values) > 0 {
		newestOther, err := NewestOf(otherFiles.Values)
		if err != nil {
			return err
		}
		if newestOther.After(targetInfo.ModTime()) {
			return createDllOrPlugin()
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
			} else if FileIsNewer(libInfo, targetInfo) {
				return createDllOrPlugin()
			}
		}
	}
	return nil
}

func Dll(target string, inputs []string, libs *Options, options *Options, otherFiles *Options, frameworks []string) error {
	return dllOrPlugin(target, inputs, libs, options, otherFiles, frameworks, platform.CreateDLL)
}

func Plugin(target string, inputs []string, libs *Options, options *Options, otherFiles *Options, frameworks []string) error {
	return dllOrPlugin(target, inputs, libs, options, otherFiles, frameworks, platform.CreatePlugin)
}
