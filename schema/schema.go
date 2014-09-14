package schema

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
	*DiffStruct
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
	Unique    bool
	DiffStruct
}

type Constraint interface {
	Name() string
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

func (d *DiffStruct) DiffMode() Diff        { return d.Diff }
func (d *DiffStruct) SetDiffMode(diff Diff) { d.Diff = diff }

func constraintNamed(name string) string {
	if name == "" {
		return ""
	}
	return "constraint " + name + " "
}

type ForeignKeyConstraint struct {
	ConstraintName   string
	SourceTableField string
	TargetTable      string
	TargetTableField string
	*DiffStruct
}

func (f ForeignKeyConstraint) Sql() string {
	return constraintNamed(f.ConstraintName) +
		"foreign key (" + f.SourceTableField + ") references " +
		f.TargetTable + " (" + f.TargetTableField + ")"
}

func (f ForeignKeyConstraint) Name() string {
	return f.ConstraintName
}

func (f ForeignKeyConstraint) DependsOnTable() string {
	return f.TargetTable
}

type PrimaryKeyConstraint struct {
	ConstraintName string
	Column         string
	*DiffStruct
}

func (c PrimaryKeyConstraint) Name() string {
	return c.ConstraintName
}

func (c PrimaryKeyConstraint) Sql() string {
	return constraintNamed(c.ConstraintName) + "primary key (" + c.Column + ")"
}

func (c PrimaryKeyConstraint) DependsOnTable() string {
	return ""
}
