# TODO

## detect overriden headers

If a new header file, say `stdio.h`, is added to a directory on the
include file search path such that the code being built _would_ now
include that file, we should re-build.

This particular situation is not common however it is possible.  Some
build enviromnents can be _messy_ and update or places things in odd
ways - MacOS, Windows and, unfortunately, some Linux-based systems
can update/install things underneath the user which can make this
happen.

Better safe than sorry.

Without compiler support (telling us which files were *not* included)
we have to figure that our ourselves.

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
