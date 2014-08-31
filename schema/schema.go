package schema

import (
	"fmt"
)

type Schema struct {
	Tables []*Table
}

type Table struct {
	Name        string
	Columns     []*Column
	Indexes     []*Index
	Constraints []Constraint
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
}

type ForeignKeyConstraint struct {
	SourceTableField string
	TargetTable      string
	TargetTableField string
}

func (f *ForeignKeyConstraint) Sql() string {
	return fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s (%s)",
		f.SourceTableField, f.TargetTable, f.TargetTableField)
}

type PrimaryKeyConstraint struct {
	Column string
}

func (c *PrimaryKeyConstraint) Sql() string {
	return "PRIMARY KEY (" + c.Column + ")"
}
