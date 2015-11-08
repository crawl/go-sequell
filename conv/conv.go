package conv

import "github.com/crawl/go-sequell/text"

// IStringMap converts v into a map[string]string, provided v is really a
// map[interface{}]interface{}. If not, returns an empty string map.
func IStringMap(v interface{}) map[string]string {
	res := map[string]string{}
	if keyMap, ok := v.(map[interface{}]interface{}); ok {
		for key, value := range keyMap {
			res[text.Str(key)] = text.Str(value)
		}
	}
	return res
}

// IStringSlice converts islice into a []string, provided islice is a
// []interface{}, or returns an empty slice if not.
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

// IStringPairs converts islice into a [][]string, where each element is a
// 2-pair []string. If islice is not a []interface{}, returns an empty
// [][]string
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

// StringSliceSet converts a []string into a map[string]bool where the keys
// in the map are values in the []string mapped to true.
func StringSliceSet(slice []string) map[string]bool {
	res := make(map[string]bool)
	for _, val := range slice {
		res[val] = true
	}
	return res
}
