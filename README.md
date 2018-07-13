# dcc - a dependency-driven C/C++ compiler driver

`dcc` is a C/C++ compiler driver _wrapper_ that that adds, parallel,
_dependency-based_ building to the underlying C or C++ compiler (which
means `gcc`, `clang`, or `icc`).

`dcc` uses compiler generated dependency information, along with
hard-coded `make`-like rules, to let it determine if compilation, or
linking, is actually required. This allows it to avoid doing work, if
possible. As with a typical `make`-based builds `dcc` only
re-compiles, or re-links, when an output file is out-of-date with
respect to its inputs, or dependencies. However, unlike make-based
builds the user doesn't have to do anything to get this behaviour.
They can use dcc as if it were cc and obtain automatic, parallel,
dependency-based builds.

The aim of moving the dependency checking into the compiler driver is
to simplify build systems. Instead of using a build tool to do the
work of correctly, and efficiently, building the program or library
using knowledge of the language dependency model, we have the
language's compiler take care of that doing that.

`dcc` adds a number of features to the compiler-driver model to
simplify the development process.  For instance, `dcc` can uses files
to store compiler options which are used as dependencies in builds
allowing automatic re-compilation when build options change.

## Building and Installation

`dcc` is written in Go and obviously requires Go installed to build
(see [golang.org](http://golang.org/)). `dcc` uses only standard Go
packages and is _trivially_ built using the `go build` command.  To
automate some other aspects of building a small `Makefile` is supplied
which provides the following targets:

- make all  
  Build `dcc`. The default target.
- make clean  
  Remove any built `dcc`.
- make install  
  Install `dcc` into `$(dest)`.


### Installation

The `make install` target copies the `dcc` executable to a
directory named by the _$(dest)_ make variable. The default
installation location is,

    $HOME/bin

This can be overriden by setting `dest` on the command line
when installing. E.g.,

    $ make install dest=/opt/bin

### `dc++`

`dcc` may be installed with the name `dc++`. This is not done by
default since it isn't strictly required - `dcc` either determines
the source language from filename extensions or can be explicitly
told via command line option.

Install `dcc` executable,

    $ cp dcc /some/where/bin/dcc

Create `dc++` link.

    $ ln /some/where/bin/dcc /some/where/bin/dc++


## Usage

`dcc` usage is similar to that of `cc(1)`, `gcc(1)` and similar
compiler drivers,

    $ dcc <option...> <pathname...>

Like `cc` et al `dcc` compiles source files to object files using the
options passed on the command line. If a `-c` is passed `dcc` stops
following compilation but if no `-c` option is supplied `dcc` runs the
linker to form an executable from the object files.

However, unlike `cc` et al `dcc` automatically generates and uses
dependency information and will only compile or link if an output
file needs to be re-created. This is entirely transparent to the
end-user. The effect being that re-compilation is far faster when
files are already up to date.

`dcc` can be used as a mostly _drop in_ replacement for `cc/c++(1)` in
existing build systems. Doing so adds additional dependency checking
to builds. There is a difference in behaviour with respect to existing
compiler drivers that may affect results, `dcc` does *not* remove
object files when no `-c` switch is used. Most build systems however
invoke the compiler for each source file passing `-c`.

### Differences to cc

Although `dcc` is similar in usage to `cc(1)1 et al, enough so to
permit it to be used directly in its place, `dcc` does behave
differently in certain situations.

#### object files without `-c`

Normally, without a `-c` option, `cc` compiles the source files,
generating object files, and runs the linker to link those object
files into an executable. It then *removes* the object files. `dcc`
does not remove the files.


## What dcc does

`dcc` _wraps_ the underlying compiler driver and passes it options to
have it output dependency information. `dcc` automatically determines
the names of the files to store this information and reads them when
re-compiling a file to obtain the dependencies.

When re-compiling a file `dcc` performs `make`-like dependency
checking to determine if compilation is actually required.  If not,
`dcc` does nothing and exits as if it had compiled the file (note,
file modification times are *not* altered). Otherwise `dcc` runs the
compiler and lets it generate its output.  Dependency generation and
checking is entirely transparent to the end-user and, `dcc` implements
additional checks on the libraries and other files used in the build.

## Command line options

The `dcc` command line consists of options for the underling compiler,
a number of `dcc`-specific options and filenames to be processed.

Options to the compiler are passed through unalterted. `dcc` does
recognize a number of options which control its behaviour or supply
dependency information (libraries).

### dcc-specific options

These options apply to `dcc` itself and are not passed on to the
compiler,

- --help 
  Get help.
- --debug 
  Enable `dcc` debug output.
- --clean 
  Remove `dcc`-generated files.
- --cpp 
  Compile source as C++ rather than C.
- --force 
  Rebuild everything, ignore dependencies.
- --quiet 
  Don't output the commands being executed.
- --exe <path> 
  Compile and link an executable called <path>.
- --dll <path> 
  Compile and create a shared library called <path>.
- --lib <path> 
  Compile and create a static library, <path>.
- -j<number> 
  Use <number> parallel compilations.

### --exe, --dll, --lib

`cc`-style compiler drivers traditionally worked in two modes. They
either compiled source files to object files or did that and linked
the object files to form an executable (and removed the object files).
Shared libraries added options to have the linker create a shared
library but the overall structure is the same as for an executable.

`dcc` has options that make these uses more explicit and adds the
feature of having the compiler driver generate a static library to
round out the various use cases.

The `dcc`-specific `--exe`, `--dll` and `--lib` options are used to
tell `dcc` what is being built and the name of the output file.

The `--exe` option means "build an executable", `--dll` means "build a
shared library" (sic) and `--lib` means "build a static library".

### Language selection

`dcc` determines the language being compiled, C or C++, using a number
of rules and uses the appropriate underlying C or C++ compiler. C++ is
selected if,

- the `dcc` program name ends with `++`, e.g `dc++`
- an input file uses a C++ extension `.cc`, `.cpp`, `.cxx`
- the `--cpp` switch was supplied

The choice of lanugage affects the choice of _options files_ (see
below).

## Dependency Files

`dcc` uses dependency information generated by the compiler itself
and information inferred from the filenames and system environment.

With gcc-style compilers `dcc` uses the `-MF` and `-MD` options to
have the compiler output make-format dependencies to a file which
`dcc` reads on the next run.

Dependency files are stored in a `.dcc.d` directory that resides in
the same directory as the object file being created. The `DCCDEPS`
environment variable can be set to use a name other than `.dcc.d` for
this directory.

## Options Files

`dcc` can read compiler and linker options stored in files called
_options files_. Options files are simple text files that contain the
options that would normally be passed on the command line.

Unlike passing options on the the comand line options files allow
options to be split across multiple lines and support '#'-based _line_
comments. Options files are also treated as dependencies and when
changed, which presumably means the options within the file have been
change, cause recompilation.  This helps ensure all files are built in
the same way.

The names adopted for options files are derived from the typical macro
names used with make(1) for the particular options,

- `CFLAGS` 
  C compiler options.
- `CXXFLAGS` 
  C++ compiler options.
- `LDFLAGS` 
  Linker options.
- `LIBS` 
  Libraries and library paths.

### Locating options files

Option files are looked for by searching the directory hierarchy
towards the root for a file with the particular name, e.g CXXFLAGS.

Files are searched for either in the specific directory or within a
`$DCCDIR` directory within that directory. `$DCCDIR` defaults to `.dcc`
but can be override by the environment variable so we call it `$DCCDIR`
even though it is rarely changed from the default `.dcc`.

Looking for the files in a `$DCCDIR` directory is a quick hack to get
the files out of the current directory and perhaps in the future some
other method may be adopted (ha ha).

### Platform-specific option files

`dcc` uses a Go-style method to support platform-specific options.
When searching for an options file `dcc` first searches for platform
and architecture specific variants of the file. `dcc` forms a file
name extension using names for for the host's architecture and
operating system and appends that extension to the filename. If a file
with that name exists it is used in place of the unadorned filename.

E.g. when searching for the `LIBS` file on a 64-bit FreeBSD host
the following files will be searched for in order,

1. `$DCCDIR/LIBS.freebsd_amd64`
2. `$DCCDIR/LIBS.freebsd`
3. `$DCCDIR/LIBS`

On a 32-bit Windows the files are,

1. `$DCCDIR/LIBS.windows_x86`
2. `$DCCDIR/LIBS.windows`
3. `$DCCDIR/LIBS`

### Libraries

The `LIBS` options file is used to define the libraries and library
directories used when linking programs and DLLs.

The `LIBS` options file behaves in a similar manner to the compiler
options and executables depend on the file and relink when it changes.

Lines starting with `-l` (elle) and `-L` (capital-elle) are special.
Any library name starting with `-l` has the `-l` removed allowing
users to use UNIX linker-style naming for familarity. _libraries_ with
names starting with `-L` are the names of of library directories.


### Option File Inheritence

Options files may contain a special line `#inherits` to indicate they
want to _inherit_ values set in by higher level, in the directory
hierarchy sense, file. It is similar to a C `#include`.

A `#inherits` directive with no arguments means _search for a file
with the same name as mine in a higher level directory and reads its
contents_. Optionally a `#inherits` may specify the filename to be
searched for. In either case it is an error to *not* inherit the
file.

## Implementation

`dcc` is written in Go and uses only standard packages in its
implementation. `dcc` should build in any supported Go environment and
be trivially cross-buildable.

`dcc` itself supports the various Linux distribtions, the BSD's, MacOS
and mostly likely other UNIX systems that use gcc, clang or
similar.

`dcc` has not really been used _in anger_ and I expect many changes if
it is used more extensively. There are many areas where I've just
hacked things in, e.g. frameworks on MacOS, which would be better
expressed in a more structured manner, i.e. more comprehensive
abstracted interfaces to the compiler and other tools to remove the
platform-specific conditiona.

The code has lots of comments. Many of them correct! The commenting
style is the result of using Visual Studio Code and its Go package's
default configuration which _golints_ your code producing lots of
annoying warnings about naming, comment style and so on. Rather than
disabling the tool like a sensible person I appeased it and wrote the
things it told me to write. That stopped it drawing little squiggles
and annoying little icons everywhere.

## License

`dcc` is  released under  the GPL,  version 2. If  you advance  dcc, and
distribute, you must share the  advancements. The reasoning being that
a  utility such  as dcc  is infrastructure  and we  should share,  and
advance, infrastructure so we all get ahead.

As per convention the license text is in the file LICENSE.


## Example

Using `dcc` in a project can vastly simplfy its build system. Instead
of implementing build rules via `make` or generating them via `cmake`
or autotools you can just use `dcc`. It takes care of the building
part.

A complete development `Makefile` for a simple program, with all
source files in one directory, can be as small as:

    .PHONY: program clean
    
    program:
        dcc $(CFLAGS) *.c -o $@
        
    clean:
        rm -f program *.o
        rm -rf .dcc.d

The `program` target builds everything using `dcc`. It is marked
marked _phony_ as we rely on `dcc` to take care of things.

Until the `dcc`'s `--clean` option is implemented the `clean` target
has to `rm -rf` a directory created by `dcc`.

## Environment Variables

- CC (or $CCFILE)  
Name of the C compiler.
- CXX (or $CXXFILE)  
Name of the C++ compiler.
- CCFILE  
Name of the file that names the C compiler.
- CXXFILE  
Name of the file that names the C++ compiler.
- CFLAGSFILE  
Name of the C options file.
- CXXFLAGSFILE  
Name of the C++ options files.
- LDFLAGSFILE  
Name of the linker options file.
- LIBSFILE  
Name of the linker _LIBS_ file.
- DCCDIR  
Name of the `.dcc` directory.
- DEPSDIR  
Name of the `.dcc.d` dependency file directory.
- OBJDIR  
Name of the object file directory.
- JOBS  
Number of compilations to run in parallel.
