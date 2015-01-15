package xlogtools

import (
	"testing"

	"github.com/greensnark/go-sequell/crawl/data"
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
	{
		xlog.Xlog{
			"race": "Red Draconian",
		},
		xlog.Xlog{
			"crace": "Draconian",
			"race":  "Red Draconian",
		},
	},
	{
		xlog.Xlog{
			"type":     "yak",
			"banisher": "red draconian",
		},
		xlog.Xlog{
			"type":      "yak",
			"banisher":  "a red draconian",
			"cbanisher": "a draconian",
		},
	},
	{
		xlog.Xlog{
			"type":     "yak",
			"banisher": "cow's ghost",
		},
		xlog.Xlog{
			"type":      "yak",
			"banisher":  "cow's ghost",
			"cbanisher": "a player ghost",
		},
	},
	{
		xlog.Xlog{
			"sk": "Transmigration",
		},
		xlog.Xlog{
			"sk": "Transmutations",
		},
	},
	{
		xlog.Xlog{
			"sk": "Translocations",
		},
		xlog.Xlog{
			"sk": "Translocations",
		},
	},
	{
		xlog.Xlog{
			"race": "Grotesk",
		},
		xlog.Xlog{
			"race":  "Grotesk",
			"crace": "Gargoyle",
		},
	},
	{
		xlog.Xlog{
			"race": "Kenku",
		},
		xlog.Xlog{
			"race":  "Kenku",
			"crace": "Tengu",
		},
	},
}

var norm = MustBuildNormalizer(data.Crawl)

func TestNormalizeLog(t *testing.T) {
	for _, test := range normXlogTest {
		log, expected := test[0], test[1]
		normalized, err := norm.NormalizeLog(log.Clone())
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
