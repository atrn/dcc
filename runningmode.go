// DO NOT EDIT.
//
// Generated: Fri Apr 19 02:39:20 2019
// From:      runningmode.enum
// By:        andy
//

package main

type RunningMode int

const (
	RunningMode_Zero_ RunningMode = iota
	ModeNotSpecified
	CompileAndLink
	CompileSourceFiles
	CompileAndMakeLib
	CompileAndMakeDLL
	CompileAndMakePlugin
)

func (v RunningMode) String() string {
	switch v {
	case RunningMode_Zero_:
		return "*!!!* UNINITIALIZED RunningMode VALUE *!!!*"
	case ModeNotSpecified:
		return "ModeNotSpecified"
	case CompileAndLink:
		return "CompileAndLink"
	case CompileSourceFiles:
		return "CompileSourceFiles"
	case CompileAndMakeLib:
		return "CompileAndMakeLib"
	case CompileAndMakeDLL:
		return "CompileAndMakeDLL"
	case CompileAndMakePlugin:
		return "CompileAndMakePlugin"
	default:
		return "*!* INVALID RunningMode VALUE *!*"
	}
}
