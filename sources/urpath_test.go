package sources

import "testing"

func TestURLTargetPath(t *testing.T) {
	for _, c := range []struct {
		alias, base, path string
		expected          string
	}{
		{"cao", "http://crawl.akrasiac.org", "allgames.txt",
			"cao/allgames.txt"},
		{"cszo", "http://dobrazupa.org", "meta/0.16/logfile",
			"cszo/meta/0.16/logfile"},
		{"cbro", "http://crawl.berotato.org/crawl",
			"meta/0.16/logfile-sprint",
			"cbro/crawl/meta/0.16/logfile-sprint"},
	} {
		result := URLTargetPath(c.alias, c.base, c.path)
		if result != c.expected {
			t.Errorf("URLTargetPath(%#v, %#v, %#v) == %#v, want %#v",
				c.alias, c.base, c.path, result, c.expected)
		}
	}
}
