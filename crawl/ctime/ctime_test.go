package ctime

import "testing"

var unixTimeTests = [][]string{
	{"20140127094533S", "20140227094533S"},
	{"20110002175502D", "20110102175502D"},
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

var timeParseTests = [][]string{
	{"20080807033000+0000", "20080807033000Z"},
	{"20080807053000+0200", "20080807033000Z"},
}

func TestSafeParseTimeWithZone(t *testing.T) {
	for _, test := range timeParseTests {
		original, expectedReformat := test[0], test[1]
		parsedTime, err := ParseTimeWithZone(original)
		if err != nil {
			t.Errorf("Unexpected error parsing %s: %s\n", original, err)
			continue
		}
		reformat := parsedTime.UTC().Format(LayoutTimeWithZone)
		if reformat != expectedReformat {
			t.Errorf("Reformatting %s (%v) = %s, expected %s\n",
				original, parsedTime, reformat, expectedReformat)
		}
	}
}

var tzParseTests = []struct {
	logtime, epoch, standardTZ, dstTZ, reformattedTime string
}{
	{"20061125235309S", "200808070330+0000", "-0500", "-0400", "20061226045309"},

	{"20061125235309D", "200808070330+0000", "-0500", "-0400", "20061226035309"},
	{"20091125235309S", "200808070330+0000", "-0500", "-0400", "20091225235309"},
}

func TestParseLogTime(t *testing.T) {
	for _, test := range tzParseTests {
		epoch := SafeParseUTCEpoch(test.epoch)
		dst, err := ParseDstLocation(test.standardTZ, test.dstTZ)
		if err != nil {
			t.Errorf("Error parsing DST locs (%s,%s): %s",
				test.standardTZ, test.dstTZ, err)
			continue
		}

		logtime, err := ParseLogTime(test.logtime, epoch, dst)
		if err != nil {
			t.Errorf("Error parsing log time %s: %s", test.logtime, err)
			continue
		}
		res := logtime.Format(LayoutUTCTime)
		if res != test.reformattedTime {
			t.Errorf("Expected log time %s to be normalized to %s, got %s",
				test.logtime, test.reformattedTime, res)
		}
	}
}
