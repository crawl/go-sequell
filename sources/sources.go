package sources

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/crawl/go-sequell/crawl/ctime"
	"github.com/crawl/go-sequell/crawl/xlogtools"
)

// Servers is the list of servers are sources for games and milestones
type Servers []*Server

// MkdirTargets creates all directories needed for all copies of
// remote logs.
func (x Servers) MkdirTargets() error {
	for _, server := range x {
		for _, log := range server.Logfiles {
			if err := log.MkdirTarget(); err != nil {
				return err
			}
		}
	}
	return nil
}

// XlogSources returns the list of all xlog sources
func (x Servers) XlogSources() []*XlogSrc {
	var sources []*XlogSrc
	addAll := func(logs []*XlogSrc) {
		for _, log := range logs {
			sources = append(sources, log)
		}
	}
	for _, server := range x {
		addAll(server.Logfiles)
	}
	return sources
}

// TargetLogDirs returns the set of target (local copy) log directories
// for all log files.
func (x Servers) TargetLogDirs() []string {
	targetDirs := []string{}
	seenDirs := map[string]bool{}
	for _, server := range x {
		for _, log := range server.Logfiles {
			if dir := log.TargetDir(); !seenDirs[dir] {
				seenDirs[dir] = true
				targetDirs = append(targetDirs, dir)
			}
		}
	}
	return targetDirs
}

// Server returns the server specified by alias.
func (x Servers) Server(alias string) *Server {
	for _, server := range x {
		if server.Name == alias || server.Aliases[alias] {
			return server
		}
	}
	return nil
}

func (x Servers) String() string {
	buf := bytes.Buffer{}
	buf.WriteString("Sources[")
	for i, server := range x {
		if i > 0 {
			buf.WriteString("; ")
		}
		buf.WriteString(server.String())
	}
	buf.WriteString("]")
	return buf.String()
}

// A Server represents an online game server that supplies logfiles and
// milestones
type Server struct {
	Name          string
	Aliases       map[string]bool
	BaseURL       string
	LocalPathBase string
	TimeZoneMap   ctime.DSTLocation
	UtcEpoch      time.Time
	Logfiles      []*XlogSrc
}

// ParseLogTime parses a timestamp as read from a server's logfile in the
// correct timezone for that server. Timezones are treated as UTC unless the
// server was operating before Crawl logfile timestamps switched to UTC.
func (s *Server) ParseLogTime(logtime string) (time.Time, error) {
	return ctime.ParseLogTime(logtime, s.UtcEpoch, s.TimeZoneMap)
}

func (s *Server) String() string {
	return fmt.Sprintf(
		"%s(aliases=%#v base=%s local=%s tz=%s epoch=%s nlog=%d",
		s.Name, s.Aliases, s.BaseURL, s.LocalPathBase, s.TimeZoneMap.String(),
		s.UtcEpoch, len(s.Logfiles))
}

// An XlogSrc is a reference to an xlogfile hosted on a server somewhere, and
// to its cached (target) path.
type XlogSrc struct {
	Server        *Server
	Name          string
	Qualifier     string
	LocalPath     string
	URL           string
	TargetPath    string
	TargetRelPath string
	CName         string
	Live          bool
	Type          xlogtools.XlogType
	Game          string
	GameVersion   string
}

func (x *XlogSrc) String() string {
	return "Src" + x.liveAsterisk() + "[" + x.Type.String() + ": " + x.URL + " > " + x.TargetPath + "]"
}

func (x *XlogSrc) liveAsterisk() string {
	if x.Live {
		return "*"
	}
	return ""
}

// TargetDir returns the directory that contains the xlogfile's local cache
func (x *XlogSrc) TargetDir() string {
	return filepath.Dir(x.TargetPath)
}

// MkdirTarget creates the parent directory of x.TargetPath.
func (x *XlogSrc) MkdirTarget() error {
	return os.MkdirAll(x.TargetDir(), os.ModeDir|0755)
}

// Local checks if this xlogfile is local to the system running Sequell. This
// can only be true if a) Sequell is running on an active Crawl server and
// reading from a live local log or b) Sequell is running in test mode with
// local copies of logs.
func (x *XlogSrc) Local() bool {
	if x.LocalPath == "" {
		return false
	}
	_, err := os.Stat(x.LocalPath)
	return err == nil
}

// TargetExists checks if the local cached logfile copy exists
func (x *XlogSrc) TargetExists() bool {
	_, err := os.Stat(x.TargetPath)
	return err == nil
}

// NeedsFetch checks if the logfile should be downloaded from the remote
// server. (It does not check if the cached copy is stale, only if this is a
// file that must be periodically downloaded.)
func (x *XlogSrc) NeedsFetch() bool {
	return x.Live && !x.Local()
}

// LinkLocal links a local logfile to the target path. Fails if x is not a
// local log, or the symlink attempt fails.
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

// DownloadURL gets the URL to download the remote xlog from, or an empty
// string if there is no known URL (such as for local files).
func (x *XlogSrc) DownloadURL() string {
	if x.Local() {
		return x.URL
	}
	return ""
}
