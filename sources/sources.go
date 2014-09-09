package sources

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/greensnark/go-sequell/crawl/ctime"
	"github.com/greensnark/go-sequell/crawl/xlogtools"
)

type Servers struct {
	Servers []*Server
}

func (x *Servers) XlogSources() []*XlogSrc {
	sources := []*XlogSrc{}
	addAll := func(logs []*XlogSrc) {
		for _, log := range logs {
			sources = append(sources, log)
		}
	}
	for _, s := range x.Servers {
		addAll(s.Logfiles)
		addAll(s.Milestones)
	}
	return sources
}

func (x *Servers) Server(alias string) *Server {
	for _, s := range x.Servers {
		if s.Name == alias || s.Aliases[alias] {
			return s
		}
	}
	return nil
}

func (x *Servers) String() string {
	buf := bytes.Buffer{}
	buf.WriteString("Sources[")
	for i, srv := range x.Servers {
		if i > 0 {
			buf.WriteString("; ")
		}
		buf.WriteString(srv.String())
	}
	buf.WriteString("]")
	return buf.String()
}

type Server struct {
	Name          string
	Aliases       map[string]bool
	BaseURL       string
	LocalPathBase string
	TimeZoneMap   ctime.DstLocation
	UtcEpoch      time.Time
	Logfiles      []*XlogSrc
	Milestones    []*XlogSrc
}

func (s *Server) String() string {
	return fmt.Sprintf(
		"%s(aliases=%#v base=%s local=%s tz=%s epoch=%s nlog=%d nmile=%d",
		s.Name, s.Aliases, s.BaseURL, s.LocalPathBase, s.TimeZoneMap.String(),
		s.UtcEpoch, len(s.Logfiles), len(s.Milestones))
}

type XlogSrc struct {
	Server      *Server
	Name        string
	Qualifier   string
	LocalPath   string
	Url         string
	TargetPath  string
	Live        bool
	Type        xlogtools.XlogType
	Game        string
	GameVersion string
}

func (x *XlogSrc) String() string {
	return "Src" + x.liveAsterisk() + "[" + x.Type.String() + ": " + x.Url + " > " + x.TargetPath + "]"
}

func (x *XlogSrc) liveAsterisk() string {
	if x.Live {
		return "*"
	}
	return ""
}

func (x *XlogSrc) Local() bool {
	if x.LocalPath == "" {
		return false
	}
	_, err := os.Stat(x.LocalPath)
	return err == nil
}

func (x *XlogSrc) TargetExists() bool {
	_, err := os.Stat(x.TargetPath)
	return err == nil
}

func (x *XlogSrc) NeedsFetch() bool {
	return x.Live && !x.Local()
}

func (x *XlogSrc) LinkLocal() error {
	if !x.Local() {
		return fmt.Errorf("%s is not a local xlog", x.String())
	}
	if x.TargetPath == "" {
		return fmt.Errorf("No target path for %s", x.String())
	}
	if x.TargetPath == x.LocalPath {
		return nil
	}
	if fi, err := os.Stat(x.TargetPath); err == nil {
		if (fi.Mode() & os.ModeSymlink) == 0 {
			return fmt.Errorf("%s: %s exists and is not a symlink", x, x.TargetPath)
		}
		return nil
	}
	return os.Symlink(x.LocalPath, x.TargetPath)
}

func (x *XlogSrc) DownloadURL() string {
	if x.Local() {
		return x.Url
	}
	return ""
}
