package stringnorm

// An ExactReplacer will substitute the text After if given the text Before,
// and will return other text unmodified.
type ExactReplacer struct {
	Before, After string
}

// Normalize returns e.After if text==e.Before, or text otherwise.
func (e *ExactReplacer) Normalize(text string) (string, error) {
	if text == e.Before {
		return e.After, nil
	}
	return text, nil
}

// ParseExactReplacers accepts a slice of string pairs, and constructs a
// List normalizer of ExactReplacers for each pair, with Before=pair[0] and
// After=pair[1].
func ParseExactReplacers(pairs [][]string) (Normalizer, error) {
	if pairs == nil {
		return nil, nil
	}
	replacers := make(List, len(pairs))
	for i, pair := range pairs {
		replacers[i] = &ExactReplacer{
			Before: pair[0],
			After:  pair[1],
		}
	}
	return replacers, nil
}
