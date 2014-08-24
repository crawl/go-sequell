package xlog

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Xlog map[string]string

func (x Xlog) Clone() Xlog {
	res := make(Xlog)
	for key, value := range x {
		res[key] = value
	}
	return res
}

func (x Xlog) Get(name string) string {
	return x[name]
}

func (x Xlog) Contains(key string) bool {
	_, exists := x[key]
	return exists
}

func (x Xlog) String() string {
	result := ""
	for key, value := range x {
		if len(result) > 0 {
			result += ":"
		}
		result += key + "=" + strings.Replace(value, ":", "::", -1)
	}
	return result
}

// Parse parses the given xlog line into a map
func Parse(line string) (Xlog, error) {
	line = strings.TrimSpace(line)
	res := make(Xlog)

	startIndex := 0
	lineByteLen := len(line)
	for startIndex < lineByteLen {
		keyValueSeparator := strings.IndexRune(line[startIndex:], '=')
		if keyValueSeparator == -1 {
			return res, &XlogParseError{
				Line:         line,
				Cause:        "trailing characters",
				ErrByteIndex: startIndex,
			}
		}

		keyValueSeparator += startIndex
		key := strings.TrimSpace(line[startIndex:keyValueSeparator])

		startIndex = keyValueSeparator + 1

		nextSeparator := FindSeparator(line[startIndex:])

		fieldEndIndex := lineByteLen
		if nextSeparator != -1 {
			fieldEndIndex = nextSeparator + startIndex
		}

		value := line[startIndex:fieldEndIndex]
		res[key] = NormalizeValue(value)

		if nextSeparator != -1 {
			startIndex += nextSeparator + 1
		} else {
			startIndex = lineByteLen
		}
	}
	return res, nil
}

// FindSeparator finds the first non-escaped occurrence of the xlog separator
// ':' in the string and returns the byte offset of that occurrence, or -1 if
// the separator is not present in the string.
func FindSeparator(s string) int {
	offset := 0
	sepRune := ':'
	for {
		sep := strings.IndexRune(s, sepRune)
		if sep == -1 {
			return -1
		}
		nextSep := strings.IndexRune(s[sep+1:], sepRune)
		if nextSep != 0 {
			return sep + offset
		}
		offset += sep + 2
		s = s[offset:]
	}
}

func NormalizeValue(value string) string {
	return strings.Replace(value, "::", ":", -1)
}

type XlogParseError struct {
	Line         string
	Cause        string
	ErrByteIndex int
}

func (e *XlogParseError) Error() string {
	return fmt.Sprintf("malformed xlogline \"%s\": %s at %d", e.Line, e.Cause,
		e.ErrRuneIndex())
}

func (e *XlogParseError) ErrRuneIndex() int {
	if e.ErrByteIndex <= 0 {
		return e.ErrByteIndex
	}
	return utf8.RuneCountInString(e.Line[:e.ErrByteIndex])
}
