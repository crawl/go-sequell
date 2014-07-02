package data

import (
	"github.com/greensnark/go-sequell/resource"
	"sync"
)

var mutex = &sync.Mutex{}
var data interface{}

func Data() interface{} {
	mutex.Lock()
	defer mutex.Unlock()
	if data == nil {
		var err error
		data, err = resource.ResourceYaml("config/crawl-data.yml")
		if err != nil {
			panic(err)
		}
	}
	return data
}

func StringArray(key string) []string {
	t, ok := Data().(map[interface{}]interface{})
	if !ok {
		return []string{}
	}

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
