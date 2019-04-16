# Building dcc

The repo contains all the .go files required to build dcc. However
the _full_ build of dcc uses my [`enums`](http://github.com/atrn/enums)
tool to generate the `runningmode.go` file from `runningmode.enum`.

`enums` is a `go generate` tool that reads `.enum` files which contain
definitions of C-style enumerated types (i.e. not tagged variants) in
a Go-like syntax. For each `.enum` file in the current directory a
corresponding `.go` file is generated.
