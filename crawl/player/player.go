package player

import (
	"github.com/greensnark/go-sequell/stringnorm"
	"github.com/greensnark/go-sequell/text"
)

var raceNorm = stringnorm.List{
	stringnorm.SR(`.*(Draconian)$`, "$1"),
}

func NormalizeRace(race string) string {
	return text.NormalizeSpace(stringnorm.NormalizeNoErr(raceNorm, race))
}
