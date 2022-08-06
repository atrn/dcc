# Version 0.0.5

- now supports Microsoft toolchain on Windows

- update options modtime when compiler and/or linker options, and
  libraries, are set via the command line so rebuilds occur when these
  "change".  This is a hack-ish way to NOT support keeping track
  of the actual options used when building, i.e. the combination
  of options read from files and the command line.

# Version 0.0.4

- use '!' as the options file directive prefix in place of '#'

- allow `!inherit` directives to define the, base, filename of
  the file to be inherited

