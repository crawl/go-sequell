package resource

import (
	"gopkg.in/v1/yaml"
	"io/ioutil"
	"os"
	"path"
)

var Root = root()

func root() string {
	root := os.Getenv("HENZELL_ROOT")
	if root != "" {
		return root
	}
	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return root
}

func ResourcePath(filepath string) string {
	return path.Join(Root, filepath)
}

func ResourceString(filepath string) (string, error) {
	bytes, err := ioutil.ReadFile(ResourcePath(filepath))
	if err != nil {
		return "", err
	}
	return string(bytes), err
}

func ResourceYaml(filepath string) (interface{}, error) {
	text, err := ResourceString(filepath)
	if err != nil {
		return nil, err
	}

	var res = make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(text), &res)
	return res, err
}
