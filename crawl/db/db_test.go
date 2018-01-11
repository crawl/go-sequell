package db

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/crawl/go-sequell/crawl/data"
)

var fieldParseCases = []struct {
	spec  string
	field Field
}{
	{"idIB%*", Field{
		Name:             "id",
		Type:             "PK",
		Features:         "%*",
		SQLName:          "id",
		SQLType:          "serial",
		DefaultString:    "",
		PrimaryKey:       true,
		Summarizable:     false,
		ForeignKeyLookup: false,
		Multivalued:      false,
		Indexed:          false,
	}},

	{"offsetIB*?&", Field{
		Name:          "offset",
		Type:          "IB",
		Features:      "*?&",
		SQLName:       "file_offset",
		SQLType:       "bigint",
		DefaultValue:  "0",
		DefaultString: "default 0",
		Indexed:       true,
		External:      true,
	}},

	{"killermapMAP??^", Field{
		Name:             "killermap",
		Type:             "MAP",
		Features:         "??^",
		SQLName:          "killermap",
		SQLType:          "citext",
		SQLRefType:       "int",
		DefaultString:    "",
		PrimaryKey:       false,
		Summarizable:     true,
		ForeignKeyLookup: true,
		Multivalued:      false,
		Indexed:          true,
		ForceIndex:       true,
	}},

	{"tiles!", Field{
		Name:             "tiles",
		Type:             "!",
		Features:         "!",
		SQLName:          "tiles",
		SQLType:          "boolean",
		DefaultString:    "default false",
		DefaultValue:     "false",
		PrimaryKey:       false,
		Summarizable:     true,
		ForeignKeyLookup: false,
		Multivalued:      false,
		Indexed:          false,
	}},

	{"nameS?^", Field{
		Name:             "name",
		Type:             "S",
		Features:         "?^",
		SQLName:          "pname",
		SQLType:          "text",
		SQLRefType:       "int",
		SQLLookupExpr:    "cast(pname as citext)",
		DefaultString:    "",
		CaseSensitive:    true,
		PrimaryKey:       false,
		Summarizable:     true,
		ForeignKeyLookup: true,
		Multivalued:      false,
		Indexed:          true,
	}},

	{"maxskills^+", Field{
		Name:             "maxskills",
		Type:             "TEXT",
		Features:         "^+",
		SQLName:          "maxskills",
		SQLType:          "citext",
		SQLRefType:       "int",
		DefaultString:    "",
		PrimaryKey:       false,
		Summarizable:     true,
		ForeignKeyLookup: true,
		Multivalued:      true,
		Indexed:          false,
	}},

	{"hash?^[uuid]", Field{
		Name:             "hash",
		Type:             "TEXT",
		Features:         "?^",
		SQLName:          "hash",
		SQLType:          "citext",
		SQLRefType:       "int",
		DefaultString:    "",
		UUID:             true,
		Summarizable:     true,
		ForeignKeyLookup: true,
		Indexed:          true,
	}},
}

func TestParseField(t *testing.T) {
	p := NewFieldParser(data.CrawlSchema().YAML)
	for _, testCase := range fieldParseCases {
		t.Run(fmt.Sprintf("ParseField(%#v)", testCase.spec), func(t *testing.T) {
			field, err := p.ParseField(testCase.spec)
			if err != nil {
				t.Fatalf("Error parsing %#v: %v", field, err)
			}
			if !reflect.DeepEqual(&testCase.field, field) {
				t.Errorf("Expected %#v to parse as %#v, but got %#v",
					testCase.spec, &testCase.field, field)
			}
		})
	}
}
