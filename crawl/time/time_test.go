package time

import "testing"

var unixTimeTests = [][]string{
	{"20140127094533S", "20140227094533"},
	{"20110002175502D", "20110102175502"},
}

func TestNormalizeUnixTime(t *testing.T) {
	for _, unixTime := range unixTimeTests {
		utime, normalizedTime := unixTime[0], unixTime[1]
		res := NormalizeUnixTime(utime)
		if res != normalizedTime {
			t.Errorf("NormalizeUnixTime(%#v) == %#v, expected %#v\n",
				utime, res, normalizedTime)
		}
	}
}
