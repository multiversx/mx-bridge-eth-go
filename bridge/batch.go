package bridge

import (
	"fmt"
	"math/big"
)

const (
	// Executed is the Executed with success status value
	Executed = uint8(3)
	// Rejected is the Rejected status value
	Rejected = uint8(4)
)

// DepositTransaction represents a deposit transaction ready to be executed on the other chain
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

// Clone will return a new copy of the current deposit transaction
func (dt *DepositTransaction) Clone() *DepositTransaction {
	return &DepositTransaction{
		To:            dt.To,
		DisplayableTo: dt.DisplayableTo,
		From:          dt.From,
		TokenAddress:  dt.TokenAddress,
		Amount:        big.NewInt(0).Set(dt.Amount),
		DepositNonce:  dt.DepositNonce,
		BlockNonce:    dt.BlockNonce,
		Status:        dt.Status,
		Error:         dt.Error,
	}
}

// Batch represents the transactions batch to be executed
type Batch struct {
	ID           BatchID
	Transactions []*DepositTransaction
}

// SetStatusOnAllTransactions will set the provided status on all existing transactions
func (batch *Batch) SetStatusOnAllTransactions(status uint8, err error) {
	for _, tx := range batch.Transactions {
		tx.Status = status
		tx.Error = err
	}
}

// Clone will copy the current batch deeply
func (batch *Batch) Clone() *Batch {
	if batch == nil {
		return nil
	}

	newBatch := &Batch{
		ID: batch.ID,
	}
	for _, tx := range batch.Transactions {
		newBatch.Transactions = append(newBatch.Transactions, tx.Clone())
	}

	return newBatch
}
