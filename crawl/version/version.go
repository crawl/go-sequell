package version

import (
	"math/big"
	"regexp"
)

var canonicalVersionRegex = regexp.MustCompile(`(\d+\.\d+)`)
var fullVersionRegex = regexp.MustCompile(`^(\d+\.\d+)($|[^\d.])`)

// MajorVersion returns the first two segments (X.Y) of a long version
// string in the form X.Y.Z
func MajorVersion(ver string) string {
	cv := canonicalVersionRegex.FindString(ver)
	if cv == "" {
		return ver
	}
	return cv
}

// FullVersion expands a short version in the form X.Y to X.Y.0
func FullVersion(ver string) string {
	return fullVersionRegex.ReplaceAllString(ver, "$1.0$2")
}

func VersionNumericId(ver string) big.Int {
	// TODO
}
