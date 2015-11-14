package unique

import (
	"regexp"
	"strings"

	"github.com/crawl/go-sequell/conv"
	"github.com/crawl/go-sequell/crawl/data"
	"github.com/crawl/go-sequell/crawl/version"
)

// Uniq normalizes and classifies Crawl unique monster names
type Uniq struct {
	c       data.Crawl
	uniques map[string]bool
	orcs    map[string]bool
}

// New creates a new Crawl unique normalizer
func New(c data.Crawl) *Uniq {
	return &Uniq{
		c:       c,
		uniques: conv.StringSliceSet(c.Uniques()),
		orcs:    conv.StringSliceSet(c.Orcs()),
	}
}

var reArticlePrefix = regexp.MustCompile(`^(?:an?|the) `)

// GenericPanLordName gets the default name for generic pandemonium lords.
func (u *Uniq) GenericPanLordName() string {
	return u.c.String("generic_panlord")
}

// IsUnique returns true if name refers to a unique monster; IsUnique will
// return true always when killerFlags includes "unique".
func (u *Uniq) IsUnique(name string, killerFlags string) bool {
	if strings.Index(strings.ToLower(killerFlags), "unique") != -1 {
		return true
	}
	return u.uniques[name]
}

// IsOrc returns true if name refers to a named orc.
func (u *Uniq) IsOrc(name string) bool {
	return u.orcs[name]
}

var panLordSuffix = regexp.MustCompile(`the pandemonium lord`)
var panLordSuffixVersion = version.NumericID("0.11")

// MaybePanLord returns true if name looks like a pandemonium lord's name,
// using cv (Crawl canonical version) and killerFlags to improve its guesses.
func (u *Uniq) MaybePanLord(cv, name, killerFlags string) bool {
	if version.CachingNumericID(cv) >= panLordSuffixVersion {
		return panLordSuffix.FindStringIndex(name) != nil
	}
	return reArticlePrefix.FindStringIndex(name) == nil && !u.IsUnique(name, killerFlags) && !u.IsOrc(name)
}
