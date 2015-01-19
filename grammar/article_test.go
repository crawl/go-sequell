package grammar

import "testing"

func TestArticle(t *testing.T) {
	for _, test := range []struct {
		word, expected string
	}{
		{"cow", "a cow"},
		{"the trees", "the trees"},
		{"umbrella", "an umbrella"},
		{"Ptolemy", "Ptolemy"},
	} {
		if actual := Article(test.word); actual != test.expected {
			t.Errorf("Article(%#v) == %#v, want %#v",
				test.word, actual, test.expected)
		}
	}
}
