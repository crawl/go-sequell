package schema

import (
	"io"
	"os"
	"strings"
)

// Select specifies what database DDL statements are selected.
type Select int

const (
	// SelTables requests only table created statements
	SelTables Select = iota

	// SelIndexesConstraints requests indexes and constraints on tables
	SelIndexesConstraints

	// SelTablesIndexesConstraints requests tables, indexes and constraints
	SelTablesIndexesConstraints

	// SelDropIndexesConstraints requests drop statements for indexes and
	// constraints. This is usually used to drop indexes and constraints for a
	// bulk data load.
	SelDropIndexesConstraints
)

// SQLCombine combines multiple SQL statements into one string, separating
// SQL statements with the ";" separator.
func SQLCombine(sqls []string) string {
	return strings.Join(sqls, ";\n") + ";\n"
}

// Write writes the SQL statements for schema, selected by sel to the writer.
func (s *Schema) Write(sel Select, writer io.Writer) (int, error) {
	return writer.Write([]byte(SQLCombine(s.SQLSel(sel))))
}

// WriteFile writes the SQL statements for schema, selected to the writer to
// the file named by filename, which will be created if possible.
func (s *Schema) WriteFile(sel Select, filename string) (int, error) {
	file, err := os.Create(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return s.Write(sel, file)
}
