package main

import (
	"context"
	"os"
	"testing"

	clix "github.com/gloo-foo/cli"
	"github.com/spf13/afero"
	urf "github.com/urfave/cli/v3"
)

// TestMain applies the wrapper's explicit help-flag configuration before the
// tests run, exactly as main does, so parse-driven tests see -h freed for
// --human-numeric-sort.
func TestMain(m *testing.M) {
	configureHelp()
	os.Exit(m.Run())
}

// parse runs args through a bare command carrying the wrapper's flags and
// returns the parsed accessor, so flag-dependent helpers are tested against real
// parsed flags.
func parse(t *testing.T, args ...string) *urf.Command {
	t.Helper()
	var got *urf.Command
	app := &urf.Command{
		Name:   name,
		Flags:  flags(),
		Action: func(_ context.Context, c *urf.Command) error { got = c; return nil },
	}
	if err := app.Run(context.Background(), args); err != nil {
		t.Fatalf("parse: %v", err)
	}
	return got
}

func TestOptions(t *testing.T) {
	all := []string{
		name, "-r", "-n", "-h", "-M", "-V", "-u", "-f", "-b", "-R", "-s", "-k", "3", "-t", ",",
	}
	cases := []struct {
		name string
		args []string
		want int
	}{
		{"none", []string{name}, 0},
		{"reverse", []string{name, "-r"}, 1},
		{"field", []string{name, "-k", "2"}, 1},
		{"delimiter", []string{name, "-t", ":"}, 1},
		{"all", all, 12},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := len(options(parse(t, tc.args...))); got != tc.want {
				t.Fatalf("options len=%d, want %d", got, tc.want)
			}
		})
	}
}

func TestBuild(t *testing.T) {
	src, filter, err := build(clix.Invocation{Args: parse(t, name), Fs: afero.NewMemMapFs()})
	if err != nil || src == nil || filter == nil {
		t.Fatalf("build: src=%v filter=%v err=%v", src, filter, err)
	}
}

// TestConfigureHelp asserts the contract: after configureHelp, urfave/cli's
// global help flag is --help with no -h alias, so -h stays free for
// --human-numeric-sort (matching GNU sort).
func TestConfigureHelp(t *testing.T) {
	orig := urf.HelpFlag
	t.Cleanup(func() { urf.HelpFlag = orig })
	urf.HelpFlag = &urf.BoolFlag{Name: flagHelp, Aliases: []string{"h"}}
	configureHelp()
	assertHelpFlagFreesH(t)
}

// assertHelpFlagFreesH fails unless the global help flag is --help with no
// aliases.
func assertHelpFlagFreesH(t *testing.T) {
	t.Helper()
	hf, ok := urf.HelpFlag.(*urf.BoolFlag)
	if !ok {
		t.Fatalf("HelpFlag=%T, want *urf.BoolFlag", urf.HelpFlag)
	}
	if hf.Name != flagHelp || len(hf.Aliases) != 0 {
		t.Fatalf("HelpFlag name=%q aliases=%v, want name %q with no aliases", hf.Name, hf.Aliases, flagHelp)
	}
}

func Test_main(t *testing.T) {
	origRun := runMain
	origHelp := urf.HelpFlag
	t.Cleanup(func() { runMain = origRun; urf.HelpFlag = origHelp })
	urf.HelpFlag = &urf.BoolFlag{Name: flagHelp, Aliases: []string{"h"}}
	var gotName clix.Name
	runMain = func(s clix.Spec, _ clix.Version) { gotName = s.Name }
	main()
	if gotName != name {
		t.Fatalf("main used spec %q, want %s", gotName, name)
	}
	assertHelpFlagFreesH(t)
}
