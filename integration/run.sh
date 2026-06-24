#!/bin/sh
# Integration checks for yup-sort, run inside a Debian (GNU coreutils) container.
#
# parity CASE  — yup-sort must produce byte-identical output to GNU `sort`.
# assert WANT  — yup-sort must produce WANT exactly (used where yup-sort diverges
#                from GNU by design; see cmd-sort COMPATIBILITY.md).
#
# LC_ALL=C is exported so GNU `sort` uses raw byte ordering (not locale
# collation), matching yup-sort's byte-lexical comparison; without it the two
# would disagree on case/locale-sensitive orderings.
set -eu
export LC_ALL=C

fails=0

# parity LABEL INPUT ARG...: feed INPUT to both yup-sort and GNU sort with the
# same ARGs and require byte-identical stdout.
parity() {
	label=$1
	input=$2
	shift 2
	ours=$(printf '%s' "$input" | yup-sort "$@" 2>/dev/null || true)
	gnu=$(printf '%s' "$input" | sort "$@" 2>/dev/null || true)
	if [ "$ours" = "$gnu" ]; then
		printf 'ok    parity  %s\n' "$label"
	else
		printf 'FAIL  parity  %s\n        gnu:  %s\n        ours: %s\n' "$label" "$gnu" "$ours"
		fails=$((fails + 1))
	fi
}

# parityref LABEL INPUT "OUR_ARGS" "GNU_ARGS": parity where yup-sort and GNU
# need different argument spellings for the same intent (e.g. yup -k N selects a
# single field, which GNU spells -kN,N).
parityref() {
	label=$1
	input=$2
	ourargs=$3
	gnuargs=$4
	ours=$(printf '%s' "$input" | yup-sort $ourargs 2>/dev/null || true)
	gnu=$(printf '%s' "$input" | sort $gnuargs 2>/dev/null || true)
	if [ "$ours" = "$gnu" ]; then
		printf 'ok    parity  %s\n' "$label"
	else
		printf 'FAIL  parity  %s\n        gnu:  %s\n        ours: %s\n' "$label" "$gnu" "$ours"
		fails=$((fails + 1))
	fi
}

# assert LABEL WANT INPUT ARG...: yup-sort must produce WANT exactly.
assert() {
	label=$1
	want=$2
	input=$3
	shift 3
	got=$(printf '%s' "$input" | yup-sort "$@" 2>/dev/null || true)
	if [ "$got" = "$want" ]; then
		printf 'ok    assert  %s\n' "$label"
	else
		printf 'FAIL  assert  %s\n        want: %s\n        got:  %s\n' "$label" "$want" "$got"
		fails=$((fails + 1))
	fi
}

# Default: byte-lexical sort of stdin.
parity 'default lexical' 'c
a
b
'

# -r: reverse comparison.
parity 'reverse -r' 'a
b
c
' -r

# -n: numeric compare by leading value.
parity 'numeric -n' '10
2
100
9
' -n

# -n -r: numeric, reversed.
parity 'numeric reverse -nr' '10
2
100
' -n -r

# -u: drop adjacent duplicate keys (input pre-sorted so "adjacent" == "all").
parity 'unique -u' 'a
a
b
c
c
' -u

# -f: fold case before comparing.
parity 'fold case -f' 'B
a
C
' -f

# -b: ignore leading blanks.
parity 'ignore blanks -b' '  c
 b
a
' -b

# -h: human-numeric (SI suffixes).
parity 'human -h' '1M
2K
1G
512
' -h

# -M: month order.
parity 'month -M' 'Mar
Jan
Feb
Dec
' -M

# -V: version/natural order.
parity 'version -V' 'v1.10
v1.2
v1.1
v2.0
' -V

# -k / -t: single-field compare. yup-sort's -k N selects ONLY field N; GNU's
# equivalent single-field spelling is -kN,N with the same -t delimiter.
parityref 'field comma -k2 -t,' 'x,3
y,1
z,2
' '-k 2 -t ,' '-k2,2 -t,'

parityref 'field colon numeric -k2 -t:' 'a:10
b:2
c:100
' '-k 2 -t : -n' '-k2,2 -t: -n'

# -s with -k: stable sort preserves input order of equal keys.
parityref 'stable field -s -k1 -t:' 'b:2
a:3
b:1
a:1
' '-s -k 1 -t :' '-s -k1,1 -t:'

# Multiple FILE operands are concatenated then sorted (parity via /dev/stdin is
# awkward in a pipe; assert the documented concatenate-then-sort contract).
printf 'two\nfour\n' >/tmp/a.txt
printf 'one\nthree\n' >/tmp/b.txt
got=$(yup-sort /tmp/a.txt /tmp/b.txt 2>/dev/null || true)
gnu=$(sort /tmp/a.txt /tmp/b.txt 2>/dev/null || true)
if [ "$got" = "$gnu" ]; then
	printf 'ok    parity  two FILE operands\n'
else
	printf 'FAIL  parity  two FILE operands\n        gnu:  %s\n        ours: %s\n' "$gnu" "$got"
	fails=$((fails + 1))
fi

if [ "$fails" -ne 0 ]; then
	printf '\n%s check(s) failed\n' "$fails"
	exit 1
fi
printf '\nall checks passed\n'
