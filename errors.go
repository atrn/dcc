// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//

package main

import "errors"

var (
	// ErrUnexpectedEOF means the dependency file is empty.
	ErrUnexpectedEOF = errors.New("unexpected end of file")

	// ErrNoColon means the first line has no ':' (target definition)
	ErrNoColon = errors.New("expected a make-target on line 1 of dependency file, no ':' found")

	// ErrMultipleTargets means more than one target file was detected.
	ErrMultipleTargets = errors.New("multiple targets found in dependency file")
)
