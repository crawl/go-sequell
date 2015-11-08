package version

import (
	"regexp"
	"strings"
	"sync"

	"github.com/crawl/go-sequell/text"
)

var canonicalVersionRegex = regexp.MustCompile(`(\d+\.\d+)`)
var fullVersionRegex = regexp.MustCompile(`^(\d+\.\d+)($|[^\d.])`)

var vnumCache = map[string]uint64{}
var vnumCacheLock = sync.Mutex{}

// If the vnum cache exceeds this, dump the entire cache. This is a
// protection against bad inputs flooding the cache, and it should
// never happen for real version numbers.
const VnumCacheDumpThreshold = 1000

var rVCSPrefix = regexp.MustCompile(`^.*?:+`)

// StripVCSQualifier removes a VCS: prefix from the head of ver.
func StripVCSQualifier(ver string) string {
	return rVCSPrefix.ReplaceAllString(ver, "")
}

// CachingNumericID returns a version ID given a Crawl version string.
// The return value is identical to VersionNumericID, but calculated values
// are cached for improved performance.
func CachingNumericID(ver string) uint64 {
	vnumCacheLock.Lock()
	defer vnumCacheLock.Unlock()
	if vnumID, exists := vnumCache[ver]; exists {
		return vnumID
	}

	if len(vnumCache) > VnumCacheDumpThreshold {
		vnumCache = map[string]uint64{}
	}
	vnumID := NumericID(ver)
	vnumCache[ver] = vnumID
	return vnumID
}

// Major returns the first two segments (X.Y) of a long version
// string in the form X.Y.Z
func Major(ver string) string {
	cv := canonicalVersionRegex.FindString(ver)
	if cv == "" {
		return ver
	}
	return cv
}

// Full expands a short version in the form X.Y to X.Y.0
func Full(ver string) string {
	return fullVersionRegex.ReplaceAllString(ver, "$1.0$2")
}

var rAlphaQualifier = regexp.MustCompile(`-[a-z]`)

// IsAlpha returns true if ver is a version with an alpha (-a, -b, etc.)
// qualifier.
func IsAlpha(ver string) bool {
	return rAlphaQualifier.FindString(ver) != ""
}

// SplitQualifier splits a hyphenated version into the parts
// before and after the first hyphen. If the string contains no
// hyphen, the first part is the entire string and the second part is
// empty.
func SplitQualifier(ver string) (string, string) {
	ver = strings.TrimSpace(ver)
	hyphenatedParts := strings.SplitN(ver, "-", 2)
	if len(hyphenatedParts) == 2 {
		return hyphenatedParts[0], hyphenatedParts[1]
	}
	return ver, ""
}

var rQualifierPrefixMajorMinor = regexp.MustCompile(`^([a-z]+)([0-9]*)(?:-(\d+))?`)
var rUnqualifiedRevCount = regexp.MustCompile(`^(\d+)-`)

// SplitQualifierPrefixMajorMinor splits a version qualifier into prefix,
// major and minor parts.
func SplitQualifierPrefixMajorMinor(qual string) (string, string, string) {
	match := rQualifierPrefixMajorMinor.FindStringSubmatch(qual)
	if match == nil {
		unqualifiedMatch := rUnqualifiedRevCount.FindStringSubmatch(qual)
		if unqualifiedMatch == nil {
			return qual, "", ""
		}
		return "", "", unqualifiedMatch[1]
	}
	return match[1], match[2], match[3]
}

// SplitDottedVersion splits a dotted version string into its parts.
func SplitDottedVersion(ver string) []string {
	return text.RightPadSlice(strings.Split(ver, "."), 3, "0")
}

// ExpandVersionKey expands a shortened Crawl version of the form
// "01", "11" etc to a 0.X form.
func ExpandVersionKey(verkey string) string {
	return "0." + strings.TrimLeft(verkey, "0")
}

// NumericID parses a Crawl version number and returns an int64
// representing the version that can be used in numeric comparisons,
// where later versions return higher numbers than older versions.
func NumericID(ver string) uint64 {
	version, qualifier := SplitQualifier(ver)
	return versionNumberize(SplitDottedVersion(version)) +
		versionQualifierNumberize(qualifier)
}

func versionNumberize(versionParts []string) uint64 {
	var base uint64 = 1e8
	var sum uint64
	for i := len(versionParts) - 1; i >= 0; i-- {
		sum += uint64(text.ParseInt(versionParts[i], 0)) * base
		base *= 1e3
	}
	return sum
}

func versionQualifierNumberize(qualifier string) uint64 {
	if qualifier == "" {
		return 99 * 1e6
	}

	prefix, major, minor := SplitQualifierPrefixMajorMinor(qualifier)

	return alphaPrefixNumberize(prefix)*1e6 +
		uint64(text.ParseInt(major, 0))*1e4 +
		uint64(text.ParseInt(minor, 0))
}

func alphaPrefixNumberize(prefix string) uint64 {
	if prefix == "" {
		return 99
	}
	return uint64(prefix[0] - 'a' + 1)
}
