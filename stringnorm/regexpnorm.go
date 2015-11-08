package stringnorm

import (
	"regexp"
)

// A RegexpNormalizer applies a regexp search+replace to text.
type RegexpNormalizer struct {
	Regexp      *regexp.Regexp
	Replacement string
}

// Reg returns a Normalizer given a regexp to search for and a replacement
// string.
func Reg(regexp *regexp.Regexp, replacement string) Normalizer {
	return &RegexpNormalizer{Regexp: regexp, Replacement: replacement}
}

// StaticReg returns a regexp Normalizer for re, with the replacment string.
// Panics if re is not a valid regexp.
func StaticReg(re, replacement string) Normalizer {
	return Reg(regexp.MustCompile(re), replacement)
}

// SR is an alias for StaticReg
func SR(re, replacement string) Normalizer {
	return StaticReg(re, replacement)
}

// Normalize applies a regexp search+replace to text, searching for r.Regexp
// and replacing any matches with r.Replacement.
func (r *RegexpNormalizer) Normalize(text string) (string, error) {
	return r.Regexp.ReplaceAllString(text, r.Replacement), nil
}
