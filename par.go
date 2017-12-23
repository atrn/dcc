// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//
package main

import "sync"

// PAR calls some number of functions concurrently and waits for them all to complete
// before returning. It is named after the occam keyword of similar function.
//
func PAR(f ...func()) {
	var w sync.WaitGroup
	for _, fn := range f {
		w.Add(1)
		go func(fn func()) {
			defer w.Done()
			fn()
		}(fn)
	}
	w.Wait()
}

// PARfor calls a single function N times concurrently, waiting for
// all calls to complete before returning.  Each call is passed its,
// unique, 'iteration number' or 'id' or 'index', an integer in the
// range [start, end).
//
func PARfor(start, end int, f func(int)) {
	var w sync.WaitGroup
	for i := start; i < end; i++ {
		w.Add(1)
		go func(i int) {
			defer w.Done()
			f(i)
		}(i)
	}
	w.Wait()
}
