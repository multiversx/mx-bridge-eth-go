package topology

import (
	"encoding/binary"

	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
)

var hasher = sha256.NewSha256()

const uint64Size = 8

type hashRandomSelector struct {
}

func (selector *hashRandomSelector) randomInt(seed uint64, max uint64) uint64 {
	if max == 0 {
		return 0
	}

	buff := make([]byte, uint64Size)
	binary.BigEndian.PutUint64(buff, seed)

	hashedSeed := hasher.Compute(string(buff))
	result := binary.BigEndian.Uint64(hashedSeed) % max

	return result
}
