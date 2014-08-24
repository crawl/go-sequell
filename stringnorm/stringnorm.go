package stringnorm

import (
	"errors"
)

var ErrNormalizeComplete error = errors.New("ErrNormalizeComplete")

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

type List []Normalizer

func (n List) Normalize(text string) (string, error) {
	for _, norm := range n {
		text, err := norm.Normalize(text)
		if err != nil {
			if err == ErrNormalizeComplete {
				return text, nil
			}
			return text, err
		}
	}
	return text, nil
}

func Normalize(normalizers []Normalizer, text string) (string, error) {
	return List(normalizers).Normalize(text)
}

func NormalizeNoErr(normalizer Normalizer, text string) string {
	res, err := normalizer.Normalize(text)
	if err != nil {
		return text
	}
	return res
}
