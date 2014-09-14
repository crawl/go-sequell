package loader

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/groupcache/lru"
	cdb "github.com/greensnark/go-sequell/crawl/db"
	"github.com/greensnark/go-sequell/ectx"
	"github.com/greensnark/go-sequell/xlog"
)

type TableLookup struct {
	Table         *cdb.LookupTable
	Lookups       map[string]LookupValue
	Capacity      int
	CaseSensitive bool

	idCache           *lru.Cache
	lookupField       *cdb.Field
	fieldNames        []string
	refFieldNames     []string
	derivedFieldNames []string
	baseQuery         string
}

type LookupValue struct {
	Value         string
	DerivedValues []string
}

// NewTableLookup returns a new table lookup object capable of caching
// up to capacity rows worth of lookups.
//
// The usage sequence is: Add all your rows, up to the lookup
// capacity, then ResolveAll once to look up uncached lookup ids, then
// SetIds all your rows.
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
}

func (t *TableLookup) constructBaseQuery() string {
	return `select id, ` + t.lookupField.SqlName +
		` from ` + t.Table.TableName() +
		` where ` + t.lookupField.SqlName + ` in `
}

// SetIds sets field_id to the lookup id for all lookup fields in the xlog.
func (t *TableLookup) SetIds(x xlog.Xlog) error {
	for i, f := range t.refFieldNames {
		field := t.fieldNames[i]
		value := x[field]
		id, err := t.Id(value)
		if err != nil {
			return ectx.Err(fmt.Sprintf("SetId(%#v)/%#v [%#v]", value, field, x), err)
		}
		x[f] = strconv.Itoa(id)
	}
	return nil
}

func (t *TableLookup) Id(value string) (int, error) {
	if id, ok := t.idCache.Get(t.LookupKey(value)); ok {
		return id.(int), nil
	}
	return 0, fmt.Errorf("value %#v not found in %s", value, t)
}

func (t *TableLookup) String() string {
	return "Lookup[" + t.Table.Name + "]"
}

func (t *TableLookup) Add(x xlog.Xlog) {
	var derivedFieldValues []string
	if t.derivedFieldNames != nil {
		derivedFieldValues = make([]string, len(t.derivedFieldNames))
		for i, n := range t.derivedFieldNames {
			derivedFieldValues[i] = x[n]
		}
	}
	for _, f := range t.fieldNames {
		t.AddLookup(x[f], derivedFieldValues, x)
	}
}

func (t *TableLookup) LookupKey(lookup string) string {
	lookup = NormalizeValue(lookup)
	if !t.CaseSensitive {
		return strings.ToLower(lookup)
	}
	return lookup
}

func (t *TableLookup) AddLookup(lookup string, derivedValues []string, x xlog.Xlog) {
	key := t.LookupKey(lookup)
	if _, ok := t.idCache.Get(key); ok {
		return
	}
	if t.IsFull() {
		panic(fmt.Sprintf("TableLookup[%s] full", t.Table.Name))
	}
	t.Lookups[key] = LookupValue{Value: lookup, DerivedValues: derivedValues}
}

// ResolveAll resolves all queued lookups into numeric ids.
func (t *TableLookup) ResolveAll(tx *sql.Tx) error {
	if err := t.queryAll(tx); err != nil {
		return err
	}
	if err := t.insertAll(tx); err != nil {
		return err
	}
	if len(t.Lookups) != 0 {
		return fmt.Errorf("%s: ResolveAll() left unresolved entries: %#v",
			t.Name(), t.Lookups)
	}
	return nil
}

func (t *TableLookup) insertAll(tx *sql.Tx) error {
	if len(t.Lookups) == 0 {
		return nil
	}
	insertQuery :=
		t.insertStatement(len(t.Lookups), len(t.derivedFieldNames)+1)
	rows, err := tx.Query(insertQuery, t.insertValues()...)
	if err != nil {
		return ectx.Err(insertQuery, err)
	}
	return t.resolveRows(rows, nil)
}

func (t *TableLookup) queryAll(tx *sql.Tx) error {
	if len(t.Lookups) == 0 {
		return nil
	}
	query := t.lookupQuery(len(t.Lookups))
	// fmt.Printf("%s lookup: %d items\n", t.Name(), len(t.Lookups))
	rows, err := tx.Query(query, t.lookupValues()...)
	if err != nil {
		return ectx.Err(query, err)
	}
	// fmt.Printf("%s lookup: resolving rows\n", t.Name())
	return t.resolveRows(rows, nil)
}

func (t *TableLookup) resolveRows(rows *sql.Rows, err error) error {
	if err != nil {
		return err
	}
	defer rows.Close()
	var id int
	var lookupValue string
	for rows.Next() {
		if err = rows.Scan(&id, &lookupValue); err != nil {
			return err
		}
		t.resolveSingleLookup(id, lookupValue)
	}
	return ectx.Err("lookup.resolveRows", rows.Err())
}

func (t *TableLookup) resolveSingleLookup(id int, lookup string) {
	key := t.LookupKey(lookup)
	t.idCache.Add(key, id)
	delete(t.Lookups, key)
}

func (t *TableLookup) lookupValues() []interface{} {
	values := make([]interface{}, len(t.Lookups))
	i := 0
	for _, v := range t.Lookups {
		values[i] = v.Value
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
	buf.WriteString(t.lookupField.SqlName)
	for _, g := range t.Table.DerivedFields {
		buf.WriteString(",")
		buf.WriteString(g.SqlName)
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
	buf.WriteString("returning id, " + t.lookupField.SqlName)
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

func (t *TableLookup) IsFull() bool {
	return len(t.Lookups) >= t.Capacity
}
