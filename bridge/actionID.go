package bridge

import (
	"fmt"
	"math/big"
)

// ActionID is the action that will signal which action should the bridge execute
type ActionID int64

// String returns the string representation of the ActionID
func (a ActionID) String() string {
	return fmt.Sprintf("%d", a)
}

// Bytes returns the action ID's bytes
func (a ActionID) Bytes() []byte {
	return big.NewInt(0).SetInt64(int64(a)).Bytes()
}

// Int64 returns the value as int64
func (a ActionID) Int64() int64 {
	return int64(a)
}
