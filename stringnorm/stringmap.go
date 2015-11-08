package stringnorm

import (
	"github.com/crawl/go-sequell/conv"
	"github.com/crawl/go-sequell/text"
)

// A MultiMapper represents a string mapping with multiple values per key.
type MultiMapper map[string][]string

// CreateMultiMapper converts a yaml map to a MultiMapper.
func CreateMultiMapper(yamlMap map[interface{}]interface{}) MultiMapper {
	mapper := MultiMapper{}
	for k, v := range yamlMap {
		key := text.Str(k)
		switch val := v.(type) {
		case string:
			mapper[key] = []string{val}
		case []interface{}:
			mapper[key] = conv.IStringSlice(val)
		default:
			continue
		}
	}
	return mapper
}

// Invert inverts m so that values become keys, and vice versa. If a key had
// multiple values: {k: [v1,v2,v3]}, Invert will return a map {v1:k,v2:k,v3:k}
func (m MultiMapper) Invert() MultiMapper {
	res := MultiMapper{}
	for k, values := range m {
		for _, v := range values {
			res[v] = append(res[v], k)
		}
	}
	return res
}

// Map calls Normalize and discards any error returned.
func (m MultiMapper) Map(text string) string {
	res, _ := m.Normalize(text)
	return res
}

// Normalize looks up the mapping for text. If there is no mapping for text, it
// is returned unmodified. If there are multiple mappings for text, the first
// mapping is returned.
func (m MultiMapper) Normalize(text string) (string, error) {
	if values, ok := m[text]; ok {
		if len(values) > 0 {
			return values[0], nil
		}
		return "", nil
	}
	return text, nil
}
