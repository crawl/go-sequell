package qyaml

import (
	"io/ioutil"
	"strings"

	"github.com/crawl/go-sequell/conv"
	"github.com/crawl/go-sequell/text"
	"gopkg.in/yaml.v2"
)

// YAML represents a YAML document.
type YAML struct {
	YAML interface{}
}

// ParseBytes parses a []byte into a YAML document.
func ParseBytes(text []byte) (YAML, error) {
	var res interface{}
	err := yaml.Unmarshal(text, &res)
	return YAML{YAML: res}, err
}

// Parse parses the file at path into a YAML document.
func Parse(path string) (YAML, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return YAML{}, err
	}
	return ParseBytes(bytes)
}

// MustParse parses the file at path into a YAML document, panicking on error.
func MustParse(path string) YAML {
	res, err := Parse(path)
	if err != nil {
		panic(err)
	}
	return res
}

// Wrap wraps value as a YAML document.
func Wrap(value interface{}) YAML { return YAML{YAML: value} }

// Key gets the value with the given key, treating ">"-separated keys as nested
// objects to descend into.
func (y YAML) Key(key string) interface{} {
	switch v := y.YAML.(type) {
	case map[interface{}]interface{}:
		return ResolveMapKey(v, key)
	}
	return nil
}

func (y YAML) String(key string) string {
	return text.Str(y.Key(key))
}

// Map gets the map object at key.
func (y YAML) Map(key string) map[interface{}]interface{} {
	if vmap, ok := y.Key(key).(map[interface{}]interface{}); ok {
		return vmap
	}
	return nil
}

// Slice gets the slice at key.
func (y YAML) Slice(key string) []interface{} {
	if arr, ok := y.Key(key).([]interface{}); ok {
		return arr
	}
	return nil
}

// StringSlice gets the string slice at key
func (y YAML) StringSlice(key string) []string {
	return conv.IStringSlice(y.Key(key))
}

// StringMap gets the map[string]string at key
func (y YAML) StringMap(key string) map[string]string {
	return conv.IStringMap(y.Key(key))
}

// ResolveMapKey looks up key in m, trying it as a direct lookup, then by
// splitting on ">" and treating each element as a sub-key to look up.
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

// SplitHierarchyKey splits a key on > into a key path.
func SplitHierarchyKey(key string) []string {
	fragments := strings.Split(key, ">")
	for i := len(fragments) - 1; i >= 0; i-- {
		fragments[i] = strings.TrimSpace(fragments[i])
	}
	return fragments
}
