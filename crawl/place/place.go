package place

import (
	"regexp"

	"github.com/greensnark/go-sequell/crawl/data"
	"github.com/greensnark/go-sequell/stringnorm"
)

var placeNormalizers stringnorm.List

func init() {
	placeFixups := data.Crawl.StringMap("place-fixups")
	placeNormalizers = make([]stringnorm.Normalizer, len(placeFixups))
	i := 0
	for k, v := range placeFixups {
		placeNormalizers[i] = stringnorm.SR("(?i)"+k, v)
		i++
	}
}

func CanonicalPlace(place string) string {
	return stringnorm.NormalizeNoErr(placeNormalizers, place)
}

var rPlaceDepth = regexp.MustCompile(`:\d+`)

func StripPlaceDepth(place string) string {
	return rPlaceDepth.ReplaceAllLiteralString(place, "")
}
