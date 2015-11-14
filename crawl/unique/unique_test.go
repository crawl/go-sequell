package unique

import (
	"testing"

	"github.com/crawl/go-sequell/crawl/data"
)

var inst = New(data.CrawlData())

func TestIsUnique(t *testing.T) {
	if !inst.IsUnique("Blork the orc", "") {
		t.Errorf("Expected Blork to be unique, but isn't")
	}

	if inst.IsUnique("Abigabaxcjd", "") {
		t.Errorf("Expected junk name to be non-unique, but is")
	}

	if !inst.IsUnique("elcid", "zombie,unique,powwow") {
		t.Errorf("Expected flagged name to be unique, but isn't")
	}
}

func TestIsOrc(t *testing.T) {
	if !inst.IsOrc("Hawl") {
		t.Errorf("Expected Hawl to be flagged an orc")
	}
	if inst.IsOrc("Tarantino") {
		t.Errorf("Expected Tarantino to be flagged a non-orc")
	}
}

func TestMaybePanLord(t *testing.T) {
	tests := []struct {
		version string
		name    string
		panLord bool
	}{
		{"0.10", "a Bogon", false},
		{"0.10", "an ufetubus", false},
		{"0.10", "the Lernaean Dogfish", false},
		{"0.10", "Hawl", false},
		{"0.10", "Fruitfly", true},
		{"0.11", "Fruitfly", false},
		{"0.11", "Cow the pandemonium lord", true},
		{"0.10", "Cow the pandemonium lord", true},
	}
	for _, test := range tests {
		isPanLord := inst.MaybePanLord(test.version, test.name, "")
		if isPanLord != test.panLord {
			t.Errorf("MaybePanLord(%#v, %#v, \"\") == %#v, want %#v", test.version, test.name, isPanLord, test.panLord)
		}
	}
}

func TestGenericPanLordName(t *testing.T) {
	if name := inst.GenericPanLordName(); name != "a pandemonium lord" {
		t.Errorf("GenericPanLordName() produced bad result: %s", name)
	}
}
