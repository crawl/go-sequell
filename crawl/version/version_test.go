package version

import (
	"strconv"
	"testing"
)

var qualSplitTests = [][]string{
	{"", "", "", ""},
	{"a", "a", "", ""},
	{"a0", "a", "0", ""},
	{"a0-263", "a", "0", "263"},
	{"pow2-263", "pow", "2", "263"},
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

var versionNumericIds = [][]string{
	{"0.1.7", "100799999999"},
	{"0.8.0-a0", "800001000000"},
	{"0.8.0-rc1", "800018010000"},
	{"0.9.0", "900099999999"},
	{"0.9", "900099999999"},
	{"0.10.4", "1000499999999"},
}

func TestVersionNumericId(t *testing.T) {
	testVersionId(t, VersionNumericId)
}

func TestCachingVersionNumericId(t *testing.T) {
	testVersionId(t, CachingVersionNumericId)
}

func BenchmarkVersionNumericId(b *testing.B) {
	benchmarkVersionNumericId(b, VersionNumericId)
}

func BenchmarkCachingVersionNumericId(b *testing.B) {
	benchmarkVersionNumericId(b, CachingVersionNumericId)
}

func benchmarkVersionNumericId(b *testing.B, impl func(string) uint64) {
	for i := 0; i < b.N; i++ {
		for _, versionStrId := range versionNumericIds {
			impl(versionStrId[0])
		}
	}
}

func testVersionId(t *testing.T, impl func(string) uint64) {
	for _, versionStrId := range versionNumericIds {
		ver, id := versionStrId[0], versionStrId[1]
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
