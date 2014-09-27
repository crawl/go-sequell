package xlogtools

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/greensnark/go-sequell/crawl/data"
	"github.com/greensnark/go-sequell/crawl/god"
	"github.com/greensnark/go-sequell/crawl/killer"
	"github.com/greensnark/go-sequell/crawl/place"
	"github.com/greensnark/go-sequell/crawl/player"
	"github.com/greensnark/go-sequell/crawl/version"
	"github.com/greensnark/go-sequell/qyaml"
	"github.com/greensnark/go-sequell/stringnorm"
	"github.com/greensnark/go-sequell/text"
	"github.com/greensnark/go-sequell/xlog"
)

type XlogType int

const (
	Unknown XlogType = iota
	Log
	Milestone
)

func (x XlogType) String() string {
	switch x {
	case Log:
		return "logfile"
	case Milestone:
		return "milestones"
	}
	return "unk"
}

func (x XlogType) BaseTable() string {
	switch x {
	case Log:
		return "logrecord"
	case Milestone:
		return "milestone"
	}
	return ""
}

func FileType(filename string) XlogType {
	if strings.Index(strings.ToLower(filename), "milestone") != -1 {
		return Milestone
	}
	return Log
}

func Type(line xlog.Xlog) XlogType {
	if _, ok := line["type"]; ok {
		return Milestone
	}
	return Log
}

func NormalizeBool(b string) string {
	if b != "" && b != "0" {
		return "t"
	}
	return "f"
}

func ValidXlog(log xlog.Xlog) bool {
	return log["start"] != "" && log["name"] != "" &&
		(log["end"] != "" || log["time"] != "")
}

type Normalizer struct {
	CrawlData qyaml.Yaml
	GodNorm   stringnorm.Normalizer
	PlaceNorm stringnorm.Normalizer
	CharNorm  *player.CharNormalizer
	FieldGens []*FieldGen
}

func MustBuildNormalizer(crawlData qyaml.Yaml) *Normalizer {
	norm, err := BuildNormalizer(crawlData)
	if err != nil {
		panic(err)
	}
	return norm
}

func BuildNormalizer(crawlData qyaml.Yaml) (*Normalizer, error) {
	fieldGen, err := ParseFieldGenerators(crawlData.Map("field-input-transforms"))
	if err != nil {
		return nil, err
	}
	return &Normalizer{
		CrawlData: crawlData,
		GodNorm:   god.Normalizer(crawlData.StringMap("god-aliases")),
		PlaceNorm: place.Normalizer(crawlData.StringMap("place-fixups")),
		CharNorm:  player.StockCharNormalizer(crawlData),
		FieldGens: fieldGen,
	}, nil
}

func (n *Normalizer) NormalizeLog(log xlog.Xlog) (xlog.Xlog, error) {
	n.CanonicalizeFields(log)

	log["v"] = version.FullVersion(log["v"])
	log["cv"] = version.MajorVersion(log["v"])
	if version.IsAlpha(log["v"]) {
		log["alpha"] = "t"
	}
	if log["alpha"] != "" {
		log["cv"] += "-a"
	}
	log["vnum"] = strconv.FormatUint(version.VersionNumericId(log["v"]), 10)
	log["cvnum"] = strconv.FormatUint(version.VersionNumericId(log["cv"]), 10)
	log["vlongnum"] =
		strconv.FormatUint(version.VersionNumericId(log["vlong"]), 10)
	log["tiles"] = NormalizeBool(log["tiles"])
	log["wiz"] = NormalizeBool(log["wiz"])
	if log["ntv"] == "" {
		log["ntv"] = "0"
	}
	log["place"] = stringnorm.NormalizeNoErr(n.PlaceNorm, log["place"])
	log["oplace"] = stringnorm.NormalizeNoErr(n.PlaceNorm, log["oplace"])
	log["br"] = stringnorm.NormalizeNoErr(n.PlaceNorm, log["br"])
	if god, err := n.GodNorm.Normalize(log["god"]); err == nil {
		log["god"] = god
	}
	log["char"] = n.CharNorm.NormalizeChar(log["crace"], log["cls"], log["char"])
	log["rstart"] = log["start"]
	log["game_key"] = log["name"] + ":" + log["src"] + ":" + log["rstart"]

	milestone := Type(log) == Milestone
	if milestone {
		log["verb"] = log["type"]
		log["noun"] = text.FirstNotEmpty(log["milestone"], "?")
		log["rtime"] = log["time"]
		log["oplace"] = text.FirstNotEmpty(log["oplace"], log["place"])
		NormalizeMilestoneFields(log)
	} else {
		log["vmsg"] = text.FirstNotEmpty(log["vmsg"], log["tmsg"])
		log["map"] = NormalizeMapName(log["map"])
		log["killermap"] = NormalizeMapName(log["killermap"])
		log["ikiller"] = text.FirstNotEmpty(log["ikiller"], log["killer"])
		log["ckiller"] =
			killer.NormalizeKiller(
				text.FirstNotEmpty(log["killer"], log["ktyp"]),
				log["killer"], log["killer_flags"])
		log["cikiller"] =
			killer.NormalizeKiller(log["ikiller"], log["ikiller"], "")
		log["kmod"] = killer.NormalizeKmod(log["killer"])
		log["ckaux"] = killer.NormalizeKaux(log["kaux"])
		log["rend"] = log["end"]
	}
	sanitizeGold(log)

	return log, nil
}

func NormalizeMapName(mapname string) string {
	return strings.Replace(mapname, ",", ";", -1)
}

var milestoneVerbMap = data.Crawl.StringMap("milestone-verb-mappings")
var rActionWord = regexp.MustCompile(`(\w+) (.*?)\.?$`)
var rGhostWord = regexp.MustCompile(`(\w+) the ghost of (\S+)`)
var rAbyssCause = regexp.MustCompile(`\((.*?)\)$`)
var rSacrificedThing = regexp.MustCompile(`sacrificed (?:an? )?(\w+)`)

func NormalizeMilestoneFields(log xlog.Xlog) {
	verb := log["verb"]
	if mappedVerb, ok := milestoneVerbMap[verb]; ok {
		log["verb"] = mappedVerb
		verb = mappedVerb
	}

	noun := log["noun"]
	switch verb {
	case "sacrifice":
		sacMatch := rSacrificedThing.FindStringSubmatch(noun)
		if sacMatch != nil {
			noun = strings.TrimSpace(sacMatch[1])
		}
	case "uniq":
		actionMatch := rActionWord.FindStringSubmatch(noun)
		if actionMatch != nil {
			actionWord, actedUpon := actionMatch[1], actionMatch[2]
			noun = actedUpon
			verb = qualifyVerbAction(verb, actionWord)
		}
	case "ghost":
		ghostMatch := rGhostWord.FindStringSubmatch(noun)
		if ghostMatch != nil {
			verb = qualifyVerbAction(verb, ghostMatch[1])
			noun = ghostMatch[2]
		}
	case "abyss.enter":
		abyssCauseMatch := rAbyssCause.FindStringSubmatch(noun)
		if abyssCauseMatch != nil {
			noun = text.FirstNotEmpty(abyssCauseMatch[1], "?")
		}
	case "br.enter", "br.end", "br.mid":
		noun = place.StripPlaceDepth(log["place"])
	case "br.exit":
		noun = place.StripPlaceDepth(log["oplace"])
	case "rune":
		noun = FoundRuneName(noun)
	case "orb":
		noun = "orb"
	case "god.mollify":
		noun = MollifiedGodName(noun)
	case "god.renounce":
		noun = RenouncedGodName(noun)
	case "god.worship":
		noun = WorshippedGodName(noun)
	case "god.maxpiety":
		noun = MaxedPietyGodName(noun)
	case "monstrous":
		noun = "demonspawn"
	case "shaft":
		noun = ShaftedPlace(noun)
	}
	log["verb"] = verb
	if noun != "" {
		log["noun"] = noun
	}
}

func qualifyVerbAction(verb string, actionWord string) string {
	if actionWord == "banished" {
		return verb + ".ban"
	}
	if actionWord == "pacified" {
		return verb + ".pac"
	}
	if actionWord == "enslaved" {
		return verb + ".ens"
	}
	return verb
}

var rFoundRune = regexp.MustCompile(`found an? (\S+) rune`)

func textReSubmatch(text string, reg *regexp.Regexp, submatch int) string {
	m := reg.FindStringSubmatch(text)
	if m != nil {
		return m[submatch]
	}
	return text
}

func FoundRuneName(found string) string {
	return textReSubmatch(found, rFoundRune, 1)
}

var rMollifiedGod = regexp.MustCompile(`^(?:partially )?mollified (.*)[.]$`)

func MollifiedGodName(mollifiedMsg string) string {
	return textReSubmatch(mollifiedMsg, rMollifiedGod, 1)
}

var rRenouncedGod = regexp.MustCompile(`^abandoned (.*)[.]$`)

func RenouncedGodName(renounceMsg string) string {
	return textReSubmatch(renounceMsg, rRenouncedGod, 1)
}

var rMaxedPietyGod = regexp.MustCompile(`^became the Champion of (.*)[.]$`)

func MaxedPietyGodName(maxpietyMsg string) string {
	return textReSubmatch(maxpietyMsg, rMaxedPietyGod, 1)
}

var rWorshippedGod = regexp.MustCompile(`^became a worshipper of (.*)[.]$`)

func WorshippedGodName(worshipMsg string) string {
	return textReSubmatch(worshipMsg, rWorshippedGod, 1)
}

var rShaftedPlace = regexp.MustCompile(`fell down a shaft to (.*)[.]$`)

func ShaftedPlace(shaftMsg string) string {
	return textReSubmatch(shaftMsg, rShaftedPlace, 1)
}

func (n *Normalizer) CanonicalizeFields(log xlog.Xlog) {
	for _, gen := range n.FieldGens {
		gen.Apply(log)
	}
}

func sanitizeGold(log xlog.Xlog) {
	if text.ParseInt(log["gold"], 0) < 0 ||
		text.ParseInt(log["goldfound"], 0) < 0 ||
		text.ParseInt(log["goldspent"], 0) < 0 {
		log["gold"] = "0"
		log["goldfound"] = "0"
		log["goldspent"] = "0"
	}
}
