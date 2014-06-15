package version

import "testing"

func TestMajorVersion(t *testing.T) {
	if MajorVersion("1.22.15") != "1.22" {
		t.Errorf("expected major version of 1.22.15 to be 1.22")
	}

	if MajorVersion("xz 0.14.2-g035434") != "0.14" {
		t.Errorf("expected major version of 0.14.2-g035434 to be 0.14")
	}
}

func TestFullVersion(t *testing.T) {
	if FullVersion("0.9") != "0.9.0" {
		t.Errorf("expected full version of 0.9 to be 0.9.0")
	}
	if FullVersion("0.9-b1") != "0.9.0-b1" {
		t.Errorf("expected full version of 0.9-b1 to be 0.9.0-b1")
	}
	if FullVersion("0.9.3-a0") != "0.9.3-a0" {
		t.Errorf("expected full version of 0.9.3-a0 to be 0.9.3-a0")
	}
}
