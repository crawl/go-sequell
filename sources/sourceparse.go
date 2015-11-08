package sources

import (
	"fmt"
	"path"
	"strings"

	"github.com/crawl/go-sequell/crawl/ctime"
	"github.com/crawl/go-sequell/crawl/xlogtools"
	"github.com/crawl/go-sequell/qyaml"
	"github.com/crawl/go-sequell/text"
)

// Sources reads a parsed sources.yml object and returns a Servers object
// with metadata on the remote servers that supply xlogs to Sequell.
func Sources(sources qyaml.YAML, cachedir string) (*Servers, error) {
	return sourceYamlParser{
		sources:  sources.Slice("sources"),
		cachedir: cachedir,
	}.Parse()
}

// DuplicateXlogTargets gets the list of duplicated xlog target paths.
// Duplicate target paths are bad news (since they overwrite each other) and
// indicate a broken configuration.
func DuplicateXlogTargets(xlogs []*XlogSrc) []string {
	pathCounts := map[string][]string{}
	for _, x := range xlogs {
		pathCounts[x.TargetPath] = append(pathCounts[x.TargetPath], x.URL)
	}
	dupes := []string{}
	for path, urls := range pathCounts {
		if len(urls) > 1 {
			dupes = append(dupes, path+" ("+strings.Join(urls, ", ")+")")
		}
	}
	return dupes
}

type sourceYamlParser struct {
	sources  []interface{}
	cachedir string
}

func (s sourceYamlParser) Parse() (*Servers, error) {
	sources := Servers{
		Servers: make([]*Server, len(s.sources)),
	}
	var err error
	for i, serverYaml := range s.sources {
		sources.Servers[i], err = serverParser{
			server:   qyaml.Wrap(serverYaml),
			cachedir: s.cachedir,
		}.Parse()
		if err != nil {
			return nil, err
		}
	}
	return &sources, nil
}

type serverParser struct {
	server   qyaml.YAML
	cachedir string
}

func (s serverParser) Parse() (*Server, error) {
	name := s.server.String("name")

	tz, err := s.ParseTimeZones(s.server.StringMap("timezones"))
	if err != nil {
		return nil, err
	}

	server := Server{
		Name:          name,
		BaseURL:       s.server.String("base"),
		LocalPathBase: s.server.String("local"),
		TimeZoneMap:   tz,
		UtcEpoch:      ctime.SafeParseTimeWithZone(s.server.String("utc-epoch")),
	}

	if server.Logfiles, err =
		s.ParseXlogRefs(&server, s.server.Slice("logs")); err != nil {
		return nil, err
	}
	return &server, nil
}

func (s serverParser) ParseTimeZones(tzdst map[string]string) (ctime.DSTLocation, error) {
	return ctime.ParseDSTLocation(tzdst["S"], tzdst["D"])
}

func (s serverParser) ParseXlogRefs(
	server *Server,
	logfiles []interface{}) ([]*XlogSrc, error) {
	xlogs, err := xlogSpecParser{
		server:   server,
		cachedir: s.cachedir,
	}.Parse(logfiles)
	if err != nil {
		return nil, err
	}
	dupes := DuplicateXlogTargets(xlogs)
	if len(dupes) > 0 {
		return nil, fmt.Errorf("Duplicate xlog targets: %s",
			strings.Join(dupes, ", "))
	}
	return xlogs, nil
}

type xlogSpecParser struct {
	server   *Server
	cachedir string
}

func (p xlogSpecParser) Parse(specs []interface{}) ([]*XlogSrc, error) {
	xlogs := []*XlogSrc{}
	for _, logspec := range specs {
		parsedXlogs, err := p.ParseXlogSpec(logspec)
		if err != nil {
			return nil, err
		}
		xlogs = append(xlogs, parsedXlogs...)
	}
	return xlogs, nil
}

func (p xlogSpecParser) ParseXlogSpec(spec interface{}) ([]*XlogSrc, error) {
	if spec == nil {
		return []*XlogSrc{}, nil
	}
	switch act := spec.(type) {
	case string:
		return p.ParseXlogNamed(act)
	case map[interface{}]interface{}:
		return p.ParseXlogAliased(act)
	default:
		return nil, fmt.Errorf("Unexpected element %#v in xlogs for %s", spec, p.server.Name)
	}
}

func (p xlogSpecParser) ParseXlogNamed(name string) ([]*XlogSrc, error) {
	expandedNames, mustSync := SplitFilenamesMustSync(name)
	res := make([]*XlogSrc, len(expandedNames))
	for i, expandedName := range expandedNames {
		res[i] = p.NewXlogSrc(expandedName, "", mustSync)
	}
	return res, nil
}

func (p xlogSpecParser) ParseXlogAliased(aliased map[interface{}]interface{}) ([]*XlogSrc, error) {
	res := []*XlogSrc{}
	for iname, iqualifier := range aliased {
		name := iname.(string)
		qualifier := iqualifier.(string)
		expandedNames, mustSync := SplitFilenamesMustSync(name)
		for _, expandedName := range expandedNames {
			res = append(res, p.NewXlogSrc(expandedName, qualifier, mustSync))
		}
	}
	return res, nil
}

func (p xlogSpecParser) NewXlogSrc(name, qualifier string, mustSync bool) *XlogSrc {
	game := xlogtools.XlogGame(name)
	gameVersion := xlogtools.XlogGameVersion(name)
	logtype := xlogtools.FileType(name)
	qualifiedName := xlogtools.XlogQualifiedName(p.server.Name, game, gameVersion, qualifier, logtype)

	targetRelPath := URLTargetPath(p.server.Name, p.server.BaseURL, name)
	targetPath := path.Join(p.cachedir, targetRelPath)
	localPath := ""
	if p.server.LocalPathBase != "" {
		localPath = path.Join(p.server.LocalPathBase, name)
	}
	return &XlogSrc{
		Server:        p.server,
		Name:          name,
		Qualifier:     qualifier,
		TargetRelPath: targetRelPath,
		TargetPath:    targetPath,
		CName:         qualifiedName,
		URL:           URLJoin(p.server.BaseURL, name),
		LocalPath:     localPath,
		Live:          mustSync,
		Type:          logtype,
		Game:          game,
		GameVersion:   gameVersion,
	}
}

// SplitFilenamesMustSync takes a name like "foo{bar,baz}*", strips
// the * that designates that the filename is in active use, and
// expands the file glob.
func SplitFilenamesMustSync(name string) ([]string, bool) {
	mustSync := false
	if strings.Index(name, "*") != -1 {
		name = strings.Replace(name, "*", "", -1)
		mustSync = true
	}
	groups, err := text.ExpandBraceGroups(name)
	if err != nil {
		return []string{name}, mustSync
	}
	return groups, mustSync
}
