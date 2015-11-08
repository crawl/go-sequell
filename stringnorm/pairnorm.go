package stringnorm

import "regexp"

// MustParseRegexpPairs parses a list of regexp search + replacement expressions
// and returns a List normalizer that applies those search+replacements in
// sequence. Panics if there is any error constructing the List normalizer.
func MustParseRegexpPairs(pairs [][]string) List {
	res, err := ParseRegexpPairs(pairs)
	if err != nil {
		panic(err)
	}
	return res
}

// ParseRegexpPairs parses a list of regexp search + replacement expressions
// and returns a List normalizer.
func ParseRegexpPairs(pairs [][]string) (List, error) {
	res := make([]Normalizer, len(pairs))
	for i, pair := range pairs {
		regex, err := regexp.Compile(pair[0])
		if err != nil {
			return nil, err
		}
		res[i] = Reg(regex, pair[1])
	}
	return List(res), nil
}
