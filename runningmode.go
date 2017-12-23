// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//
package main

// RunningMode defines the in which dcc was invoked.
//
// We derive the mode from the compiler and dcc options and can then
// just check the mode to determine actions.
//
type RunningMode int

const (
	// ModeNotSpecified means the user hasn't explicitly defined a
	// running mode, yet.  This is the initial, default, mode.
	ModeNotSpecified RunningMode = iota

	// CompileAndLink means "build an exectable", the usual
	// default cc(1) behaviour.
	CompileAndLink

	// CompileSourceFiles means the user supplied a -c switch
	// and no linking is done, only compilation.
	CompileSourceFiles

	// CompileAndMakeLib means the user supplied the dcc --lib switch
	// to create a static library.
	CompileAndMakeLib

	// CompileAndMakeDLL means the user supplied the dcc --dll switch.
	CompileAndMakeDLL

	// CleanupOutputFiles means the user supplied the dcc --clean switch.
	CleanupOutputFiles // --clean
)

func (rm RunningMode) String() string {
	switch rm {
	case ModeNotSpecified:
		return "ModeNotSpecified"
	case CompileSourceFiles:
		return "CompileSourceFiles"
	case CompileAndLink:
		return "CompileAndLink"
	case CompileAndMakeLib:
		return "CompileAndMakeLib"
	case CompileAndMakeDLL:
		return "CompileAndMakeDLL"
	case CleanupOutputFiles:
		return "CleanupOutputFiles"
	default:
		panic("invalid RunningMode value")
	}
}
