package player

import (
	"strings"

	"github.com/crawl/go-sequell/qyaml"
	"github.com/crawl/go-sequell/stringnorm"
)

type CharNormalizer struct {
	SpeciesAbbrNameMap stringnorm.MultiMapper
	SpeciesNameAbbrMap stringnorm.MultiMapper
	ClassAbbrNameMap   stringnorm.MultiMapper
	ClassNameAbbrMap   stringnorm.MultiMapper
}

func StockCharNormalizer(yaml qyaml.Yaml) *CharNormalizer {
	return NewCharNormalizer(yaml.Map("species"), yaml.Map("classes"))
}

func NewCharNormalizer(species, classes map[interface{}]interface{}) *CharNormalizer {
	specAbbrMap := CreateMultiMapper(species)
	classAbbrMap := CreateMultiMapper(classes)
	return &CharNormalizer{
		SpeciesAbbrNameMap: specAbbrMap,
		SpeciesNameAbbrMap: specAbbrMap.Invert(),
		ClassAbbrNameMap:   classAbbrMap,
		ClassNameAbbrMap:   classAbbrMap.Invert(),
	}
}

func (c *CharNormalizer) NormalizeChar(race, class, existingChar string) string {
	if raceAbbr, ok := c.SpeciesNameAbbrMap[race]; ok {
		if classAbbr, ok := c.ClassNameAbbrMap[class]; ok {
			return raceAbbr[0] + classAbbr[0]
		}
	}
	return existingChar
}

func CreateMultiMapper(value map[interface{}]interface{}) stringnorm.MultiMapper {
	res := stringnorm.CreateMultiMapper(value)
	for _, v := range res {
		for i := range v {
			v[i] = strings.Replace(v[i], "*", "", -1)
		}
	}
	return res
}
