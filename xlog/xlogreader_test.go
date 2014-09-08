package xlog

import (
	"testing"
)

func TestReader(t *testing.T) {
	file := "cszo-git.log"
	reader := Reader(file)
	lines, err := reader.ReadAll()
	if err != nil {
		t.Errorf("Unexpected error reading %s: %s", file, err)
	}
	if len(lines) != 9 {
		t.Errorf("Expected to read 9 lines from %s, got %d",
			file, len(lines))
	}

	var tests = []struct {
		line int
		xlog Xlog
	}{
		{0, Xlog{
			"v":     "0.16-a0",
			"vlong": "0.16-a0-341-gb66f077",
			"name":  "OneEyedJack",
			"end":   "20140808162913S",
			"tmsg":  "slain by a jackal",
		}},
		{8, Xlog{
			"vmsg": "mangled by a gnoll (a +2 whip of electrocution)",
			"name": "Inkie",
		}},
	}
	for _, test := range tests {
		actual := lines[test.line]
		for k, v := range test.xlog {
			if actual[k] != v {
				t.Errorf("Unexpected xlog line on line %d: %#v, expected %#v",
					test.line+1, actual, test.xlog)
			}
		}
	}
}
