# TODO

## `cl.exe`

Microsoft cl support is planned for but not yet done. There are some
initial bits of code but it hasn't been used.

`cl.exe` is the reason for most of the internal abstractions around
compilers. Obtaining dependency information from `cl.exe` requires a
different approach to that used with the gcc-style compilers. With
gcc, clang and icc (Intel) you just pass some extra options and they
do what we need. With `cl.exe` we have to _scrape_ the output produced
by the `/showIncludes` option to obtain the names of dependent files.
I haven't implemented that as yet.

## `#include` in options files

The inheritence mechanism for options files needs a good and proper
specification and implementation.

## improved `.dcc` directory support

Options file searching doesn't work as expected when a `.dcc`
directory is used.

## put object files in a directory

`dcc` can do something to keep file systems _clean_. As with `.d`
files it should put `.o` files into a directory automatically. That
makes it less _drop-in_ replacement but nicer to use.
