package sql

import "bytes"

// A Table is a SQL query fragment referencing a single table.
type Table struct {
	Name  string
	Expr  *Select
	Alias string
}

// QualifiedName qualifies the column name with the table's query alias.
func (t *Table) QualifiedName(name string) string {
	return t.EffectiveAlias() + "." + name
}

// SimpleTable returns true if there are no explicit columns selected from
// this table, viz. a simple reference to the underlying table.
func (t *Table) SimpleTable() bool {
	return t.Expr == nil
}

// EffectiveAlias gets the alias for this table in the broader query. If no
// explicit alias has been specified, returns the table name itself.
func (t *Table) EffectiveAlias() string {
	if t.Alias == "" {
		return t.Name
	}
	return t.Alias
}

// SQL ges the SQL clause for this query fragment.
func (t *Table) SQL() string {
	var sql string
	if t.SimpleTable() {
		sql = t.Name
	} else {
		sql = "(" + t.Expr.SQL() + ")"
	}
	if t.Alias != "" {
		sql += " as " + t.Alias
	}
	return sql
}

// A TableExpr specifies a table and its join relationship to the set of
// existing tables in the query.
type TableExpr struct {
	*Table
	JoinType       string
	JoinConditions string
}

// JoinSQL gets the SQL join clause for this joined table.
func (t *TableExpr) JoinSQL() string {
	if t.JoinConditions != "" && t.JoinType != "" {
		return " " + t.JoinType + " " + t.Table.SQL() + " on " +
			t.JoinConditions
	}
	return ", " + t.Table.SQL()
}

// A Column represents a selected expression in a SQL query and its alias.
type Column struct {
	Expr  string
	Alias string
}

// SQL gets the SQL expression for the column expression with its optional
// alias.
func (c *Column) SQL() string {
	sql := c.Expr
	if c.Alias != "" {
		sql += "as " + c.Alias
	}
	return sql
}

// A Select represents a SQL select statement with a list of columns to
// select and a list of tables being selected from.
type Select struct {
	Columns    []*Column
	TableExprs []*TableExpr
}

// AddColumnExpr adds expr as a selected column.
func (q *Select) AddColumnExpr(expr string) *Column {
	return q.AddAliasedColumn(expr, "")
}

// AddAliasedColumn adds expr as a selected column with alias.
func (q *Select) AddAliasedColumn(expr, alias string) *Column {
	return q.AddColumn(&Column{Expr: expr, Alias: alias})
}

// AddColumn adds col as a column.
func (q *Select) AddColumn(col *Column) *Column {
	q.Columns = append(q.Columns, col)
	return col
}

// AddNamedTable adds name as a selected table.
func (q *Select) AddNamedTable(name string) *TableExpr {
	return q.AddNamedAliasedTable(name, "")
}

// AddTableExpr adds expr as a joined table expression.
func (q *Select) AddTableExpr(expr *TableExpr) *TableExpr {
	q.TableExprs = append(q.TableExprs, expr)
	return expr
}

// AddNamedAliasedTable adds name as a selected table with a table alias.
func (q *Select) AddNamedAliasedTable(name, alias string) *TableExpr {
	return q.AddTableExpr(&TableExpr{
		Table: &Table{Name: name, Alias: alias},
	})
}

// AddTable adds table as a selected table.
func (q *Select) AddTable(table *Table) *TableExpr {
	return q.AddTableExpr(&TableExpr{
		Table: table,
	})
}

// AddJoinTable adds table as a join table with joinExpr as the join condition
// and joinType specifying the type of join.
func (q *Select) AddJoinTable(table *Table, joinExpr string, joinType string) *TableExpr {
	return q.AddTableExpr(&TableExpr{
		Table:          table,
		JoinConditions: joinExpr,
		JoinType:       joinType,
	})
}

// SQL gets the query SQL as a string
func (q *Select) SQL() string {
	return "select " + q.ColumnSQL() + " from " + q.FromClauses()
}

// ColumnSQL gets the SQL fragment for the columns being selected.
func (q *Select) ColumnSQL() string {
	res := bytes.Buffer{}
	for i, c := range q.Columns {
		if i > 0 {
			res.WriteString(", ")
		}
		res.WriteString(c.SQL())
	}
	return res.String()
}

// FromClauses gets the SQL FROM clauses for this query.
func (q *Select) FromClauses() string {
	res := bytes.Buffer{}
	for i, expr := range q.TableExprs {
		if i > 0 {
			res.WriteString(expr.JoinSQL())
		} else {
			res.WriteString(expr.SQL())
		}
	}
	return res.String()
}
