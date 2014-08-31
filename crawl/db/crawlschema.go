package db

import (
	"github.com/greensnark/go-sequell/qyaml"
	"github.com/greensnark/go-sequell/schema"
)

type CrawlSchema struct {
	FieldParser  *FieldParser
	Tables       []*CrawlTable
	LookupTables []*LookupTable
}

type CrawlTable struct {
	Name   string
	Fields []*Field
}

type LookupTable CrawlTable

func LoadSchema(schemaDef qyaml.Yaml) *CrawlSchema {
	schema := CrawlSchema{FieldParser: NewFieldParser(schemaDef)}
	for _, table := range schemaDef.StringSlice("query-tables") {
		schema.ParseTable(table, schemaDef)
	}
	for iname, definition := range schemaDef.Map("lookup-tables") {
		name := iname.(string)
		schema.ParseLookupTable(name, definition)
	}
	return &schema
}

func (s *CrawlSchema) ParseTable(name string, schemaDef qyaml.Yaml) {

}

func (s *CrawlSchema) ParseLookupTable(name string, defn interface{}) {
}

func (s *CrawlSchema) Schema() *schema.Schema {
	return nil
}
