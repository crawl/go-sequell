package data

import (
	"github.com/greensnark/go-sequell/qyaml"
	"github.com/greensnark/go-sequell/resource"
)

var Crawl qyaml.Yaml = resource.ResourceYamlMustExist("config/crawl-data.yml")
var Schema = Crawl

func Uniques() []string {
	return Crawl.StringSlice("uniques")
}

func Orcs() []string {
	return Crawl.StringSlice("orcs")
}
