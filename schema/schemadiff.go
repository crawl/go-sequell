package schema

import (
	"fmt"
	"io"
)

func (s *Schema) PrintDelta(out io.Writer) {
	for _, t := range s.Tables {
		t.PrintDelta(out)
	}
}

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
			fmt.Fprintf(out, "\t%s: constraint: %s\n", Added.Sigil(), c.Sql())
		}
		for _, i := range t.Indexes {
			i.PrintDelta(out)
		}
	}
}

func (c *Column) PrintDelta(out io.Writer) {
	if c.Diff != NoDiff {
		fmt.Fprintf(out, "\t%s: %s\n", c.Diff.Sigil(), c.Sql())
	}
}

func (c *Index) PrintDelta(out io.Writer) {
	if c.Diff != NoDiff {
		fmt.Fprintf(out, "\t%s: %s\n", c.Diff.Sigil(), c.Sql())
	}
}

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

func (t *Table) DiffConstraints(old *Table) []Constraint {
	cmap := map[string]bool{}
	for _, c := range old.Constraints {
		cmap[c.Sql()] = true
	}
	res := []Constraint(nil)
	for _, c := range t.Constraints {
		if !cmap[c.Sql()] {
			res = append(res, c)
		}
	}
	return res
}

func (c *Column) DiffSchema(old *Column) *Column {
	copyCol := func(change Diff) *Column {
		col := *c
		col.SetDiffMode(change)
		return &col
	}
	if old == nil {
		return copyCol(Added)
	}
	if old.SqlType != c.SqlType || old.Default != c.Default {
		return copyCol(Changed)
	}
	return nil
}
