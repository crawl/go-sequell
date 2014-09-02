package xlogtools

import (
	"regexp"
	"strings"

	"github.com/greensnark/go-sequell/crawl/data"
	"github.com/greensnark/go-sequell/crawl/version"
	"github.com/greensnark/go-sequell/qyaml"
)

var gameTypeMatcher = createTextTypeMatcher(data.Crawl.Map("game-type-tags"))

// XlogGame guesses what kind of games a given logfile or milestone
// filename contains.
func XlogGame(filename string) string {
	return gameTypeMatcher.FindType(filename)
}

var rGitVersion = regexp.MustCompile(`(?i)\b(?:git|svn|master|trunk)\b`)
var rEmbeddedVersion = regexp.MustCompile(`\d+[.]\d+`)
var rEmbeddedVersionKey = regexp.MustCompile(`\d{2}\b`)

func XlogGameVersion(filename string) string {
	if rGitVersion.MatchString(filename) {
		return "git"
	}
	if ver := rEmbeddedVersion.FindString(filename); ver != "" {
		return ver
	}
	if ver := rEmbeddedVersionKey.FindString(filename); ver != "" {
		return version.ExpandVersionKey(ver)
	}
	return "any"
}

func XlogQualifiedName(server, game, version, qualifier string, xlogtype XlogType) string {
	base := "remote." + server + "-" + xlogtype.String() + "-" + version
	if game != "" {
		base += "-" + game
	}
	if qualifier != "" {
		return base + "-" + qualifier
	}
	return base
}

type typeMatcher struct {
	*regexp.Regexp
	typeName string
}

type TextTypeLookup struct {
	TypeMatchers []typeMatcher
	DefaultType  string
}

func (g TextTypeLookup) FindType(filename string) string {
	for _, m := range g.TypeMatchers {
		if m.MatchString(filename) {
			return m.typeName
		}
	}
	return g.DefaultType
}

func createTextTypeMatcher(typeTagsMap map[interface{}]interface{}) TextTypeLookup {
	lookup := TextTypeLookup{}
	matchers := []typeMatcher{}
	for k, v := range typeTagsMap {
		textType := k.(string)
		if textType == "DEFAULT" {
			lookup.DefaultType = v.(string)
			continue
		}
		switch tags := v.(type) {
		case string:
			matchers = append(matchers, createTypeMatcher([]string{tags}, textType))
		case []interface{}:
			matchers = append(matchers, createTypeMatcher(qyaml.IStringSlice(tags), textType))
		}
	}
	lookup.TypeMatchers = matchers
	return lookup
}

func createTypeMatcher(tags []string, tagType string) typeMatcher {
	for i := range tags {
		tags[i] = regexp.QuoteMeta(tags[i])
	}
	return typeMatcher{
		Regexp:   regexp.MustCompile(`\b(?:` + strings.Join(tags, "|") + `)\b`),
		typeName: tagType,
	}
}
