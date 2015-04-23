package loader

import (
	"fmt"
	"testing"

	"github.com/crawl/go-sequell/crawl/data"
	cdb "github.com/crawl/go-sequell/crawl/db"
	"github.com/crawl/go-sequell/pg"
	"github.com/crawl/go-sequell/xlog"
)

var testSchema = cdb.MustLoadSchema(data.CrawlData())

func testConn() pg.DB {
	db, err := pg.OpenDBUser("sequell_test", "sequell", "sequell")
	if err != nil {
		panic(err)
	}
	return db
}

func createLookup() *TableLookup {
	return NewTableLookup(testSchema.LookupTable("sk"), 3)
}

func TestLookupCI(t *testing.T) {
	var tests = []struct {
		table         string
		caseSensitive bool
	}{
		{"killer", true},
		{"sk", false},
	}
	for _, test := range tests {
		l := NewTableLookup(testSchema.LookupTable(test.table), 3)
		actual := l.CaseSensitive
		if actual != test.caseSensitive {
			t.Errorf("Expected lookup case-sensitive=%t, got %t",
				test.table, test.caseSensitive, actual)
		}
	}
}

func TestTableLookup(t *testing.T) {
	DB := testConn()
	lookup := createLookup()
	file := "cszo-git.log"
	reader := xlog.Reader(file, file)

	purgeTables(DB)
	rows := []xlog.Xlog{}
	for i := 0; i < 3; i++ {
		testXlog, err := reader.Next()
		if err != nil {
			t.Errorf("Error reading %s: %s", file, err)
			return
		}
		rows = append(rows, testXlog)
		lookup.Add(testXlog)
	}
	tx, err := DB.Begin()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
		return
	}
	if err := lookup.ResolveAll(tx); err != nil {
		t.Errorf("Error resolving lookups: %s", err)
		tx.Rollback()
		return
	}
	if err := tx.Commit(); err != nil {
		t.Errorf("Error committing tx: %s", err)
		return
	}

	for _, row := range rows {
		if err := lookup.SetIds(row); err != nil {
			t.Errorf("SetIds(%#v) failed: %s", row, err)
		}
		if row["sk_id"] == "" {
			t.Errorf("SetIds(%#v) did not set sk_id", row)
		}
		fmt.Printf("Resolved row: %s\n", row)
	}
}
