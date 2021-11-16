package bridge

import (
	"fmt"
	"math/big"
)

// Nonce is the batch's current nonce
type Nonce int64

// String returns the string representation of the Nonce
func (n Nonce) String() string {
	return fmt.Sprintf("%d", n)
}

// Bytes returns the nonce's bytes
func (n Nonce) Bytes() []byte {
	return big.NewInt(0).SetInt64(int64(n)).Bytes()
}

// Int64 returns the value as int64
func (n Nonce) Int64() int64 {
	return int64(n)
}
