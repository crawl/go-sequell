package version

import (
	"fmt"
	"strconv"
	"testing"
)

func TestIsVersionLike(t *testing.T) {
	for _, test := range []struct {
		text        string
		versionLike bool
	}{
		{"0.15.0-9-g1a96a59", true},
		{"0.15", true},
		{"0.3.2", true},
		{"cannonball", false},
	} {
		t.Run(fmt.Sprintf("IsVersionLike(%#v)==%#v", test.text, test.versionLike), func(t *testing.T) {
			if versionLike := IsVersionLike(test.text); versionLike != test.versionLike {
				t.Errorf("IsVersionLike(%#v) = %#v", test.text, versionLike)
			}
		})
	}
}

func TestVnumOrder(t *testing.T) {
	var vnumOrderTests = [][]string{
		{"0.15.0-9-g1a96a59", "0.15.0-34-g666c3a1", "0.15.1"},
		{"0.15.0", "0.15.0-9-g1a96a59", "0.15.0-34-g666c3a1"},
	}
	for _, vnumOrder := range vnumOrderTests {
		var lastNum uint64
		for i, ver := range vnumOrder {
			verNum := NumericID(ver)
			if verNum <= lastNum {
				t.Errorf("Version %s < %s (%d < %d), expected %s > %s",
					ver, vnumOrder[i-1],
					verNum, lastNum,
					ver, vnumOrder[i-1])
			}
			lastNum = verNum
		}
	}
}

var qualSplitTests = [][]string{
	{"", "", "", ""},
	{"a", "a", "", ""},
	{"a0", "a", "0", ""},
	{"a0-263", "a", "0", "263"},
	{"pow2-263", "pow", "2", "263"},
	{"34-g666c3a1", "", "", "34"},
}

func TestSplitQualifier(t *testing.T) {
	for _, qualSplitTest := range qualSplitTests {
		prefix, major, minor := SplitQualifierPrefixMajorMinor(qualSplitTest[0])
		if prefix != qualSplitTest[1] || major != qualSplitTest[2] ||
			minor != qualSplitTest[3] {
			t.Errorf("SplitQualifierPrefixMajorMinor(%#v) = %#v,%#v,%#v, expected %#v\n", qualSplitTest[0], prefix, major, minor, qualSplitTest[1:])
		}
	}
}

func TestMajor(t *testing.T) {
	if Major("1.22.15") != "1.22" {
		t.Errorf("expected major version of 1.22.15 to be 1.22")
	}

	if Major("xz 0.14.2-g035434") != "0.14" {
		t.Errorf("expected major version of 0.14.2-g035434 to be 0.14")
	}
}

func TestFullVersion(t *testing.T) {
	if Full("0.9") != "0.9.0" {
		t.Errorf("expected full version of 0.9 to be 0.9.0")
	}
	if Full("0.9-b1") != "0.9.0-b1" {
		t.Errorf("expected full version of 0.9-b1 to be 0.9.0-b1")
	}
	if Full("0.9.3-a0") != "0.9.3-a0" {
		t.Errorf("expected full version of 0.9.3-a0 to be 0.9.3-a0")
	}
}

var versionNumericIDs = [][]string{
	{"0.1.7", "100799000000"},
	{"0.8.0-a0", "800001000000"},
	{"0.8.0-rc1", "800018010000"},
	{"0.9.0", "900099000000"},
	{"0.9", "900099000000"},
	{"0.10.4", "1000499000000"},
}

func TestVersionNumericId(t *testing.T) {
	testVersionID(t, NumericID)
}

func TestCachingVersionNumericId(t *testing.T) {
	testVersionID(t, CachingNumericID)
}

func BenchmarkVersionNumericId(b *testing.B) {
	benchmarkNumericID(b, NumericID)
}

func BenchmarkCachingVersionNumericId(b *testing.B) {
	benchmarkNumericID(b, CachingNumericID)
}

func benchmarkNumericID(b *testing.B, impl func(string) uint64) {
	for i := 0; i < b.N; i++ {
		for _, versionStrID := range versionNumericIDs {
			impl(versionStrID[0])
		}
	}
}

func testVersionID(t *testing.T, impl func(string) uint64) {
	for _, versionStrID := range versionNumericIDs {
		ver, id := versionStrID[0], versionStrID[1]
		res := strconv.FormatUint(impl(ver), 10)
		if res != id {
			t.Errorf("VersionNumericId(%s) == %s, expected %s\n",
				ver, res, id)
		}
	}
}

func TestExpandVersionKey(t *testing.T) {
	tests := [][]string{
		{"01", "0.1"},
		{"10", "0.10"},
	}
	for _, test := range tests {
		actual := ExpandVersionKey(test[0])
		if actual != test[1] {
			t.Errorf("ExpandVersionKey(%s) = %s, expected %s",
				test[0], actual, test[1])
		}
	}
}

func TestSplitVersionQualifier(t *testing.T) {
	var tests = [][]string{
		{"0.15.0-34-g666c3a1", "0.15.0", "34-g666c3a1"},
		{"0.8.0-rc1", "0.8.0", "rc1"},
		{"0.1.7", "0.1.7", ""},
	}
	for _, test := range tests {
		v := test[0]
		a, b := SplitQualifier(v)
		if a != test[1] || b != test[2] {
			t.Errorf("SplitQualifier(%#v) = (%#v, %#v); want (%#v, %#v)",
				v, a, b, test[1], test[2])
		}
	}
}
