package bridge

import "math/big"

type ActionId uint

// TODO: refactor to big *big.Int
type Nonce uint

type DepositTransaction struct {
	To           string
	From         string
	TokenAddress string
	Amount       *big.Int
	DepositNonce Nonce
}
