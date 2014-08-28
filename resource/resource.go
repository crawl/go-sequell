package resource

import (
	"gopkg.in/v1/yaml"
	"io/ioutil"
	"os"
	"path"

	"github.com/greensnark/go-sequell/qyaml"
)

var Root = root()

func firstSet(vars ...string) string {
	for _, value := range vars {
		if value != "" {
			return value
		}
	}
	return ""
}

func root() string {
	root := firstSet(os.Getenv("SEQUELL_ROOT"), os.Getenv("HENZELL_ROOT"))
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

func ResourceYaml(filepath string) (qyaml.Yaml, error) {
	text, err := ResourceString(filepath)
	if err != nil {
		return qyaml.Yaml{}, err
	}

	return StringYaml(text)
}

func ResourceYamlMustExist(filepath string) qyaml.Yaml {
	yaml, err := ResourceYaml(filepath)
	if err != nil {
		panic(err)
	}
	return yaml
}

func StringYaml(text string) (qyaml.Yaml, error) {
	var res interface{}
	err := yaml.Unmarshal([]byte(text), &res)
	return qyaml.Yaml{res}, err
}
