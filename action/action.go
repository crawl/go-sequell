package action

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/crawl/go-sequell/crawl/data"
	"github.com/crawl/go-sequell/flock"
	"github.com/crawl/go-sequell/isync"
	"github.com/crawl/go-sequell/logfetch"
	"github.com/crawl/go-sequell/pg"
	"github.com/crawl/go-sequell/resource"
	"github.com/crawl/go-sequell/sources"
)

// Root is the root directory for Sequell's config
var Root = resource.Root

// LogCache is the cache directory that holds logfiles retrieved from remote
// servers.
var LogCache = Root.Path("server-xlogs")

// DBLock is the lock file used to guard write access to the DB.
var DBLock = flock.New(Root.Path(".seq.db.lock"))

// FetchLock is the lock file used to guard files downloaded to the cache.
var FetchLock = flock.New(Root.Path(".seq.fetch.lock"))

func logProc(act func(sources.Servers) error) error {
	src, err := sources.Sources(data.Sources(), data.CrawlData(), LogCache)
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

// LinkLogs links cached logs to their correct modern locations.
func LinkLogs() error {
	return logProc(func(src sources.Servers) error {
		for _, xl := range src.XlogSources() {
			oldPath := filepath.Join(LogCache, xl.CName)
			if fi, err := os.Stat(oldPath); err == nil && fi.Mode().IsRegular() {
				// If target exists, skip.
				if _, err := os.Stat(xl.TargetPath); err == nil {
					continue
				}

				if err = xl.MkdirTarget(); err != nil {
					return err
				}

				fmt.Fprintln(os.Stderr, "ln", oldPath, xl.TargetPath)
				if err = os.Link(oldPath, xl.TargetPath); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func xlogFilter(filters []string) func([]*sources.XlogSrc) []*sources.XlogSrc {
	accept := func(file *sources.XlogSrc) bool {
		fileDesc := file.String()
		for _, filter := range filters {
			if strings.Index(fileDesc, filter) != -1 {
				return true
			}
		}
		return false
	}
	return func(files []*sources.XlogSrc) []*sources.XlogSrc {
		if filters == nil || len(filters) == 0 {
			return files
		}
		res := make([]*sources.XlogSrc, 0, len(files))
		for _, file := range files {
			if accept(file) {
				res = append(res, file)
			}
		}
		return res
	}
}

// ShowSourceURLs shows all remote Xlog URLs, including Xlogs that are no
// longer live.
func ShowSourceURLs() error {
	src, err := sources.Sources(data.Sources(), data.CrawlData(), LogCache)
	if err != nil {
		return err
	}
	for _, xlog := range src.XlogSources() {
		fmt.Println(xlog.URL, xlog.TargetPath)
	}
	return nil
}

// DownloadLogs downloads all logfiles, possibly filtered to a subset. If
// incremental, ignores files that are no longer live.
func DownloadLogs(incremental bool, filters []string) error {
	src, err := sources.Sources(data.Sources(), data.CrawlData(), LogCache)
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

	logfetch.New().DownloadAndWait(xlogFilter(filters)(src.XlogSources()), incremental)
	return nil
}

// Isync runs the isync process that runs as a slave to Sequell and periodically
// downloads and loads logs.
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
