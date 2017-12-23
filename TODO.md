# TODO

## `cl.exe`

Microsoft cl support is planned for but not yet done. There are some
initial bits of code but it hasn't been used.

`cl.exe` is the reason for much internal abstraction and
complexity. Obtaining dependency information from `cl.exe` requires a
vastly different approach to that used with the gcc-style compilers,
which just require some extra options to do what we need. `cl.exe`
needs an option, `/showIncludes`, and have the resultant output
_scraped_ to obtain the names of the included, dependent, files.

## `#include` in options files

The inheritence mechanism for options files needs proper specification
and implementation.

## improved `.dcc` directory support

For instance options file searching doesn't work as expected. It
needs to use a $DCCDIR and to not use $DCCDIR during the search.

## put object files in a directory

`dcc` can do something to keep file systems _clean_. As with `.d`
files it should put `.o` files into a directory automatically. That
makes it less _drop-in_ but far nicer to use.
