package jsontestsuite_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/waiyan13/rfc-8259-go/internal/jsontestsuite"
)

func stdlib(input []byte) error {
	if json.Valid(input) {
		return nil
	}
	return errors.New("encoding/json: invalid")
}

func TestCorpusCounts(t *testing.T) {
	cases, err := jsontestsuite.Cases()
	if err != nil {
		t.Fatal(err)
	}
	got := map[jsontestsuite.Kind]int{}
	for _, c := range cases {
		got[c.Kind]++
	}
	want := map[jsontestsuite.Kind]int{
		jsontestsuite.MustAccept:  95,
		jsontestsuite.MustReject:  188,
		jsontestsuite.ImplDefined: 35,
	}
	for k, n := range want {
		if got[k] != n {
			t.Errorf("%s cases: got %d, want %d", k, got[k], n)
		}
	}
}

func TestCheckCatchesWrongVerdicts(t *testing.T) {
	acceptAll := func([]byte) error { return nil }
	rejectAll := func([]byte) error { return errors.New("no") }
	panics := func([]byte) error { panic("boom") }

	for _, tc := range []struct {
		name    string
		kind    jsontestsuite.Kind
		parse   jsontestsuite.Parse
		wantErr bool
	}{
		{"y_ rejected", jsontestsuite.MustAccept, rejectAll, true},
		{"y_ accepted", jsontestsuite.MustAccept, acceptAll, false},
		{"n_ accepted", jsontestsuite.MustReject, acceptAll, true},
		{"n_ rejected", jsontestsuite.MustReject, rejectAll, false},
		{"i_ accepted", jsontestsuite.ImplDefined, acceptAll, false},
		{"i_ rejected", jsontestsuite.ImplDefined, rejectAll, false},
		{"y_ panic", jsontestsuite.MustAccept, panics, true},
		{"n_ panic", jsontestsuite.MustReject, panics, true},
		{"i_ panic", jsontestsuite.ImplDefined, panics, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := jsontestsuite.Check(jsontestsuite.Case{Name: "x", Kind: tc.kind, Input: []byte("[]")}, tc.parse)
			if (err != nil) != tc.wantErr {
				t.Errorf("err = %v, wantErr %t", err, tc.wantErr)
			}
		})
	}
}

// encoding/json is the reference: the harness must pass it on every y_ and n_ case.
func TestRunAgainstEncodingJSON(t *testing.T) {
	rep := jsontestsuite.Run(t, stdlib)
	if len(rep.Implementation) != 35 {
		t.Errorf("i_ results: got %d, want 35", len(rep.Implementation))
	}
}
