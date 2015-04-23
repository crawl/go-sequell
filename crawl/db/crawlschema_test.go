package db

import (
	"fmt"
	"testing"

	"github.com/crawl/go-sequell/crawl/data"
	"github.com/crawl/go-sequell/schema"
)

func TestLoadSchema(t *testing.T) {
	cschema, err := LoadSchema(data.Schema)
	if err != nil {
		t.Errorf("Error loading schema: %v\n", err)
		return
	}
	size, err := cschema.Schema().Sort().WriteFile(schema.SelTablesIndexesConstraints, "test.sql")
	if err != nil {
		t.Errorf("Error saving test.sql: %v\n", err)
		return
	}
	fmt.Println("Wrote", size, "bytes to test.sql")
}
