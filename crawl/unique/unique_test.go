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
