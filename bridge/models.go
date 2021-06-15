package bridge

import (
	"math/big"
)

const (
	Executed = uint8(3)
	Rejected = uint8(4)
)

type ActionId *big.Int
type Nonce *big.Int
type BatchId *big.Int

func NewNonce(value int64) Nonce {
	return big.NewInt(value)
}

func NewBatchId(value int64) BatchId {
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
	BlockNonce   Nonce
	Status       uint8
	Error        error
}

type Batch struct {
	Id           BatchId
	Transactions []*DepositTransaction
}
