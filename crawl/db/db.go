package db

import (
	"fmt"
	"regexp"
	"strings"
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
}

var rFieldSpec = regexp.MustCompile(`^([a-z]+)([A-Z]*)([^\w]*)$`)

func ParseField(spec string) (*Field, error) {
	match := rFieldSpec.FindStringSubmatch(strings.TrimSpace(spec))
	if match == nil {
		return nil, fmt.Errorf("malformed field spec \"%s\"", spec)
	}

	field := &Field{Name: match[1], Type: match[2], Features: match[2]}
	field.initialize()
	return field, nil
}

func (f *Field) initialize() {

}
