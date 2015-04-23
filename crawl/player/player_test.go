package player

import (
	"testing"

	"github.com/crawl/go-sequell/crawl/data"
)

var norm = NewCharNormalizer(
	data.Crawl.Map("species"),
	data.Crawl.Map("classes"))

func TestNormalizeChar(t *testing.T) {
	var tests = []struct {
		race, class, expected string
	}{
		{"Draconian", "Reaver", "DrRe"},
		{"Ghoul", "Skald", "GhSk"},
		{"Gherkin", "Fighter", ""},
	}
	for _, test := range tests {
		actual := norm.NormalizeChar(test.race, test.class, "")
		if actual != test.expected {
			t.Errorf("NormalizeChar(%#v,%#v,%#v) = %#v; want %#v",
				test.race, test.class, "", actual, test.expected)
		}
	}
}
