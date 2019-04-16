# TODO

## detect override headers and force re-build

If a new header file, say `stdio.h`, is added to a directory on
the include file search path so it is found *before* the previously
used version of the header a re-build needs to occur. `dcc` won't
detect that situation.

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

## improved `.dcc` directory support

Options file searching doesn't work as expected when a `.dcc`
directory is used.
