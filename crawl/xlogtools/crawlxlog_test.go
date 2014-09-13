package xlogtools

import (
	"testing"

	"github.com/greensnark/go-sequell/xlog"
)

var normXlogTest = [][]xlog.Xlog{
	{
		xlog.Xlog{
			"src":       "cao",
			"v":         "0.10-a0",
			"alpha":     "y",
			"sk":        "Translocation",
			"gold":      "-10",
			"type":      "unique",
			"milestone": "pacified Sigmund",
			"start":     "20140001123755S",
		},
		xlog.Xlog{
			"v":         "0.10.0-a0",
			"cv":        "0.10-a",
			"sk":        "Translocations",
			"gold":      "0",
			"goldfound": "0",
			"goldspent": "0",
			"verb":      "uniq.pac",
			"noun":      "Sigmund",
			"rstart":    "20140001123755S",
			"start":     "20140001123755S",
		},
	},
}

func TestNormalizeLog(t *testing.T) {
	for _, test := range normXlogTest {
		log, expected := test[0], test[1]
		normalized, err := NormalizeLog(log.Clone())
		if err != nil {
			t.Errorf("NormalizeLog(%v) failed: %v\n", log, err)
			continue
		}

		for key, expectedValue := range expected {
			actual := normalized[key]
			if actual != expectedValue {
				t.Errorf("NormalizeLog(%#v): original: %#v, expected: %#v, actual: %#v\n", key, log[key], expectedValue, actual)
			}
		}
	}
}
