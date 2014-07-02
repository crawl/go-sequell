package unique

import (
	"github.com/greensnark/go-sequell/xlog"
	"testing"
)

func TestIsUnique(t *testing.T) {
	if !IsUnique("Blork the orc", xlog.Xlog{}) {
		t.Errorf("Expected Blork to be unique, but isn't")
	}

	if IsUnique("Abigabaxcjd", xlog.Xlog{}) {
		t.Errorf("Expected junk name to be non-unique, but is")
	}

	if !IsUnique("elcid", xlog.Xlog{"killer_flags": "zombie,unique,powwow"}) {
		t.Errorf("Expected flagged name to be unique, but isn't")
	}
}

func TestIsOrc(t *testing.T) {
	if !IsOrc("Hawl") {
		t.Errorf("Expected Hawl to be flagged an orc")
	}
	if IsOrc("Tarantino") {
		t.Errorf("Expected Tarantino to be flagged a non-orc")
	}
}

func TestMaybePanLord(t *testing.T) {
	for _, fake := range []string{"a Bogon", "an ufetubus", "the Lernaean Dogfish"} {
		if MaybePanLord(fake, xlog.Xlog{}) {
			t.Errorf("MaybePanLord: %s flagged a panlord, but isn't", fake)
		}
	}

	if MaybePanLord("Hawl", xlog.Xlog{}) {
		t.Errorf("MaybePanLord: Hawl incorrectly flagged as a panlord")
	}

	if !MaybePanLord("Fruitfly", xlog.Xlog{}) {
		t.Errorf("Fruitfly not flagged as a panlord")
	}
}

func TestGenericPanLordName(t *testing.T) {
	if GenericPanLordName() != "a pandemonium lord" {
		t.Errorf("GenericPanLordName() produced bad result: %s",
			GenericPanLordName())
	}
}
