package stringnorm

import (
	"errors"
)

// ErrNormalizeComplete is a sentinel value returned by a normalizer to
// request that no other normalizers be run.
var ErrNormalizeComplete = errors.New("ErrNormalizeComplete")

// A Normalizer normalizes a string value.
type Normalizer interface {
	// Normalize copies the given text and normalizes and returns the copy.
	// If the normalizer does not recognize the text, it must return the
	// original text.
	// If the normalizer wishes to declare its result final, it must return
	// the next text and ErrNormalizeComplete.
	// If the normalizer wishes to reject the text as invalid, it may return
	// any other error.
	Normalize(text string) (string, error)
}

// A List of Normalizers, which applies each Normalizer in order.
type List []Normalizer

// Normalize applies each normalizer in n to text, returning the final value.
// If any normalizer returns ErrNormalizeComplete, the remaining normalizers
// are short-circuited.
func (n List) Normalize(text string) (string, error) {
	var err error
	for _, norm := range n {
		text, err = norm.Normalize(text)
		if err != nil {
			if err == ErrNormalizeComplete {
				return text, nil
			}
			return text, err
		}
	}
	return text, nil
}

// Normalize applies the list of string normalizers to text.
func Normalize(normalizers []Normalizer, text string) (string, error) {
	return List(normalizers).Normalize(text)
}

// Combine combines a list of normalizers into a single Normalizer
// instance that applies each normalizer in order as a List does.
func Combine(normalizers ...Normalizer) Normalizer {
	combined := make(List, 0, len(normalizers))
	for _, norm := range normalizers {
		if norm != nil {
			combined = append(combined, norm)
		}
	}
	if len(combined) == 1 {
		return combined[0]
	}
	return combined
}

// NormalizeNoErr applies normalizer to text; errors are silently ignored,
// and the original text is returned on error.
func NormalizeNoErr(normalizer Normalizer, text string) string {
	res, err := normalizer.Normalize(text)
	if err != nil {
		return text
	}
	return res
}
