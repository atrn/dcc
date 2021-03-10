/*
 *  Copyright © 2013. Andy Newman.
 */

package par

import (
	"sync"
)

//  DO calls its arguments, functions, concurrently. Each
//  argument is called from within a separate goroutine with DO only
//  returning when all functions return. If no arguments are passed
//  passed PAR returns immediately. And, as a small optimisation, if a
//  single function is passed it is called directly from the caller's
//  goroutine. Only if more than one function is supplied are any
//  goroutines created.
//
//  Internally DO uses a sync.WaitGroup to synchronize function
//  return/goroutine termination with panics capatured to ensure
//  correct wait group management.
//
func DO(f ...func()) {
	switch len(f) {
	case 0:
		return
	case 1:
		if f[0] != nil {
			f[0]()
		}
	default:
		var w sync.WaitGroup
		for _, fn := range f {
			if fn != nil {
				w.Add(1)
				go func(fn func()) {
					defer w.Done()
					fn()
				}(fn)
			}
		}
		w.Wait()
	}
}

//  FOR is an  iterative  variaant of  DO, a  "parallel"
//  for-loop. FOR accepts an integer range - [start, limit) - and calls
//  a  function,  concurrently, for  each  value  in that  range.   The
//  function being  passed a single  argument, it's "index"  within the
//  range (an integer). Like DO() the FOR function returns when all its
//  calls to the supplied function return.
//
func FOR(start, limit int, f func(int)) {
	if f == nil {
		return
	}
	switch n := limit - start; {
	case n < 1:
		return
	case n == 1:
		f(start)
		return
	}
	var w sync.WaitGroup
	for i := start; i < limit; i++ {
		w.Add(1)
		go func(i int) {
			defer w.Done()
			f(i)
		}(i)
	}
	w.Wait()
}

//  DOfn returns a function that, when called, invokes the
//  DO function (aka PAR) over the arguments passed to DOfn. This
//  function is for used to create functions for calls to DO.
//
func DOfn(f ...func()) func() {
	return func() {
		DO(f...)
	}
}

//  FORfn function returns a function that, when called, invokes the
//  FOR "control structure" function (aka PAR_FOR) over the arguments.
//  This function is for used in creating functions for calls to DO.
//
func FORfn(start, limit int, f func(int)) func() {
	return func() {
		FOR(start, limit, f)
	}
}
