package god

import (
	"strings"

	"github.com/crawl/go-sequell/crawl/data"
)

type Normalizer map[string]string

func (n Normalizer) Normalize(god string) (string, error) {
	if canonical, exists := n[strings.ToLower(god)]; exists {
		return canonical, nil
	}
	return god, nil
}

var godAliases = data.Crawl.StringMap("god-aliases")

func CanonicalGod(god string) string {
	if canonical, exists := godAliases[strings.ToLower(god)]; exists {
		return canonical
	}
	return god
}
