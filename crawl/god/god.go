package god

import (
	"strings"

	"github.com/crawl/go-sequell/crawl/data"
)

// A Normalizer normalizes god names.
type Normalizer map[string]string

// Normalize normalizes the god name
func (n Normalizer) Normalize(god string) (string, error) {
	if canonical, exists := n[strings.ToLower(god)]; exists {
		return canonical, nil
	}
	return god, nil
}

var godAliases = data.Crawl.StringMap("god-aliases")

// CanonicalGod returns the canonical name for god.
func CanonicalGod(god string) string {
	if canonical, exists := godAliases[strings.ToLower(god)]; exists {
		return canonical
	}
	return god
}
