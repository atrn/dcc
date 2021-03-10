/*
 *  Copyright © 2013. Andy Newman.
 */

// Package par provides functions to structure concurrent programs.
//
// Package par defines the functions par.DO and par.FOR that implement
// synchronized "processes" (goroutines) in a manner similar to the
// occam programming language's PAR and replicated-PAR concurrent
// control structures.
//
// par.DO mimics occam's PAR, calling some number of functions
// concurrently then waiting for them to complete before it oompletes
// and returns to its caller. par.FOR mimics occam's replicated-PAR
// structure and provides a form of concurrent for-loop with the same
// synchronization semantics as par.DO, only returning when all of its
// goroutines complete.
//
// The process structures implemented using par.DO and par.FOR are
// common.  Synchronizing upon goroutine complementation being common
// enough to be idiomatically implemented using sync.WaitGroup. The
// fact that the idiom exists being all the evidence required.
//
// Implementation
//
// par.DO and par.FOR are implemented using sync.WaitGroup and
// consolidate the various WaitGroup manipulations within the package
// helping remove repetition from user code. Subjectively the
// functions also improve readability and remove the clutter caused by
// the WaitGroup manipulations which can often obscure the actual
// code.
//
// Exanple Usage
//
// 	par.DO(
//	    controlFuelRods,
//	    monitorCoolant,
//	    moveDials,
//	    flashLights,
//	    func() {
//		par.FOR(0, 4, func(i int) {
// 		    ...  run generator i
// 		})
//          },
//      )
//
//
// Nesting PARs
//
// The above example shows nesting of PARs, the function literal
// calling par.FOR. This pattern, a func() that just calls par.FOR is
// common but Go allows us to make it simpler.
//
//     func DOfn(f ...func()) func()
//     func FORfn(start, limit int, f func(int)) func()
//
// The "fn" functions return a func() intended to be passed to par.DO and
// are used in created nested process structures.  FORfn is the most
// useful.
//
// Armed with FORfn we can now write the above code to avoid one
// level of function,
//
// 	par.DO(
// 	    controlFuelRods,
// 	    monitorCoolant,
// 	    moveDials,
// 	    flashLights,
// 	    par.FORfn(0, 4, func(i int) {
// 	        ...  run generator i
// 	    }),
//     )
//
// Groups, dynamic PAR
//
// To support dynamic use, the package defines the type Group that
// wraps sync.WaitGroup and uses methods specific to starting
// goroutines and waiting for them.
//
// The user defines a variable of type par.Group and then calls the
// Add and Wait methods to start new goroutines and wait for them to
// complete.
//
//	var g par.Group
//	for in := range channel {
//		g.Add(fn, in)
//	}
//	g.Wait()
//
// More Advanced Example
//
// A common process structure is "fan-out", dividing the processing of
// "work" among some number of "workers" and collecting the results of
// that processing. Go of course provides all the means to implement
// such structures but doing so using standard packages can lead to
// cluttered and messy code which obscures the actual work being
// done.
//
// par.DO and par.FOR help avoid that clutter and help visualize the
// process structure. And in the case of our "fan-out" example also
// allow for correct resource management, i.e. we close channels at
// the correct times.
//
// A some-what generic "fan-out" based process structure to process
// the work held in the workToBeDone collection using Nworkers to Do()
// the work. The call to par.DO returns when all work is done.
//
//
//	toWorkers := make(chan Work)	  // for some type of Work
//	fromWorkers := make(chan Result)  // for some type of Result
//
//	par.DO(
//		func() { // feed the workers work
//			for _, work := range workToBeDone {
//				toWorkers <- work
//			}
//			close(toWorkers)
//		},
//		func() { // process work using Nworkers
//			par.FOR(0, Nworkers, func(int) {
//				for work := range toWorkers {
//					fromWorkers <- Do(work)
//				}
//			})
//			close(fromWorkers)
//		},
//		func() { // collect the results
//			for result := range fromWorkers {
//				...  do something with each result
//			}
//		},
//	)
//
//
package par
