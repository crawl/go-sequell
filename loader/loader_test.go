package loader

import (
	"testing"

	"github.com/greensnark/go-sequell/crawl/data"
	"github.com/greensnark/go-sequell/crawl/xlogtools"
	"github.com/greensnark/go-sequell/pg"
	"github.com/greensnark/go-sequell/sources"
	"github.com/greensnark/go-sequell/xlog"
)

func createLoader() (*Loader, error) {
	srv, err := sources.Sources(data.Sources(), "server-xlogs")
	if err != nil {
		return nil, err
	}
	return New(testConn(), srv, testSchema,
		data.Crawl.StringMap("game-type-prefixes")), nil
}

func TestTableName(t *testing.T) {
	ldr, err := createLoader()
	if err != nil {
		t.Errorf("Error creating loader: %s\n", err)
	}

	var tests = []struct {
		game          string
		logtype       xlogtools.XlogType
		expectedTable string
	}{
		{"", xlogtools.Log, "logrecord"},
		{"sprint", xlogtools.Milestone, "spr_milestone"},
		{"zotdef", xlogtools.Log, "zot_logrecord"},
		{"nostalgia", xlogtools.Log, "nostalgia_logrecord"},
	}

	for _, test := range tests {
		actual := ldr.TableName(&sources.XlogSrc{
			Game: test.game,
			Type: test.logtype,
		})
		if actual != test.expectedTable {
			t.Errorf("ldr.TableName({Game:%s,Type:%s}) = %#v, expected %#v",
				test.game, test.logtype, actual, test.expectedTable)
		}
	}
}

func createSingleFileLoader(file string) (*Loader, error) {
	ldr, err := createLoader()
	if err != nil {
		return nil, err
	}
	ldr.init()

	src := &sources.XlogSrc{
		Server: &sources.Server{
			Name: "cszo",
		},
		Name:        file,
		Live:        true,
		Game:        "",
		GameVersion: "git",
		Type:        xlogtools.Log,
		TargetPath:  file,
	}
	ldr.Readers = []Reader{
		Reader{
			XlogReader: xlog.Reader(src.TargetPath),
			XlogSrc:    src,
			Table:      ldr.TableName(src),
		},
	}
	return ldr, nil
}

func purgeTables(db pg.DB) error {
	sch := testSchema.Schema().Sort()
	for i := len(sch.Tables) - 1; i >= 0; i-- {
		t := sch.Tables[i]
		if _, err := db.Exec("delete from " + t.Name); err != nil {
			return err
		}
	}
	return nil
}

func TestLoader(t *testing.T) {
	ldr, err := createSingleFileLoader("cszo-git.log")
	if err != nil {
		t.Errorf("Error creating loader: %s", err)
		return
	}

	if err := purgeTables(ldr.DB); err != nil {
		t.Errorf("Error purging tables: %s", err)
	}

	if err := ldr.LoadCommit(); err != nil {
		t.Errorf("Error loading logs: %s\n", err)
	}
}
