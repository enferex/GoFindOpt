### GoFindOpt: Scan ELF objects looking for getopt strings.

#### What
GoFindOpt is a tool for inferring potential command line options from a given
ELF object.  The idea that this tool relies on is the following.  Many programs
use POSIX 'getopt' command for dealing with command line parsing.  One can
imagine that some options are not documented publicly, perhaps they are in
test or unlock magic powers.  This tool aims to find the getopt string in a ELF
binary.

#### How
GoFindOpt takes, as input, filenames to ELF object files. These objects can be
programs, shared libraries, or any other ELF item.  Next, the symbol table for
each ELF is scanned looking for use of the 'getopt' symbol.  If that symbol is
found, then the .rodata section is scanned for any strings that match a subset
of the getopt 'optstring' definition.  For an idea of what those strings might
look like, check out 'man 3 getopt' and look for the 'optstring' definition.

#### Output
GoFindOut presents its findings as CSV where each line is a potential getopt
optstring.
   # elf object name , potential optstring , Vaild or Not Valid

Validity is with respect to the ELF.  If the item is not really an ELF file then
it is not valid.  This is reported to the user for informational purposes.

#### Caveat
This tool can produce a large number of false positives:  Use and or tweak at
your own discretion.  Hopefully, I've planted a seed, now water this plant.

#### Contact
https://github.com/enferex
