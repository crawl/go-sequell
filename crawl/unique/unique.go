package unique

import (
	"github.com/greensnark/go-sequell/crawl/data"
	"github.com/greensnark/go-sequell/xlog"
	"regexp"
	"strings"
	"sync"
)

var dataLock = &sync.Mutex{}
var uniqueMapData map[string]bool
var orcMapData map[string]bool

var reArticlePrefix = regexp.MustCompile(`^(?:an?|the) `)

func uniqueMap() map[string]bool {
	dataLock.Lock()
	defer dataLock.Unlock()

	if uniqueMapData == nil {
		uniqueMapData = data.StringSliceSet(data.Uniques())
	}
	return uniqueMapData
}

func orcMap() map[string]bool {
	dataLock.Lock()
	defer dataLock.Unlock()

	if orcMapData == nil {
		orcMapData = data.StringSliceSet(data.Orcs())
	}
	return orcMapData
}

func GenericPanLordName() string {
	return data.String("generic_panlord")
}

func IsUnique(name string, rec xlog.Xlog) bool {
	if strings.Index(strings.ToLower(rec["killer_flags"]), "unique") != -1 {
		return true
	}
	return uniqueMap()[name]
}

func IsOrc(name string) bool {
	return orcMap()[name]
}

func MaybePanLord(name string, rec xlog.Xlog) bool {
	return reArticlePrefix.FindStringIndex(name) == nil && !IsUnique(name, rec) && !IsOrc(name)
}
