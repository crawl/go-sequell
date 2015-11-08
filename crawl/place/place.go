package place

import (
	"regexp"

	"github.com/crawl/go-sequell/stringnorm"
)

// Normalizer returns a string normalizer for places.
func Normalizer(placeFixups map[string]string) stringnorm.List {
	placeNormalizers := make([]stringnorm.Normalizer, len(placeFixups))
	i := 0
	for k, v := range placeFixups {
		placeNormalizers[i] = stringnorm.SR("(?i)"+k, v)
		i++
	}
	return placeNormalizers
}

var rPlaceDepth = regexp.MustCompile(`:\d+`)

// StripPlaceDepth removes the depth qualifier in a place name. eg.
// StripPlaceDepth("Snake:5") == "Snake"
func StripPlaceDepth(place string) string {
	return rPlaceDepth.ReplaceAllLiteralString(place, "")
}
