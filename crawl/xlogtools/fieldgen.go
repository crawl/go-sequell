package xlogtools

import (
	"fmt"
	"log"

	"github.com/crawl/go-sequell/conv"
	"github.com/crawl/go-sequell/stringnorm"
	"github.com/crawl/go-sequell/text"
	"github.com/crawl/go-sequell/xlog"
	"github.com/pkg/errors"
)

var (
	// ErrBadFieldIfMatchCondition means an if-match field condition in a
	// crawl.yml field-input-transforms / if-match condition was malformed.
	ErrBadFieldIfMatchCondition = errors.New("bad if-match condition")
)

// FieldGenCondition matches xlogs that need a field to be generated.
type FieldGenCondition interface {
	Matches(x xlog.Xlog) bool
}

// FieldGen describes how a target field is generated from a source field.
type FieldGen struct {
	SourceField string
	TargetField string

	Conditions []FieldGenCondition
	Transforms stringnorm.Normalizer
}

// Matches checks if the xlog x needs a new generated field
func (f *FieldGen) Matches(x xlog.Xlog) bool {
	if len(f.Conditions) == 0 {
		return true
	}
	for _, c := range f.Conditions {
		if !c.Matches(x) {
			return false
		}
	}
	return true
}

// Apply generates the target field in x based on the source field.
func (f *FieldGen) Apply(x xlog.Xlog) {
	if !f.Matches(x) {
		return
	}
	src := x[f.SourceField]
	result, err := f.Transforms.Normalize(src)
	if err != nil {
		log.Printf("Error gen %s[%s]->%s: %s in %#v\n",
			f.SourceField, src, f.TargetField, err, x)
		return
	}
	x[f.TargetField] = result
}

type fieldEqValueCondition struct {
	field string
	value string
}

func (f fieldEqValueCondition) Matches(x xlog.Xlog) bool {
	return x.Get(f.field) == f.value
}

// parseFieldMatchCondition parses an individual field match condition of the form
// { field: X, equal: Y }
func parseFieldMatchCondition(ifCond interface{}) (FieldGenCondition, error) {
	condErr := func() error {
		return errors.Wrapf(ErrBadFieldIfMatchCondition, "parseFieldMatchCondition(%#v)", ifCond)
	}

	fieldMatch, ok := ifCond.(map[interface{}]interface{})
	if !ok {
		return nil, condErr()
	}

	fieldName, ok := fieldMatch["field"].(string)
	if !ok {
		return nil, condErr()
	}

	expectedValue, ok := fieldMatch["equal"].(string)
	if !ok {
		return nil, condErr()
	}

	return fieldEqValueCondition{field: fieldName, value: expectedValue}, nil
}

// ParseFieldMatchConditions parses a list of condition block in the form: [{"field":
// "field-name", "equal": "value"}, ...] into a list of FieldGenConditions
func ParseFieldMatchConditions(ifMatchList interface{}) (fieldGenConditions []FieldGenCondition, err error) {
	if ifMatchList == nil {
		return nil, nil
	}

	matchConditions, ok := ifMatchList.([]interface{})
	if !ok {
		return nil, errors.Wrapf(ErrBadFieldIfMatchCondition, "ParseFieldMatchConditions(%#v)", ifMatchList)
	}

	for _, matchCondition := range matchConditions {
		cond, err := parseFieldMatchCondition(matchCondition)
		if err != nil {
			return nil, errors.Wrapf(err, "ParseFieldMatchConditions(%#v) (in %#v)", ifMatchList, matchCondition)
		}
		fieldGenConditions = append(fieldGenConditions, cond)
	}

	return fieldGenConditions, nil
}

// ParseFieldGenerators parses a list of field generators from a
// field-input-transforms definition in crawl-data.yml
func ParseFieldGenerators(fieldGenExpressions []interface{}) ([]*FieldGen, error) {
	res := make([]*FieldGen, 0, len(fieldGenExpressions))
	for _, fieldGenExpr := range fieldGenExpressions {
		fieldGens, ok := fieldGenExpr.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("Unexpected config for generated fields: %#v (in %#v)", fieldGenExpr, fieldGenExpressions)
		}

		for targetFieldI, cfg := range fieldGens {
			cfgMap, ok := cfg.(map[interface{}]interface{})
			if !ok {
				return nil, fmt.Errorf("Unexpected config for generated field %s: %#v", targetFieldI, cfg)
			}
			targetField := text.Str(targetFieldI)
			sourceField := text.FirstNotEmpty(text.Str(cfgMap["source"]), targetField)

			stringForms, err := ParseStringReplacements(cfgMap["string-replace"])
			if err != nil {
				return nil, err
			}
			regexpForms, err := ParseRegexpReplacements(cfgMap["regexp-replace"])
			if err != nil {
				return nil, err
			}

			conditionChecks, err := ParseFieldMatchConditions(cfgMap["if-match"])
			if err != nil {
				return nil, err
			}

			if stringForms == nil && regexpForms == nil {
				return nil, fmt.Errorf("No transforms for field %s in %#v",
					targetField, cfg)
			}

			xforms := stringnorm.Combine(stringForms, regexpForms)
			res = append(res, &FieldGen{
				SourceField: sourceField,
				TargetField: targetField,
				Conditions:  conditionChecks,
				Transforms:  xforms,
			})
		}
	}

	return res, nil
}

// ParseRegexpReplacements parses regexp replacements into a string normalizer.
func ParseRegexpReplacements(regexps interface{}) (stringnorm.Normalizer, error) {
	if regexps == nil {
		return nil, nil
	}
	return stringnorm.ParseRegexpPairs(conv.IStringPairs(regexps))
}

// ParseStringReplacements parses string replacements into a string normalizer.
func ParseStringReplacements(replacements interface{}) (stringnorm.Normalizer, error) {
	if replacements == nil {
		return nil, nil
	}
	return stringnorm.ParseExactReplacers(conv.IStringPairs(replacements))
}
