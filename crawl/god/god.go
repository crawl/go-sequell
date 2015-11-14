package god

import "strings"

// A Normalizer normalizes god names.
type Normalizer map[string]string

// Normalize normalizes the god name
func (n Normalizer) Normalize(god string) (string, error) {
	if canonical, exists := n[strings.ToLower(god)]; exists {
		return canonical, nil
	}
	return god, nil
}
