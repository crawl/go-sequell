package db

import (
	"reflect"
	"testing"
)

var fieldParseCases = []struct {
	spec  string
	field Field
}{
	{"idIB%*", Field{
		Name:             "id",
		Type:             "IB",
		Features:         "%*",
		SqlName:          "id",
		SqlType:          "bigint",
		DefaultString:    "",
		PrimaryKey:       true,
		Summarizable:     false,
		ForeignKeyLookup: false,
		Multivalued:      false,
		Indexed:          false,
	}},

	{"killermapMAP?^", Field{
		Name:             "killermap",
		Type:             "MAP",
		Features:         "?^",
		SqlName:          "killermap",
		SqlType:          "citext",
		DefaultString:    "",
		PrimaryKey:       false,
		Summarizable:     true,
		ForeignKeyLookup: true,
		Multivalued:      false,
		Indexed:          true,
	}},

	{"tiles!", Field{
		Name:             "tiles",
		Type:             "!",
		Features:         "!",
		SqlName:          "tiles",
		SqlType:          "boolean",
		DefaultString:    "",
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
		SqlName:          "pname",
		SqlType:          "text",
		DefaultString:    "",
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
		SqlName:          "maxskills",
		SqlType:          "citext",
		DefaultString:    "",
		PrimaryKey:       false,
		Summarizable:     true,
		ForeignKeyLookup: true,
		Multivalued:      true,
		Indexed:          false,
	}},
}

func TestParseField(t *testing.T) {
	for _, testCase := range fieldParseCases {
		field, err := ParseField(testCase.spec)
		if err != nil {
			t.Errorf("Error parsing %#v: %v", field, err)
			continue
		}
		if !reflect.DeepEqual(&testCase.field, field) {
			t.Errorf("Expected %#v to parse as %#v, but got %#v",
				testCase.spec, testCase.field, field)
		}
	}
}
