package killer

import (
	"regexp"
	"strings"

	"github.com/crawl/go-sequell/grammar"
	"github.com/crawl/go-sequell/stringnorm"
	"github.com/crawl/go-sequell/text"
)

type articleNorm struct {
	specialCases []string
}

func ArticleNormalizer(specialCases []string) *articleNorm {
	return &articleNorm{specialCases: specialCases}
}

// Normalize prefixes the given killer name with "a" or "an" if it seems
// appropriate. Similar to grammar.Article, but returns apostrophised
// names and other special-cases unmodified.
func (a *articleNorm) Normalize(killer string) (string, error) {
	if strings.IndexRune(killer, '\'') != -1 {
		return killer, nil
	}
	for _, specialCase := range a.specialCases {
		if strings.Index(killer, specialCase) != -1 {
			return killer, nil
		}
	}
	return grammar.Article(killer), nil
}

func NormalizeKiller(killer, rawKiller, killerFlags string) string {
	var err error
	for _, norm := range normalizers {
		killer, err = norm.NormalizeKiller(killer, rawKiller, killerFlags)
		if err != nil {
			return killer
		}
	}
	return killer
}

var rSpectralThing = regexp.MustCompile(`spectral (\w+)`)
var rDerivedUndead = regexp.MustCompile(`(?i)(zombie|skeleton|simulacrum)$`)

func NormalizeKmod(killer string) string {
	spectralMatch := rSpectralThing.FindStringSubmatch(killer)
	if spectralMatch != nil && spectralMatch[1] != "warrior" {
		return "spectre"
	}

	if strings.Index(killer, "shapeshifter") != -1 {
		return "shapeshifter"
	}

	derivedUndeadMatch := rDerivedUndead.FindStringSubmatch(killer)
	if derivedUndeadMatch != nil {
		return derivedUndeadMatch[1]
	}
	return ""
}

var kauxNormalizers = stringnorm.List{
	stringnorm.SR(`^Hit by (.*) thrown .*$`, "$1"),
	stringnorm.SR(`^Shot with (.*) by .*$`, "$1"),
	stringnorm.SR(`\{.*?\}`, ""),
	stringnorm.SR(`\(.*?\)`, ""),
	stringnorm.SR(`[+-]\d+,?\s*`, ""),
	stringnorm.SR(`^an? `, ""),
	stringnorm.SR(`(?:elven|orcish|dwarven) `, ""),
	stringnorm.SR(`\b(?:un)?cursed `, ""),
}

func NormalizeKaux(kaux string) string {
	return text.NormalizeSpace(stringnorm.NormalizeNoErr(kauxNormalizers, kaux))
}

type killerNormalizer interface {
	NormalizeKiller(killer, killerRawValue, killerFlags string) (string, error)
}

type killerNormFunc func(string, string, string) (string, error)

func (k killerNormFunc) NormalizeKiller(killer, killerRaw, killerFlags string) (string, error) {
	return k(killer, killerRaw, killerFlags)
}

var normalizers = []killerNormalizer{
	reNorm(`^an? \w+-headed (hydra.*)$`, "a $1"),
	reNorm(`^the \w+-headed ((?:Lernaean )?hydra.*)$`, "the $1"),
	reNorm(`^.*'s? ghost$`, "a player ghost"),
	reNorm(`^.*'s? illusion$`, "a player illusion"),
	reNorm(`^an? \w+ (draconian.*)`, "a $1"),
	killerNormFunc(func(killer, raw, flags string) (string, error) {
		if strings.Index(killer, "very ugly thing") != -1 {
			return "a very ugly thing", nil
		}
		if strings.Index(killer, "ugly thing") != -1 {
			return "an ugly thing", nil
		}
		return killer, nil
	}),
	reNorm(`^an? .* \(((?:glowing )?shapeshifter)\)$`, "a $1"),
	reNorm(`^the .* shaped (.*)$`, "the $1"),
	reNorm(`.*\bmiscasting\b.*$`, "miscast"),
	reNorm(`.*\bunwield\b.*`, "unwield"),
	&uniqueNormalizer{},
}

type simpleKillerNormalizer struct {
	norm stringnorm.Normalizer
}

func (n *simpleKillerNormalizer) NormalizeKiller(killer, rawKiller, killerFlags string) (string, error) {
	return n.norm.Normalize(killer)
}

func reNorm(re, repl string) killerNormalizer {
	return &simpleKillerNormalizer{stringnorm.StaticReg(re, repl)}
}
