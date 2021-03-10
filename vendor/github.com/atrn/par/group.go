/*
 *  Copyright © 2013. Andy Newman.
 */

package par

import "sync"

// Group is a group of goroutines that may be waited upon.
//
type Group struct {
	wg sync.WaitGroup
}

// Add adds a function to the group of goroutines by
// calling the function in a new goroutine.
//
func (g *Group) Add(fn func(interface{}), arg interface{}) {
	if fn == nil {
		return
	}
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		fn(arg)
	}()
}

// Wait blocks the caller until all goroutines in the group are
// complete.
//
func (g *Group) Wait() {
	g.wg.Wait()
}
