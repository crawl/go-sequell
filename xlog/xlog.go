// Package xlog is an xlogfile parsing and manipulation library.
package xlog

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/crawl/go-sequell/text"
)

// An Xlog is a mapping of xlog keys to values, where both keys and values are
// strings.
type Xlog map[string]string

// Clone returns a copy of x.
func (x Xlog) Clone() Xlog {
	res := make(Xlog)
	for key, value := range x {
		res[key] = value
	}
	return res
}

// Get gets the xlog value for the field name.
func (x Xlog) Get(name string) string {
	return x[name]
}

// Contains checks if x contains key.
func (x Xlog) Contains(key string) bool {
	_, exists := x[key]
	return exists
}

func (x Xlog) String() string {
	result := ""
	for key, value := range x {
		if IsKeyHidden(key) {
			continue
		}
		if len(result) > 0 {
			result += ":"
		}
		result += key + "=" + QuoteValue(value)
	}
	return result
}

// IsKeyHidden checks if key is a *hidden* xlog fieldname
func IsKeyHidden(key string) bool {
	return key == "" || key[0] == ':'
}

// Parse parses the given xlog line into an Xlog object.
func Parse(line, sourceKey string) (Xlog, error) {
	line = strings.TrimSpace(line)

	parsedXlog := Xlog{
		"hash": text.Hash(sourceKey + ": " + line),
	}

	startIndex := 0
	lineByteLen := len(line)
	for startIndex < lineByteLen {
		keyValueSeparator := strings.IndexRune(line[startIndex:], '=')
		if keyValueSeparator == -1 {
			return parsedXlog, &ParseError{
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
		parsedXlog[key] = UnquoteValue(value)

		if nextSeparator != -1 {
			startIndex += nextSeparator + 1
		} else {
			startIndex = lineByteLen
		}
	}
	return parsedXlog, nil
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

// IsPotentialXlogLine returns true if line looks like it might be a
// valid Xlog line. This is a convenient shortcut to discard trivial
// invalid lines such as blank lines and lines starting with colons.
//
// It is the caller's responsibility to strip any extraneous leading
// and trailing space, including trailing newlines.
func IsPotentialXlogLine(line string) bool {
	return len(line) > 0 && line[0] != ':'
}

// A ParseError is an error in parsing an xlog line.
type ParseError struct {
	Line         string
	Cause        string
	ErrByteIndex int
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("malformed xlogline \"%s\": %s at %d", e.Line, e.Cause,
		e.ErrRuneIndex())
}

// ErrRuneIndex gets the rune index of the parse error in the xlog Line.
func (e *ParseError) ErrRuneIndex() int {
	if e.ErrByteIndex <= 0 {
		return e.ErrByteIndex
	}
	return utf8.RuneCountInString(e.Line[:e.ErrByteIndex])
}
