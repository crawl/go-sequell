package resource

import (
	"github.com/greensnark/go-sequell/qyaml"
	"github.com/greensnark/go-sequell/root"
)

var Root = root.New("", "SEQUELL_ROOT", "HENZELL_ROOT")

func Yaml(path string) (qyaml.Yaml, error) {
	return qyaml.Parse(Root.Path(path))
}

func YamlMustParse(path string) qyaml.Yaml {
	yaml, err := Yaml(path)
	if err != nil {
		panic(err)
	}
	return yaml
}
