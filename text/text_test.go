package text

import (
	"reflect"
	"testing"
)

var braceExpandTests = []struct {
	text       string
	expansions []string
}{
	{"yoda", []string{"yoda"}},
	{"", []string{""}},
	{"yak{}cow", []string{"yakcow"}},
	{"foo{bar,}bletch", []string{"foobarbletch", "foobletch"}},
	{"a{b,c,d,x{e,f,g}}h{i,j}",
		[]string{
			"abhi", "abhj", "achi", "achj", "adhi", "adhj",
			"axehi", "axehj", "axfhi", "axfhj", "axghi", "axghj",
		}},
}

func TestBraceExpand(t *testing.T) {
	for _, test := range braceExpandTests {
		res, err := ExpandBraceGroups(test.text)
		if err != nil {
			t.Errorf("Unexpected error expanding %s: %s", test.text, err)
			continue
		}
		if !reflect.DeepEqual(res, test.expansions) {
			t.Errorf("BraceExpand(%#v) == %#v, expected %#v",
				test.text, res, test.expansions)
		}
	}
}
