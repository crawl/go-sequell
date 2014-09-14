package schema

import (
	"strings"
)

func (s *Schema) SqlSel(sel SchemaSelect) []string {
	switch sel {
	case SelTablesIndexesConstraints:
		return s.Sql()
	case SelTables:
		return s.SqlNoIndexesConstraints()
	case SelIndexesConstraints:
		return s.IndexConstraintSql()
	case SelDropIndexesConstraints:
		return s.DropIndexConstraintSql()
	}
	return nil
}

func (s *Schema) Sql() []string {
	return append(
		s.sqlTableRevMap((*Table).DropSql),
		s.sqlTableMap((*Table).Sql)...)
}

func (s *Schema) DropIndexConstraintSql() []string {
	return s.sqlTableMap((*Table).DropIndexConstraintSql)
}

func (s *Schema) IndexConstraintSql() []string {
	return s.sqlTableMap((*Table).IndexConstraintSql)
}

func (s *Schema) SqlNoIndexesConstraints() []string {
	return append(
		s.sqlTableRevMap((*Table).DropSql),
		s.sqlTableMap((*Table).SqlNoIndexesConstraints)...)
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
	return append(t.SqlNoIndexesConstraints(), t.IndexConstraintSql()...)
}

func (t *Table) DropSql() []string {
	return []string{"drop table if exists " + t.Name}
}

func (t *Table) SqlNoIndexesConstraints() []string {
	return []string{t.CreateTableSql()}
}

func (t *Table) IndexConstraintSql() []string {
	return t.CreateIndexConstraintSqls()
}

func (t *Table) DropIndexConstraintSql() []string {
	ncons := len(t.Constraints)
	sqls := make([]string, ncons+len(t.Indexes))
	for i, c := range t.Constraints {
		sqls[i] = "alter table " + t.Name + " drop " + c.Sql()
	}
	for i, index := range t.Indexes {
		sqls[i+ncons] = index.DropSql()
	}
	return sqls
}

func (t *Table) CreateTableSql() string {
	colsConstraints := t.CreateColumnClauses()
	return "create table " + t.Name +
		" (\n" + strings.Join(colsConstraints, ",\n") + "\n)"
}

func (t *Table) CreateIndexConstraintSqls() []string {
	nindex := len(t.Indexes)
	sqls := make([]string, nindex+len(t.Constraints))
	for i, index := range t.Indexes {
		sqls[i] = index.Sql()
	}
	for i, c := range t.Constraints {
		sqls[i+nindex] = "alter table " + t.Name + " add " + c.Sql()
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

func (c *Column) Sql() string {
	base := c.Name + " " + c.SqlType
	if c.Default != "" {
		base += " " + c.Default
	}
	return base
}

func (i *Index) Sql() string {
	createClause := "create index"
	if i.Unique {
		createClause = "create unique index"
	}
	return createClause + " " + i.Name + " on " + i.TableName +
		" (" + strings.Join(i.Columns, ", ") + ")"
}

func (i *Index) DropSql() string {
	return "drop index if exists " + i.Name
}
