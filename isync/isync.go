package isync

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/greensnark/go-sequell/crawl/data"
	"github.com/greensnark/go-sequell/crawl/db"
	"github.com/greensnark/go-sequell/loader"
	"github.com/greensnark/go-sequell/pg"
	"github.com/greensnark/go-sequell/qyaml"
	"github.com/greensnark/go-sequell/sources"
)

const CommandFetch = "fetch"
const CommandCommit = "commit"
const CommandExit = "exit"

var ErrExit = errors.New("exit")

type Loader struct {
	ConnSpec  pg.ConnSpec
	DB        pg.DB
	CacheDir  string
	Loader    *loader.Loader
	Servers   *sources.Servers
	Schema    *db.CrawlSchema
	CrawlData qyaml.Yaml

	changedFiles chan string
}

func New(c pg.ConnSpec, cachedir string) (*Loader, error) {
	DB, err := c.Open()
	if err != nil {
		return nil, err
	}
	l := &Loader{
		ConnSpec:     c,
		DB:           DB,
		CacheDir:     cachedir,
		CrawlData:    data.Crawl,
		changedFiles: make(chan string),
	}
	if err = l.init(); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *Loader) init() error {
	if err := l.setServers(); err != nil {
		return err
	}
	return l.setSchema()
}

func (l *Loader) setServers() error {
	servers, err := sources.Sources(data.Sources(), l.CacheDir)
	if err != nil {
		return err
	}
	l.Servers = servers
	return nil
}

func (l *Loader) setSchema() error {
	schema, err := db.LoadSchema(data.CrawlData())
	if err != nil {
		return err
	}
	l.Schema = schema
	return nil
}

func (l *Loader) gameTypePrefixes() map[string]string {
	return l.CrawlData.StringMap("game-type-prefixes")
}

func (l *Loader) newLoader() *loader.Loader {
	return loader.New(l.DB, l.Servers, l.Schema, l.gameTypePrefixes())
}

// Run monitors stdin for commands.
func (l *Loader) Run() error {
	go l.runLoader()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("isync:")

	normalShutdown := func() error {
		fmt.Println("Exiting...")
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

func (l *Loader) runCommand(cmd string) error {
	switch strings.ToLower(cmd) {
	case "fetch":
		//l.Fetcher.TriggerFetch()
	case "exit":
		return ErrExit
	}
	return nil
}

func (l *Loader) runLoader() {
	log.Println("Loading logs into", l.ConnSpec.Database)
	l.Loader = l.newLoader()
	l.Loader.LoadCommit()

	// for file := range l.changedFiles {
	// 	l.Loader.LoadCommitLog(file)
	// }

	log.Println("loader exiting")
}
