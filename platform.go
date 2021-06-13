// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//

package main

// The Platform type defines a number of platform-specific values and
// functions. Code defines a single global value, platform, of type
// Platform which is defined in the platform_xxx.go file used for the
// target system.
//
type Platform struct {
	DefaultCC         string
	DefaultCXX        string
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
