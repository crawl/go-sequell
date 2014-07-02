package killer

import (
	"github.com/greensnark/go-sequell/stringnorm"
	"github.com/greensnark/go-sequell/xlog"
	"regexp"
)

func NormalizeKiller(killer string, rec xlog.Xlog) string {
	for _, norm := range normalizers {
		killer, err := norm.NormalizeKiller(killer, rec)
		if err != nil {
			return killer
		}
	}
	return killer
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
	uniqueNormalizer{},
}

type simpleKillerNormalizer stringnorm.Normalizer

func (n simpleKillerNormalizer) NormalizeKiller(killer string, rec xlog.Xlog) (string, error) {
	return stringnorm.Normalizer(n).Normalize(killer)
}

func reNorm(re, repl string) killerNormalizer {
	return simpleKillerNormalizer(stringnorm.StaticRegexpNormalizer(re, repl))
}