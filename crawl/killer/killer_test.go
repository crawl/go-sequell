package killer

import (
	"github.com/greensnark/go-sequell/xlog"
	"testing"
)

var killerTests = []struct {
	start  string
	finish string
}{
	{"a two-headed hydra", "a hydra"},
	{"a 12-headed hydra", "a hydra"},
	{"the 27-headed Lernaean hydra zombie", "the Lernaean hydra zombie"},
	{"foobar's ghost", "a player ghost"},
	{"ackbar's illusion", "a player illusion"},
	{"a red draconian shifter", "a draconian shifter"},
	{"a yak (shapeshifter)", "a shapeshifter"},
	{"a camel (glowing shapeshifter)", "a glowing shapeshifter"},
	{"Grum", "Grum"},
	{"Bogdan the orc", "an orc"},
	{"Ghib", "a pandemonium lord"},
}

func TestNormalizeKiller(t *testing.T) {
	for _, test := range killerTests {
		res := NormalizeKiller(test.start, xlog.Xlog{"killer": test.start})
		if res != test.finish {
			t.Errorf("Expected '%s' to normalize to '%s', but got '%s'", test.start, test.finish, res)
		}
	}
}
