package stringnorm

import (
	"regexp"
)

type RegexpNormalizer struct {
	Regexp      *regexp.Regexp
	Replacement string
}

func Reg(regexp *regexp.Regexp, replacement string) Normalizer {
	return &RegexpNormalizer{Regexp: regexp, Replacement: replacement}
}

func StaticReg(re, replacement string) Normalizer {
	return Reg(regexp.MustCompile(re), replacement)
}

func SR(re, replacement string) Normalizer {
	return StaticReg(re, replacement)
}

func (r *RegexpNormalizer) Normalize(text string) (string, error) {
	return r.Regexp.ReplaceAllString(text, r.Replacement), nil
}
