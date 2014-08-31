package db

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/greensnark/go-sequell/qyaml"
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

type FieldParser struct {
	yaml             qyaml.Yaml
	sqlFieldNameMap  map[string]string
	sqlFieldTypes    map[string]string
	sqlFieldDefaults map[string]string
}

func NewFieldParser(spec qyaml.Yaml) *FieldParser {
	return &FieldParser{
		yaml:             spec,
		sqlFieldNameMap:  spec.StringMap("sql-field-names"),
		sqlFieldTypes:    spec.StringMap("field-types > sql"),
		sqlFieldDefaults: spec.StringMap("field-types > defaults"),
	}
}

var rFieldSpec = regexp.MustCompile(`^([a-z]+)([A-Z]*)([^\w]*)$`)

func (f *FieldParser) ParseField(spec string) (*Field, error) {
	match := rFieldSpec.FindStringSubmatch(strings.TrimSpace(spec))
	if match == nil {
		return nil, fmt.Errorf("malformed field spec \"%s\"", spec)
	}

	field := &Field{Name: match[1], Type: match[2], Features: match[3]}
	err := field.initialize(f)
	if err != nil {
		return nil, err
	}
	return field, nil
}

func (f *Field) initialize(parser *FieldParser) (err error) {
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
	f.SqlName = parser.FieldSqlName(f.Name)
	f.SqlType, err = parser.FieldSqlType(f.Type)
	if err != nil {
		return
	}
	if !f.PrimaryKey {
		f.DefaultString = parser.FieldSqlDefault(f.Type)
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

func (p *FieldParser) FieldSqlName(name string) string {
	if sqlName, ok := p.sqlFieldNameMap[name]; ok {
		return sqlName
	}
	return name
}

func (p *FieldParser) FieldSqlType(annotatedType string) (string, error) {
	if sqlType, ok := p.sqlFieldTypes[annotatedType]; ok {
		return sqlType, nil
	}
	return "", fmt.Errorf("FieldSqlType(%#v) undefined", annotatedType)
}

func (p *FieldParser) FieldSqlDefault(annotatedType string) string {
	return p.sqlFieldDefaults[annotatedType]
}
