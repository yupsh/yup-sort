# yup-sort

```
NAME:
   sort - sort lines of text files

USAGE:
   sort [OPTIONS] [FILE...]

   Write sorted concatenation of all FILE(s) to standard output.
   With no FILE, or when FILE is -, read standard input.

VERSION:
   dev

GLOBAL OPTIONS:
   --reverse, -r                        reverse the result of comparisons
   --numeric-sort, -n                   compare according to string numerical value
   --human-numeric-sort, -h             compare human readable numbers (e.g., 2K 1G)
   --month-sort, -M                     compare (unknown) < 'JAN' < ... < 'DEC'
   --version-sort, -V                   natural sort of (version) numbers within text
   --unique, -u                         output only the first of an equal run
   --ignore-case, -f                    fold lower case to upper case characters
   --ignore-leading-blanks, -b          ignore leading blanks
   --random-sort, -R                    shuffle, but group identical keys
   --stable, -s                         stabilize sort by disabling last-resort comparison
   --key int, -k int                    sort via a key; KEYDEF gives location and type (default: 0)
   --field-separator string, -t string  use SEP instead of non-blank to blank transition
   --help                               show help
   --version                            print version information and exit
```
