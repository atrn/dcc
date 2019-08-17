# Building dcc

The repo contains everything required. A simple `go build` builds
`dcc` nd `go install` puts it in a _bin_ directory.

The real _full_ build of dcc uses my tool,
[enums](http://github.com/atrn/enums)`, to generate the implementation
of the RunningMode type.

`enums` is a `go generate`` tool that uses a Go-like lanauge to define
enumerated types and generates Go code to implement those
types. `enums` supports a variety of interface implementations to
transform textual enumerators to their numeric values and vice-cerase.

`dcc` uses `enums` to create the file `runningmode.go` file from
`runningmode.enum`.

## enums

You can install `enums` via,

    $ go get github.com/atrn/enums

Given that dcc only has a single enum using `enums` is a little
gratutious and it is really a way to get enums exposure. I do however
use it more intensively and it can remove a lot of repetition,
especially with databases where enums makes it trivial to use
string-valued database columns but integral _enum_ values in Go code
with automatic translation performed in the database layer.
