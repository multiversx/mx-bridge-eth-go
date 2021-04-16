package bridge

import "math/big"

type DepositTransaction struct {
	Hash         string
	From         string
	TokenAddress string
	Amount       *big.Int
}
