package db

import (
	"fmt"
	"strings"

	"github.com/greensnark/go-sequell/conv"
	"github.com/greensnark/go-sequell/qyaml"
	"github.com/greensnark/go-sequell/schema"
)

type CrawlSchema struct {
	FieldParser             *FieldParser
	Tables                  []*CrawlTable
	TableVariantPrefixes    []string
	VariantNamePrefixMap    map[string]string
	LookupTables            []*LookupTable
	FieldNameLookupTableMap map[string]*LookupTable
}

type CrawlTable struct {
	Name             string
	Fields           []*Field
	PrimaryKeyField  *Field
	CompositeIndexes []*Index
}

type Index struct {
	Columns []string
}

type LookupTable struct {
	CrawlTable
	ReferencingFields []*Field
	DerivedFields     []*Field
}

func MustLoadSchema(schemaDef qyaml.Yaml) *CrawlSchema {
	sch, err := LoadSchema(schemaDef)
	if err != nil {
		panic(err)
	}
	return sch
}

func LoadSchema(schemaDef qyaml.Yaml) (*CrawlSchema, error) {
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

func (s *CrawlSchema) LookupTable(name string) *LookupTable {
	for _, lt := range s.LookupTables {
		if lt.Name == name {
			return lt
		}
	}
	return nil
}

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

func (s *CrawlSchema) ParseTable(name string, schemaDef qyaml.Yaml) (err error) {
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

func (s *CrawlSchema) ParseTableFields(name string, schemaDef qyaml.Yaml) (tableFields []*Field, err error) {
	annotatedFieldNames := schemaDef.StringSlice(name + "-fields-with-type")
	return s.ParseFields(annotatedFieldNames)
}

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

func (s *CrawlSchema) ParseCompositeIndexes(name string, fields []*Field, schemaDef qyaml.Yaml) ([]*Index, error) {
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

	fieldSqlNames := func(names []string) ([]string, error) {
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
		sqlNames, err := fieldSqlNames(fields)
		if err != nil {
			return nil, err
		}
		compositeIndexes[i] = &Index{
			Columns: sqlNames,
		}
	}
	return compositeIndexes, nil
}

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

func (s *CrawlSchema) FindLookupTableForField(fieldName string) *LookupTable {
	return s.FieldNameLookupTableMap[fieldName]
}

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

func IndexName(table string, fields []string, unique bool) string {
	base := "ind_" + table
	if unique {
		base += "_uniq"
	}
	return base + "_" + strings.Join(fields, "_")
}

func (s *CrawlSchema) Schema() *schema.Schema {
	return &schema.Schema{
		Tables: s.SchemaTables(),
	}
}

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

func (t *CrawlTable) FindField(name string) *Field {
	for _, f := range t.Fields {
		if f.Name == name {
			return f
		}
	}
	return nil
}

func (t *CrawlTable) SchemaTablePrefixed(prefix string) *schema.Table {
	tableName := prefix + t.Name
	return &schema.Table{
		Name:        tableName,
		Columns:     t.SchemaColumns(),
		Indexes:     t.SchemaIndexes(prefix),
		Constraints: t.SchemaConstraints(tableName),
	}
}

func (t *CrawlTable) SchemaColumns() []*schema.Column {
	cols := make([]*schema.Column, len(t.Fields))
	for i, field := range t.Fields {
		cols[i] = field.SchemaColumn()
	}
	return cols
}

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

func (t *CrawlTable) SchemaIndexName(prefix string, columns []string, unique bool) string {
	return IndexName(prefix+t.Name, columns, unique)
}

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
			Column:         t.PrimaryKeyField.SqlName,
		})
	}
	for _, field := range t.Fields {
		if field.ForeignKeyLookup {
			add(field.ForeignKeyConstraint(table))
		}
	}

	return constraints
}

func (l *LookupTable) CaseSensitive() bool {
	return l.LookupField().CaseSensitive
}

// ReferencingFieldCount returns the number of fields in a single row
// that may refer to the lookup table.
func (l *LookupTable) ReferencingFieldCount() int {
	return len(l.ReferencingFields)
}

func (l *LookupTable) LookupField() *Field {
	return l.Fields[1]
}

func (l *LookupTable) TableName() string {
	return "l_" + l.Name
}

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

func (l *LookupTable) SchemaColumns() []*schema.Column {
	cols := make([]*schema.Column, len(l.Fields))
	for i, field := range l.Fields {
		cols[i] = field.LookupSchemaColumn()
	}
	return cols
}

func (l *LookupTable) SchemaIndexes() []*schema.Index {
	indexes := []*schema.Index{
		&schema.Index{
			Name:      IndexName(l.TableName(), []string{l.Fields[1].SqlName}, true),
			TableName: l.TableName(),
			Columns:   []string{l.Fields[1].SqlName},
			Unique:    true,
			Force:     true,
		},
	}
	for _, field := range l.Fields {
		if field.SqlLookupExpr != "" {
			indexes = append(indexes, &schema.Index{
				Name:      IndexName(l.TableName(), []string{field.SqlName}, false),
				TableName: l.TableName(),
				Columns:   []string{field.SqlLookupExpr},
			})
		}
	}
	return indexes
}
