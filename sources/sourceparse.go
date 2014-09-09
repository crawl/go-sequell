package sources

import (
	"fmt"
	"path"
	"strings"

	"github.com/greensnark/go-sequell/crawl/ctime"
	"github.com/greensnark/go-sequell/crawl/xlogtools"
	"github.com/greensnark/go-sequell/qyaml"
	"github.com/greensnark/go-sequell/text"
)

func Sources(sources qyaml.Yaml, cachedir string) (*Servers, error) {
	return sourceYamlParser{
		sources:  sources.Slice("sources"),
		cachedir: cachedir,
	}.Parse()
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
	server   qyaml.Yaml
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
		s.ParseXlogRefs(&server, s.server.Slice("logfiles"), xlogtools.Log); err != nil {
		return nil, err
	}
	if server.Milestones, err =
		s.ParseXlogRefs(&server, s.server.Slice("milestones"), xlogtools.Milestone); err != nil {
		return nil, err
	}
	return &server, nil
}

func (s serverParser) ParseTimeZones(tzdst map[string]string) (ctime.DstLocation, error) {
	return ctime.ParseDstLocation(tzdst["S"], tzdst["D"])
}

func (s serverParser) ParseXlogRefs(
	server *Server,
	logfiles []interface{},
	logtype xlogtools.XlogType) ([]*XlogSrc, error) {
	return xlogSpecParser{
		server:   server,
		logtype:  logtype,
		cachedir: s.cachedir,
	}.Parse(logfiles)
}

type xlogSpecParser struct {
	server   *Server
	logtype  xlogtools.XlogType
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
	qualifiedName := xlogtools.XlogQualifiedName(p.server.Name, game, gameVersion, qualifier, p.logtype)
	localPath := ""
	if p.server.LocalPathBase != "" {
		localPath = path.Join(p.server.LocalPathBase, name)
	}
	return &XlogSrc{
		Server:      p.server,
		Name:        name,
		Qualifier:   qualifier,
		TargetPath:  path.Join(p.cachedir, qualifiedName),
		Url:         UrlJoin(p.server.BaseURL, name),
		LocalPath:   localPath,
		Live:        mustSync,
		Type:        p.logtype,
		Game:        game,
		GameVersion: gameVersion,
	}
}

// UrlJoin joins two URL path segments.
func UrlJoin(base, path string) string {
	if base == "" {
		return path
	}
	if base[len(base)-1] == '/' {
		return base + path
	}
	return base + "/" + path
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
