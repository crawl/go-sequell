package data

import (
	"github.com/crawl/go-sequell/qyaml"
	"github.com/crawl/go-sequell/resource"
)

// Crawl is the crawl-specific configuration data.
var Crawl = CrawlData()

// Schema is the Sequell database schema definition.
var Schema = Crawl

// CrawlData parses the crawl-data.yml file into a YAML object, panicking on error.
func CrawlData() qyaml.YAML {
	return resource.MustParseYAML("config/crawl-data.yml")
}

// Sources parses the sources.yml file into a YAML object, panicking on error.
func Sources() qyaml.YAML {
	return resource.MustParseYAML("config/sources.yml")
}

// Uniques gets the list of uniques defined in crawl-data.yml
func Uniques() []string {
	return Crawl.StringSlice("uniques")
}

// Orcs gets the list of orcs defined in crawl-data.yml
func Orcs() []string {
	return Crawl.StringSlice("orcs")
}
