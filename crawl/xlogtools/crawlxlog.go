package xlogtools

import (
	"github.com/greensnark/go-sequell/crawl/place"
	"github.com/greensnark/go-sequell/crawl/version"
	"github.com/greensnark/go-sequell/xlog"
	"strings"
)

type XlogType int

const (
	Unknown XlogType = iota
	Log
	Milestone
)

func Type(line xlog.Xlog) XlogType {
	if _, ok := line["type"]; ok {
		return Milestone
	}
	return Log
}

func Normalize(log xlog.Xlog) xlog.Xlog {
	if Type(log) == Milestone {
		return NormalizeMilestone(log)
	} else {
		return NormalizeLog(log)
	}
}

func NormalizeBool(b string) string {
	if b != "" {
		return "y"
	}
	return b
}

func NormalizeLog(log xlog.XlogLine) xlog.XlogLine {
	log["v"] = version.FullVersion(log["v"])
	log["cv"] = version.MajorVersion(log["v"])
	if log["alpha"] != "" {
		log["cv"] += "-a"
	}
	log["vnum"] = version.VersionNumericId(log["v"])
	log["cvnum"] = version.VersionNumericId(log["cv"])
	log["tiles"] = NormalizeBool(log["tiles"])
	if log["ntv"] == "" {
		log["ntv"] = "0"
	}
	log["place"] = place.CanonicalPlace(log["place"])
	log["oplace"] = place.CanonicalPlace(log["oplace"])
	log["br"] = place.CanonicalPlace(log["br"])
	log["god"] = god.CanonicalGod(log["god"])

	milestone := Type(log) == Milestone
	if milestone {
		log["oplace"] = FirstNotEmpty(log["oplace"], log["place"])
	}

	if !milestone {
		log["vmsg"] = FirstNotEmpty(log["vsmg"], log["tmsg"])
		log["map"] = NormalizeMapName(log["map"])
		log["killermap"] = NormalizeMapName(log["killermap"])
		log["ikiller"] = FirstNotEmpty(log["ikiller"], log["killer"])
		log["ckiller"] =
			killer.NormalizeKiller(
				FirstNotEmpty(log["killer"], log["ktyp"]), log["killer"])
	}
	return log
}

func FirstNotEmpty(choices ...string) string {
	for _, val := range choices {
		if val != "" {
			return val
		}
	}
	return ""
}

func NormalizeMapName(mapname string) string {
	return strings.Replace(mapname, ",", ";", -1)
}

func NormalizeMilestone(mile xlog.Xlog) xlog.Xlog {
	return NormalizeLog(mile)
}
