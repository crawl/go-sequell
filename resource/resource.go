package resource

import (
	"github.com/greensnark/go-sequell/ectx"
	"github.com/greensnark/go-sequell/qyaml"
	"github.com/greensnark/go-sequell/root"
)

var Root = root.New("", "SEQUELL_ROOT", "HENZELL_ROOT")

func Yaml(path string) (qyaml.Yaml, error) {
	yaml, err := qyaml.Parse(Root.Path(path))
	if err != nil {
		return yaml, ectx.Err("yaml: "+path, err)
	}
	return yaml, nil
}

func YamlMustParse(path string) qyaml.Yaml {
	yaml, err := Yaml(path)
	if err != nil {
		panic(err)
	}
	return yaml
}
