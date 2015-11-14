package killer

import (
	"testing"

	"github.com/crawl/go-sequell/crawl/data"
)

var killerTests = []struct {
	start  string
	finish string
}{
	{"a two-headed hydra", "a hydra"},
	{"a 12-headed hydra", "a hydra"},
	{"the 27-headed Lernaean hydra zombie", "the Lernaean hydra zombie"},
	{"Foobar's ghost", "a player ghost"},
	{"ackbar's illusion", "a player illusion"},
	{"a red draconian shifter", "a draconian shifter"},
	{"a yak (shapeshifter)", "a shapeshifter"},
	{"a camel (glowing shapeshifter)", "a glowing shapeshifter"},
	{"a red ugly thing", "an ugly thing"},
	{"a green very ugly thing", "a very ugly thing"},
	{"Grum", "Grum"},
	{"Bogdan the orc", "an orc"},
	{"Ghib", "a pandemonium lord"},
}

func TestNormalizeKiller(t *testing.T) {
	k := NewNormalizer(data.CrawlData())
	for _, test := range killerTests {
		res := k.NormalizeKiller("0.10", test.start, test.start, "")
		if res != test.finish {
			t.Errorf("Expected '%s' to normalize to '%s', but got '%s'", test.start, test.finish, res)
		}
	}
}

var kmodTests = [][]string{
	{"a spectral warrior", ""},
	{"a spectral hydra", "spectre"},
	{"a dragon (shapeshifter)", "shapeshifter"},
	{"a glowing shapeshifter", "shapeshifter"},
	{"an 18-headed hydra zombie", "zombie"},
	{"a bat skeleton", "skeleton"},
	{"a goblin simulacrum", "simulacrum"},
	{"a kobold", ""},
}

func TestNormalizeKmod(t *testing.T) {
	for _, test := range kmodTests {
		killer, kmod := test[0], test[1]
		actual := NormalizeKmod(killer)
		if actual != kmod {
			t.Errorf("NormalizeKmod(%s) == %s, expected %s\n",
				killer, actual, kmod)
		}
	}
}

var kauxTests = [][]string{
	{"a longsword {cow}", "longsword"},
	{"a +2 sling (mu)", "sling"},
	{"an ox", "ox"},
	{"an elven crossbow", "crossbow"},
	{"Hit by a dart thrown by a horse", "dart"},
	{"Shot with an arrow by an elf", "arrow"},
	{"a cursed dagger", "dagger"},
}

func TestNormalizeKaux(t *testing.T) {
	for _, kauxTest := range kauxTests {
		kaux, ckaux := kauxTest[0], kauxTest[1]
		res := NormalizeKaux(kaux)
		if res != ckaux {
			t.Errorf("NormalizeKaux(%#v) == %#v, expected %#v\n",
				kaux, res, ckaux)
		}
	}
}
