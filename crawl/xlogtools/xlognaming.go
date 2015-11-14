package xlogtools

import (
	"regexp"
	"strings"

	"github.com/crawl/go-sequell/conv"
	"github.com/crawl/go-sequell/crawl/data"
	"github.com/crawl/go-sequell/crawl/version"
)

// NewGameMatcher creates a GameMatcher using the game-type-tags property of c.
func NewGameMatcher(c data.Crawl) *GameMatcher {
	return &GameMatcher{
		t: createTextTypeMatcher(c.Map("game-type-tags")),
	}
}

// A GameMatcher identifies the type of a game by examining its log filename.
type GameMatcher struct {
	t TextTypeLookup
}

// XlogGame guesses what kind of games a given logfile or milestone
// filename contains.
func (g *GameMatcher) XlogGame(filename string) string {
	return g.t.FindType(filename)
}

// XlogServerType parses an Xlog qualified name and returns the Xlog
// server, game type, and Xlog file type.
func (g *GameMatcher) XlogServerType(filename string) (server, game string, xlogtype XlogType) {
	m := rXlogQualifiedName.FindStringSubmatch(filename)
	if m == nil {
		return "", "", Unknown
	}
	return m[1], g.XlogGame(filename), XlogFileType(filename)
}

var rGitVersion = regexp.MustCompile(`(?i)\b(?:git|svn|master|trunk)\b`)
var rEmbeddedVersion = regexp.MustCompile(`\d+[.]\d+`)
var rEmbeddedVersionKey = regexp.MustCompile(`\d{2}\b`)

// XlogGameVersion guesses the game version from an xlog filename.
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

var rXlogQualifiedName = regexp.MustCompile(`^remote.(\w+)-[\w.-]+$`)

// IsXlogQualifiedName returns true if the filename is correctly
// formatted as a canonical qualified name.
func IsXlogQualifiedName(filename string) bool {
	return rXlogQualifiedName.FindString(filename) != ""
}

// XlogFileType guesses the type of xlog file based on the name.
func XlogFileType(filename string) XlogType {
	if strings.Index(filename, "logfile") != -1 {
		return Log
	} else if strings.Index(filename, "milestone") != -1 {
		return Milestone
	}
	return Unknown
}

// XlogQualifiedName gets the canonical qualified name for an xlog file, given
// the server, game, Crawl version, qualifier, and xlog type.
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

// TextTypeLookup guesses the type of a thing based on its name.
type TextTypeLookup struct {
	TypeMatchers []typeMatcher
	DefaultType  string
}

// FindType guesses the type of filename based on its name.
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
			matchers = append(matchers, createTypeMatcher(conv.IStringSlice(tags), textType))
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
