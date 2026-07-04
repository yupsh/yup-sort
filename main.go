// Command yup-sort is the CLI wrapper around github.com/gloo-foo/cmd-sort.
package main

import (
	clix "github.com/gloo-foo/cli"
	command "github.com/gloo-foo/cmd-sort"
	urf "github.com/urfave/cli/v3"
)

// version is the build version. It defaults to "dev" for local builds and is
// overridden at release time via the linker: -ldflags "-X main.version=<v>".
var version = "dev"

const (
	name                    = "sort"
	flagHelp                = "help"
	flagReverse             = "reverse"
	flagNumeric             = "numeric-sort"
	flagHumanNumeric        = "human-numeric-sort"
	flagMonth               = "month-sort"
	flagVersion             = "version-sort"
	flagUnique              = "unique"
	flagIgnoreCase          = "ignore-case"
	flagIgnoreLeadingBlanks = "ignore-leading-blanks"
	flagRandom              = "random-sort"
	flagStableSort          = "stable"
	flagField               = "key"
	flagDelimiter           = "field-separator"
)

// synopsis is the multi-line --help usage block; urfave/cli indents it three
// spaces, so the lines stay flush-left.
const synopsis = `sort [OPTIONS] [FILE...]

Write sorted concatenation of all FILE(s) to standard output.
With no FILE, or when FILE is -, read standard input.`

// configureHelp drops urfave/cli's default -h help alias so -h is free for
// --human-numeric-sort, matching GNU sort; --help still shows help. It is
// called explicitly from main (mirroring how the clix runner replaces
// urf.VersionFlag at app construction) rather than hidden in an init.
func configureHelp() {
	urf.HelpFlag = &urf.BoolFlag{Name: flagHelp, Usage: "show help"}
}

// spec declares the sort wrapper: a file-or-stdin filter with GNU sort's flags.
var spec = clix.Spec{
	Name:     name,
	Summary:  "sort lines of text files",
	Synopsis: synopsis,
	Build:    build,
	Flags:    flags(),
}

// flags returns fresh flag instances. It is a constructor rather than a shared
// slice because urfave/cli flag structs retain per-parse state, so each parse
// (including tests) must build its own.
func flags() []urf.Flag {
	return []urf.Flag{
		&urf.BoolFlag{Name: flagReverse, Aliases: []string{"r"}, Usage: "reverse the result of comparisons"},
		&urf.BoolFlag{Name: flagNumeric, Aliases: []string{"n"}, Usage: "compare according to string numerical value"},
		&urf.BoolFlag{
			Name:    flagHumanNumeric,
			Aliases: []string{"h"},
			Usage:   "compare human readable numbers (e.g., 2K 1G)",
		},
		&urf.BoolFlag{Name: flagMonth, Aliases: []string{"M"}, Usage: "compare (unknown) < 'JAN' < ... < 'DEC'"},
		&urf.BoolFlag{
			Name:    flagVersion,
			Aliases: []string{"V"},
			Usage:   "natural sort of (version) numbers within text",
		},
		&urf.BoolFlag{Name: flagUnique, Aliases: []string{"u"}, Usage: "output only the first of an equal run"},
		&urf.BoolFlag{Name: flagIgnoreCase, Aliases: []string{"f"}, Usage: "fold lower case to upper case characters"},
		&urf.BoolFlag{Name: flagIgnoreLeadingBlanks, Aliases: []string{"b"}, Usage: "ignore leading blanks"},
		&urf.BoolFlag{Name: flagRandom, Aliases: []string{"R"}, Usage: "shuffle, but group identical keys"},
		&urf.BoolFlag{
			Name:    flagStableSort,
			Aliases: []string{"s"},
			Usage:   "stabilize sort by disabling last-resort comparison",
		},
		&urf.IntFlag{Name: flagField, Aliases: []string{"k"}, Usage: "sort via a key; KEYDEF gives location and type"},
		&urf.StringFlag{
			Name:    flagDelimiter,
			Aliases: []string{"t"},
			Usage:   "use SEP instead of non-blank to blank transition",
		},
	}
}

// build maps the invocation to sort's pipeline: a file-or-stdin source into the
// sort command configured by the flags.
func build(inv clix.Invocation) (clix.Source, clix.Command, error) {
	return clix.OperandsOrStdin(inv), command.Sort(options(inv.Args)...), nil
}

// flagOption pairs a boolean flag name with the library option it enables.
type flagOption struct {
	option any
	name   string
}

// boolOptions lists the boolean flag-to-option mappings folded by options.
func boolOptions() []flagOption {
	return []flagOption{
		{name: flagReverse, option: command.SortReverse},
		{name: flagNumeric, option: command.SortNumeric},
		{name: flagHumanNumeric, option: command.SortHumanNumeric},
		{name: flagMonth, option: command.SortMonthSort},
		{name: flagVersion, option: command.SortVersionSort},
		{name: flagUnique, option: command.SortUnique},
		{name: flagIgnoreCase, option: command.SortIgnoreCase},
		{name: flagIgnoreLeadingBlanks, option: command.SortIgnoreLeadingBlanks},
		{name: flagRandom, option: command.SortRandom},
		{name: flagStableSort, option: command.SortStableSort},
	}
}

// options folds the parsed flags into sort's option values.
func options(c *urf.Command) []any {
	var opts []any
	for _, fo := range boolOptions() {
		if c.Bool(fo.name) {
			opts = append(opts, fo.option)
		}
	}
	if c.IsSet(flagField) {
		opts = append(opts, command.SortField(c.Int(flagField)))
	}
	if c.IsSet(flagDelimiter) {
		opts = append(opts, command.SortDelimiter(c.String(flagDelimiter)))
	}
	return opts
}

// runMain is an indirection seam so main's wiring is testable without spawning
// the process; a test swaps it and restores it.
var runMain = clix.Main

func main() {
	configureHelp()
	runMain(spec, version)
}
