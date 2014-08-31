package qyaml

import (
	"strings"

	"github.com/greensnark/go-sequell/text"
)

type Yaml struct {
	Yaml interface{}
}

func (y Yaml) Key(key string) interface{} {
	switch v := y.Yaml.(type) {
	case map[interface{}]interface{}:
		return ResolveMapKey(v, key)
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

func (y Yaml) Slice(key string) []interface{} {
	if arr, ok := y.Key(key).([]interface{}); ok {
		return arr
	}
	return nil
}

func (y Yaml) StringSlice(key string) []string {
	return IStringSlice(y.Key(key))
}

func (y Yaml) StringMap(key string) map[string]string {
	return IStringMap(y.Key(key))
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

func IStringSlice(islice interface{}) []string {
	if islice == nil {
		return nil
	}
	if slice, ok := islice.([]interface{}); ok {
		sarr := make([]string, len(slice))
		for i, v := range slice {
			sarr[i] = text.Str(v)
		}
		return sarr
	}
	return nil
}

func StringSliceSet(slice []string) map[string]bool {
	res := make(map[string]bool)
	for _, val := range slice {
		res[val] = true
	}
	return res
}

func ResolveMapKey(m map[interface{}]interface{}, key string) interface{} {
	if directLookup, ok := m[key]; ok {
		return directLookup
	}

	hierarchy := SplitHierarchyKey(key)
	last := len(hierarchy) - 1
	for i, fragment := range hierarchy {
		if value, ok := m[fragment]; ok {
			if i == last {
				return value
			}
			if m, ok = value.(map[interface{}]interface{}); !ok {
				return nil
			}
			continue
		}
		break
	}
	return nil
}

func SplitHierarchyKey(key string) []string {
	fragments := strings.Split(key, ">")
	for i := len(fragments) - 1; i >= 0; i-- {
		fragments[i] = strings.TrimSpace(fragments[i])
	}
	return fragments
}
