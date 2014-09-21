package stringnorm

type ExactReplacer struct {
	Before, After string
}

func (e *ExactReplacer) Normalize(text string) (string, error) {
	if text == e.Before {
		return e.After, nil
	}
	return text, nil
}

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
