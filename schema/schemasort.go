package schema

import (
	"sort"
)

type TableSort []*Table

func (t TableSort) Len() int      { return len(t) }
func (t TableSort) Swap(i, j int) { t[j], t[i] = t[i], t[j] }
func (t TableSort) Less(i, j int) bool {
	// If i depends on j, i > j
	if t[i].DependsOn(t[j]) {
		return false
	}
	if t[j].DependsOn(t[i]) {
		return true
	}
	// Otherwise, lexicographic sort:
	return t[i].Name < t[j].Name
}

func (s *Schema) Sort() *Schema {
	sort.Sort(TableSort(s.Tables))
	return s
}

func (t *Table) DependsOn(other *Table) bool {
	cachedDeps := t.knownDeps != nil
	if cachedDeps {
		if dependent, ok := t.knownDeps[other.Name]; ok {
			return dependent
		}
	}

	dependent := false
	for _, c := range t.Constraints {
		if c.DependsOnTable() == other.Name {
			dependent = true
			break
		}
	}
	if !cachedDeps {
		t.knownDeps = map[string]bool{}
	}
	t.knownDeps[other.Name] = dependent
	return dependent
}
