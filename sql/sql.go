package sql

import "bytes"

type Table struct {
	Name  string
	Expr  *Select
	Alias string
}

func (t *Table) QualifiedName(name string) string {
	return t.EffectiveAlias() + "." + name
}

func (t *Table) SimpleTable() bool {
	return t.Expr == nil
}

func (t *Table) EffectiveAlias() string {
	if t.Alias == "" {
		return t.Name
	}
	return t.Alias
}

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

type TableExpr struct {
	*Table
	JoinType       string
	JoinConditions string
}

func (t *TableExpr) JoinSQL() string {
	if t.JoinConditions != "" && t.JoinType != "" {
		return " " + t.JoinType + " " + t.Table.SQL() + " on " +
			t.JoinConditions
	} else {
		return ", " + t.Table.SQL()
	}
}

type Column struct {
	Expr  string
	Alias string
}

func (c *Column) SQL() string {
	sql := c.Expr
	if c.Alias != "" {
		sql += "as " + c.Alias
	}
	return sql
}

type Select struct {
	Columns    []*Column
	TableExprs []*TableExpr
}

func (q *Select) AddColumnExpr(expr string) *Column {
	return q.AddAliasedColumn(expr, "")
}

func (q *Select) AddAliasedColumn(expr, alias string) *Column {
	return q.AddColumn(&Column{Expr: expr, Alias: alias})
}

func (q *Select) AddColumn(col *Column) *Column {
	q.Columns = append(q.Columns, col)
	return col
}

func (q *Select) AddNamedTable(name string) *TableExpr {
	return q.AddNamedAliasedTable(name, "")
}

func (q *Select) AddTableExpr(expr *TableExpr) *TableExpr {
	q.TableExprs = append(q.TableExprs, expr)
	return expr
}

func (q *Select) AddNamedAliasedTable(name, alias string) *TableExpr {
	return q.AddTableExpr(&TableExpr{
		Table: &Table{Name: name, Alias: alias},
	})
}

func (q *Select) AddTable(table *Table) *TableExpr {
	return q.AddTableExpr(&TableExpr{
		Table: table,
	})
}

func (q *Select) AddJoinTable(table *Table, joinExpr string, joinType string) *TableExpr {
	return q.AddTableExpr(&TableExpr{
		Table:          table,
		JoinConditions: joinExpr,
		JoinType:       joinType,
	})
}

func (q *Select) SQL() string {
	return "select " + q.ColumnSQL() + " from " + q.FromClauses()
}

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
