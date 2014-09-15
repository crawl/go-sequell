package action

import (
	"os"

	"github.com/greensnark/go-sequell/crawl/data"
	"github.com/greensnark/go-sequell/flock"
	"github.com/greensnark/go-sequell/isync"
	"github.com/greensnark/go-sequell/logfetch"
	"github.com/greensnark/go-sequell/pg"
	"github.com/greensnark/go-sequell/resource"
	"github.com/greensnark/go-sequell/sources"
)

var Root = resource.Root
var LogCache = Root.Path("server-xlogs")

var DBLock = flock.New(Root.Path(".seq.db.lock"))
var FetchLock = flock.New(Root.Path(".seq.fetch.lock"))

func DownloadLogs(incremental bool) error {
	src, err := sources.Sources(data.Sources(), LogCache)
	if err != nil {
		return err
	}
	err = os.MkdirAll(LogCache, os.ModePerm)
	if err != nil {
		return err
	}
	if err := FetchLock.Lock(false); err != nil {
		return err
	}
	defer FetchLock.Unlock()

	logfetch.New().DownloadAndWait(src, incremental)
	return nil
}

func Isync(db pg.ConnSpec) error {
	if err := os.MkdirAll(LogCache, os.ModePerm); err != nil {
		return err
	}

	if err := DBLock.Lock(false); err != nil {
		return err
	}
	defer DBLock.Unlock()
	if err := FetchLock.Lock(false); err != nil {
		return err
	}
	defer FetchLock.Unlock()

	sync, err := isync.New(db, LogCache)
	if err != nil {
		return err
	}
	return sync.Run()
}
