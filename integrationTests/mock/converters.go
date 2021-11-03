package mock

import (
	"encoding/hex"
	"math/big"
	"strconv"
)

// BoolToByteSlice -
func BoolToByteSlice(val bool) []byte {
	if val {
		return []byte{1}
	}

	return []byte{0}
}

// Uint64ByteSlice -
func Uint64ByteSlice(val uint64) []byte {
	b := big.NewInt(int64(val))
	return b.Bytes()
}

// Uint64BytesFromHash -
func Uint64BytesFromHash(hash string) []byte {
	return []byte(hash[:8])
}

// HashToActionID -
func HashToActionID(hash string) *big.Int {
	bytes := Uint64BytesFromHash(hash)
	result, err := strconv.ParseUint(hex.EncodeToString(bytes), 16, 64)
	if err != nil {
		panic(err)
	}

	val := big.NewInt(0).SetInt64(int64(result))
	if val.Cmp(big.NewInt(0)) < 0 {
		val = val.Mul(val, big.NewInt(-1))
	}

	return val
}
