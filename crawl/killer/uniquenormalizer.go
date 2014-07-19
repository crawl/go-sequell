package killer

import (
	"github.com/greensnark/go-sequell/crawl/unique"
	"github.com/greensnark/go-sequell/grammar"
	"github.com/greensnark/go-sequell/xlog"
	"regexp"
)

type uniqueNormalizer struct{}

var reStartsWithUppercase = regexp.MustCompile(`^\p{Lu}`)
var reProperName = regexp.MustCompile(`^(\p{Lu}[\p{L}\p{N}']*(?: \p{Lu}[\p{L}\p{N}']+)*)`)

var reNameWithTitle = regexp.MustCompile(` the (.*)$`)

func (u uniqueNormalizer) NormalizeKiller(killer string, xlog xlog.Xlog) (string, error) {
	killerName := xlog["killer"]
	if killerName != "" && reStartsWithUppercase.FindString(killerName) != "" {
		properName := reProperName.FindString(killer)
		titleMatch := reNameWithTitle.FindStringSubmatch(killer)
		if titleMatch != nil {
			if unique.IsUnique(properName, xlog) {
				killer = properName
			} else {
				// Orcs
				killer = grammar.Article(titleMatch[1])
			}
		} else {
			if unique.MaybePanLord(properName, xlog) {
				return unique.GenericPanLordName(), nil
			}
		}
	}
	return killer, nil
}
