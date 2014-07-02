package unique

import (
	"github.com/greensnark/go-sequell/crawl/data"
	"github.com/greensnark/go-sequell/xlog"
	"strings"
	"sync"
)

var dataLock = &sync.Mutex{}
var uniqueMapData map[string]bool

func uniqueMap() map[string]bool {
	dataLock.Lock()
	defer dataLock.Unlock()

	if uniqueMapData == nil {
		uniqueMapData = data.StringSliceSet(data.Uniques())
	}
	return uniqueMapData
}

func IsUnique(name string, rec xlog.Xlog) bool {
	if strings.Index(strings.ToLower(rec["killer_flags"]), "unique") != -1 {
		return true
	}
	return uniqueMap()[name]
}

func MaybePanLord(name string, rec xlog.Xlog) {
}
