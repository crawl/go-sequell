package ctime

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const LayoutTimeWithZone = "20060102150405Z0700"
const LayoutUTCTime = "20060102150405"
const LayoutUTCEpoch = "200601021504Z0700"
const LayoutDBTime = "2006-01-02 15:04:05"
const LayoutLogTime = LayoutUTCTime

// Explicit 0-9: don't match Unicode digits.
var rYearMonth = regexp.MustCompile(`^[0-9]{6}`)
var rDaylightSavingSuffix = regexp.MustCompile(`[SD]$`)

type DstLocation struct {
	loc    *time.Location
	dstLoc *time.Location
}

func ParseDstLocation(loc string, dst string) (DstLocation, error) {
	if loc == "" {
		return DstLocation{}, nil
	}
	if dst == "" {
		dst = loc
	}

	baseLocation, err := ParseTZ(loc)
	if err != nil {
		return DstLocation{}, err
	}
	dstLocation, err := ParseTZ(dst)
	if err != nil {
		return DstLocation{}, err
	}
	return DstLocation{loc: baseLocation, dstLoc: dstLocation}, nil
}

func (d DstLocation) IsZero() bool {
	return d.loc == nil || d.dstLoc == nil
}

func (d DstLocation) Location(dst bool) *time.Location {
	if dst {
		return d.dstLoc
	} else {
		return d.loc
	}
}

func (d DstLocation) String() string {
	return "{" + d.loc.String() + " " + d.dstLoc.String() + "}"
}

// ParseTZ parses a time zone string of the form "-0700" or "-07:00"
// into a time.Location. "Z" is recognized as UTC.
func ParseTZ(tz string) (*time.Location, error) {
	if tz == "Z" {
		return time.UTC, nil
	}
	tz = strings.Replace(tz, ":", "", 1)
	perr := func() error {
		return fmt.Errorf("Malformed timezone spec: '%s'", tz)
	}

	if len(tz) != 5 {
		return nil, perr()
	}
	offsetSign, hours, minutes := tz[0], tz[1:3], tz[3:5]
	nhour, err := strconv.ParseInt(hours, 10, 32)
	if err != nil {
		return nil, perr()
	}
	nmin, err := strconv.ParseInt(minutes, 10, 32)
	if err != nil {
		return nil, perr()
	}
	var mul int64 = 1
	if offsetSign == '-' {
		mul = -1
	}
	return time.FixedZone(tz, int(mul*nhour*60*60+nmin*60)), nil
}

// NormalizeUnixTime formats a Unix time with DST qualifier into a standard
// time formatted as 20060102150405[DS]
func NormalizeUnixTime(time string) string {
	return rYearMonth.ReplaceAllStringFunc(
		time,
		func(yearMonth string) string {
			monthString := yearMonth[4:]
			month, _ := strconv.ParseInt(monthString, 10, 32)
			return fmt.Sprintf("%s%02d", yearMonth[0:4], month+1)
		})
}

// ParseTimeWithZone parses a Crawl format time with a timezone into a
// time object. If the timezone is unspecified, assumes UTC.
func ParseTimeWithZone(stime string) (time.Time, error) {
	t, err := time.Parse(LayoutTimeWithZone, stime)
	if err != nil {
		t, err = time.Parse(LayoutUTCTime, stime)
		if err != nil {
			return time.Time{}, err
		}
		return t, nil
	}
	return t, nil
}

// SplitDstQualifier splits a Crawl log time string into a time string
// and a DST qualifier that is true for Daylight Savings Time and false
// otherwise.
func SplitDstQualifier(logtime string) (string, bool) {
	if logtime == "" {
		return "", false
	}

	qualifierIndex := len(logtime) - 1
	qualifier := logtime[qualifierIndex]
	if qualifier == 'D' || qualifier == 'S' {
		return logtime[:qualifierIndex], dstQualifier(qualifier)
	}
	return logtime, false
}

func dstQualifier(qualifier byte) bool {
	return qualifier == 'D'
}

// ParseLogTime parses a Crawl log time to a UTC time. If a UTC epoch is
// provided that is after the given log time, then the time is parsed in the
// server's local time zone as specified by dstlocations.
func ParseLogTime(
	logtime string,
	utcepoch time.Time,
	dstlocations DstLocation) (time.Time, error) {

	logtimestr, dst := SplitDstQualifier(NormalizeUnixTime(logtime))
	utcTime, err := time.Parse(LayoutLogTime, logtimestr)
	if err != nil {
		return time.Time{}, err
	}

	if utcepoch.IsZero() || dstlocations.IsZero() || utcTime.After(utcepoch) {
		if dst {
			return time.Time{}, fmt.Errorf("Unexpected DST in %s with no TZ conversion available", logtime)
		}
		return utcTime, nil
	}

	localTime, err := time.ParseInLocation(LayoutLogTime, logtimestr,
		dstlocations.Location(dst))
	if err != nil {
		return time.Time{}, err
	}
	return localTime.UTC(), nil
}

// SafeParseUTCEpoch parses a Crawl UTC epoch formatted as
// 200601021504Z0700 into a UTC time.
func SafeParseUTCEpoch(stime string) time.Time {
	t, err := time.Parse(LayoutUTCEpoch, stime)
	if err != nil {
		return time.Time{}
	}
	return t.UTC()
}

// SafeParseTimeWithZone behaves like ParseTimeWithZone, but returns a
// zero time instead of an error if the parse fails.
func SafeParseTimeWithZone(stime string) time.Time {
	t, _ := ParseTimeWithZone(stime)
	return t
}
