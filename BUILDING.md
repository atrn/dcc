# Building dcc

The repo contains everything required. A simple `go build` builds
`dcc` and `go install` puts it in a user's (Go) _bin_ directory.

    $ go build
    $ go install

The complete build of dcc uses [enums](http://github.com/atrn/enums)`
to generate the runningmode,go file.

## enums

`enums` is a `go generate`` tool that uses a Go-like lanauge to define
C-like enumerated types. enums generates the Go code to implement the
types. `dcc` uses `enums` to create the file `runningmode.go` file
(from `runningmode.enum`).

You can install `enums` via,

    $ go get github.com/atrn/enums

Given that dcc only has a single enum using `enums` is a little
gratutious but I use it elsewhere and its _natural_ to me.

## Complete Build

    $ go get github.com/atrn/enums
    $ go generate
    $ go build
    $ go install
