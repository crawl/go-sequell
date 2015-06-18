package xlogtools

import (
	"testing"

	"github.com/crawl/go-sequell/crawl/data"
	"github.com/crawl/go-sequell/xlog"
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
			"type":     "yak",
			"banisher": "Peony",
		},
		xlog.Xlog{
			"type":      "yak",
			"banisher":  "Peony",
			"cbanisher": "a pandemonium lord",
		},
	},
	{
		xlog.Xlog{
			"type":     "yak",
			"banisher": "you",
		},
		xlog.Xlog{
			"type":      "yak",
			"banisher":  "you",
			"cbanisher": "you",
		},
	},
	{
		xlog.Xlog{
			"type":     "yak",
			"banisher": "miscasting Shatter",
		},
		xlog.Xlog{
			"type":      "yak",
			"banisher":  "miscasting Shatter",
			"cbanisher": "miscast",
		},
	},
	{
		xlog.Xlog{
			"type":     "yak",
			"banisher": "distortion unwield",
		},
		xlog.Xlog{
			"type":      "yak",
			"banisher":  "distortion unwield",
			"cbanisher": "unwield",
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
			"type": "god.ecumenical",
			"god":  "Makhleb",
		},
		xlog.Xlog{
			"verb": "god.ecumenical",
			"noun": "Makhleb",
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
	{
		xlog.Xlog{
			"type":      "unique",
			"milestone": "slimified Maurice",
		},
		xlog.Xlog{
			"type":      "unique",
			"verb":      "uniq.slime",
			"milestone": "slimified Maurice",
			"noun":      "Maurice",
		},
	},
	{
		xlog.Xlog{
			"explbr": "HEAD",
		},
		xlog.Xlog{
			"explbr": "",
		},
	},
	{
		xlog.Xlog{
			"explbr": "crawl-0.16",
		},
		xlog.Xlog{
			"explbr": "",
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
