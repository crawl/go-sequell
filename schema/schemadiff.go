package schema

import (
	"fmt"
	"io"
)

// PrintDelta writes the schema delta to out.
func (s *Schema) PrintDelta(out io.Writer) {
	for _, t := range s.Tables {
		t.PrintDelta(out)
	}
}

// PrintDelta writes the table delta to out.
func (t *Table) PrintDelta(out io.Writer) {
	switch t.Diff {
	case Added, Removed:
		fmt.Fprintf(out, "%s: table %s\n", t.Diff.Sigil(), t.Name)
	case Changed:
		fmt.Fprintf(out, "%s: table %s (%d col, %d ind, %d cons):\n",
			t.Diff.Sigil(), t.Name, len(t.Columns), len(t.Indexes), len(t.Constraints))
		for _, c := range t.Columns {
			c.PrintDelta(out)
		}
		for _, c := range t.Constraints {
			fmt.Fprintf(out, "\t%s: constraint: %s\n", Added.Sigil(), c.SQL())
		}
		for _, i := range t.Indexes {
			i.PrintDelta(out)
		}
	}
}

// PrintDelta writes the column delta to out.
func (c *Column) PrintDelta(out io.Writer) {
	if c.Diff != NoDiff {
		fmt.Fprintf(out, "\t%s: %s\n", c.Diff.Sigil(), c.SQL())
	}
}

// PrintDelta writes the index delta to out.
func (c *Index) PrintDelta(out io.Writer) {
	if c.Diff != NoDiff {
		fmt.Fprintf(out, "\t%s: %s\n", c.Diff.Sigil(), c.SQL())
	}
}

// DiffSchema compares s to old and returns a diff schema.
func (s *Schema) DiffSchema(old *Schema) *Schema {
	diffSchema := Schema{}
	for _, table := range s.Tables {
		otherTable := old.Table(table.Name)
		if diffTable := table.DiffSchema(otherTable); diffTable != nil {
			diffSchema.Tables = append(diffSchema.Tables, diffTable)
		}
	}
	return &diffSchema
}

// DiffSchema compares t to old and returns a diff table.
func (t *Table) DiffSchema(old *Table) *Table {
	if old == nil {
		return &Table{
			Name:       t.Name,
			DiffStruct: &DiffStruct{Added},
		}
	}

	diffTable := Table{
		Name:       t.Name,
		DiffStruct: &DiffStruct{Changed},
	}
	for _, col := range t.Columns {
		if diffCol := col.DiffSchema(old.Column(col.Name)); diffCol != nil {
			diffTable.Columns = append(diffTable.Columns, diffCol)
		}
	}
	for _, ind := range t.Indexes {
		if otherInd := old.Index(ind.Name); otherInd == nil {
			missingIndex := *ind
			missingIndex.SetDiffMode(Added)
			diffTable.Indexes = append(diffTable.Indexes, &missingIndex)
		}
	}
	diffTable.Constraints = t.DiffConstraints(old)
	if diffTable.Columns != nil || diffTable.Indexes != nil ||
		diffTable.Constraints != nil {
		return &diffTable
	}
	return nil
}

// DiffConstraints compares the constraints on t and old and returns the list of
// constraint diffs.
func (t *Table) DiffConstraints(old *Table) []Constraint {
	cmap := map[string]bool{}
	for _, c := range old.Constraints {
		cmap[c.SQL()] = true
	}
	res := []Constraint(nil)
	for _, c := range t.Constraints {
		if !cmap[c.SQL()] {
			res = append(res, c)
		}
	}
	return res
}

// DiffSchema compares c and old and returns a diff column.
func (c *Column) DiffSchema(old *Column) *Column {
	copyCol := func(change Diff) *Column {
		col := *c
		col.SetDiffMode(change)
		return &col
	}
	if old == nil {
		return copyCol(Added)
	}
	if old.SQLType != c.SQLType || old.Default != c.Default {
		return copyCol(Changed)
	}
	return nil
}
