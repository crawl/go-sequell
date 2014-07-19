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

func Normalize(text string, normalizers []Normalizer) (string, error) {
	for _, norm := range normalizers {
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
