// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//

package main

import (
	"log"
	"runtime"
)

// The Platform type defines a number of platform-specific values and
// functions. Code defines a single global value, platform, of type
// Platform which is defined in the platform_xxx.go file used for the
// target system.
//
type Platform struct {
	DefaultCC         string
	DefaultCXX        string
	DefaultDepsDir    string
	ObjectFileSuffix  string
	StaticLibPrefix   string
	StaticLibSuffix   string
	DynamicLibPrefix  string
	DynamicLibSuffix  string
	PluginPrefix      string
	PluginSuffix      string
	DefaultExecutable string
	LibraryPaths      []string
	CreateLibrary     func(string, []string) error
	CreateDLL         func(string, []string, []string, []string, []string) error
	CreatePlugin      func(string, []string, []string, []string, []string) error
	SelectTarget      func(*Platform, string) error
}

// StaticLibrary transforms a filename "stem" to the name of a
// static library on the host platform.
//
func (p *Platform) StaticLibrary(name string) string {
	return p.StaticLibPrefix + name + p.StaticLibSuffix
}

// DynamicLibrary transforms a filename "stem" to the name of
// a dynamic library on the host platform.
//
func (p *Platform) DynamicLibrary(name string) string {
	return p.DynamicLibPrefix + name + p.DynamicLibSuffix
}

// PluginFile transforms a filename "stem" to the name of
// a "plugin" on the host platform.
//
func (p *Platform) PluginFile(name string) string {
	return p.PluginPrefix + name + p.PluginSuffix
}

// PlatformSpecific is a pathname filter function used with the
// MustFindFile function to transform pathnames to so-called platform-
// specific versions so they may be used in preference to other files.
//
func PlatformSpecific(path string) string {
	try := func(path string) (string, bool) {
		if DebugFind {
			log.Printf("DEBUG FIND: Trying %q", path)
		}
		_, err := Stat(path)
		return path, err == nil
	}
	if p, ok := try(path + "." + runtime.GOOS + "_" + runtime.GOARCH); ok {
		return p
	}
	if p, ok := try(path + "." + runtime.GOOS); ok {
		return p
	}
	return path
}
