package schema

import (
	"fmt"
)

const UnknownColumn = "§"

type Diff int

const (
	NoDiff Diff = iota
	Added
	Removed
	Changed
)

func (d Diff) Sigil() string {
	switch d {
	case Added:
		return "A"
	case Removed:
		return "D"
	case Changed:
		return "M"
	}
	return ""
}

type Schema struct {
	Tables []*Table
}

func (s *Schema) Table(name string) *Table {
	for _, table := range s.Tables {
		if table.Name == name {
			return table
		}
	}
	return nil
}

type Table struct {
	Name        string
	Columns     []*Column
	Indexes     []*Index
	Constraints []Constraint
	knownDeps   map[string]bool
	DiffStruct
}

func (t *Table) Column(name string) *Column {
	for _, c := range t.Columns {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func (t *Table) Index(name string) *Index {
	for _, i := range t.Indexes {
		if i.Name == name {
			return i
		}
	}
	return nil
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
	DiffMode() Diff
	SetDiffMode(diff Diff)
}

type DiffStruct struct {
	Diff Diff
}

func (d DiffStruct) DiffMode() Diff        { return d.Diff }
func (d DiffStruct) SetDiffMode(diff Diff) { d.Diff = diff }

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
