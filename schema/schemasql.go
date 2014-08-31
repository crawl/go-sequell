package schema

import (
	"strings"
)

func (s *Schema) SqlSel(sel SchemaSelect) []string {
	switch sel {
	case SelTablesIndexes:
		return s.Sql()
	case SelTables:
		return s.SqlNoIndex()
	case SelIndexes:
		return s.IndexSql()
	}
	return nil
}

func (s *Schema) Sql() []string {
	return append(
		s.sqlTableRevMap((*Table).DropSql),
		s.sqlTableMap((*Table).Sql)...)
}

func (s *Schema) IndexSql() []string {
	return s.sqlTableMap((*Table).IndexSql)
}

func (s *Schema) SqlNoIndex() []string {
	return append(
		s.sqlTableRevMap((*Table).DropSql),
		s.sqlTableMap((*Table).SqlNoIndex)...)
}

func (s *Schema) sqlTableRevMap(tsql func(t *Table) []string) []string {
	res := []string{}
	for i := len(s.Tables) - 1; i >= 0; i-- {
		res = append(res, tsql(s.Tables[i])...)
	}
	return res
}

func (s *Schema) sqlTableMap(tsql func(t *Table) []string) []string {
	res := []string{}
	for _, table := range s.Tables {
		res = append(res, tsql(table)...)
	}
	return res
}

func (t *Table) Sql() []string {
	return append(t.SqlNoIndex(), t.IndexSql()...)
}

func (t *Table) DropSql() []string {
	return []string{"drop table if exists " + t.Name}
}

func (t *Table) SqlNoIndex() []string {
	return []string{t.CreateTableSql()}
}

func (t *Table) IndexSql() []string {
	return t.CreateIndexSqls()
}

func (t *Table) CreateTableSql() string {
	colsConstraints :=
		append(t.CreateColumnClauses(), t.CreateConstraintClauses()...)
	return "create table " + t.Name +
		" (\n" + strings.Join(colsConstraints, ",\n") + "\n)"
}

func (t *Table) CreateIndexSqls() []string {
	sqls := make([]string, len(t.Indexes))
	for i, index := range t.Indexes {
		sqls[i] = index.Sql()
	}
	return sqls
}

func (t *Table) CreateColumnClauses() []string {
	pieces := make([]string, len(t.Columns))
	for i, col := range t.Columns {
		pieces[i] = "  " + col.Sql()
	}
	return pieces
}

func (t *Table) CreateConstraintClauses() []string {
	constraints := make([]string, len(t.Constraints))
	for i, cons := range t.Constraints {
		constraints[i] = "  " + cons.Sql()
	}
	return constraints
}

func (c *Column) Sql() string {
	base := c.Name + " " + c.SqlType
	if c.Default != "" {
		base += " " + c.Default
	}
	return base
}

func (i *Index) Sql() string {
	return "create index " + i.Name + " on " + i.TableName +
		" (" + strings.Join(i.Columns, ", ") + ")"
}
