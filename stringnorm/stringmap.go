package stringnorm

import (
	"github.com/greensnark/go-sequell/conv"
	"github.com/greensnark/go-sequell/text"
)

type MultiMapper map[string][]string

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

func (m MultiMapper) Invert() MultiMapper {
	res := MultiMapper{}
	for k, values := range m {
		for _, v := range values {
			res[v] = append(res[v], k)
		}
	}
	return res
}

func (m MultiMapper) Map(text string) string {
	res, _ := m.Normalize(text)
	return res
}

func (m MultiMapper) Normalize(text string) (string, error) {
	if values, ok := m[text]; ok {
		if len(values) > 0 {
			return values[0], nil
		}
		return "", nil
	}
	return text, nil
}
