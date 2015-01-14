package grammar

import "regexp"

var reStartsWithArticle = regexp.MustCompile(`^(?:an?|the) `)

func Article(thing string) string {
	if reStartsWithArticle.FindString(thing) != "" {
		return thing
	}
	for _, rune := range thing {
		if IsVowel(rune) {
			return "an " + thing
		} else {
			return "a " + thing
		}
	}
	return thing
}

func IsVowel(r rune) bool {
	return r == 'a' || r == 'e' || r == 'i' || r == 'o' || r == 'u'
}
