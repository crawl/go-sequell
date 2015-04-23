package pg

import (
	"fmt"
	"testing"

	"github.com/crawl/go-sequell/schema"
)

func TestOpenDBUser(t *testing.T) {
	db, err := OpenDBUser("henzell", "henzell", "henzell")
	if err != nil {
		t.Errorf("Database connection failed: %s\n", err)
	}

	count := 0
	row := db.QueryRow("select count(*) from logrecord")
	err = row.Scan(&count)
	if err != nil {
		t.Errorf("Error reading logrecord count: %s\n", err)
	}
	fmt.Printf("logrecord count = %d\n", count)

	fmt.Printf("Introspecting schema\n")
	sch, err := db.IntrospectSchema()
	if err != nil {
		t.Errorf("Error introspecting schema: %s\n", err)
		return
	}

	sch.Sort()
	_, err = sch.WriteFile(schema.SelTablesIndexesConstraints, "introspect.sql")
	if err != nil {
		t.Errorf("Error writing schema to introspect.sql: %s\n", err)
	}
	fmt.Printf("Wrote schema to introspect.sql\n")
}
