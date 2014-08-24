package time

import (
	"fmt"
	"regexp"
	"strconv"
)

// Explicit 0-9: don't match Unicode digits.
var rYearMonth = regexp.MustCompile(`^[0-9]{6}`)
var rDaylightSavingSuffix = regexp.MustCompile(`[SD]$`)

func NormalizeUnixTime(time string) string {
	return rYearMonth.ReplaceAllStringFunc(
		rDaylightSavingSuffix.ReplaceAllLiteralString(time, ""),
		func(yearMonth string) string {
			monthString := yearMonth[4:]
			month, _ := strconv.ParseInt(monthString, 10, 32)
			return fmt.Sprintf("%s%02d", yearMonth[0:4], month+1)
		})
}
