package resource

import (
	"github.com/crawl/go-sequell/ectx"
	"github.com/crawl/go-sequell/qyaml"
	"github.com/crawl/go-sequell/root"
)

// Root is the resource root containing Sequell's config
var Root = root.New("", "SEQUELL_ROOT", "HENZELL_ROOT")

// YAML reads the YAML file at path.
func YAML(path string) (qyaml.YAML, error) {
	yaml, err := qyaml.Parse(Root.Path(path))
	if err != nil {
		return yaml, ectx.Err("yaml: "+path, err)
	}
	return yaml, nil
}

// MustParseYAML reads the YAML at path, panicking if there is an error.
func MustParseYAML(path string) qyaml.YAML {
	yaml, err := YAML(path)
	if err != nil {
		panic(err)
	}
	return yaml
}
