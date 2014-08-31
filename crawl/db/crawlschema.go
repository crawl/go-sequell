package db

import (
	"fmt"
	"strings"

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

type LookupTable CrawlTable

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

	indexes, err := s.ParseCompositeIndexes(name, schemaDef)
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

func (s *CrawlSchema) ParseCompositeIndexes(name string, schemaDef qyaml.Yaml) ([]*Index, error) {
	indexDefs := schemaDef.Slice(name + "-indexes")
	compositeIndexes := make([]*Index, len(indexDefs))
	for i, def := range indexDefs {
		fields := qyaml.IStringSlice(def)
		if len(fields) == 0 {
			return nil, fmt.Errorf("No fields defined for index on %s with spec %#v\n",
				name, def)
		}
		compositeIndexes[i] = &Index{
			Columns: s.FieldParser.FieldSqlNames(fields),
		}
	}
	return compositeIndexes, nil
}

func (s *CrawlSchema) ParseLookupTable(name string, defn interface{}) error {
	switch tdef := defn.(type) {
	case []interface{}:
		fields, err := s.ParseFields(qyaml.IStringSlice(tdef))
		if err != nil {
			return err
		}
		s.AddLookupTable(name, fields, nil)
	case map[interface{}]interface{}:
		fields, err := s.ParseFields(qyaml.IStringSlice(tdef["fields"]))
		if err != nil {
			return err
		}
		generatedFields, err := s.ParseFields(qyaml.IStringSlice(tdef["generated-fields"]))
		if err != nil {
			return err
		}
		s.AddLookupTable(name, fields, generatedFields)
	}
	return nil
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
	fields = append([]*Field{idField}, fields...)
	if generatedFields != nil {
		fields = append(fields, generatedFields...)
	}
	lookupTable := &LookupTable{
		Name:   name,
		Fields: fields,
	}
	for _, field := range fields {
		s.FieldNameLookupTableMap[field.Name] = lookupTable
	}
	s.LookupTables = append(s.LookupTables, lookupTable)
	return lookupTable
}

func IndexName(table string, fields []string) string {
	return "ind_" + table + "_" + strings.Join(fields, "_")
}

func (s *CrawlSchema) Schema() *schema.Schema {
	return &schema.Schema{
		Tables: s.SchemaTables(),
	}
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

func (t *CrawlTable) SchemaTablePrefixed(prefix string) *schema.Table {
	return &schema.Table{
		Name:        prefix + t.Name,
		Columns:     t.SchemaColumns(),
		Indexes:     t.SchemaIndexes(prefix),
		Constraints: t.SchemaConstraints(),
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
			Name:      t.SchemaIndexName(prefix, compIndex.Columns),
			TableName: tableName,
			Columns:   compIndex.Columns,
		})
	}

	for _, field := range t.Fields {
		if field.NeedsIndex() {
			cols := []string{field.RefName()}
			add(&schema.Index{
				Name:      t.SchemaIndexName(prefix, cols),
				TableName: tableName,
				Columns:   cols,
			})
		}
	}

	return indexes
}

func (t *CrawlTable) SchemaIndexName(prefix string, columns []string) string {
	return "ind_" + prefix + t.Name + "_" + strings.Join(columns, "_")
}

func (t *CrawlTable) SchemaConstraints() []schema.Constraint {
	constraints := []schema.Constraint{}
	add := func(c schema.Constraint) {
		if c != nil {
			constraints = append(constraints, c)
		}
	}

	if t.PrimaryKeyField != nil {
		add(schema.PrimaryKeyConstraint{Column: t.PrimaryKeyField.SqlName})
	}
	for _, field := range t.Fields {
		if field.ForeignKeyLookup {
			add(field.ForeignKeyConstraint())
		}
	}

	return constraints
}

func (l *LookupTable) TableName() string {
	return "l_" + l.Name
}

func (l *LookupTable) SchemaTable() *schema.Table {
	return &schema.Table{
		Name:    l.TableName(),
		Columns: l.SchemaColumns(),
		Constraints: []schema.Constraint{
			schema.PrimaryKeyConstraint{Column: "id"},
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
	indexes := []*schema.Index{}
	for _, field := range l.Fields {
		if field.SqlLookupExpr != "" {
			indexes = append(indexes, &schema.Index{
				Name:      IndexName(l.TableName(), []string{field.SqlName}),
				TableName: l.TableName(),
				Columns:   []string{field.SqlLookupExpr},
			})
		}
	}
	return indexes
}
