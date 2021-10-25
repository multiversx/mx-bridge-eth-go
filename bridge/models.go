package bridge

import (
	"fmt"
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
	To            string
	DisplayableTo string
	From          string
	TokenAddress  string
	Amount        *big.Int
	DepositNonce  Nonce
	BlockNonce    Nonce
	Status        uint8
	Error         error
}

// String will convert the deposit transaction to a string
func (dt *DepositTransaction) String() string {
	return fmt.Sprintf("to: %s, from: %s, token address: %s, amount: %v, deposit nonce: %v, block nonce: %v, "+
		"status: %d, error: %v", dt.DisplayableTo, dt.From, dt.TokenAddress, dt.Amount, dt.DepositNonce, dt.BlockNonce, dt.Status, dt.Error)
}

type Batch struct {
	Id           BatchId
	Transactions []*DepositTransaction
}

func (batch *Batch) SetStatusOnAllTransactions(status uint8, err error) {
	for _, tx := range batch.Transactions {
		tx.Status = status
		tx.Error = err
	}
}
