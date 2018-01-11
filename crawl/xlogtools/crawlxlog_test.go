package xlogtools

import (
	"strconv"
	"testing"

	"github.com/crawl/go-sequell/crawl/data"
	"github.com/crawl/go-sequell/crawl/version"
	"github.com/crawl/go-sequell/xlog"
)

type xinput xlog.Xlog
type xoutput xlog.Xlog

var normXlogTest = []struct {
	xinput
	xoutput
}{
	{
		xinput{
			"src":       "cao",
			"v":         "0.10-a0",
			"alpha":     "y",
			"sk":        "Translocation",
			"gold":      "-10",
			"type":      "unique",
			"milestone": "pacified Sigmund",
			"start":     "20140001123755S",
		},
		xoutput{
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
		xinput{
			"vsavrv": "Git::0.10.0-a0",
		},
		xoutput{
			"vsavrv":    "0.10.0-a0",
			"vsavrvnum": strconv.FormatUint(version.NumericID("0.10.0-a0"), 10),
		},
	},
	{
		xinput{
			"vsav": "34.150",
		},
		xoutput{
			"vsav":    "34.150",
			"vsavnum": strconv.FormatUint(version.NumericID("34.150"), 10),
		},
	},
	{
		xinput{
			"race": "Red Draconian",
		},
		xoutput{
			"crace": "Draconian",
			"race":  "Red Draconian",
		},
	},
	{
		xinput{
			"type":     "yak",
			"banisher": "red draconian",
		},
		xoutput{
			"type":      "yak",
			"banisher":  "a red draconian",
			"cbanisher": "a draconian",
		},
	},
	{
		xinput{
			"type":     "yak",
			"banisher": "cow's ghost",
		},
		xoutput{
			"type":      "yak",
			"banisher":  "cow's ghost",
			"cbanisher": "a player ghost",
		},
	},
	{
		xinput{
			"type":     "yak",
			"banisher": "Peony",
		},
		xoutput{
			"type":      "yak",
			"banisher":  "Peony",
			"cbanisher": "a pandemonium lord",
		},
	},
	{
		xinput{
			"type":     "yak",
			"banisher": "you",
		},
		xoutput{
			"type":      "yak",
			"banisher":  "you",
			"cbanisher": "you",
		},
	},
	{
		xinput{
			"type":     "yak",
			"banisher": "miscasting Shatter",
		},
		xoutput{
			"type":      "yak",
			"banisher":  "miscasting Shatter",
			"cbanisher": "miscast",
		},
	},
	{
		xinput{
			"type":     "yak",
			"banisher": "distortion unwield",
		},
		xoutput{
			"type":      "yak",
			"banisher":  "distortion unwield",
			"cbanisher": "unwield",
		},
	},
	{
		xinput{
			"sk": "Transmigration",
		},
		xoutput{
			"sk": "Transmutations",
		},
	},
	{
		xinput{
			"sk": "Translocations",
		},
		xoutput{
			"sk": "Translocations",
		},
	},
	{
		xinput{
			"race": "Grotesk",
		},
		xoutput{
			"race":  "Grotesk",
			"crace": "Gargoyle",
		},
	},
	{
		xinput{
			"type": "god.ecumenical",
			"god":  "Makhleb",
		},
		xoutput{
			"verb": "god.ecumenical",
			"noun": "Makhleb",
		},
	},
	{
		xinput{
			"race": "Kenku",
		},
		xoutput{
			"race":  "Kenku",
			"crace": "Tengu",
		},
	},
	{
		xinput{
			"type":      "unique",
			"milestone": "slimified Maurice",
		},
		xoutput{
			"type":      "unique",
			"verb":      "uniq.slime",
			"milestone": "slimified Maurice",
			"noun":      "Maurice",
		},
	},
	{
		xinput{
			"explbr": "HEAD",
		},
		xoutput{
			"explbr": "",
		},
	},
	{
		xinput{
			"explbr": "crawl-0.16",
		},
		xoutput{
			"explbr": "",
		},
	},
	{
		xinput{
			"type":      "ancestor.class",
			"milestone": "remembered their ancestor Yon as a battlemage.",
		},
		xoutput{
			"type":      "ancestor.class",
			"verb":      "ancestor.class",
			"milestone": "remembered their ancestor Yon as a battlemage.",
			"noun":      "battlemage",
		},
	},
	{
		xinput{
			"type":      "ancestor.special",
			"milestone": "remembered their ancestor Cihuaton casting Metabolic Englaciation.",
		},
		xoutput{
			"type":      "ancestor.special",
			"verb":      "ancestor.special",
			"milestone": "remembered their ancestor Cihuaton casting Metabolic Englaciation.",
			"noun":      "Metabolic Englaciation",
		},
	},
	{
		xinput{
			"type":      "ancestor.special",
			"milestone": "remembered their ancestor Servius wielding a demon trident.",
		},
		xoutput{
			"type":      "ancestor.special",
			"verb":      "ancestor.special",
			"milestone": "remembered their ancestor Servius wielding a demon trident.",
			"noun":      "demon trident",
		},
	},
	{
		xinput{
			"race": "Gnome",
			"char": "GnFE",
		},
		xoutput{
			"race":  "Gnome",
			"crace": "Gnome",
			"char":  "GmFE",
		},
	},
	{
		xinput{
			"race": "Bultungin",
			"char": "BuMo",
		},
		xoutput{
			"race":  "Gnoll",
			"crace": "Gnoll",
			"char":  "GnMo",
		},
	},
	{
		xinput{
			"race": "Gnoll",
			"char": "GnMo",
		},
		xoutput{
			"race":  "Gnoll",
			"crace": "Gnoll",
			"char":  "GnMo",
		},
	},
}

var norm = MustBuildNormalizer(data.CrawlData().YAML)

func TestNormalizeLog(t *testing.T) {
	for _, test := range normXlogTest {
		inputXlog, expectedXlog := test.xinput, test.xoutput

		normalizedXlog := xlog.Xlog(inputXlog).Clone()
		if err := norm.NormalizeLog(normalizedXlog); err != nil {
			t.Errorf("NormalizeLog(%v) failed: %v\n", inputXlog, err)
			continue
		}

		for key, expectedValue := range expectedXlog {
			actualValue := normalizedXlog[key]
			if actualValue != expectedValue {
				t.Errorf("NormalizeLog(%#v: %#v): original: %#v, expected: %#v, actual: %#v\n", key, inputXlog, inputXlog[key], expectedValue, actualValue)
			}
		}
	}
}
