package player

import "testing"

var raceNormTests = [][]string{
	{"Ogre", "Ogre"},
	{"Red Draconian", "Draconian"},
}

func TestNormalizeRace(t *testing.T) {
	for _, raceNorm := range raceNormTests {
		race, normalized := raceNorm[0], raceNorm[1]
		res := NormalizeRace(race)
		if res != normalized {
			t.Errorf("NormalizeRace(%#v) == %#v, expected %#v",
				race, res, normalized)
		}
	}
}
