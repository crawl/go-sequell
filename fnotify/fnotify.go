package fnotify

import (
	"log"
	"time"

	"gopkg.in/fsnotify.v1"
)

type Notifier struct {
	name     string
	watcher  *fsnotify.Watcher
	Debounce time.Duration
	shutdown chan bool
}

const DefaultDebounce = 250 * time.Millisecond

func New(name string) *Notifier {
	return &Notifier{
		name:     name,
		Debounce: DefaultDebounce,
		shutdown: make(chan bool, 1),
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

	throttler := time.NewTimer(n.Debounce)
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
			throttler.Reset(n.Debounce)
		case <-throttleChan():
			for file := range pendingChanges {
				delete(pendingChanges, file)
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
