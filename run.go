package main

import (
	"context"
	"fmt"
	"io"

	command "github.com/gloo-foo/cmd-sort"
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

const name = "sort"

const (
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

// usageText is the command's multi-line usage synopsis, shown in --help.
// cli/v3 indents the whole block by 3 spaces, so these lines are flush-left to
// stay aligned in the rendered output.
const usageText = `sort [OPTIONS] [FILE...]

Write sorted concatenation of all FILE(s) to standard output.
With no FILE, or when FILE is -, read standard input.`

// buildVersion is the binary's build version threaded from main's ldflags
// target (`var version`) into the CLI. It is an alias, not a defined type:
// cli.Command.Version is a plain string and must be wired as the bare
// `version` identifier (no conversion) for --version to stay verifiably
// bound to the ldflags symbol.
type buildVersion = string

// run builds and executes the sort CLI against the injected version, I/O, and
// filesystem, returning the process exit code.
func run(version buildVersion, args []string, stdin io.Reader, stdout, stderr io.Writer, fs afero.Fs) int {
	cmd := newCommand(version, stdin, stdout, fs)
	cmd.Writer = stdout
	cmd.ErrWriter = stderr
	if err := cmd.Run(context.Background(), args); err != nil {
		_, _ = fmt.Fprintf(stderr, name+": %v\n", err)
		return 1
	}
	return 0
}

func newCommand(version buildVersion, stdin io.Reader, stdout io.Writer, fs afero.Fs) *cli.Command {
	// Replace urfave/cli's default --version/-v flag with a --version-only
	// flag, freeing the single-letter -v for command flags while still
	// exposing the injected build version, and drop the default -h help alias
	// so -h is available for --human-numeric-sort, matching GNU sort. Done
	// here rather than in func init so construction stays explicit.
	cli.VersionFlag = &cli.BoolFlag{Name: "version", Usage: "print version information and exit"}
	cli.HelpFlag = &cli.BoolFlag{Name: "help", Usage: "show help"}
	return &cli.Command{
		Name:            name,
		Version:         version,
		Usage:           "sort lines of text files",
		UsageText:       usageText,
		HideHelpCommand: true,
		// Keep exit handling in run() rather than letting urfave/cli call
		// os.Exit, so the exit code stays testable.
		ExitErrHandler: func(context.Context, *cli.Command, error) {},
		Flags:          flags(),
		Action:         action(stdin, stdout, fs),
	}
}

func flags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{Name: flagReverse, Aliases: []string{"r"}, Usage: "reverse the result of comparisons"},
		&cli.BoolFlag{Name: flagNumeric, Aliases: []string{"n"}, Usage: "compare according to string numerical value"},
		&cli.BoolFlag{
			Name:    flagHumanNumeric,
			Aliases: []string{"h"},
			Usage:   "compare human readable numbers (e.g., 2K 1G)",
		},
		&cli.BoolFlag{Name: flagMonth, Aliases: []string{"M"}, Usage: "compare (unknown) < 'JAN' < ... < 'DEC'"},
		&cli.BoolFlag{
			Name:    flagVersion,
			Aliases: []string{"V"},
			Usage:   "natural sort of (version) numbers within text",
		},
		&cli.BoolFlag{Name: flagUnique, Aliases: []string{"u"}, Usage: "output only the first of an equal run"},
		&cli.BoolFlag{Name: flagIgnoreCase, Aliases: []string{"f"}, Usage: "fold lower case to upper case characters"},
		&cli.BoolFlag{Name: flagIgnoreLeadingBlanks, Aliases: []string{"b"}, Usage: "ignore leading blanks"},
		&cli.BoolFlag{Name: flagRandom, Aliases: []string{"R"}, Usage: "shuffle, but group identical keys"},
		&cli.BoolFlag{
			Name:    flagStableSort,
			Aliases: []string{"s"},
			Usage:   "stabilize sort by disabling last-resort comparison",
		},
		&cli.IntFlag{Name: flagField, Aliases: []string{"k"}, Usage: "sort via a key; KEYDEF gives location and type"},
		&cli.StringFlag{
			Name:    flagDelimiter,
			Aliases: []string{"t"},
			Usage:   "use SEP instead of non-blank to blank transition",
		},
	}
}

func action(stdin io.Reader, stdout io.Writer, fs afero.Fs) cli.ActionFunc {
	return func(_ context.Context, c *cli.Command) error {
		_, err := gloo.Run(source(c, stdin, fs), gloo.ByteWriteTo(stdout), command.Sort(options(c)...))
		return err
	}
}

func source(c *cli.Command, stdin io.Reader, fs afero.Fs) any {
	if c.NArg() == 0 {
		return gloo.ByteReaderSource([]io.Reader{stdin})
	}
	files := make([]gloo.File, c.NArg())
	for i := range files {
		files[i] = gloo.File(c.Args().Get(i))
	}
	return gloo.ByteFileSource(fs, files)
}

// flagOption pairs a boolean CLI flag name with the library option it enables.
type flagOption struct {
	option any
	name   string
}

func flagOptions() []flagOption {
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

func options(c *cli.Command) []any {
	var opts []any
	for _, fo := range flagOptions() {
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
