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

func DownloadLogs(incremental bool) error {
	src, err := sources.Sources(data.Sources(), LogCache)
	if err != nil {
		return err
	}
	err = os.MkdirAll(LogCache, os.ModePerm)
	if err != nil {
		return err
	}
	logfetch.New().DownloadAndWait(src, incremental)
	return nil
}

func Isync(db pg.ConnSpec) error {
	if err := os.MkdirAll(LogCache, os.ModePerm); err != nil {
		return err
	}

	lock := flock.New(Root.Path("isync.lock"))
	if err := lock.Lock(false); err != nil {
		return err
	}
	defer lock.Unlock()

	sync, err := isync.New(db, LogCache)
	if err != nil {
		return err
	}
	return sync.Run()
}
