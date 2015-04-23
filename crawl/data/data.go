package data

import (
	"github.com/crawl/go-sequell/qyaml"
	"github.com/crawl/go-sequell/resource"
)

var Crawl qyaml.Yaml = CrawlData()
var Schema = Crawl

func CrawlData() qyaml.Yaml {
	return resource.YamlMustParse("config/crawl-data.yml")
}

func Sources() qyaml.Yaml {
	return resource.YamlMustParse("config/sources.yml")
}

func Uniques() []string {
	return Crawl.StringSlice("uniques")
}

func Orcs() []string {
	return Crawl.StringSlice("orcs")
}
