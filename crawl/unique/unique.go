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

func GenericPanLordName() string {
	return data.Crawl.String("generic_panlord")
}

func IsUnique(name string, killerFlags string) bool {
	if strings.Index(strings.ToLower(killerFlags), "unique") != -1 {
		return true
	}
	return uniqueMap()[name]
}

func IsOrc(name string) bool {
	return orcMap()[name]
}

var panLordSuffix = regexp.MustCompile(`the pandemonium lord`)
var panLordSuffixVersion = version.VersionNumericId("0.11")

func MaybePanLord(cv, name, killerFlags string) bool {
	if version.CachingVersionNumericId(cv) >= panLordSuffixVersion {
		return panLordSuffix.FindStringIndex(name) != nil
	}
	return reArticlePrefix.FindStringIndex(name) == nil && !IsUnique(name, killerFlags) && !IsOrc(name)
}
