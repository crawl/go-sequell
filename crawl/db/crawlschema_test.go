package db

import (
	"fmt"
	"testing"

	"github.com/greensnark/go-sequell/crawl/data"
	"github.com/greensnark/go-sequell/schema"
)

func TestLoadSchema(t *testing.T) {
	cschema, err := LoadSchema(data.Schema)
	if err != nil {
		t.Errorf("Error loading schema: %v\n", err)
		return
	}
	size, err := cschema.Schema().WriteFile(schema.SelTablesIndexes, "test.sql")
	if err != nil {
		t.Errorf("Error saving test.sql: %v\n", err)
		return
	}
	fmt.Println("Wrote", size, "bytes to test.sql")
}
