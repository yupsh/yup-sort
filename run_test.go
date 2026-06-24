package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestRun(t *testing.T) {
	cases := []struct {
		name       string
		version    string
		args       []string
		stdin      string
		files      map[string]string
		wantOut    string
		wantCode   int
		wantErrSub string
	}{
		{
			name:    "default sort",
			args:    []string{"sort"},
			stdin:   "c\na\nb\n",
			wantOut: "a\nb\nc\n",
		},
		{
			name:    "reverse",
			args:    []string{"sort", "-r"},
			stdin:   "a\nb\nc\n",
			wantOut: "c\nb\na\n",
		},
		{
			name:    "numeric",
			args:    []string{"sort", "-n"},
			stdin:   "10\n2\n100\n",
			wantOut: "2\n10\n100\n",
		},
		{
			name:    "human numeric",
			args:    []string{"sort", "-h"},
			stdin:   "1M\n2K\n1G\n",
			wantOut: "2K\n1M\n1G\n",
		},
		{
			name:    "month",
			args:    []string{"sort", "-M"},
			stdin:   "Mar\nJan\nFeb\n",
			wantOut: "Jan\nFeb\nMar\n",
		},
		{
			name:    "version",
			args:    []string{"sort", "-V"},
			stdin:   "v1.10\nv1.2\nv1.1\n",
			wantOut: "v1.1\nv1.2\nv1.10\n",
		},
		{
			name:    "unique",
			args:    []string{"sort", "-u"},
			stdin:   "a\nb\na\nc\n",
			wantOut: "a\nb\nc\n",
		},
		{
			name:    "fold case",
			args:    []string{"sort", "-f"},
			stdin:   "B\na\nC\n",
			wantOut: "a\nB\nC\n",
		},
		{
			name:    "ignore leading blanks",
			args:    []string{"sort", "-b"},
			stdin:   "  c\n b\na\n",
			wantOut: "a\n b\n  c\n",
		},
		{
			name:    "random groups all lines",
			args:    []string{"sort", "-R"},
			stdin:   "only\n",
			wantOut: "only\n",
		},
		{
			name:    "stable by field",
			args:    []string{"sort", "-s", "-k", "1", "-t", ":"},
			stdin:   "b:2\na:3\nb:1\na:1\n",
			wantOut: "a:3\na:1\nb:2\nb:1\n",
		},
		{
			name:    "key and delimiter",
			args:    []string{"sort", "-k", "2", "-t", ","},
			stdin:   "x,3\ny,1\nz,2\n",
			wantOut: "y,1\nz,2\nx,3\n",
		},
		{
			name:    "file source",
			args:    []string{"sort", "/in.txt"},
			files:   map[string]string{"/in.txt": "two\none\n"},
			wantOut: "one\ntwo\n",
		},
		{
			name:    "version flag reports injected version",
			version: "1.2.3",
			args:    []string{"sort", "--version"},
			wantOut: "sort version 1.2.3\n",
		},
		{
			name:       "unknown flag errors",
			args:       []string{"sort", "--nope"},
			wantCode:   1,
			wantErrSub: "sort:",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			for path, content := range tc.files {
				if err := afero.WriteFile(fs, path, []byte(content), 0o644); err != nil {
					t.Fatalf("write fixture %s: %v", path, err)
				}
			}

			var out, errOut bytes.Buffer
			code := run(tc.version, tc.args, strings.NewReader(tc.stdin), &out, &errOut, fs)

			if code != tc.wantCode {
				t.Fatalf("exit code = %d, want %d (stderr=%q)", code, tc.wantCode, errOut.String())
			}
			if tc.wantErrSub == "" && out.String() != tc.wantOut {
				t.Fatalf("stdout = %q, want %q", out.String(), tc.wantOut)
			}
			if tc.wantErrSub != "" && !strings.Contains(errOut.String(), tc.wantErrSub) {
				t.Fatalf("stderr = %q, want substring %q", errOut.String(), tc.wantErrSub)
			}
		})
	}
}

func Test_main(t *testing.T) {
	origExit, origRun := osExit, runCLI
	t.Cleanup(func() { osExit, runCLI = origExit, origRun })

	gotCode := -1
	osExit = func(code int) { gotCode = code }
	runCLI = func(string, []string, io.Reader, io.Writer, io.Writer, afero.Fs) int { return 7 }

	main()

	if gotCode != 7 {
		t.Fatalf("main propagated exit code %d, want 7", gotCode)
	}
}
