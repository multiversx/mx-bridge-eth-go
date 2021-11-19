package bridge

import (
	"fmt"
	"math/big"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("bridge/models")

const (
	// Executed is the Executed with success status value
	Executed = uint8(3)
	// Rejected is the Rejected status value
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

// DepositTransaction represents a deposit transaction ready to be executed on the other chain
type DepositTransaction struct {
	To            string
	DisplayableTo string
	From          string
	TokenAddress  string
	Amount        *big.Int
	DepositNonce  Nonce
	BlockNonce    Nonce
	Error         error
}

// String will convert the deposit transaction to a string
func (dt *DepositTransaction) String() string {
	return fmt.Sprintf("to: %s, from: %s, token address: %s, amount: %v, deposit nonce: %v, block nonce: %v, "+
		"error: %v", dt.DisplayableTo, dt.From, dt.TokenAddress, dt.Amount, dt.DepositNonce, dt.BlockNonce, dt.Error)
}

// Batch represents the transactions batch to be executed
type Batch struct {
	Id           BatchId
	Transactions []*DepositTransaction
	Statuses     []byte
}

// SetStatusOnAllTransactions will set the provided status on all existing transactions
func (batch *Batch) SetStatusOnAllTransactions(status byte, err error) {
	for _, tx := range batch.Transactions {
		tx.Error = err
	}

	for i := 0; i < len(batch.Statuses); i++ {
		batch.Statuses[i] = status
	}
}

// ResolveNewDeposits will add new statuses as rejected if the newNumDeposits exceeds the number of the deposits
func (batch *Batch) ResolveNewDeposits(newNumDeposits int) {
	oldLen := len(batch.Statuses)
	if newNumDeposits == oldLen {
		log.Debug("num statuses ok", "len statuses", oldLen)
		return
	}

	if newNumDeposits < oldLen {
		log.Error("num statuses unrecoverable", "len statuses", oldLen, "new num deposits", newNumDeposits)
		return
	}

	for newNumDeposits > len(batch.Statuses) {
		batch.Statuses = append(batch.Statuses, Rejected)
	}

	log.Warn("recovered num statuses", "len statuses", oldLen, "new num deposits", newNumDeposits)
}
