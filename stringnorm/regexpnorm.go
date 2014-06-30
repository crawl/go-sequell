package stringnorm

import (
	"regexp"
)

type regexpNormalizer struct {
	regexp      *regexp.Regexp
	replacement string
}

func RegexpNormalizer(regexp *regexp.Regexp, replacement string) Normalizer {
	return &regexpNormalizer{regexp: regexp, replacement: replacement}
}

func StaticRegexpNormalizer(re, replacement string) Normalizer {
	return RegexpNormalizer(regexp.MustCompile(re), replacement)
}

func (r *regexpNormalizer) Normalize(text string) (string, error) {
	return r.regexp.ReplaceAllString(text, r.replacement), nil
}
