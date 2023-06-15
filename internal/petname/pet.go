package petname

import (
	"github.com/lucasepe/codename"
)

func PetName(tokenLength int) string {
	rng, err := codename.DefaultRNG()
	if err != nil {
		panic(err)
	}
	return codename.Generate(rng, tokenLength)
}
