package version

import (
	"regexp"
	"strings"
	"sync"

	"github.com/greensnark/go-sequell/text"
)

var canonicalVersionRegex = regexp.MustCompile(`(\d+\.\d+)`)
var fullVersionRegex = regexp.MustCompile(`^(\d+\.\d+)($|[^\d.])`)

var vnumCache = map[string]uint64{}
var vnumCacheLock = sync.Mutex{}

// If the vnum cache exceeds this, dump the entire cache. This is a
// protection against bad inputs flooding the cache, and it should
// never happen for real version numbers.
const VnumCacheDumpThreshold = 1000

func CachingVersionNumericId(ver string) uint64 {
	vnumCacheLock.Lock()
	defer vnumCacheLock.Unlock()
	if vnumId, exists := vnumCache[ver]; exists {
		return vnumId
	}

	if len(vnumCache) > VnumCacheDumpThreshold {
		vnumCache = map[string]uint64{}
	}
	vnumId := VersionNumericId(ver)
	vnumCache[ver] = vnumId
	return vnumId
}

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

var rAlphaQualifier = regexp.MustCompile(`-[a-z]`)

func IsAlpha(ver string) bool {
	return rAlphaQualifier.FindString(ver) != ""
}

func SplitVersionQualifier(ver string) (string, string) {
	ver = strings.TrimSpace(ver)
	hyphenatedParts := strings.SplitN(ver, "-", 2)
	if len(hyphenatedParts) == 2 {
		return hyphenatedParts[0], hyphenatedParts[1]
	} else {
		return ver, ""
	}
}

var rQualifierPrefixMajorMinor = regexp.MustCompile(`^([a-z]+)([0-9]*)(?:-(\d+))?`)

func SplitQualifierPrefixMajorMinor(qual string) (string, string, string) {
	match := rQualifierPrefixMajorMinor.FindStringSubmatch(qual)
	if match == nil {
		return qual, "", ""
	}
	return match[1], match[2], match[3]
}

func SplitDottedVersion(ver string) []string {
	return text.RightPadSlice(strings.Split(ver, "."), 3, "0")
}

func VersionNumericId(ver string) uint64 {
	version, qualifier := SplitVersionQualifier(ver)
	return versionNumberize(SplitDottedVersion(version)) +
		versionQualifierNumberize(qualifier)
}

func versionNumberize(versionParts []string) uint64 {
	var base uint64 = 1e8
	var sum uint64 = 0
	for i := len(versionParts) - 1; i >= 0; i-- {
		sum += uint64(text.ParseInt(versionParts[i], 0)) * base
		base *= 1e3
	}
	return sum
}

func versionQualifierNumberize(qualifier string) uint64 {
	if qualifier == "" {
		return 1e8 - 1
	}

	prefix, major, minor := SplitQualifierPrefixMajorMinor(qualifier)

	return alphaPrefixNumberize(prefix)*1e6 +
		uint64(text.ParseInt(major, 0))*1e4 +
		uint64(text.ParseInt(minor, 0))
}

func alphaPrefixNumberize(prefix string) uint64 {
	if prefix == "" {
		return 0
	}
	return uint64(prefix[0] - 'a' + 1)
}
