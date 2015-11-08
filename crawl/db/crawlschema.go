package db

import (
	"fmt"
	"strings"

	"github.com/crawl/go-sequell/conv"
	"github.com/crawl/go-sequell/qyaml"
	"github.com/crawl/go-sequell/schema"
)

// A CrawlSchema is a schema modeling the tables that track games and milestones
type CrawlSchema struct {
	FieldParser             *FieldParser
	Tables                  []*CrawlTable
	TableVariantPrefixes    []string
	VariantNamePrefixMap    map[string]string
	LookupTables            []*LookupTable
	FieldNameLookupTableMap map[string]*LookupTable
}

// A CrawlTable represents a Crawl fact or dimension table. Game/Milestone
// tables are fact tables, lookup tables are dimension tables.
type CrawlTable struct {
	Name             string
	Fields           []*Field
	PrimaryKeyField  *Field
	CompositeIndexes []*Index
}

// An Index represents a SQL index on a Crawl table.
type Index struct {
	Columns []string
}

// A LookupTable is a dimension table
type LookupTable struct {
	CrawlTable
	ReferencingFields []*Field
	DerivedFields     []*Field
}

// MustLoadSchema loads a Crawl schema from the schemaDef YAML, panicking on
// error.
func MustLoadSchema(schemaDef qyaml.YAML) *CrawlSchema {
	sch, err := LoadSchema(schemaDef)
	if err != nil {
		panic(err)
	}
	return sch
}

// LoadSchema loads a CrawlSchema from the schemaDef YAML.
func LoadSchema(schemaDef qyaml.YAML) (*CrawlSchema, error) {
	schema := CrawlSchema{
		FieldParser:             NewFieldParser(schemaDef),
		FieldNameLookupTableMap: map[string]*LookupTable{},
	}
	schema.ParseVariants(schemaDef.StringMap("game-type-prefixes"))
	for iname, definition := range schemaDef.Map("lookup-tables") {
		name := iname.(string)
		err := schema.ParseLookupTable(name, definition)
		if err != nil {
			return nil, err
		}
	}
	for _, table := range schemaDef.StringSlice("query-tables") {
		if err := schema.ParseTable(table, schemaDef); err != nil {
			return nil, err
		}
	}
	return &schema, nil
}

// PrefixedTablesWithField returns the list of tables containing field,
// prefixing them with their table variant prefixes.
func (s *CrawlSchema) PrefixedTablesWithField(field string) []*CrawlTable {
	res := []*CrawlTable{}
	for _, t := range s.Tables {
		if t.FindField(field) != nil {
			for _, variant := range s.TableVariantPrefixes {
				copy := *t
				copy.Name = variant + copy.Name
				res = append(res, &copy)
			}
		}
	}
	return res
}

// LookupTable gets a lookup table by name.
func (s *CrawlSchema) LookupTable(name string) *LookupTable {
	for _, lt := range s.LookupTables {
		if lt.Name == name {
			return lt
		}
	}
	return nil
}

// ParseVariants parses the game variant map.
func (s *CrawlSchema) ParseVariants(variantMap map[string]string) {
	s.VariantNamePrefixMap = variantMap
	s.TableVariantPrefixes = make([]string, len(variantMap))
	i := 0
	for _, prefix := range variantMap {
		s.TableVariantPrefixes[i] = prefix
		i++
	}
	if len(s.TableVariantPrefixes) == 0 {
		s.TableVariantPrefixes = []string{""}
	}
}

// ParseTable reads a table schema from the schema file for the given table name.
func (s *CrawlSchema) ParseTable(name string, schemaDef qyaml.YAML) (err error) {
	fields, err := s.ParseTableFields(name, schemaDef)
	if err != nil {
		return err
	}

	indexes, err := s.ParseCompositeIndexes(name, fields, schemaDef)
	if err != nil {
		return err
	}
	table := &CrawlTable{
		Name:             name,
		Fields:           fields,
		CompositeIndexes: indexes,
	}
	for _, field := range fields {
		if field.PrimaryKey {
			table.PrimaryKeyField = field
			break
		}
	}
	s.RegisterLookupTables(fields)
	s.Tables = append(s.Tables, table)
	return nil
}

// ParseTableFields parses the table field defs from the schema for the named table.
func (s *CrawlSchema) ParseTableFields(name string, schemaDef qyaml.YAML) (tableFields []*Field, err error) {
	annotatedFieldNames := schemaDef.StringSlice(name + "-fields-with-type")
	return s.ParseFields(annotatedFieldNames)
}

// ParseFields parses a list of annotated field names into Field objects.
func (s *CrawlSchema) ParseFields(annotatedFieldNames []string) (tableFields []*Field, err error) {
	tableFields = make([]*Field, len(annotatedFieldNames))
	for i, field := range annotatedFieldNames {
		tableFields[i], err = s.FieldParser.ParseField(field)
		if err != nil {
			return nil, err
		}
	}
	return
}

// ParseCompositeIndexes parses composite index definitions from the schema.
func (s *CrawlSchema) ParseCompositeIndexes(name string, fields []*Field, schemaDef qyaml.YAML) ([]*Index, error) {
	indexDefs := schemaDef.Slice(name + "-indexes")
	compositeIndexes := make([]*Index, len(indexDefs))

	findField := func(name string) *Field {
		for _, f := range fields {
			if f.Name == name {
				return f
			}
		}
		return nil
	}

	fieldSQLNames := func(names []string) ([]string, error) {
		res := make([]string, len(names))
		for i, name := range names {
			field := findField(name)
			if field == nil {
				return nil, fmt.Errorf("no field definition for '%s'", name)
			}
			res[i] = field.RefName()
		}
		return res, nil
	}

	for i, def := range indexDefs {
		fields := conv.IStringSlice(def)
		if len(fields) == 0 {
			return nil, fmt.Errorf("No fields defined for index on %s with spec %#v\n",
				name, def)
		}
		sqlNames, err := fieldSQLNames(fields)
		if err != nil {
			return nil, err
		}
		compositeIndexes[i] = &Index{
			Columns: sqlNames,
		}
	}
	return compositeIndexes, nil
}

// ParseLookupTable parses a lookup table definition
func (s *CrawlSchema) ParseLookupTable(name string, defn interface{}) error {
	switch tdef := defn.(type) {
	case []interface{}:
		fields, err := s.ParseFields(conv.IStringSlice(tdef))
		if err != nil {
			return err
		}
		s.AddLookupTable(name, s.markReferenceFields(fields), nil)
	case map[interface{}]interface{}:
		fields, err := s.ParseFields(conv.IStringSlice(tdef["fields"]))
		if err != nil {
			return err
		}
		generatedFields, err := s.ParseFields(
			conv.IStringSlice(tdef["generated-fields"]))
		if err != nil {
			return err
		}
		s.AddLookupTable(name, s.markReferenceFields(fields), generatedFields)
	}
	return nil
}

func (s *CrawlSchema) markReferenceFields(fields []*Field) []*Field {
	for _, f := range fields {
		f.ForeignKeyLookup = true
	}
	return fields
}

// RegisterLookupTables registers lookup tables for all fields that belong
// in a lookup table.
func (s *CrawlSchema) RegisterLookupTables(fields []*Field) {
	for _, field := range fields {
		if field.ForeignKeyLookup {
			lookup := s.FindLookupTableForField(field.Name)
			if lookup == nil {
				lookup = s.AddLookupTable(field.Name, []*Field{field}, nil)
			}
			field.ForeignKeyTable = lookup.TableName()
		}
	}
}

// FindLookupTableForField returns the lookup table for fieldName, or nil if
// no lookup table exists.
func (s *CrawlSchema) FindLookupTableForField(fieldName string) *LookupTable {
	return s.FieldNameLookupTableMap[fieldName]
}

// AddLookupTable creates and registers a lookup table, returning the new table.
func (s *CrawlSchema) AddLookupTable(name string, fields []*Field, generatedFields []*Field) *LookupTable {
	idField, err := s.FieldParser.ParseField("idIB%*")
	if err != nil {
		panic(err)
	}
	tableFields := make([]*Field, 2+len(generatedFields))
	tableFields[0] = idField
	tableFields[1] = fields[0]
	for i, g := range generatedFields {
		tableFields[i+2] = g
	}

	logDerivedFields := make([]*Field, 0, len(generatedFields))
	for _, f := range generatedFields {
		if !f.External {
			logDerivedFields = append(logDerivedFields, f)
		}
	}
	lookupTable := &LookupTable{
		CrawlTable: CrawlTable{
			Name:   name,
			Fields: tableFields,
		},
		ReferencingFields: fields,
		DerivedFields:     logDerivedFields,
	}
	for _, field := range fields {
		s.FieldNameLookupTableMap[field.Name] = lookupTable
	}
	s.LookupTables = append(s.LookupTables, lookupTable)
	return lookupTable
}

// IndexName gets the name for the index on table (fields...).
func IndexName(table string, fields []string, unique bool) string {
	base := "ind_" + table
	if unique {
		base += "_uniq"
	}
	return base + "_" + strings.Join(fields, "_")
}

// Schema creates a SQL schema for this Crawl schema.
func (s *CrawlSchema) Schema() *schema.Schema {
	return &schema.Schema{
		Tables: s.SchemaTables(),
	}
}

// PrimaryTableNames returns the names of the fact tables (viz. games and
// milestones), with all Crawl game variants (i.e. logrecord, spr_logrecord,
// etc.)
func (s *CrawlSchema) PrimaryTableNames() []string {
	names := make([]string, len(s.Tables)*len(s.TableVariantPrefixes))
	i := 0
	for _, prefix := range s.TableVariantPrefixes {
		for _, table := range s.Tables {
			names[i] = prefix + table.Name
			i++
		}
	}
	return names
}

// Table gets the CrawlTable object for the named table.
func (s *CrawlSchema) Table(name string) *CrawlTable {
	for _, table := range s.Tables {
		if table.Name == name {
			return table
		}
	}
	return nil
}

// SchemaTables gets the list of table SQL schema objects.
func (s *CrawlSchema) SchemaTables() []*schema.Table {
	tables := make([]*schema.Table,
		len(s.LookupTables)+len(s.Tables)*len(s.TableVariantPrefixes))
	for i, lookup := range s.LookupTables {
		tables[i] = lookup.SchemaTable()
	}

	i := len(s.LookupTables)
	for _, prefix := range s.TableVariantPrefixes {
		for _, table := range s.Tables {
			tables[i] = table.SchemaTablePrefixed(prefix)
			i++
		}
	}

	return tables
}

// FindField looks up the field definition given a field name.
func (t *CrawlTable) FindField(name string) *Field {
	for _, f := range t.Fields {
		if f.Name == name {
			return f
		}
	}
	return nil
}

// SchemaTablePrefixed gets the schema table object for this table, prefixing
// the name with prefix.
func (t *CrawlTable) SchemaTablePrefixed(prefix string) *schema.Table {
	tableName := prefix + t.Name
	return &schema.Table{
		Name:        tableName,
		Columns:     t.SchemaColumns(),
		Indexes:     t.SchemaIndexes(prefix),
		Constraints: t.SchemaConstraints(tableName),
	}
}

// SchemaColumns gets the list of columns in this table.
func (t *CrawlTable) SchemaColumns() []*schema.Column {
	cols := make([]*schema.Column, len(t.Fields))
	for i, field := range t.Fields {
		cols[i] = field.SchemaColumn()
	}
	return cols
}

// SchemaIndexes gets the list of indexes for this table.
func (t *CrawlTable) SchemaIndexes(prefix string) []*schema.Index {
	indexes := []*schema.Index{}
	add := func(i *schema.Index) {
		if i != nil {
			indexes = append(indexes, i)
		}
	}

	tableName := prefix + t.Name
	for _, compIndex := range t.CompositeIndexes {
		add(&schema.Index{
			Name:      t.SchemaIndexName(prefix, compIndex.Columns, false),
			TableName: tableName,
			Columns:   compIndex.Columns,
		})
	}

	for _, field := range t.Fields {
		if field.NeedsIndex() {
			cols := []string{field.RefName()}
			add(&schema.Index{
				Name:      t.SchemaIndexName(prefix, cols, false),
				TableName: tableName,
				Columns:   cols,
			})
		}
	}

	return indexes
}

// SchemaIndexName gets the name for the index for the given columns.
func (t *CrawlTable) SchemaIndexName(prefix string, columns []string, unique bool) string {
	return IndexName(prefix+t.Name, columns, unique)
}

// SchemaConstraints gets the list of schema constraints for this table.
func (t *CrawlTable) SchemaConstraints(table string) []schema.Constraint {
	constraints := []schema.Constraint{}
	add := func(c schema.Constraint) {
		if c != nil {
			constraints = append(constraints, c)
		}
	}

	if t.PrimaryKeyField != nil {
		add(schema.PrimaryKeyConstraint{
			ConstraintName: table + "_pk",
			Column:         t.PrimaryKeyField.SQLName,
		})
	}
	for _, field := range t.Fields {
		if field.ForeignKeyLookup {
			add(field.ForeignKeyConstraint(table))
		}
	}

	return constraints
}

// CaseSensitive checks if this table's lookup field is case sensitive.
func (l *LookupTable) CaseSensitive() bool {
	return l.LookupField().CaseSensitive
}

// ReferencingFieldCount returns the number of fields in a single row
// that may refer to the lookup table.
func (l *LookupTable) ReferencingFieldCount() int {
	return len(l.ReferencingFields)
}

// LookupField gets the primary field for this lookup table.
func (l *LookupTable) LookupField() *Field {
	return l.Fields[1]
}

// TableName gets the SQL name of this lookup table.
func (l *LookupTable) TableName() string {
	return "l_" + l.Name
}

// SchemaTable gets the table definition for this table.
func (l *LookupTable) SchemaTable() *schema.Table {
	return &schema.Table{
		Name:    l.TableName(),
		Columns: l.SchemaColumns(),
		Constraints: []schema.Constraint{
			schema.PrimaryKeyConstraint{
				ConstraintName: l.TableName() + "_pk",
				Column:         "id",
			},
		},
		Indexes: l.SchemaIndexes(),
	}
}

// SchemaColumns gets the list of columns in this lookup table.
func (l *LookupTable) SchemaColumns() []*schema.Column {
	cols := make([]*schema.Column, len(l.Fields))
	for i, field := range l.Fields {
		cols[i] = field.LookupSchemaColumn()
	}
	return cols
}

// SchemaIndexes gets the list of indexes for this lookup table.
func (l *LookupTable) SchemaIndexes() []*schema.Index {
	indexes := []*schema.Index{
		&schema.Index{
			Name:      IndexName(l.TableName(), []string{l.Fields[1].SQLName}, true),
			TableName: l.TableName(),
			Columns:   []string{l.Fields[1].SQLName},
			Unique:    true,
			Force:     true,
		},
	}
	for _, field := range l.Fields {
		if field.SQLLookupExpr != "" {
			indexes = append(indexes, &schema.Index{
				Name:      IndexName(l.TableName(), []string{field.SQLName}, false),
				TableName: l.TableName(),
				Columns:   []string{field.SQLLookupExpr},
			})
		}
	}
	return indexes
}
