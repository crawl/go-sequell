package schema

import (
	"io"
	"os"
	"strings"
)

type SchemaSelect int

const (
	SelTables        SchemaSelect = 1
	SelIndexes       SchemaSelect = 2
	SelTablesIndexes SchemaSelect = 4
)

func SqlCombine(sqls []string) string {
	return strings.Join(sqls, ";\n") + ";\n"
}

func (s *Schema) Write(sel SchemaSelect, writer io.Writer) (int, error) {
	return writer.Write([]byte(SqlCombine(s.SqlSel(sel))))
}

func (s *Schema) WriteFile(sel SchemaSelect, filename string) (int, error) {
	file, err := os.Create(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return s.Write(sel, file)
}
