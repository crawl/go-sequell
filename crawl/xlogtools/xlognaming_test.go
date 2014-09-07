package xlogtools

import (
	"testing"
)

func TestXlogGame(t *testing.T) {
	tests := [][]string{
		{"logfile04", ""},
		{"logfile11-zotdef", "zotdef"},
		{"logfile15-sprint", "sprint"},
		{"allgames-spr", "sprint"},
		{"allgames-zd", "zotdef"},
		{"meta/nostalgia/logfile", "nostalgia"},
	}
	for _, test := range tests {
		actual := XlogGame(test[0])
		if actual != test[1] {
			t.Errorf("XlogGame(%s) = %s, expected %s", test[0], actual, test[1])
		}
	}
}

func TestXlogGameVersion(t *testing.T) {
	tests := [][]string{
		{"logfile", "any"},
		{"logfile-git", "git"},
		{"allgames-trunk", "git"},
		{"cow-svn", "git"},
		{"zap-master", "git"},
		{"yak-0.13", "0.13"},
		{"foo/0.9/bar", "0.9"},
		{"logfile11", "0.11"},
		{"yak02-pow", "0.2"},
	}
	for _, test := range tests {
		if actual := XlogGameVersion(test[0]); actual != test[1] {
			t.Errorf("XlogGameVersion(%s) = %s, expected %s",
				test[0], actual, test[1])
		}
	}
}
