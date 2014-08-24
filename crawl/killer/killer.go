package killer

import (
	"regexp"
	"strings"

	"github.com/greensnark/go-sequell/stringnorm"
	"github.com/greensnark/go-sequell/text"
	"github.com/greensnark/go-sequell/xlog"
)

func NormalizeKiller(killer string, rec xlog.Xlog) string {
	var err error
	for _, norm := range normalizers {
		killer, err = norm.NormalizeKiller(killer, rec)
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
	NormalizeKiller(killer string, rec xlog.Xlog) (string, error)
}

var normalizers = []killerNormalizer{
	reNorm(`^an? \w+-headed (hydra.*)$`, "a $1"),
	reNorm(`^the \w+-headed ((?:Lernaean )?hydra.*)$`, "the $1"),
	reNorm(`^.*'s? ghost$`, "a player ghost"),
	reNorm(`^.*'s? illusion$`, "a player illusion"),
	reNorm(`^an? \w+ (draconian.*)`, "a $1"),
	reNorm(`^an? .* \(((?:glowing )?shapeshifter)\)$`, "a $1"),
	reNorm(`^the .* shaped (.*)$`, "the $1"),
	&uniqueNormalizer{},
}

type simpleKillerNormalizer struct {
	norm stringnorm.Normalizer
}

func (n *simpleKillerNormalizer) NormalizeKiller(killer string, rec xlog.Xlog) (string, error) {
	return n.norm.Normalize(killer)
}

func reNorm(re, repl string) killerNormalizer {
	return &simpleKillerNormalizer{stringnorm.StaticReg(re, repl)}
}
