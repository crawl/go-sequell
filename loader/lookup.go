package loader

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	cdb "github.com/crawl/go-sequell/crawl/db"
	"github.com/crawl/go-sequell/xlog"
	"github.com/golang/groupcache/lru"
	"github.com/pkg/errors"
)

// A TableLookup collects a list of lookup field values that must be inserted
// into a lookup table, and replaced by their corresponding foreign keys.
type TableLookup struct {
	Table         *cdb.LookupTable
	Lookups       map[string]LookupValue
	Capacity      int
	CaseSensitive bool

	// globallyUnique means that the lookup field must be globally unique across
	// all logfiles and milestones. A duplicate value for a GloballyUnique field
	// means that the entire xlog row is a duplicate and should be discarded.
	globallyUnique bool

	// duplicateGlobalLookupIDs is the list of lookup values that are duplicated
	// when GloballyUnique == true
	duplicateGlobalLookupIDs map[string]bool

	idCache           *lru.Cache
	lookupField       *cdb.Field
	fieldNames        []string
	refFieldNames     []string
	derivedFieldNames []string
	baseQuery         string
}

// A LookupValue is a string Value that belongs in a lookup table, along with
// any derived values that can be calculated from the Value.
type LookupValue struct {
	Value         string
	DerivedValues []string
}

// NewTableLookup returns a new table lookup object capable of caching
// up to capacity rows worth of lookups.
//
// The usage sequence is: Reset(), Add() all your rows, up to the lookup
// capacity, then ResolveAll once to look up uncached lookup ids, then
// SetIDs all your rows.
func NewTableLookup(table *cdb.LookupTable, capacity int) *TableLookup {
	tl := &TableLookup{
		Table:         table,
		CaseSensitive: table.CaseSensitive(),
		Lookups:       map[string]LookupValue{},
		Capacity:      capacity,
	}
	tl.init()
	return tl
}

// Reset clears t and prepares it for a fresh batch of rows.
func (t *TableLookup) Reset() {
	if len(t.duplicateGlobalLookupIDs) == 0 {
		return
	}

	t.clearGlobalLookupIDs()
}

func (t *TableLookup) clearGlobalLookupIDs() {
	t.duplicateGlobalLookupIDs = map[string]bool{}
}

// GloballyUnique returns true if t is a lookup of values that must be globally
// unique, where duplicate values imply duplicat xlog entries that must be
// discarded.
func (t *TableLookup) GloballyUnique() bool {
	return t.globallyUnique
}

// SetGloballyUnique flags t as requiring globally unique values and returns t.
func (t *TableLookup) SetGloballyUnique(unique bool) *TableLookup {
	t.globallyUnique = unique
	if unique {
		t.clearGlobalLookupIDs()
	}
	return t
}

// Name returns the name of the lookup table.
func (t *TableLookup) Name() string {
	return t.Table.Name
}

func (t *TableLookup) init() {
	cardinality := t.Table.ReferencingFieldCount()
	t.Capacity *= cardinality
	t.lookupField = t.Table.LookupField()
	t.fieldNames = make([]string, cardinality)
	t.refFieldNames = make([]string, cardinality)
	for i, f := range t.Table.ReferencingFields {
		t.fieldNames[i] = f.Name
		t.refFieldNames[i] = f.RefName()
	}

	t.derivedFieldNames = make([]string, len(t.Table.DerivedFields))
	for i, f := range t.Table.DerivedFields {
		t.derivedFieldNames[i] = f.Name
	}

	if len(t.fieldNames) > 1 && len(t.derivedFieldNames) > 0 {
		panic("Cannot have a compound lookup table with derived fields")
	}

	t.baseQuery = t.constructBaseQuery()
	t.idCache = lru.New(t.Capacity)

	t.SetGloballyUnique(t.lookupField.UUID)
}

func (t *TableLookup) constructBaseQuery() string {
	return `select id, ` + t.lookupField.SQLName +
		` from ` + t.Table.TableName() +
		` where ` + t.lookupField.SQLName + ` in `
}

// SetIDs sets [field]_id to the lookup id for all lookup fields in the xlog.
// You must call ResolveAll before using SetIds.
func (t *TableLookup) SetIDs(x xlog.Xlog) error {
	for i, f := range t.refFieldNames {
		field := t.fieldNames[i]
		value := x[field]
		id, err := t.ID(value)
		if err != nil {
			return errors.Wrapf(err, "SetIDs(%#v)/%#v [%#v]", value, field, x)
		}
		x[f] = strconv.Itoa(id)
	}
	return nil
}

// ID retrieves the foreign key value for the given lookup value. ID must be
// used after a call to ResolveAll to load lookup values into the lookup table,
// or find the existing ids for lookup values that are already in the lookup
// table.
func (t *TableLookup) ID(value string) (int, error) {
	valueLookup := t.lookupKey(value)

	if t.globallyUnique && t.duplicateGlobalLookupIDs[valueLookup] {
		return 0, ErrDuplicateRow
	}

	if id, ok := t.idCache.Get(valueLookup); ok {
		if t.globallyUnique {
			t.duplicateGlobalLookupIDs[valueLookup] = true
		}
		return id.(int), nil
	}
	return 0, fmt.Errorf("value %#v not found in %s", value, t)
}

func (t *TableLookup) String() string {
	return "Lookup[" + t.Table.Name + "]"
}

// Add adds the lookup fields for this table and any derived values from the
// xlog to the list of values to be resolved/inserted to the lookup table.
func (t *TableLookup) Add(x xlog.Xlog) {
	var derivedFieldValues []string
	if t.derivedFieldNames != nil {
		derivedFieldValues = make([]string, len(t.derivedFieldNames))
		for i, n := range t.derivedFieldNames {
			derivedFieldValues[i] = x[n]
		}
	}
	for _, f := range t.fieldNames {
		t.AddLookup(x[f], derivedFieldValues)
	}
}

// AddAll resets t and adds all the given logs to be resolved
func (t *TableLookup) AddAll(logs []xlog.Xlog) {
	t.Reset()
	for _, log := range logs {
		t.Add(log)
	}
}

// LookupKey normalizes a lookupValue into its canonical form
func LookupKey(lookupValue string, caseSensitive bool) (key string) {
	lookupValue = NormalizeValue(lookupValue)
	if caseSensitive {
		return lookupValue
	}
	return strings.ToLower(lookupValue)
}

// lookupKey transforms a lookup value into its canonical form in the id map.
func (t *TableLookup) lookupKey(lookup string) string {
	return LookupKey(lookup, t.CaseSensitive)
}

// AddLookup adds the lookup value and the list of values derived from it to
// t, to be resolved by the next call to ResolveAll.
func (t *TableLookup) AddLookup(lookup string, derivedValues []string) {
	key := t.lookupKey(lookup)
	if _, ok := t.idCache.Get(key); ok {
		return
	}
	if t.IsFull() {
		panic(fmt.Sprintf("TableLookup[%s] full", t.Table.Name))
	}
	t.Lookups[key] = LookupValue{Value: lookup, DerivedValues: derivedValues}
}

// ResolveAll resolves all lookups in the given xlogs logs into numeric ids. For
// instance, given killer names like "an ogre", "a kobold", etc. looks up the
// corresponding IDs in the lookup table (l_killer) and saves the IDs for each
// value in the ID lookup cache.
func (t *TableLookup) ResolveAll(tx *sql.Tx, logs []xlog.Xlog) error {
	t.AddAll(logs)
	return t.ResolveQueued(tx)
}

// ResolveQueued resolves all lookups field values previously queued using
// t.AddAll or t.Add.
func (t *TableLookup) ResolveQueued(tx *sql.Tx) error {
	if err := t.findExistingValueIDs(tx); err != nil {
		return err
	}
	if err := t.insertNewValues(tx); err != nil {
		return err
	}
	if len(t.Lookups) != 0 {
		return fmt.Errorf("%s: ResolveAll() left unresolved entries: %#v",
			t.Name(), t.Lookups)
	}
	return nil
}

type fieldValueLookup int

const (
	fieldValueLookupNew fieldValueLookup = iota
	fieldValueLookupExisting
)

func (t *TableLookup) insertNewValues(tx *sql.Tx) error {
	if len(t.Lookups) == 0 {
		return nil
	}
	insertQuery :=
		t.insertStatement(len(t.Lookups), len(t.derivedFieldNames)+1)
	values := t.insertValues()
	rows, err := tx.Query(insertQuery, values...)
	if err != nil {
		return errors.Wrapf(err, "Query: %s, binds: %#v", insertQuery, values)
	}
	return t.resolveRows(rows, fieldValueLookupNew)
}

func (t *TableLookup) findExistingValueIDs(tx *sql.Tx) error {
	if len(t.Lookups) == 0 {
		return nil
	}
	query := t.lookupQuery(len(t.Lookups))
	rows, err := tx.Query(query, t.lookupValues()...)
	if err != nil {
		return errors.Wrap(err, query)
	}
	return t.resolveRows(rows, fieldValueLookupExisting)
}

func (t *TableLookup) resolveRows(rows *sql.Rows, lookupResult fieldValueLookup) error {
	defer rows.Close()
	var id int
	var lookupValue string
	for rows.Next() {
		if err := rows.Scan(&id, &lookupValue); err != nil {
			return err
		}
		t.resolveSingleLookup(id, lookupValue, lookupResult)
	}
	return errors.Wrap(rows.Err(), "lookup.resolveRows")
}

func (t *TableLookup) resolveSingleLookup(id int, lookup string, lookupResult fieldValueLookup) {
	key := t.lookupKey(lookup)

	if t.globallyUnique && lookupResult == fieldValueLookupExisting {
		t.duplicateGlobalLookupIDs[key] = true
	} else {
		t.idCache.Add(key, id)
	}
	delete(t.Lookups, key)
}

func (t *TableLookup) lookupValues() []interface{} {
	values := make([]interface{}, len(t.Lookups))
	i := 0
	for _, v := range t.Lookups {
		values[i] = NormalizeValue(v.Value)
		i++
	}
	return values
}

func (t *TableLookup) lookupQuery(nbinds int) string {
	var buf bytes.Buffer
	buf.WriteString(t.baseQuery)
	buf.WriteString("(")
	for i := 1; i <= nbinds; i++ {
		if i > 1 {
			buf.WriteString(",")
		}
		buf.WriteString("$")
		buf.WriteString(strconv.Itoa(i))
	}
	buf.WriteString(")")
	return buf.String()
}

func (t *TableLookup) insertStatement(nrows int, bindsPerRow int) string {
	var buf bytes.Buffer
	buf.WriteString("insert into " + t.Table.TableName() + " (")
	buf.WriteString(t.lookupField.SQLName)
	for _, g := range t.Table.DerivedFields {
		buf.WriteString(",")
		buf.WriteString(g.SQLName)
	}
	buf.WriteString(")\nvalues ")

	i := 1
	for row := 0; row < nrows; row++ {
		if row > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString("(")
		for bind := 0; bind < bindsPerRow; bind++ {
			if bind > 0 {
				buf.WriteString(",")
			}
			buf.WriteString("$")
			buf.WriteString(strconv.Itoa(i))
			i++
		}
		buf.WriteString(")\n")
	}
	buf.WriteString("returning id, " + t.lookupField.SQLName)
	return buf.String()
}

func (t *TableLookup) insertValues() []interface{} {
	res := make([]interface{}, len(t.Lookups)*(1+len(t.derivedFieldNames)))
	i := 0
	for _, v := range t.Lookups {
		res[i] = NormalizeValue(v.Value)
		i++
		for _, v := range v.DerivedValues {
			res[i] = NormalizeValue(v)
			i++
		}
	}
	return res
}

// IsFull returns true if t is at its capacity (this is usually the point you'd
// call ResolveAll())
func (t *TableLookup) IsFull() bool {
	return len(t.Lookups) >= t.Capacity
}
