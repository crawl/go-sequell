package killer

import (
	"github.com/greensnark/go-sequell/crawl/unique"
	"github.com/greensnark/go-sequell/grammar"
	"regexp"
)

type uniqueNormalizer struct{}

var reStartsWithUppercase = regexp.MustCompile(`^\p{Lu}`)
var reProperName = regexp.MustCompile(`^(\p{Lu}[\p{L}\p{N}']*(?: \p{Lu}[\p{L}\p{N}']+)*)`)

var reNameWithTitle = regexp.MustCompile(` the (.*)$`)

func (u uniqueNormalizer) NormalizeKiller(killer, killerName string) (string, error) {
	if killerName != "" && reStartsWithUppercase.FindString(killerName) != nil {
		properName := reProperName.FindString(killer)
		titleMatch := reNameWithTitle.FindStringSubmatch(killer)
		if titleMatch != nil {
			if unique.IsUnique(properName, rec) {
				killer = properName
			} else {
				// Orcs
				killer = grammar.Article(titleMatch[1])
			}
		} else {
			if unique.MaybePanLord(properName, rec) {
				return unique.GenericPanLordName
			}
		}
	}
}
