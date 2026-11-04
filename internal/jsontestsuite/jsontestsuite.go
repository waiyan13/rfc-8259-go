// Package jsontestsuite runs the test_parsing corpus of nst/JSONTestSuite
// against a parser. The corpus is in testdata/; see testdata/README.md.
package jsontestsuite

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"testing"
)

//go:embed testdata/test_parsing
var corpus embed.FS

const corpusDir = "testdata/test_parsing"

// Kind is the expected result for a case, from the first letter of its file name.
type Kind byte

const (
	MustAccept  Kind = 'y'
	MustReject  Kind = 'n'
	ImplDefined Kind = 'i'
)

func (k Kind) String() string { return string(k) + "_" }

// Case is one file of the corpus.
type Case struct {
	Name  string // file name without ".json"
	Kind  Kind
	Input []byte
}

// Parse returns nil when the parser accepts input as one JSON text.
type Parse func(input []byte) error

// Cases returns all cases in file-name order.
func Cases() ([]Case, error) {
	entries, err := fs.ReadDir(corpus, corpusDir)
	if err != nil {
		return nil, err
	}
	cases := make([]Case, 0, len(entries))
	for _, e := range entries {
		name, ok := strings.CutSuffix(e.Name(), ".json")
		if !ok || len(name) < 2 || name[1] != '_' {
			return nil, fmt.Errorf("unexpected corpus file %q", e.Name())
		}
		kind := Kind(name[0])
		if kind != MustAccept && kind != MustReject && kind != ImplDefined {
			return nil, fmt.Errorf("unknown kind in corpus file %q", e.Name())
		}
		input, err := fs.ReadFile(corpus, path.Join(corpusDir, e.Name()))
		if err != nil {
			return nil, err
		}
		cases = append(cases, Case{Name: name, Kind: kind, Input: input})
	}
	return cases, nil
}

// Check runs parse on c. The error is non-nil when the parser panics, or when its
// result breaks the rule for c.Kind. An ImplDefined case fails only on a panic.
func Check(c Case, parse Parse) (accepted bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			accepted, err = false, fmt.Errorf("parser panicked: %v\ninput: %s", r, preview(c.Input))
		}
	}()
	perr := parse(c.Input)
	accepted = perr == nil
	switch {
	case c.Kind == MustAccept && !accepted:
		return accepted, fmt.Errorf("rejected valid input: %v\ninput: %s", perr, preview(c.Input))
	case c.Kind == MustReject && accepted:
		return accepted, fmt.Errorf("accepted invalid input: %s", preview(c.Input))
	}
	return accepted, nil
}

// Report holds the parser result for each ImplDefined case, by case name.
type Report struct {
	Implementation map[string]bool
}

// Run checks every case in its own subtest and logs a summary.
func Run(t *testing.T, parse Parse) Report {
	t.Helper()
	cases, err := Cases()
	if err != nil {
		t.Fatal(err)
	}
	rep := Report{Implementation: map[string]bool{}}
	total := map[Kind]int{}
	failed := map[Kind]int{}
	for _, c := range cases {
		total[c.Kind]++
		t.Run(c.Name, func(t *testing.T) {
			accepted, err := Check(c, parse)
			if err != nil {
				failed[c.Kind]++
				t.Error(err)
				return
			}
			if c.Kind == ImplDefined {
				rep.Implementation[c.Name] = accepted
				t.Logf("accepted=%t", accepted)
			}
		})
	}
	accepted := 0
	for _, a := range rep.Implementation {
		if a {
			accepted++
		}
	}
	t.Logf("y_: %d/%d pass, n_: %d/%d pass, i_: %d accepted, %d rejected, %d panicked",
		total[MustAccept]-failed[MustAccept], total[MustAccept],
		total[MustReject]-failed[MustReject], total[MustReject],
		accepted, len(rep.Implementation)-accepted, failed[ImplDefined])
	return rep
}

func preview(b []byte) string {
	const limit = 80
	if len(b) <= limit {
		return fmt.Sprintf("%q", b)
	}
	return fmt.Sprintf("%q... (%d bytes)", b[:limit], len(b))
}
