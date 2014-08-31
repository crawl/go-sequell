package schema

import (
	"fmt"
)

const UnknownColumn = "§"

type Schema struct {
	Tables []*Table
}

type Table struct {
	Name        string
	Columns     []*Column
	Indexes     []*Index
	Constraints []Constraint
	knownDeps   map[string]bool
}

type Column struct {
	Name    string
	SqlType string
	Default string
	Alias   string
}

type Index struct {
	Name      string
	TableName string
	Columns   []string
}

type Constraint interface {
	Sql() string
	DependsOnTable() string
}

type ForeignKeyConstraint struct {
	SourceTableField string
	TargetTable      string
	TargetTableField string
}

func (f ForeignKeyConstraint) Sql() string {
	return fmt.Sprintf("foreign key (%s) references %s (%s)",
		f.SourceTableField, f.TargetTable, f.TargetTableField)
}

func (f ForeignKeyConstraint) DependsOnTable() string {
	return f.TargetTable
}

type PrimaryKeyConstraint struct {
	Column string
}

func (c PrimaryKeyConstraint) Sql() string {
	return "primary key (" + c.Column + ")"
}

func (c PrimaryKeyConstraint) DependsOnTable() string {
	return ""
}
