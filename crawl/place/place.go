package place

import (
	"regexp"

	"github.com/greensnark/go-sequell/stringnorm"
)

var placeNormalizers = stringnorm.List{
	stringnorm.StaticReg(`^Vault\b`, "Vaults"),
	stringnorm.StaticReg(`^Shoal\b`, "Shoals"),
}

func CanonicalPlace(place string) string {
	return stringnorm.NormalizeNoErr(placeNormalizers, place)
}

var rPlaceDepth = regexp.MustCompile(`:\d+`)

func StripPlaceDepth(place string) string {
	return rPlaceDepth.ReplaceAllLiteralString(place, "")
}
