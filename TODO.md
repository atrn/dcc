# TODO

## Detect Overriden Headers

The issue:

If a new header file is added with the same name as a header already
included by some source file, say a new `stdio.h`, we should rebuild
as the re-compilation will be use the new header. This means we need
to be able to detect that by emulating the compiler's search for
included files so we can determine if this has occurred.

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

## Re-work Options Files

The current implementation is a bit of a mess (hacked) and could do
with a re-work.

### Conditionsls

It would be good to have some sort of conditional processing to better
accomodate different types of builds.  The use of separate files using
the Go-style platform specific names works reasonably well for the
platform granularity but doesn't work well for things like debug
vs. release or shared object vs static library.

Adding some directives that allow options files to contain
conditional sections would likely solve the issue.

### Searches

The search for options files needs to obey POLA.

## Linux-based OS Library Searching

Linux distributions like to use all manner of directories
to hold libraries and compiler-version dependent libraries
so the current hard-coded paths are (a) not sufficient and
(b) incorrect.

The real fix is to either obtain the paths from the compiler,
e.g. via its "specs" file or the like.
