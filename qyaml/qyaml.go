package qyaml

import (
	"github.com/greensnark/go-sequell/text"
)

type Yaml struct {
	Yaml interface{}
}

func IStringMap(v interface{}) map[string]string {
	res := map[string]string{}
	if keyMap, ok := v.(map[interface{}]interface{}); ok {
		for key, value := range keyMap {
			res[text.Str(key)] = text.Str(value)
		}
	}
	return res
}

func (y Yaml) Key(key string) interface{} {
	switch v := y.Yaml.(type) {
	case map[interface{}]interface{}:
		return v[key]
	}
	return nil
}

func (y Yaml) String(key string) string {
	return text.Str(y.Key(key))
}

func (y Yaml) Map(key string) map[interface{}]interface{} {
	if vmap, ok := y.Key(key).(map[interface{}]interface{}); ok {
		return vmap
	}
	return nil
}

func (y Yaml) Array(key string) []interface{} {
	if arr, ok := y.Key(key).([]interface{}); ok {
		return arr
	}
	return nil
}

func (y Yaml) StringArray(key string) []string {
	arr := y.Array(key)
	if arr == nil {
		return nil
	}

	sarr := make([]string, len(arr))
	for i, v := range arr {
		sarr[i] = text.Str(v)
	}
	return sarr
}

func (y Yaml) StringMap(key string) map[string]string {
	return IStringMap(y.Key(key))
}

func StringSliceSet(slice []string) map[string]bool {
	res := make(map[string]bool)
	for _, val := range slice {
		res[val] = true
	}
	return res
}
