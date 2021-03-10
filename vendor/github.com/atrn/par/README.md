# par - occam-style concurrency _primitives_

The par  package provides functions that implement  occam-style PAR
and replicated-PAR control structures. These provide synchronization
upon goroutine completion in the same way as idiomatic `sync.WaitGroup`
usage.

The `par.DO` function calls some number of functions, concurrently, and waits for
them to complete before it returns. The `par.FOR` function calls a _single_ function
a number of times defined by two integers. Each call occurs concurrently and as with
`par.DO`, `par.FOR` only returns when all of its function _calls_ complete.

`par.DO` mimics the occam `PAR` keyword and `par.FOR` the
_replicated-PAR_ construct (a concurrent _for_-loop). In Go
the functions are implemented around `sync.WaitGroup` and hide
the repetitive _clutter_ of wait group manipulations.

## An example

Imagine we have some functions that run loops to do some control
operation. In our system we run these concurrently, perhaps they
communicate but that's detail we can ignore for the time being.
We run then concurrently and wait for them to finish.

	par.DO(
		ControlFuelRods,
		MonitorCoolant,
		MoveDials,
		FlashLights,
		ControlSirens,
		func() {
			par.FOR(0, 10, func(i int) {
				MonitorDoor(i+1)
			})
		},
        )


## Hiding sync.WaitGroup

The `par` functions encapsulate  the, now common,  idiom of using  a 
`sync.WaitGroup`  to synchronize  goroutine  completion.   The `par`
functions offer  no actual new  functionality over what direct  use of
sync.WaitGroup affords, and actually provide  less, but their use does
make  for cleaner  code by  hiding the  implementation details  of the
synchronization.  The  functions eliminate clutter making  the process
structure  more obvious  and  therefore more  easily comprehended  and
maintained (i.e. not broken).

## Abusing import

We can abuse Go's `import .` to let us use the package's functions
without qualification. This makes them seem a little more like using
a language construct.

Importing the package using,

	import . "github.com/atrn/par"

lets us write,

	DO(
		ControlFuelRods,
		MonitorCoolant,
		MoveDials,
		FlashLights,
		ControlSirens,
		func() {
			FOR(0, 10, func(number int) {
				MonitorDoor(number)
		},
        )


That looks okay, if you accept the namespace pollution, but
DO() and FOR() are a little too generic and not that descriptive.

## Synonyms, PAR and PAR_FOR 

The package define synonyms for DO and FOR, PAR and PAR_FOR.
Using these the code becomes,

	PAR(
		ControlFuelRods,
		MonitorCoolant,
		MoveDials,
		FlashLights,
		ControlSirens,
		func() {
			PAR_FOR(0, 10, func(number int) {
				MonitorDoor(number)
		},
        )
        

## Nested PARs

Each of the above examples shows nesting of PAR via the
function literal calling par.FOR aka PAR_FOR. This pattern,
a func() that just calls par.FOR is common, luckily Go lets
us simplify it.

The package defines what it refers to as _fn_ function
(I never thought of a good name).

	func FORfn(start, limit int, f func(int)) func()

The returned function calls par.FOR using the supplied
arguments and is passed to par.DO as one of its functions
to call concurrently. As with par.DO and par.FOR, par.FORfn
has a synonym intended to be used via `import .` - PAR_FORfn.

Armed with PAR_FORfn we can write,

	PAR(
	    	ControlFuelRods,
	    	MonitorCoolant,
	    	MoveDials,
	    	FlashLights,
		ControlSirens,
	    	PAR_FORfn(0, 10, func(number int) {
			MonitorDoor(number)
		}),
        )


## Classic fanout

	jobs := make(chan Work)
	results := make(chan Result)
	par.DO(
		func() {
			for job := range Jobs() {
				jobs <- job
			}
			close(work)
		},
		func() {
			par.FOR(0, Nworkers, func(int) {
				for job := range jobs {
					results <- Process(job)
				}
			}
			close(results)
		},
		func() {
			for result := range results {
				Consume(result)
			}
		},
	)

Removing the explicit sync.WaitGroup use makes the
process structure easier to comprehend (and may
help stop the endless complaints about multiple
channel closes).
