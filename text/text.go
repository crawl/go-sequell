// Package text provides convenience functions for text manipulation.
package text

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Str converts any object into a string, either by a direct cast, a
// call to Stringer.String, falling back to fmt.Sprintf("%s", object).
// nil converts to "".
func Str(any interface{}) string {
	if any == nil {
		return ""
	}
	switch t := any.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	case int:
		return strconv.Itoa(t)
	case float32:
		return strconv.FormatFloat(float64(t), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
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

// ExpandBraceGroups expands {foo,bar} groups in text by returning every
// permutation of brace expansions, similar to shell brace expansion.
func ExpandBraceGroups(text string) ([]string, error) {
	firstBrace := strings.Index(text, "{")
	if firstBrace == -1 {
		return []string{text}, nil
	}
	group, end, err := scanBracedGroup(text[firstBrace:])
	if err != nil {
		return nil, err
	}
	tailExpansions, err := ExpandBraceGroups(text[end+firstBrace:])
	if err != nil {
		return nil, err
	}
	expansions, err := group.Expand()
	if err != nil {
		return nil, err
	}
	leader := text[:firstBrace]
	res := make([]string, len(expansions)*len(tailExpansions))
	i := 0
	for _, expansion := range expansions {
		prefix := leader + expansion
		for _, tailExp := range tailExpansions {
			res[i] = prefix + tailExp
			i++
		}
	}
	return res, nil
}

type braceGroup struct {
	groups []string
}

func (b braceGroup) Expand() ([]string, error) {
	res := make([]string, 0, len(b.groups))
	for _, g := range b.groups {
		groupExp, err := ExpandBraceGroups(g)
		if err != nil {
			return nil, err
		}
		res = append(res, groupExp...)
	}
	return res, nil
}

func scanBracedGroup(text string) (grp braceGroup, end int, err error) {
	groups := []string{}
	order := 0
	end = len(text)

	var groupStart int
	for i, c := range text {
		if order == 0 && i > 0 {
			end = i
			break
		}
		if (c == ',' || c == '}') && order == 1 {
			groups = append(groups, text[groupStart:i])
		}
		switch c {
		case '{':
			order++
			if order == 1 {
				groupStart = i + 1
			}
		case ',':
			if order == 1 {
				groupStart = i + 1
			}
		case '}':
			order--
		}
	}
	if order > 0 {
		return braceGroup{}, 0, fmt.Errorf("Mismatched brace group: %s", text)
	}
	return braceGroup{groups: groups}, end, nil
}
