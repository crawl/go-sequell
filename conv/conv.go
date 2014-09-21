package conv

import "github.com/greensnark/go-sequell/text"

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

func IStringPairs(islice interface{}) [][]string {
	if islice == nil {
		return nil
	}
	if slice, ok := islice.([]interface{}); ok {
		res := make([][]string, 0, len(slice))
		for _, thing := range slice {
			if pair, ok := thing.([]interface{}); ok {
				spair := make([]string, 2)
				spair[0] = text.Str(pair[0])
				spair[1] = text.Str(pair[1])
				res = append(res, spair)
			}
		}
		return res
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
