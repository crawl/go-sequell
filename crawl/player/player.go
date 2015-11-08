package player

import (
	"strings"

	"github.com/crawl/go-sequell/qyaml"
	"github.com/crawl/go-sequell/stringnorm"
)

// CharNormalizer normalizes Crawl char abbreviations (such as TeCK)
type CharNormalizer struct {
	SpeciesAbbrNameMap stringnorm.MultiMapper
	SpeciesNameAbbrMap stringnorm.MultiMapper
	ClassAbbrNameMap   stringnorm.MultiMapper
	ClassNameAbbrMap   stringnorm.MultiMapper
}

// StockCharNormalizer creates a normalizer using the species and classes data
// in yaml
func StockCharNormalizer(yaml qyaml.YAML) *CharNormalizer {
	return NewCharNormalizer(yaml.Map("species"), yaml.Map("classes"))
}

// NewCharNormalizer creates a normalizer using the species and class
// abbreviation maps supplied.
func NewCharNormalizer(species, classes map[interface{}]interface{}) *CharNormalizer {
	specAbbrMap := createMultiMapper(species)
	classAbbrMap := createMultiMapper(classes)
	return &CharNormalizer{
		SpeciesAbbrNameMap: specAbbrMap,
		SpeciesNameAbbrMap: specAbbrMap.Invert(),
		ClassAbbrNameMap:   classAbbrMap,
		ClassNameAbbrMap:   classAbbrMap.Invert(),
	}
}

// NormalizeChar normalizes a species:class char abbreviation
func (c *CharNormalizer) NormalizeChar(race, class, existingChar string) string {
	if raceAbbr, ok := c.SpeciesNameAbbrMap[race]; ok {
		if classAbbr, ok := c.ClassNameAbbrMap[class]; ok {
			return raceAbbr[0] + classAbbr[0]
		}
	}
	return existingChar
}

func createMultiMapper(value map[interface{}]interface{}) stringnorm.MultiMapper {
	res := stringnorm.CreateMultiMapper(value)
	for _, v := range res {
		for i := range v {
			v[i] = strings.Replace(v[i], "*", "", -1)
		}
	}
	return res
}
