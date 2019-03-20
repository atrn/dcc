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
)

// Lib ceates a static library from the inputs iff the inputs are
// newer than any existing target.
//
func Lib(target string, inputs []string) error {
	createLib := func() error {
		return platform.CreateLibrary(target, inputs)
	}
	if IgnoreDependencies {
		return createLib()
	}
	targetInfo, err := Stat(target)
	if os.IsNotExist(err) {
		return createLib()
	} else if err != nil {
		return err
	}
	newestInput, err := NewestOf(inputs)
	if err == nil {
		if newestInput.After(targetInfo.ModTime()) {
			return createLib()
		}
	}
	return err
}
