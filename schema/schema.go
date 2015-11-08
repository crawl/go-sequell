package schema

// UnknownColumn is the character that represents a column that cannot be
// introspected from the database.
const UnknownColumn = "§"

// Diff specifies the nature of a schema change
type Diff int

// Diff types
const (
	NoDiff Diff = iota
	Added
	Removed
	Changed
)

// Sigil returns a short description of this diff type.
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

// A Schema is a set of tables
type Schema struct {
	Tables []*Table
}

// Table gets the table object given a table name.
func (s *Schema) Table(name string) *Table {
	for _, table := range s.Tables {
		if table.Name == name {
			return table
		}
	}
	return nil
}

// A Table represents a single table's columns, indexes, constaints and
// dependencies.
type Table struct {
	Name        string
	Columns     []*Column
	Indexes     []*Index
	Constraints []Constraint
	knownDeps   map[string]bool
	*DiffStruct
}

// Column gets the column object given the column name
func (t *Table) Column(name string) *Column {
	for _, c := range t.Columns {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// Index finds the index on this table with the given name
func (t *Table) Index(name string) *Index {
	for _, i := range t.Indexes {
		if i.Name == name {
			return i
		}
	}
	return nil
}

// A Column represents a single column in a table
type Column struct {
	Name    string
	SQLType string
	Default string
	Alias   string
	DiffStruct
}

// An Index represents a table index.
type Index struct {
	Name      string
	TableName string
	Columns   []string
	Unique    bool
	Force     bool // Always created, even if other indexes are omitted
	DiffStruct
}

// A Constraint is a table constraint such as a unique column or foreign-key.
type Constraint interface {
	Name() string
	SQL() string
	DependsOnTable() string
	Differ
}

// A Differ controls how a schema diff is performed. (TODO: rename)
type Differ interface {
	DiffMode() Diff
	SetDiffMode(diff Diff)
}

// A DiffStruct represents a diff.
type DiffStruct struct {
	Diff Diff
}

// DiffMode returns the current diff mode.
func (d *DiffStruct) DiffMode() Diff { return d.Diff }

// SetDiffMode sets the diff mode to diff.
func (d *DiffStruct) SetDiffMode(diff Diff) { d.Diff = diff }

func constraintNamed(name string) string {
	if name == "" {
		return ""
	}
	return "constraint " + name + " "
}

// A ForeignKeyConstraint links a foreign key column in a table to a (primary)
// key in a target table.
type ForeignKeyConstraint struct {
	ConstraintName   string
	SourceTableField string
	TargetTable      string
	TargetTableField string
	*DiffStruct
}

// SQL gets the SQL clause for the constraint f.
func (f ForeignKeyConstraint) SQL() string {
	return constraintNamed(f.ConstraintName) +
		"foreign key (" + f.SourceTableField + ") references " +
		f.TargetTable + " (" + f.TargetTableField + ") on delete cascade"
}

// Name gets the name of the foreign key constraint f
func (f ForeignKeyConstraint) Name() string {
	return f.ConstraintName
}

// DependsOnTable gets the name of the table this constraint depends on.
func (f ForeignKeyConstraint) DependsOnTable() string {
	return f.TargetTable
}

// A PrimaryKeyConstraint represents the primary key constraint for a table.
type PrimaryKeyConstraint struct {
	ConstraintName string
	Column         string
	*DiffStruct
}

// Name returns the constraint name for c.
func (c PrimaryKeyConstraint) Name() string {
	return c.ConstraintName
}

// SQL returns the SQL for c.
func (c PrimaryKeyConstraint) SQL() string {
	return constraintNamed(c.ConstraintName) + "primary key (" + c.Column + ")"
}

// DependsOnTable returns "" always (primary-key constraints can have no
// dependencies on other tables).
func (c PrimaryKeyConstraint) DependsOnTable() string {
	return ""
}
