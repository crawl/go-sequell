package xlogtools

import (
	"github.com/greensnark/go-sequell/crawl/version"
	"github.com/greensnark/go-sequell/xlog"
)

type XlogType int

const (
	Unknown XlogType = iota
	Log
	Milestone
)

func Type(line xlog.XlogLine) XlogType {
	if _, ok := line["type"]; ok {
		return Milestone
	}
	return Log
}

func Normalize(log xlog.XlogLine) xlog.XlogLine {
	if Type(log) == Milestone {
		return NormalizeMilestone(log)
	} else {
		return NormalizeLog(log)
	}
}

func NormalizeLog(log xlog.XlogLine) xlog.XlogLine {
	log["cv"] = version.CanonicalVersion(log["v"])
}
