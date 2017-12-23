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
	"sync"
)

var (
	// StatCache is a cache of os.Stat results.
	//
	StatCache = make(map[string]os.FileInfo)

	// StatCacheMutex protects StatCache
	//
	StatCacheMutex sync.Mutex
)

// Stat wraps os.Stat and caches the results.
//
func Stat(path string) (os.FileInfo, error) {

	StatCacheMutex.Lock()
	info, found := StatCache[path]
	StatCacheMutex.Unlock()

	if found {
		return info, nil
	}

	//  We don't hold the mutex while calling os.Stat and this
	//  introduces a race - multiple routines Stat'ing the same
	//  path can overlap and do the same work - call os.Stat and
	//  add the resultant info to the cache. Other than doing
	//  things more than once this doesn't really matter.
	//
	// Given the nature of the paths being Stat'd the FileInfo
	// will be identical. The real effect of the race will be to
	// call os.Stat more than once for a given path, updating the
	// cache with identical data if it succeeds with the
	// associated locking overhead.
	//
	// The likelihood of this race occurring depends the ordering
	// and timing of Stat calls which at first would appear to be
	// essentiallly random but it should be noted that generated
	// dependencies will often list paths in the same order and
	// that the initial compilation jobs will all start at about
	// the same time.
	//
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	StatCacheMutex.Lock()
	if _, found = StatCache[path]; !found {
		StatCache[path] = info
	}
	StatCacheMutex.Unlock()

	return info, nil
}

// ClearCachedStat removes an entry from the cache. This is required after compilation
// to ensure we re-stat the file and get its new modtime.
//
func ClearCachedStat(path string) {
	StatCacheMutex.Lock()
	delete(StatCache, path)
	StatCacheMutex.Unlock()
}
