package sources

import (
	"fmt"
	"testing"

	"github.com/greensnark/go-sequell/qyaml"
)

func TestURLJoin(t *testing.T) {
	var testCases = []struct {
		base, path string
		expected   string
	}{
		{"http://crawl.berotato.org/crawl", "meta/nostalgia/logfile",
			"http://crawl.berotato.org/crawl/meta/nostalgia/logfile"},
		{"http://yak.foo", "http://bar.foo", "http://bar.foo"},
	}
	for _, test := range testCases {
		actual := URLJoin(test.base, test.path)
		if actual != test.expected {
			t.Errorf("URLJoin(%#v, %#v) = %#v, expected %#v",
				test.base, test.path, actual, test.expected)
		}
	}
}

func TestSources(t *testing.T) {
	schema, err := qyaml.Parse("test-sources.yml")
	if err != nil {
		t.Errorf("Error parsing yaml: %s", err)
		return
	}
	src, err := Sources(schema, "test")
	if err != nil {
		t.Errorf("Error parsing sources: %s", err)
		return
	}
	expectedCount := 11
	if len(src.Servers) != expectedCount {
		t.Errorf("Expected %d sources, got %d", expectedCount, len(src.Servers))
		return
	}

	cao := src.Server("cao")
	if cao == nil {
		t.Errorf("Couldn't find source cao in %s", src)
	}
	expectedUrl := "http://crawl.akrasiac.org/allgames.txt"
	if cao.Logfiles[0].URL != expectedUrl {
		t.Errorf("Expected CAO first URL to be %s, was %s", expectedUrl,
			cao.Logfiles[0].URL)
	}

	if cao.TimeZoneMap.IsZero() {
		t.Errorf("CAO has no tz map")
	} else {
		fmt.Printf("CAO tz maps: %#v\n", cao.TimeZoneMap)
	}
	if cao.UtcEpoch.IsZero() {
		t.Errorf("CAO has no UTC epoch")
	} else {
		fmt.Printf("CAO UTC epoch: %s\n", cao.UtcEpoch)
	}

	for _, srv := range src.Servers {
		fmt.Println()
		fmt.Println(srv.Name)
		for i, log := range srv.Logfiles {
			fmt.Printf("%02d) %s\n", i+1, log)
		}
	}
}
