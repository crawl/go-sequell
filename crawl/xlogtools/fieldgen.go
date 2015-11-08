package xlogtools

import (
	"fmt"
	"log"

	"github.com/crawl/go-sequell/conv"
	"github.com/crawl/go-sequell/stringnorm"
	"github.com/crawl/go-sequell/text"
	"github.com/crawl/go-sequell/xlog"
)

// FieldGenCondition matches xlogs that need a field to be generated.
type FieldGenCondition interface {
	Matches(x xlog.Xlog) bool
}

// FieldGen describes how a target field is generated from a source field.
type FieldGen struct {
	SourceField string
	TargetField string
	Conditions  []FieldGenCondition
	Transforms  stringnorm.Normalizer
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

// ParseFieldGenerators parses a list of field generators from a
// field-input-transforms definition in crawl-data.yml
func ParseFieldGenerators(genmap map[interface{}]interface{}) ([]*FieldGen, error) {
	res := make([]*FieldGen, 0, len(genmap))
	for targetF, cfg := range genmap {
		cfgMap, ok := cfg.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("Unexpected config for generated field %s: %#v", targetF, cfg)
		}
		target := text.Str(targetF)
		source := text.Str(cfgMap["source"])
		if source == "" {
			source = target
		}

		stringForms, err := ParseStringReplacements(cfgMap["string-replace"])
		if err != nil {
			return nil, err
		}
		regexpForms, err := ParseRegexpReplacements(cfgMap["regexp-replace"])
		if err != nil {
			return nil, err
		}

		xforms := stringnorm.Combine(stringForms, regexpForms)
		if xforms == nil {
			return nil, fmt.Errorf("No transforms for field %s in %#v",
				targetF, cfg)
		}
		res = append(res, &FieldGen{
			SourceField: source,
			TargetField: target,
			Transforms:  xforms,
		})
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
