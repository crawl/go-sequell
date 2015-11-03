package fnotify

import (
	"log"
	"time"

	"gopkg.in/fsnotify.v1"
)

type Notifier struct {
	name           string
	watcher        *fsnotify.Watcher
	Debounce       time.Duration
	RemonitorDelay time.Duration
	shutdown       chan bool
}

const DefaultDebounce = 250 * time.Millisecond
const DefaultRemonitorDelay = 500 * time.Millisecond

func New(name string) *Notifier {
	return &Notifier{
		name:           name,
		Debounce:       DefaultDebounce,
		RemonitorDelay: DefaultRemonitorDelay,
		shutdown:       make(chan bool, 1),
	}
}

func (n *Notifier) Close() {
	n.shutdown <- true
}

func (n *Notifier) Notify(files []string, res chan<- string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	for _, f := range files {
		if err := watcher.Add(f); err != nil {
			return err
		}
	}

	pendingChanges := map[string]bool{}
	filesToRemonitor := map[string]bool{}

	throttler := time.NewTimer(n.Debounce)
	remonitorTimer := time.NewTimer(n.RemonitorDelay)
	throttleChan := func() <-chan time.Time {
		if len(pendingChanges) == 0 {
			return nil
		}
		return throttler.C
	}
	remonitorChan := func() <-chan time.Time {
		if len(filesToRemonitor) == 0 {
			return nil
		}
		return remonitorTimer.C
	}

selectLoop:
	for {
		select {
		case event := <-watcher.Events:
			if event.Name == "" {
				break selectLoop
			}
			throttler.Reset(n.Debounce)
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename|fsnotify.Remove) != 0 {
				pendingChanges[event.Name] = true

				if (event.Op & fsnotify.Remove) != 0 {
					filesToRemonitor[event.Name] = true
					remonitorTimer.Reset(n.RemonitorDelay)
				}
			}
		case <-remonitorChan():
			for file := range filesToRemonitor {
				if err := watcher.Add(file); err != nil {
					log.Println("watcher", n.name, "cannot re-monitor", file, err)
				}
				delete(filesToRemonitor, file)
			}
		case <-throttleChan():
			for file := range pendingChanges {
				delete(pendingChanges, file)
				log.Println("file changed:", file)
				res <- file
			}
		case err := <-watcher.Errors:
			if err != nil {
				log.Println("watcher", n.name, "error:", err)
				break
			}
			break selectLoop
		case <-n.shutdown:
			break selectLoop
		}
	}
	throttler.Stop()
	return nil
}
