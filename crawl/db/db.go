package db

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/crawl/go-sequell/qyaml"
	"github.com/crawl/go-sequell/schema"
)

// A Field represents a field in a game or milestone table.
type Field struct {
	Name             string
	Type             string
	Features         string
	UUID             bool
	SQLName          string
	SQLType          string
	SQLRefType       string
	SQLLookupExpr    string
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

// NeedsIndex checks if a field needs an index on it.
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
		return f.SQLName + "_id"
	}
	return f.SQLName
}

// RefType returns the foreign key type for this field if the field belongs in
// a lookup table, or the direct type if not.
func (f *Field) RefType() string {
	if f.ForeignKeyLookup {
		return f.SQLRefType
	}
	return f.SQLType
}

// RefDefault returns the default unless this is a foreign key.
func (f *Field) RefDefault() string {
	if f.ForeignKeyLookup {
		return ""
	}
	return f.DefaultString
}

// ForeignKeyConstraint returns the foreign key constraint for the field f, if
// f belongs in a lookup table. For fields that are stored directly in the
// fact table, returns nil.
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

// SchemaColumn returns the SQL column spec for this field.
func (f *Field) SchemaColumn() *schema.Column {
	return &schema.Column{
		Name:    f.RefName(),
		SQLType: f.RefType(),
		Default: f.RefDefault(),
	}
}

// LookupSchemaColumn gets the schema column for this field as it would be
// defined in a lookup table. Viz. this is always the true name and type of the
// field, even if the field would normally be stored as a foreign key in a
// fact table.
func (f *Field) LookupSchemaColumn() *schema.Column {
	return &schema.Column{
		Name:    f.SQLName,
		SQLType: f.SQLType,
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
	f.SQLName = parser.FieldSQLName(f.Name)
	f.SQLType, err = parser.FieldSQLType(f.Type)
	f.SQLLookupExpr = parser.FieldSQLLookupExpr(f.SQLName, f.Type)
	if err != nil {
		return
	}

	if f.ForeignKeyLookup {
		f.SQLRefType, err = parser.FieldSQLType("REF")
		if err != nil {
			return
		}
	}

	if !f.PrimaryKey {
		f.DefaultValue = parser.FieldSQLDefault(f.Type)
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

// FieldParser parses SQL field specs from the schema YAML.
type FieldParser struct {
	yaml                qyaml.YAML
	sqlFieldNameMap     map[string]string
	sqlFieldTypes       map[string]string
	sqlFieldDefaults    map[string]string
	sqlFieldLookupExprs map[string]string
}

// NewFieldParser creates a field parser object given the schema YAML.
func NewFieldParser(spec qyaml.YAML) *FieldParser {
	return &FieldParser{
		yaml:                spec,
		sqlFieldNameMap:     spec.StringMap("sql-field-names"),
		sqlFieldTypes:       spec.StringMap("field-types > sql"),
		sqlFieldDefaults:    spec.StringMap("field-types > defaults"),
		sqlFieldLookupExprs: spec.StringMap("field-types > lookup"),
	}
}

var rFieldSpec = regexp.MustCompile(`^([a-z_]+)([A-Z]*)([^\w]*)((?:\[uuid\])?)$`)

// ParseField parses a field spec
func (f *FieldParser) ParseField(spec string) (*Field, error) {
	match := rFieldSpec.FindStringSubmatch(strings.TrimSpace(spec))
	if match == nil {
		return nil, fmt.Errorf("malformed field spec \"%s\"", spec)
	}

	field := &Field{Name: match[1], Type: match[2], Features: match[3], UUID: match[4] != ""}
	err := field.initialize(f)
	if err != nil {
		return nil, err
	}
	return field, nil
}

// FieldSQLLookupExpr gets the lookup table expression for the field named sqlName
func (f *FieldParser) FieldSQLLookupExpr(sqlName string, typeKey string) string {
	lookupExpr := f.sqlFieldLookupExprs[typeKey]
	if lookupExpr == "" {
		return ""
	}
	return strings.Replace(lookupExpr, "%s", sqlName, -1)
}

// FieldSQLName gets the SQL column name for the field named name.
func (f *FieldParser) FieldSQLName(name string) string {
	if sqlName, ok := f.sqlFieldNameMap[name]; ok {
		return sqlName
	}
	return name
}

// FieldSQLNames gets the list of SQL column names for the field names.
func (f *FieldParser) FieldSQLNames(names []string) []string {
	res := make([]string, len(names))
	for i, name := range names {
		res[i] = f.FieldSQLName(name)
	}
	return res
}

// FieldSQLType gets the SQL type of a field given a schema type sigil string.
func (f *FieldParser) FieldSQLType(annotatedType string) (string, error) {
	if sqlType, ok := f.sqlFieldTypes[annotatedType]; ok {
		return sqlType, nil
	}
	return "", fmt.Errorf("FieldSQLType(%#v) undefined", annotatedType)
}

// FieldSQLDefault gets the default value for a field given its sigil type.
func (f *FieldParser) FieldSQLDefault(annotatedType string) string {
	return f.sqlFieldDefaults[annotatedType]
}
