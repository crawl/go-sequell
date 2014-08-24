package god

import (
	"strings"

	"github.com/greensnark/go-sequell/crawl/data"
)

var godAliases = data.StringMap("god-aliases")

func CanonicalGod(god string) string {
	if canonical, exists := godAliases[strings.ToLower(god)]; exists {
		return canonical
	}
	return god
}
