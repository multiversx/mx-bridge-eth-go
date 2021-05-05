package bridge

import "math/big"

const (
	Executed = uint8(3)
	Rejected = uint8(4)
)

type TokenMap map[string]string
type ActionId *big.Int
type Nonce *big.Int

func NewNonce(value int64) Nonce {
	return big.NewInt(value)
}

func NewActionId(value int64) ActionId {
	return big.NewInt(value)
}

type DepositTransaction struct {
	To           string
	From         string
	TokenAddress string
	Amount       *big.Int
	DepositNonce Nonce
}
