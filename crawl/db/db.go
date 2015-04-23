package db

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/crawl/go-sequell/qyaml"
	"github.com/crawl/go-sequell/schema"
)

type Field struct {
	Name             string
	Type             string
	Features         string
	SqlName          string
	SqlType          string
	SqlRefType       string
	SqlLookupExpr    string
	DefaultValue     string
	DefaultString    string
	PrimaryKey       bool
	CaseSensitive    bool
	Summarizable     bool
	ForeignKeyLookup bool
	ForeignKeyTable  string
	Multivalued      bool
	Indexed          bool
	ForceIndex       bool
	External         bool
}

func (f *Field) NeedsIndex() bool {
	return f.ForeignKeyLookup || f.Indexed
}

// ResolvedKey returns the key name in the xlog for this field after
// lookups have been resolved, viz. the RefName() for foreign key
// fields and the simple name for other fields.
func (f *Field) ResolvedKey() string {
	if f.ForeignKeyLookup {
		return f.RefName()
	}
	return f.Name
}

// RefName returns the SQL column name for this field in the primary
// table, being the foreign key column name for reference fields and
// the simple SQL name for other fields.
func (f *Field) RefName() string {
	if f.ForeignKeyLookup {
		return f.SqlName + "_id"
	}
	return f.SqlName
}

func (f *Field) RefType() string {
	if f.ForeignKeyLookup {
		return f.SqlRefType
	}
	return f.SqlType
}

func (f *Field) RefDefault() string {
	if f.ForeignKeyLookup {
		return ""
	}
	return f.DefaultString
}

func (f *Field) ForeignKeyConstraint(table string) schema.Constraint {
	if !f.ForeignKeyLookup || f.ForeignKeyTable == "" {
		return nil
	}
	return schema.ForeignKeyConstraint{
		ConstraintName:   table + "_" + f.RefName() + "_fk",
		SourceTableField: f.RefName(),
		TargetTable:      f.ForeignKeyTable,
		TargetTableField: "id",
	}
}

func (f *Field) SchemaColumn() *schema.Column {
	return &schema.Column{
		Name:    f.RefName(),
		SqlType: f.RefType(),
		Default: f.RefDefault(),
	}
}

func (f *Field) LookupSchemaColumn() *schema.Column {
	return &schema.Column{
		Name:    f.SqlName,
		SqlType: f.SqlType,
	}
}

func (f *Field) initialize(parser *FieldParser) (err error) {
	f.Summarizable = true
	for _, feat := range f.Features {
		f.applyFeature(feat)
	}
	if f.Type == "" {
		f.Type = "TEXT"
	}
	if f.Type == "S" {
		f.CaseSensitive = true
	}
	if f.PrimaryKey {
		f.Type = "PK"
	} else if f.Type == "PK" {
		f.PrimaryKey = true
	}
	f.SqlName = parser.FieldSqlName(f.Name)
	f.SqlType, err = parser.FieldSqlType(f.Type)
	f.SqlLookupExpr = parser.FieldSqlLookupExpr(f.SqlName, f.Type)
	if err != nil {
		return
	}

	if f.ForeignKeyLookup {
		f.SqlRefType, err = parser.FieldSqlType("REF")
		if err != nil {
			return
		}
	}

	if !f.PrimaryKey {
		f.DefaultValue = parser.FieldSqlDefault(f.Type)
		if f.DefaultValue != "" {
			f.DefaultString = "default " + f.DefaultValue
		}
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
	case '&':
		f.External = true
	case '?':
		if f.Indexed {
			f.ForceIndex = true
		}
		f.Indexed = true
	}
}

type FieldParser struct {
	yaml                qyaml.Yaml
	sqlFieldNameMap     map[string]string
	sqlFieldTypes       map[string]string
	sqlFieldDefaults    map[string]string
	sqlFieldLookupExprs map[string]string
}

func NewFieldParser(spec qyaml.Yaml) *FieldParser {
	return &FieldParser{
		yaml:                spec,
		sqlFieldNameMap:     spec.StringMap("sql-field-names"),
		sqlFieldTypes:       spec.StringMap("field-types > sql"),
		sqlFieldDefaults:    spec.StringMap("field-types > defaults"),
		sqlFieldLookupExprs: spec.StringMap("field-types > lookup"),
	}
}

var rFieldSpec = regexp.MustCompile(`^([a-z_]+)([A-Z]*)([^\w]*)$`)

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

func (p *FieldParser) FieldSqlLookupExpr(sqlName string, typeKey string) string {
	lookupExpr := p.sqlFieldLookupExprs[typeKey]
	if lookupExpr == "" {
		return ""
	}
	return strings.Replace(lookupExpr, "%s", sqlName, -1)
}

func (p *FieldParser) FieldSqlName(name string) string {
	if sqlName, ok := p.sqlFieldNameMap[name]; ok {
		return sqlName
	}
	return name
}

func (p *FieldParser) FieldSqlNames(names []string) []string {
	res := make([]string, len(names))
	for i, name := range names {
		res[i] = p.FieldSqlName(name)
	}
	return res
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
