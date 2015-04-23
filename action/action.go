package action

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/crawl/go-sequell/crawl/data"
	"github.com/crawl/go-sequell/flock"
	"github.com/crawl/go-sequell/isync"
	"github.com/crawl/go-sequell/logfetch"
	"github.com/crawl/go-sequell/pg"
	"github.com/crawl/go-sequell/resource"
	"github.com/crawl/go-sequell/sources"
)

var Root = resource.Root
var LogCache = Root.Path("server-xlogs")

var DBLock = flock.New(Root.Path(".seq.db.lock"))
var FetchLock = flock.New(Root.Path(".seq.fetch.lock"))

func logProc(act func(*sources.Servers) error) error {
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
	return act(src)
}

func LinkLogs() error {
	return logProc(func(src *sources.Servers) error {
		for _, xl := range src.XlogSources() {
			oldPath := filepath.Join(LogCache, xl.CName)
			if fi, err := os.Stat(oldPath); err == nil && fi.Mode().IsRegular() {
				// If target exists, skip.
				if _, err := os.Stat(xl.TargetPath); err == nil {
					fmt.Fprintln(os.Stderr, "Skipping target", xl.TargetPath)
					continue
				}

				if err = xl.MkdirTarget(); err != nil {
					return err
				}

				if err = os.Link(oldPath, xl.TargetPath); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

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
