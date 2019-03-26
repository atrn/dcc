// DO NOT EDIT.
// Generated: 2019-03-26 17:15:17.083389 +1100 AEDT m=+0.000946167
// From: runningmode.enum

package main

type RunningMode int

const (
	RunningMode_Zero_ RunningMode = iota
	ModeNotSpecified
	CompileAndLink
	CompileSourceFiles
	CompileAndMakeLib
	CompileAndMakeDLL
	CleanupOutputFiles
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
	case CleanupOutputFiles:
		return "CleanupOutputFiles"
	default:
		return "*!* INVALID RunningMode VALUE *!*"
	}
}


