package stringnorm

import (
	"regexp"
)

type regexpNormalizer struct {
	regexp      *regexp.Regexp
	replacement string
}

func Reg(regexp *regexp.Regexp, replacement string) Normalizer {
	return &regexpNormalizer{regexp: regexp, replacement: replacement}
}

func StaticReg(re, replacement string) Normalizer {
	return Reg(regexp.MustCompile(re), replacement)
}

func SR(re, replacement string) Normalizer {
	return StaticReg(re, replacement)
}

func (r *regexpNormalizer) Normalize(text string) (string, error) {
	return r.regexp.ReplaceAllString(text, r.replacement), nil
}
