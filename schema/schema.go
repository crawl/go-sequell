package schema

import (
	"fmt"
)

const UnknownColumn = "§"

type Diff int

const (
	Added Diff = iota
	Removed
	Changed
)

type Schema struct {
	Tables []*Table
}

type Table struct {
	Name        string
	Columns     []*Column
	Indexes     []*Index
	Constraints []Constraint
	knownDeps   map[string]bool
	DiffStruct
}

type Column struct {
	Name    string
	SqlType string
	Default string
	Alias   string
	DiffStruct
}

type Index struct {
	Name      string
	TableName string
	Columns   []string
	DiffStruct
}

type Constraint interface {
	Sql() string
	DependsOnTable() string
	Differ
}

type Differ interface {
	Diff() Diff
	SetDiff(diff Diff)
}

type DiffStruct struct {
	diff Diff
}

func (d DiffStruct) Diff() Diff        { return d.diff }
func (d DiffStruct) SetDiff(diff Diff) { d.diff = diff }

type ForeignKeyConstraint struct {
	SourceTableField string
	TargetTable      string
	TargetTableField string
	DiffStruct
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
	DiffStruct
}

func (c PrimaryKeyConstraint) Sql() string {
	return "primary key (" + c.Column + ")"
}

func (c PrimaryKeyConstraint) DependsOnTable() string {
	return ""
}
