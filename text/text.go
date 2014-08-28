package text

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func Str(any interface{}) string {
	if any == nil {
		return ""
	}
	switch t := any.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	default:
		return fmt.Sprintf("%s", t)
	}
}

var rSpaceRegexp = regexp.MustCompile(`\s+`)

func NormalizeSpace(text string) string {
	return rSpaceRegexp.ReplaceAllLiteralString(strings.TrimSpace(text), " ")
}

func FirstNotEmpty(choices ...string) string {
	for _, val := range choices {
		if val != "" {
			return val
		}
	}
	return ""
}

// ParseInt parses the integer from the text; in case of error,
// returns the default value.
func ParseInt(text string, defval int) int {
	v, err := strconv.ParseInt(text, 10, 32)
	if err != nil {
		return defval
	}
	return int(v)
}

// RightPadSlice returns a slice that is at least nparts elements,
// padding out with new elements equal to `pad`. The original slice
// may be returned unmodified, or a new slice may be allocated. The
// original is never modified.
func RightPadSlice(slice []string, nparts int, pad string) []string {
	sliceLen := len(slice)
	if sliceLen >= nparts {
		return slice
	}

	paddedSlice := make([]string, nparts)
	for i := 0; i < sliceLen; i++ {
		paddedSlice[i] = slice[i]
	}

	for i := sliceLen; i < nparts; i++ {
		paddedSlice[i] = pad
	}
	return paddedSlice
}
