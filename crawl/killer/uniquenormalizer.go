package killer

import (
	"regexp"

	"github.com/crawl/go-sequell/crawl/data"
	"github.com/crawl/go-sequell/crawl/unique"
	"github.com/crawl/go-sequell/grammar"
)

type uniqueNormalizer struct {
	c data.Crawl
	u *unique.Uniq
}

var reStartsWithUppercase = regexp.MustCompile(`^\p{Lu}`)
var reProperName = regexp.MustCompile(`^(\p{Lu}[\p{L}\p{N}']*(?: \p{Lu}[\p{L}\p{N}']+)*)`)

var reNameWithTitle = regexp.MustCompile(` the (.*)$`)

func (u *uniqueNormalizer) NormalizeKiller(cv, killer, killerName, killerFlags string) (string, error) {
	// Reject ghosts and illusions:
	if killer == "a player ghost" || killer == "a player illusion" {
		return killer, nil
	}
	if killerName != "" && reStartsWithUppercase.FindString(killerName) != "" {
		properName := reProperName.FindString(killer)
		titleMatch := reNameWithTitle.FindStringSubmatch(killer)
		if titleMatch != nil {
			if u.u.IsUnique(properName, killerFlags) {
				killer = properName
			} else {
				// Orcs
				killer = grammar.Article(titleMatch[1])
			}
		} else {
			if u.u.MaybePanLord(cv, properName, killerFlags) {
				return u.u.GenericPanLordName(), nil
			}
		}
	}
	return killer, nil
}
