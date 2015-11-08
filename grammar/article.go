package grammar

import (
	"regexp"
	"unicode"
)

var reStartsWithArticle = regexp.MustCompile(`^(?:an?|the) `)

// Article prefixes thing with "a" or "an" as appropriate.
func Article(thing string) string {
	if reStartsWithArticle.FindString(thing) != "" {
		return thing
	}
	for _, rune := range thing {
		if unicode.IsUpper(rune) {
			return thing
		}
		if IsVowel(rune) {
			return "an " + thing
		}

		return "a " + thing
	}
	return thing
}

// IsVowel checks if r is an English vowel.
func IsVowel(r rune) bool {
	return r == 'a' || r == 'e' || r == 'i' || r == 'o' || r == 'u'
}
