package isync

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/greensnark/go-sequell/crawl/data"
	"github.com/greensnark/go-sequell/crawl/db"
	"github.com/greensnark/go-sequell/fnotify"
	"github.com/greensnark/go-sequell/loader"
	"github.com/greensnark/go-sequell/logfetch"
	"github.com/greensnark/go-sequell/pg"
	"github.com/greensnark/go-sequell/qyaml"
	"github.com/greensnark/go-sequell/resource"
	"github.com/greensnark/go-sequell/sources"
	"gopkg.in/fsnotify.v1"
)

const CommandFetch = "fetch"
const CommandCommit = "commit"
const CommandExit = "exit"

var ErrExit = errors.New("exit")

type Sync struct {
	ConnSpec  pg.ConnSpec
	DB        pg.DB
	CacheDir  string
	Loader    *loader.Loader
	Servers   *sources.Servers
	Schema    *db.CrawlSchema
	CrawlData qyaml.Yaml
	Fetcher   *logfetch.Fetcher

	logFileWatcher     *fnotify.Notifier
	configWatcher      *fnotify.Notifier
	slaveWaitGroup     sync.WaitGroup
	masterWaitGroup    sync.WaitGroup
	fetchRequests      chan bool
	changedLogFiles    chan string
	changedConfigFiles chan string
}

func New(c pg.ConnSpec, cachedir string) (*Sync, error) {
	DB, err := c.Open()
	if err != nil {
		return nil, err
	}
	l := &Sync{
		ConnSpec:           c,
		DB:                 DB,
		CacheDir:           cachedir,
		CrawlData:          data.Crawl,
		Fetcher:            logfetch.New(),
		changedLogFiles:    make(chan string),
		changedConfigFiles: make(chan string),
		fetchRequests:      make(chan bool, 1),
	}
	if err = l.init(); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *Sync) init() error {
	if err := l.setServers(); err != nil {
		return err
	}
	return l.setSchema()
}

func (l *Sync) setServers() error {
	servers, err := sources.Sources(data.Sources(), l.CacheDir)
	if err != nil {
		return err
	}
	l.Servers = servers
	return nil
}

func (l *Sync) setSchema() error {
	schema, err := db.LoadSchema(l.CrawlData)
	if err != nil {
		return err
	}
	l.Schema = schema
	return nil
}

func (l *Sync) gameTypePrefixes() map[string]string {
	return l.CrawlData.StringMap("game-type-prefixes")
}

func (l *Sync) newLoader() *loader.Loader {
	return loader.New(l.DB, l.Servers, l.Schema, l.gameTypePrefixes())
}

// Run monitors stdin for commands.
func (l *Sync) Run() error {
	l.startBackgroundTasks()
	reader := bufio.NewReader(os.Stdin)

	normalShutdown := func() error {
		fmt.Println("Cleaning up...")
		l.Shutdown()
		fmt.Println("Exiting.")
		return nil
	}
	for {
		line, err := reader.ReadString('\n')
		if line != "" {
			if err := l.runCommand(strings.TrimSpace(line)); err != nil {
				if err == ErrExit {
					return normalShutdown()
				}
				return err
			}
		}
		if err != nil {
			if err == io.EOF {
				return normalShutdown()
			}
			return err
		}
	}
}

func (l *Sync) Shutdown() {
	l.stopAllTasks()
}

func (l *Sync) runCommand(cmd string) error {
	switch strings.ToLower(cmd) {
	case "fetch":
		select {
		case l.fetchRequests <- true:
			log.Println("Fetch requested")
		default:
		}
	case "exit":
		return ErrExit
	default:
		log.Println("Unknown command: try \"fetch\" or \"exit\"")
	}
	return nil
}

func (l *Sync) startBackgroundTasks() {
	l.startMasterTasks()
	l.startSlaveTasks()
}

func (l *Sync) startMasterTasks() {
	l.masterWaitGroup.Add(2)
	l.monitorConfigs()
	go l.reloadConfigs()
}

func (l *Sync) startSlaveTasks() {
	// The three main goroutines that need to be restarted when a
	// config changes:
	l.slaveWaitGroup.Add(3)
	l.monitorLogs()
	go l.readLogs()
	go l.monitorFetchRequests()
}

func (l *Sync) monitorFetchRequests() {
	for fetch := range l.fetchRequests {
		if !fetch {
			break
		}
		l.Fetcher.Download(l.Servers, true)
	}
	log.Println("fetch request monitor exiting")
	l.slaveWaitGroup.Done()
}

func (l *Sync) stopAllTasks() {
	l.stopSlaveTasks()
	l.stopMasterTasks()
}

func (l *Sync) stopMasterTasks() {
	l.configWatcher.Close()
	l.changedConfigFiles <- ""
	l.masterWaitGroup.Wait()
}

func (l *Sync) stopSlaveTasks() {
	// Sentinel to shut down the log reader and fetch monitor:
	l.changedLogFiles <- ""
	l.fetchRequests <- false
	l.logFileWatcher.Close()
	l.slaveWaitGroup.Wait()
}

func (l *Sync) reloadConfigs() {
	for cfg := range l.changedConfigFiles {
		if cfg == "" {
			break
		}
		log.Printf("Config %s changed, reloading\n", cfg)
		l.stopSlaveTasks()
		l.CrawlData = data.CrawlData()
		data.Crawl = l.CrawlData
		l.setServers()
		l.setSchema()
		l.startSlaveTasks()
	}
	log.Println("config reload monitor exiting")
	l.masterWaitGroup.Done()
}

func (l *Sync) readLogs() {
	log.Println("Loading logs into", l.ConnSpec.Database)
	l.Loader = l.newLoader()
	if err := l.Loader.LoadCommit(); err != nil {
		log.Println("Error preloading logs:", err)
	}

	for file := range l.changedLogFiles {
		// Sentinel:
		if file == "" {
			break
		}
		log.Println("Reading changed log", file)
		if err := l.Loader.LoadCommitLog(file); err != nil {
			log.Printf("Error reading changed log %s: %s\n", file, err)
		}
	}
	log.Println("log loader exiting")
	l.slaveWaitGroup.Done()
}

func (l *Sync) monitorConfigs() {
	configs := []string{
		resource.Root.Path("config/sources.yml"),
		resource.Root.Path("config/crawl-data.yml"),
	}
	l.configWatcher = fnotify.New("config")
	l.configWatcher.Debounce = time.Millisecond * 5000
	go func() {
		l.configWatcher.Notify(configs, l.changedConfigFiles)
		log.Println("config monitor exiting")
		l.masterWaitGroup.Done()
	}()
}

func (l *Sync) monitorLogs() {
	l.logFileWatcher = fnotify.New("logs")
	go func() {
		l.logFileWatcher.Notify([]string{l.CacheDir}, l.changedLogFiles)
		log.Println("log monitor exiting")
		l.slaveWaitGroup.Done()
	}()
}

func (l *Sync) monitorFiles(name string, files []string, res chan<- string, waitGroup *sync.WaitGroup) *fsnotify.Watcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if err := watcher.Add(f); err != nil {
			panic(err)
		}
	}
	go func() {
		pendingChanges := map[string]bool{}

		throttleDelay := 250 * time.Millisecond
		throttler := time.NewTimer(throttleDelay)
		throttleChan := func() <-chan time.Time {
			if len(pendingChanges) == 0 {
				return nil
			}
			return throttler.C
		}

	selectLoop:
		for {
			select {
			case event := <-watcher.Events:
				if event.Name == "" {
					break selectLoop
				}
				if event.Op&(fsnotify.Create|fsnotify.Write) != 0 {
					pendingChanges[event.Name] = true
				}
				throttler.Reset(throttleDelay)
			case <-throttleChan():
				for file := range pendingChanges {
					delete(pendingChanges, file)
					res <- file
				}
			case err := <-watcher.Errors:
				if err != nil {
					log.Println("watcher", name, "error:", err)
					break
				}
				break selectLoop
			}
		}
		throttler.Stop()
		if waitGroup != nil {
			log.Println("watcher", name, "exiting")
			waitGroup.Done()
		}
	}()
	return watcher
}
