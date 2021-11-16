package bridge

import (
	"fmt"
	"math/big"
)

// BatchID is the batch's ID
type BatchID int64

// String returns the string representation of the Nonce
func (b BatchID) String() string {
	return fmt.Sprintf("%d", b)
}

// Bytes returns the nonce's bytes
func (b BatchID) Bytes() []byte {
	return big.NewInt(0).SetInt64(int64(b)).Bytes()
}

// Int64 returns the value as int64
func (b BatchID) Int64() int64 {
	return int64(b)
}
