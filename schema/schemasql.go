package schema

import (
	"strings"
)

// SQLSel returns the SQL DDL statements selected by sel
func (s *Schema) SQLSel(sel Select) []string {
	switch sel {
	case SelTablesIndexesConstraints:
		return s.SQL()
	case SelTables:
		return s.SQLNoIndexesConstraints()
	case SelIndexesConstraints:
		return s.IndexConstraintSQL()
	case SelDropIndexesConstraints:
		return s.DropIndexConstraintSQL()
	}
	return nil
}

// SQL returns the list of DDL statements for this schema.
func (s *Schema) SQL() []string {
	return append(
		s.sqlTableRevMap((*Table).DropSQL),
		s.sqlTableMap((*Table).SQL)...)
}

// DropIndexConstraintSQL returns the DDL to drop indexes and constraints,
// usually to prepare for a bulk load.
func (s *Schema) DropIndexConstraintSQL() []string {
	return s.sqlTableMap((*Table).DropIndexConstraintSQL)
}

// IndexConstraintSQL returns the DDL for table indexes and constraints.
func (s *Schema) IndexConstraintSQL() []string {
	return s.sqlTableMap((*Table).IndexConstraintSQL)
}

// SQLNoIndexesConstraints returns the schema SQL without index and constraints.
func (s *Schema) SQLNoIndexesConstraints() []string {
	return append(
		s.sqlTableRevMap((*Table).DropSQL),
		s.sqlTableMap((*Table).SQLNoIndexesConstraints)...)
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

// SQL returns the DDL to create the table t, along with indexes and
// constraints.
func (t *Table) SQL() []string {
	return append(t.SQLNoIndexesConstraints(), t.IndexConstraintSQL()...)
}

// DropSQL returns the SQL to drop t
func (t *Table) DropSQL() []string {
	return []string{"drop table if exists " + t.Name}
}

// SQLNoIndexesConstraints returns the table's DDL SQL, ignoring indexes and
// constraints.
func (t *Table) SQLNoIndexesConstraints() []string {
	return append([]string{t.CreateTableSQL()}, t.CreateForceIndexSQL()...)
}

// CreateForceIndexSQL returns SQL statements for force-created indexes. Forced
// indexes are necessary for fast bulk-loads (usually indexes on lookup tables),
// and must be returned even when if the caller requested no regular indexes.
func (t *Table) CreateForceIndexSQL() []string {
	indexSQLs := []string{}
	for _, index := range t.Indexes {
		if index.Force {
			indexSQLs = append(indexSQLs, index.SQL())
		}
	}
	return indexSQLs
}

// IndexConstraintSQL returns SQL statements to create indexes and constraints
func (t *Table) IndexConstraintSQL() []string {
	return t.CreateIndexConstraintSQLs()
}

// DropIndexConstraintSQL returns the DDL to drop this table's indexes and
// constraints.
func (t *Table) DropIndexConstraintSQL() []string {
	sqls := []string{}
	for _, c := range t.Constraints {
		sqls = append(sqls, "alter table "+t.Name+" drop "+c.SQL())
	}
	for _, index := range t.Indexes {
		if !index.Force {
			sqls = append(sqls, index.DropSQL())
		}
	}
	return sqls
}

// CreateTableSQL returns the DDL SQL to create this table.
func (t *Table) CreateTableSQL() string {
	colsConstraints := t.CreateColumnClauses()
	return "create table " + t.Name +
		" (\n" + strings.Join(colsConstraints, ",\n") + "\n)"
}

// CreateIndexConstraintSQLs returns the DDL statements to create the table's
// indexes and constraints.
func (t *Table) CreateIndexConstraintSQLs() []string {
	sqls := []string{}
	for _, index := range t.Indexes {
		if !index.Force {
			sqls = append(sqls, index.SQL())
		}
	}
	for _, c := range t.Constraints {
		sqls = append(sqls, "alter table "+t.Name+" add "+c.SQL())
	}
	return sqls
}

// CreateColumnClauses returns the list of column SQL expressions.
func (t *Table) CreateColumnClauses() []string {
	pieces := make([]string, len(t.Columns))
	for i, col := range t.Columns {
		pieces[i] = "  " + col.SQL()
	}
	return pieces
}

// SQL returns the SQL expression for the column c
func (c *Column) SQL() string {
	base := c.Name + " " + c.SQLType
	if c.Default != "" {
		base += " " + c.Default
	}
	return base
}

// SQL returns the DDL to create the index i.
func (i *Index) SQL() string {
	createClause := "create index"
	if i.Unique {
		createClause = "create unique index"
	}
	return createClause + " " + i.Name + " on " + i.TableName +
		" (" + strings.Join(i.Columns, ", ") + ")"
}

// DropSQL returns the DDL statement to drop the index i.
func (i *Index) DropSQL() string {
	return "drop index if exists " + i.Name
}
