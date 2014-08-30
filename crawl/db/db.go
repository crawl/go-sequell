package db

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/greensnark/go-sequell/crawl/data"
)

type Field struct {
	Name             string
	Type             string
	Features         string
	SqlName          string
	SqlType          string
	DefaultString    string
	PrimaryKey       bool
	Summarizable     bool
	ForeignKeyLookup bool
	Multivalued      bool
	Indexed          bool
}

var rFieldSpec = regexp.MustCompile(`^([a-z]+)([A-Z]*)([^\w]*)$`)

func ParseField(spec string) (*Field, error) {
	match := rFieldSpec.FindStringSubmatch(strings.TrimSpace(spec))
	if match == nil {
		return nil, fmt.Errorf("malformed field spec \"%s\"", spec)
	}

	field := &Field{Name: match[1], Type: match[2], Features: match[3]}
	err := field.initialize()
	if err != nil {
		return nil, err
	}
	return field, nil
}

func (f *Field) initialize() (err error) {
	f.Summarizable = true
	for _, feat := range f.Features {
		f.applyFeature(feat)
	}
	if f.Type == "" {
		f.Type = "TEXT"
	}

	if f.Type == "PK" {
		f.PrimaryKey = true
	}
	f.SqlName = FieldSqlName(f.Name)
	f.SqlType, err = FieldSqlType(f.Type)
	if err != nil {
		return
	}
	if !f.PrimaryKey {
		f.DefaultString = FieldSqlDefault(f.Type)
	}
	return
}

func (f *Field) applyFeature(feat rune) {
	switch feat {
	case '%':
		f.PrimaryKey = true
	case '!':
		f.Type = "!"
	case '^':
		f.ForeignKeyLookup = true
	case '+':
		f.Multivalued = true
	case '*':
		f.Summarizable = false
	case '?':
		f.Indexed = true
	}
}

var sqlFieldNameMap = data.Schema.StringMap("sql-field-names")

func FieldSqlName(name string) string {
	if sqlName, ok := sqlFieldNameMap[name]; ok {
		return sqlName
	}
	return name
}

var sqlFieldTypes = data.Schema.StringMap("field-types > sql")

func FieldSqlType(annotatedType string) (string, error) {
	if sqlType, ok := sqlFieldTypes[annotatedType]; ok {
		return sqlType, nil
	}
	return "", fmt.Errorf("FieldSqlType(%#v) undefined", annotatedType)
}

var sqlFieldDefaults = data.Schema.StringMap("field-types > defaults")

func FieldSqlDefault(annotatedType string) string {
	return sqlFieldDefaults[annotatedType]
}
