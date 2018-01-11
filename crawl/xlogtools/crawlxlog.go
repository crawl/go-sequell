package xlogtools

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/crawl/go-sequell/crawl/data"
	"github.com/crawl/go-sequell/crawl/god"
	"github.com/crawl/go-sequell/crawl/killer"
	"github.com/crawl/go-sequell/crawl/place"
	"github.com/crawl/go-sequell/crawl/player"
	"github.com/crawl/go-sequell/crawl/version"
	"github.com/crawl/go-sequell/qyaml"
	"github.com/crawl/go-sequell/stringnorm"
	"github.com/crawl/go-sequell/text"
	"github.com/crawl/go-sequell/xlog"
)

// XlogType is the type of an xlog file: a logfile, a milestones file, or
// unknown.
type XlogType int

// XlogType constants
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

// BaseTable returns the base table name (logrecord or milestone) for x
func (x XlogType) BaseTable() string {
	switch x {
	case Log:
		return "logrecord"
	case Milestone:
		return "milestone"
	}
	return ""
}

// FileType guesses whether the filename supplied is a logfile or milestone
func FileType(filename string) XlogType {
	if strings.Index(strings.ToLower(filename), "milestone") != -1 {
		return Milestone
	}
	return Log
}

// Type guesses whether a given xlog line is a milestone or a logfile entry
func Type(line xlog.Xlog) XlogType {
	if _, ok := line["type"]; ok {
		return Milestone
	}
	return Log
}

// NormalizeBool converts an arbitrary string b into a boolean "t" or "f". Any
// value of b other than the empty string, "0" and "f" are treated as true.
func NormalizeBool(b string) string {
	if b != "" && b != "0" && b != "f" {
		return "t"
	}
	return "f"
}

// ValidXlog returns true if log seems to be a valid xlog line.
func ValidXlog(log xlog.Xlog) bool {
	return log["start"] != "" && log["name"] != "" &&
		(log["end"] != "" || log["time"] != "")
}

// Normalizer is a set of normalizations that are applied to xlog fields to
// canonicalize an xlog record
type Normalizer struct {
	CrawlData        qyaml.YAML
	GodNorm          stringnorm.Normalizer
	PlaceNorm        stringnorm.Normalizer
	CharNorm         *player.CharNormalizer
	KillerArticle    stringnorm.Normalizer
	FieldGens        []*FieldGen
	milestoneVerbMap map[string]string
	knorm            *killer.Normalizer
}

// MustBuildNormalizer creates an xlog normalizer given a crawlData config. It
// panics when it cannot create a normalizer because of config or other errors.
func MustBuildNormalizer(crawlData qyaml.YAML) *Normalizer {
	norm, err := BuildNormalizer(crawlData)
	if err != nil {
		panic(err)
	}
	return norm
}

// BuildNormalizer creates an xlog Normalizer given the crawlData YAML
// structure.
func BuildNormalizer(crawlData qyaml.YAML) (*Normalizer, error) {
	fieldGen, err := ParseFieldGenerators(crawlData.Slice("field-input-transforms"))
	if err != nil {
		return nil, err
	}
	return &Normalizer{
		CrawlData:        crawlData,
		GodNorm:          god.Normalizer(crawlData.StringMap("god-aliases")),
		PlaceNorm:        place.Normalizer(crawlData.StringMap("place-fixups")),
		CharNorm:         player.StockCharNormalizer(crawlData),
		KillerArticle:    killer.ArticleNormalizer(crawlData.StringSlice("special-killer-phrases")),
		FieldGens:        fieldGen,
		milestoneVerbMap: crawlData.StringMap("milestone-verb-mappings"),
		knorm:            killer.NewNormalizer(data.Crawl{YAML: crawlData}),
	}, nil
}

// NormalizeLog normalizes an xlog record, cleaning up fields and generating
// fields that Sequell wants.
func (n *Normalizer) NormalizeLog(log xlog.Xlog) error {
	n.CanonicalizeFields(log)

	log["v"] = version.Full(log["v"])
	cv := version.Major(log["v"])
	if version.IsAlpha(log["v"]) {
		log["alpha"] = "t"
		cv += "-a"
	}
	log["cv"] = cv
	log["vnum"] = strconv.FormatUint(version.NumericID(log["v"]), 10)
	log["cvnum"] = strconv.FormatUint(version.NumericID(cv), 10)
	log["vsavrv"] = version.StripVCSQualifier(log["vsavrv"])
	log["vsavnum"] = strconv.FormatUint(version.NumericID(log["vsav"]), 10)
	log["vsavrvnum"] = strconv.FormatUint(version.NumericID(log["vsavrv"]), 10)
	log["vlongnum"] =
		strconv.FormatUint(version.NumericID(log["vlong"]), 10)
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

	if banisher, ok := log["banisher"]; ok {
		var err error
		if banisher, err = n.KillerArticle.Normalize(banisher); err != nil {
			return err
		}
		log["banisher"] = banisher
		log["cbanisher"] = n.knorm.NormalizeKiller(cv, banisher, banisher, "")
	}

	milestone := Type(log) == Milestone
	if milestone {
		log["verb"] = log["type"]
		log["noun"] = text.FirstNotEmpty(log["milestone"], "?")
		log["rtime"] = log["time"]
		log["oplace"] = text.FirstNotEmpty(log["oplace"], log["place"])
		n.NormalizeMilestoneFields(log)
	} else {
		log["vmsg"] = text.FirstNotEmpty(log["vmsg"], log["tmsg"])
		log["map"] = NormalizeMapName(log["map"])
		log["killermap"] = NormalizeMapName(log["killermap"])
		log["ikiller"] = text.FirstNotEmpty(log["ikiller"], log["killer"])
		log["ckiller"] =
			n.knorm.NormalizeKiller(cv,
				text.FirstNotEmpty(log["killer"], log["ktyp"]),
				log["killer"], log["killer_flags"])
		log["cikiller"] =
			n.knorm.NormalizeKiller(cv, log["ikiller"], log["ikiller"], "")
		log["kmod"] = killer.NormalizeKmod(log["killer"])
		log["ckaux"] = killer.NormalizeKaux(log["kaux"])
		log["rend"] = log["end"]
	}
	sanitizeGold(log)

	return nil
}

// NormalizeMapName cleans up a map field, converting ","-separated map names
// to ";"-separated names.
func NormalizeMapName(mapname string) string {
	return strings.Replace(mapname, ",", ";", -1)
}

var rActionWord = regexp.MustCompile(`(\w+) (.*?)\.?$`)
var rGhostWord = regexp.MustCompile(`(\w+) the ghost of (\S+)`)
var rAbyssCause = regexp.MustCompile(`\((.*?)\)$`)
var rSacrificedThing = regexp.MustCompile(`sacrificed (?:an? )?(\w+)`)
var rAncestorType = regexp.MustCompile(`.* as a (.+)[.]$`)
var rAncestorSpecial = regexp.MustCompile(`.*(?:casting|wielding an?) (.+)[.]$`)

// NormalizeMilestoneFields normalizes a milestone record, cleaning up fields
// and converting them to the canonical forms Sequell expects.
func (n *Normalizer) NormalizeMilestoneFields(log xlog.Xlog) {
	verb := log["verb"]
	if mappedVerb, ok := n.milestoneVerbMap[verb]; ok {
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
	case "ancestor.class":
		ancestorTypeMatch := rAncestorType.FindStringSubmatch(noun)
		if ancestorTypeMatch != nil {
			noun = ancestorTypeMatch[1]
		}
	case "ancestor.special":
		ancestorSpecialMatch := rAncestorSpecial.FindStringSubmatch(noun)
		if ancestorSpecialMatch != nil {
			noun = ancestorSpecialMatch[1]
		}
	case "br.enter", "br.end", "br.mid":
		noun = place.StripPlaceDepth(log["place"])
	case "br.exit":
		noun = place.StripPlaceDepth(log["oplace"])
	case "rune":
		noun = foundRuneName(noun)
	case "orb":
		noun = "orb"
	case "god.ecumenical":
		noun = log["god"]
	case "god.mollify":
		noun = mollifiedGodName(noun)
	case "god.renounce":
		noun = renouncedGodName(noun)
	case "god.worship":
		noun = worshippedGodName(noun)
	case "god.maxpiety":
		noun = maxedPietyGodName(noun)
	case "monstrous":
		noun = "demonspawn"
	case "shaft":
		noun = shaftedPlace(noun)
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
	if actionWord == "slimified" {
		return verb + ".slime"
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

func foundRuneName(found string) string {
	return textReSubmatch(found, rFoundRune, 1)
}

var rMollifiedGod = regexp.MustCompile(`^(?:partially )?mollified (.*)[.]$`)

func mollifiedGodName(mollifiedMsg string) string {
	return textReSubmatch(mollifiedMsg, rMollifiedGod, 1)
}

var rRenouncedGod = regexp.MustCompile(`^abandoned (.*)[.]$`)

// renouncedGodName gets the god name from a god.renounce message.
func renouncedGodName(renounceMsg string) string {
	return textReSubmatch(renounceMsg, rRenouncedGod, 1)
}

var rMaxedPietyGod = regexp.MustCompile(`^became the Champion of (.*)[.]$`)

func maxedPietyGodName(maxpietyMsg string) string {
	return textReSubmatch(maxpietyMsg, rMaxedPietyGod, 1)
}

var rWorshippedGod = regexp.MustCompile(`^became a worshipper of (.*)[.]$`)

func worshippedGodName(worshipMsg string) string {
	return textReSubmatch(worshipMsg, rWorshippedGod, 1)
}

var rShaftedPlace = regexp.MustCompile(`fell down a shaft to (.*)[.]$`)

func shaftedPlace(shaftMsg string) string {
	return textReSubmatch(shaftMsg, rShaftedPlace, 1)
}

// CanonicalizeFields applies all the normalizers in n to the xlog record log.
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
