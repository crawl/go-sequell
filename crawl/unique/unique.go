package unique

import (
	"regexp"
	"strings"
	"sync"

	"github.com/crawl/go-sequell/conv"
	"github.com/crawl/go-sequell/crawl/data"
	"github.com/crawl/go-sequell/crawl/version"
)

var dataLock = &sync.Mutex{}
var uniqueMapData map[string]bool
var orcMapData map[string]bool

var reArticlePrefix = regexp.MustCompile(`^(?:an?|the) `)

func uniqueMap() map[string]bool {
	dataLock.Lock()
	defer dataLock.Unlock()

	if uniqueMapData == nil {
		uniqueMapData = conv.StringSliceSet(data.Uniques())
	}
	return uniqueMapData
}

func orcMap() map[string]bool {
	dataLock.Lock()
	defer dataLock.Unlock()

	if orcMapData == nil {
		orcMapData = conv.StringSliceSet(data.Orcs())
	}
	return orcMapData
}

// GenericPanLordName gets the default name for generic pandemonium lords.
func GenericPanLordName() string {
	return data.Crawl.String("generic_panlord")
}

// IsUnique returns true if name refers to a unique monster; IsUnique will
// return true always when killerFlags includes "unique".
func IsUnique(name string, killerFlags string) bool {
	if strings.Index(strings.ToLower(killerFlags), "unique") != -1 {
		return true
	}
	return uniqueMap()[name]
}

// IsOrc returns true if name refers to a named orc.
func IsOrc(name string) bool {
	return orcMap()[name]
}

var panLordSuffix = regexp.MustCompile(`the pandemonium lord`)
var panLordSuffixVersion = version.NumericID("0.11")

// MaybePanLord returns true if name looks like a pandemonium lord's name,
// using cv (Crawl canonical version) and killerFlags to improve its guesses.
func MaybePanLord(cv, name, killerFlags string) bool {
	if version.CachingNumericID(cv) >= panLordSuffixVersion {
		return panLordSuffix.FindStringIndex(name) != nil
	}
	return reArticlePrefix.FindStringIndex(name) == nil && !IsUnique(name, killerFlags) && !IsOrc(name)
}
