package version

import (
	"math/big"
	"regexp"
	"strings"
	"sync"

	"github.com/golang/groupcache/lru"
	"github.com/greensnark/go-sequell/text"
)

var canonicalVersionRegex = regexp.MustCompile(`(\d+\.\d+)`)
var fullVersionRegex = regexp.MustCompile(`^(\d+\.\d+)($|[^\d.])`)

var vnumCache = lru.New(50)
var vnumCacheLock = sync.Mutex{}

func CachingVersionNumericId(ver string) *big.Int {
	vnumCacheLock.Lock()
	defer vnumCacheLock.Unlock()
	if vnumId, exists := vnumCache.Get(ver); exists {
		return vnumId.(*big.Int)
	}

	vnumId := VersionNumericId(ver)
	vnumCache.Add(ver, vnumId)
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

func SplitVersionQualifier(ver string) (string, string) {
	ver = strings.TrimSpace(ver)
	hyphenatedParts := strings.SplitN(ver, "-", 2)
	if len(hyphenatedParts) == 2 {
		return hyphenatedParts[0], hyphenatedParts[1]
	} else {
		return ver, ""
	}
}

var rQualifierPrefixIndex = regexp.MustCompile(`^([a-z]+)([0-9]*)`)

func SplitQualifierPrefixIndex(qual string) (string, string) {
	match := rQualifierPrefixIndex.FindStringSubmatch(qual)
	if match == nil {
		return qual, ""
	}
	return match[1], match[2]
}

func SplitDottedVersion(ver string) []string {
	return text.RightPadSlice(strings.Split(ver, "."), 4, "0")
}

func VersionNumericId(ver string) *big.Int {
	version, qualifier := SplitVersionQualifier(ver)
	vnum := versionNumberize(SplitDottedVersion(version))
	vnum.Add(vnum, versionQualifierNumberize(qualifier))
	return vnum
}

func versionNumberize(versionParts []string) *big.Int {
	base := big.NewInt(1000000)
	mul := big.NewInt(0)
	sum := big.NewInt(0)
	for i := len(versionParts) - 1; i >= 0; i-- {
		mul.SetInt64(int64(text.ParseInt(versionParts[i], 0)))

		mul.Mul(mul, base)
		sum.Add(sum, mul)

		mul.SetInt64(1000)
		base.Mul(base, mul)
	}
	return sum
}

func versionQualifierNumberize(qualifier string) *big.Int {
	if qualifier == "" {
		return big.NewInt(999 * 999)
	}

	prefix, index := SplitQualifierPrefixIndex(qualifier)

	return big.NewInt(int64(
		alphaPrefixNumberize(prefix) + text.ParseInt(index, 0)))
}

func alphaPrefixNumberize(prefix string) int {
	return 1000 * int(prefix[0])
}
