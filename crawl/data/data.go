package data

import (
	"github.com/crawl/go-sequell/qyaml"
	"github.com/crawl/go-sequell/resource"
)

const crawlDataFile = "config/crawl-data.yml"

// Crawl is the crawl-specific configuration data.
type Crawl struct {
	qyaml.YAML
}

// Schema is the Sequell database schema definition.
type Schema struct {
	qyaml.YAML
}

// CrawlData parses the crawl-data.yml file into a Crawl YAML object, panicking on error.
func CrawlData() Crawl {
	return Crawl{resource.MustParseYAML(crawlDataFile)}
}

// CrawlSchema parses the crawl-data.yml file into a Schema YAML object, panicking on error.
func CrawlSchema() Schema {
	return Schema{resource.MustParseYAML(crawlDataFile)}
}

// Sources parses the sources.yml file into a YAML object, panicking on error.
func Sources() qyaml.YAML {
	return resource.MustParseYAML("config/sources.yml")
}

// Uniques gets the list of uniques defined in crawl-data.yml
func (c Crawl) Uniques() []string {
	return c.StringSlice("uniques")
}

// Orcs gets the list of orcs defined in crawl-data.yml
func (c Crawl) Orcs() []string {
	return c.StringSlice("orcs")
}
