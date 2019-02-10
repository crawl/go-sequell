package loader

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/crawl/go-sequell/crawl/data"
	cdb "github.com/crawl/go-sequell/crawl/db"
	"github.com/crawl/go-sequell/pg"
	"github.com/crawl/go-sequell/xlog"
	"github.com/pkg/errors"
)

var testSchema = cdb.MustLoadSchema(data.CrawlData().YAML)

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
			t.Errorf("Expected lookup (%s) case-sensitive=%t, got %t",
				test.table, test.caseSensitive, actual)
		}
	}
}

func cleanTestDB() pg.DB {
	db := testConn()
	purgeTables(db)
	return db
}

func testInTransaction(testUsingDB func(tx *sql.Tx)) {
	db := cleanTestDB()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		panic("unable to start transaction on test DB")
	}

	defer tx.Rollback()
	testUsingDB(tx)
}

func TestDuplicateXlogRejection(t *testing.T) {
	testInTransaction(func(tx *sql.Tx) {
		hashLookup := NewTableLookup(testSchema.LookupTable("hash"), 4)
		if !hashLookup.GloballyUnique() {
			t.Errorf("hash lookup should be UUID, but isn't")
			return
		}

		// The second iteration should declare everything duplicated,
		// because all the values are loaded into the ID cache.
		for i := 0; i < 2; i++ {
			err := hashLookup.ResolveAll(tx, []xlog.Xlog{
				{"hash": "abc"},
				{"hash": "def"},
				{"hash": "abc"},
				{"hash": "xyz"},
			})

			if err != nil {
				t.Errorf("hash resolve failed: %s", err)
				return
			}

			for i, testCase := range []struct {
				hash           string
				rejectedAsDupe bool
			}{
				{hash: "abc", rejectedAsDupe: i == 1},
				{hash: "def", rejectedAsDupe: i == 1},
				{hash: "abc", rejectedAsDupe: true},
				{hash: "xyz", rejectedAsDupe: i == 1},
			} {
				t.Run(fmt.Sprintf("%s.%d", testCase.hash, i), func(t *testing.T) {
					_, err := hashLookup.ID(testCase.hash)
					rejectedAsDupe := errors.Cause(err) == ErrDuplicateRow
					if rejectedAsDupe != testCase.rejectedAsDupe {
						t.Errorf("hashlookup.ID(%#v) == dupe:%t, want dupe:%t", testCase.hash, rejectedAsDupe, testCase.rejectedAsDupe)
					}
				})
			}
		}
	})
}

func TestTableLookup(t *testing.T) {
	DB := testConn()
	defer DB.Close()

	purgeTables(DB)

	lookup := createLookup()
	file := "cszo-git.log"
	reader := xlog.NewReader("cszo", file, file)

	rows := []xlog.Xlog{}
	for i := 0; i < 3; i++ {
		testXlog, err := reader.Next()
		if err != nil {
			t.Errorf("Error reading %s: %s", file, err)
			return
		}
		rows = append(rows, testXlog)
	}
	tx, err := DB.Begin()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
		return
	}
	if err := lookup.ResolveAll(tx, rows); err != nil {
		t.Errorf("Error resolving lookups: %s", err)
		tx.Rollback()
		return
	}
	if err := tx.Commit(); err != nil {
		t.Errorf("Error committing tx: %s", err)
		return
	}

	for _, row := range rows {
		if err := lookup.SetIDs(row); err != nil {
			t.Errorf("SetIds(%#v) failed: %s", row, err)
		}
		if row["sk_id"] == "" {
			t.Errorf("SetIds(%#v) did not set sk_id", row)
		}
		fmt.Printf("Resolved row: %s\n", row)
	}
}
