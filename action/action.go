package action

import (
	"fmt"
	"os"

	"github.com/greensnark/go-sequell/crawl/data"
	cdb "github.com/greensnark/go-sequell/crawl/db"
	"github.com/greensnark/go-sequell/logfetch"
	"github.com/greensnark/go-sequell/pg"
	"github.com/greensnark/go-sequell/schema"
	"github.com/greensnark/go-sequell/sources"
)

const LogCache = "server-xlogs"

func DownloadLogs(incremental bool) error {
	src, err := sources.Sources(data.Sources(), LogCache)
	if err != nil {
		return err
	}
	err = os.MkdirAll(LogCache, os.ModePerm)
	if err != nil {
		return err
	}
	return logfetch.Download(src, incremental)
}

func CrawlSchema() *schema.Schema {
	schema, err := cdb.LoadSchema(data.CrawlData())
	if err != nil {
		panic(err)
	}
	return schema.Schema()
}

func PrintSchema(skipIndexes, dropIndexes, createIndexes bool) {
	s := CrawlSchema()
	sel := schema.SelTablesIndexes
	if skipIndexes {
		sel = schema.SelTables
	}
	if dropIndexes {
		sel = schema.SelDropIndexes
	}
	if createIndexes {
		sel = schema.SelIndexes
	}
	s.Sort().Write(sel, os.Stdout)
}

func CheckDBSchema(dbspec pg.ConnSpec, applyDelta bool) error {
	db, err := dbspec.Open()
	if err != nil {
		return err
	}
	actualSchema, err := db.IntrospectSchema()
	if err != nil {
		return err
	}
	wantedSchema := CrawlSchema()
	diff := wantedSchema.DiffSchema(actualSchema)
	if len(diff.Tables) == 0 {
		fmt.Fprintf(os.Stderr, "Schema is up-to-date.\n")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Schema delta:\n")
	diff.PrintDelta(os.Stderr)
	if applyDelta {
		return nil
	}
	return nil
}
