package data

import (
	"github.com/greensnark/go-sequell/qyaml"
	"github.com/greensnark/go-sequell/resource"
)

var Crawl qyaml.Yaml = CrawlData()
var Schema = Crawl

func CrawlData() qyaml.Yaml {
	return resource.ResourceYamlMustExist("config/crawl-data.yml")
}

func Sources() qyaml.Yaml {
	return resource.ResourceYamlMustExist("config/sources.yml")
}

func Uniques() []string {
	return Crawl.StringSlice("uniques")
}

func Orcs() []string {
	return Crawl.StringSlice("orcs")
}
