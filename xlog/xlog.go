// Package xlog is an xlogfile parsing and manipulation library.
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
		result += key + "=" + QuoteValue(value)
	}
	return result
}

// Parse parses the given xlog line into an Xlog object.
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
		res[key] = UnquoteValue(value)

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
		nextSep := false
		for _, c := range s[sep+1:] {
			nextSep = c == sepRune
			break
		}
		if !nextSep {
			return sep + offset
		}
		offset += sep + 2
		s = s[sep+2:]
	}
}

// QuoteValue quotes an Xlog value field by escaping embedded ":" as "::".
func QuoteValue(value string) string {
	return strings.Replace(value, ":", "::", -1)
}

// UnquoteValue unquotes an Xlog value, replacing "::" with ":".
func UnquoteValue(value string) string {
	return strings.Replace(value, "::", ":", -1)
}

// An XlogParseError is an error in parsing an xlog line.
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
