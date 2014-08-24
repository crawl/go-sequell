package version

import (
	"strconv"
	"testing"
)

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
	{"0.1.7", "1007000998001"},
	{"0.8.0-a0", "8000000097000"},
	{"0.8.0-rc1", "8000000114001"},
	{"0.9.0", "9000000998001"},
	{"0.9", "9000000998001"},
	{"0.10.4", "10004000998001"},
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
