package data

import (
	"fmt"
	"github.com/greensnark/go-sequell/resource"
	"sync"
)

var mutex = &sync.Mutex{}
var data map[interface{}]interface{}

func Data() map[interface{}]interface{} {
	mutex.Lock()
	defer mutex.Unlock()
	if data == nil {
		tmp, err := resource.ResourceYaml("config/crawl-data.yml")
		if err != nil {
			panic(err)
		}

		var ok bool
		data, ok = tmp.(map[interface{}]interface{})
		if !ok {
			panic(fmt.Sprintf("unexpected data: %v", tmp))
		}
	}
	return data
}

func String(key string) string {
	s, ok := Data()[key].(string)
	if ok {
		return s
	}
	return ""
}

func StringArray(key string) []string {
	t := Data()
	arr := t[key].([]interface{})
	res := make([]string, len(arr))
	for i, v := range arr {
		res[i] = v.(string)
	}
	return res
}

func Uniques() []string {
	return StringArray("uniques")
}

func Orcs() []string {
	return StringArray("orcs")
}

func StringSliceSet(slice []string) map[string]bool {
	res := make(map[string]bool)
	for _, val := range slice {
		res[val] = true
	}
	return res
}
