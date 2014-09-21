package qyaml

import (
	"io/ioutil"
	"strings"

	"github.com/greensnark/go-sequell/conv"
	"github.com/greensnark/go-sequell/text"
	"gopkg.in/v1/yaml"
)

type Yaml struct {
	Yaml interface{}
}

func ParseBytes(text []byte) (Yaml, error) {
	var res interface{}
	err := yaml.Unmarshal(text, &res)
	return Yaml{Yaml: res}, err
}

func Parse(path string) (Yaml, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return Yaml{}, err
	}
	return ParseBytes(bytes)
}

func MustParse(path string) Yaml {
	res, err := Parse(path)
	if err != nil {
		panic(err)
	}
	return res
}

func Wrap(value interface{}) Yaml { return Yaml{Yaml: value} }

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
	return conv.IStringSlice(y.Key(key))
}

func (y Yaml) StringMap(key string) map[string]string {
	return conv.IStringMap(y.Key(key))
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
