package stringnorm

import (
	"testing"
)

var reReplCases = []struct {
	regex    string
	repl     string
	input    string
	expected string
}{
	{`^an? \w+-headed (.*)$`, "a $1", "a two-headed hydra", "a hydra"},
}

func TestRegexpNormalizer(t *testing.T) {
	for _, replCase := range reReplCases {
		norm := StaticRegexpNormalizer(replCase.regex, replCase.repl)
		res, _ := norm.Normalize(replCase.input)
		if res != replCase.expected {
			t.Errorf("Expected %s =~ s/%s/%s/ == %s, but got %s", replCase.input, replCase.regex, replCase.repl, replCase.expected, res)
		}
	}
}
