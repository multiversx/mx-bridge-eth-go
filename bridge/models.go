package bridge

import "math/big"

const (
	Executed = 3
	Rejected = 4
)

type ActionId uint
type Nonce uint

type DepositTransaction struct {
	To           string
	From         string
	TokenAddress string
	Amount       *big.Int
	DepositNonce Nonce
}
