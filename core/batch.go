package core

import (
	"encoding/hex"
	"fmt"
	"math/big"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("clients")

// TransferBatch is the transfer batch structure agnostic of any chain implementation
type TransferBatch struct {
	ID          uint64             `json:"batchId"`
	BlockNumber uint64             `json:"blockNumber"`
	Deposits    []*DepositTransfer `json:"deposits"`
	Statuses    []byte             `json:"statuses"`
}

// Clone will deep clone the current TransferBatch instance
func (tb *TransferBatch) Clone() *TransferBatch {
	cloned := &TransferBatch{
		ID:       tb.ID,
		Deposits: make([]*DepositTransfer, 0, len(tb.Deposits)),
		Statuses: make([]byte, len(tb.Statuses)),
	}

	for _, dt := range tb.Deposits {
		cloned.Deposits = append(cloned.Deposits, dt.Clone())
	}
	copy(cloned.Statuses, tb.Statuses)

	return cloned
}

// String will convert the transfer batch to a string
func (tb *TransferBatch) String() string {
	str := fmt.Sprintf("Batch id %d:", tb.ID)
	for _, dt := range tb.Deposits {
		str += "\n  " + dt.String()
	}
	str += "\nStatuses: " + hex.EncodeToString(tb.Statuses)

	return str
}

// ResolveNewDeposits will add new statuses as rejected if the newNumDeposits exceeds the number of the deposits
func (tb *TransferBatch) ResolveNewDeposits(newNumDeposits int) {
	oldLen := len(tb.Statuses)
	if newNumDeposits == oldLen {
		log.Debug("num statuses ok", "len statuses", oldLen)
		return
	}

	for i := newNumDeposits; i < oldLen; i++ {
		tb.Statuses[i] = Rejected
	}

	for newNumDeposits > len(tb.Statuses) {
		tb.Statuses = append(tb.Statuses, Rejected)
	}

	log.Warn("recovered num statuses", "len statuses", oldLen, "new num deposits", newNumDeposits)
}

// DepositTransfer is the deposit transfer structure agnostic of any chain implementation
type DepositTransfer struct {
	Nonce                 uint64   `json:"nonce"`
	ToBytes               []byte   `json:"-"`
	DisplayableTo         string   `json:"to"`
	FromBytes             []byte   `json:"-"`
	DisplayableFrom       string   `json:"from"`
	SourceTokenBytes      []byte   `json:"-"`
	DestinationTokenBytes []byte   `json:"-"`
	DisplayableToken      string   `json:"token"`
	Amount                *big.Int `json:"amount"`
	Data                  []byte   `json:"-"`
	DisplayableData       string   `json:"data"`
}

// String will convert the deposit transfer to a string
func (dt *DepositTransfer) String() string {
	return fmt.Sprintf("to: %s, from: %s, token address: %s, amount: %v, deposit nonce: %d, data: %s",
		dt.DisplayableTo, dt.DisplayableFrom, dt.DisplayableToken, dt.Amount, dt.Nonce, dt.DisplayableData)
}

// Clone will deeply clone the current DepositTransfer instance
func (dt *DepositTransfer) Clone() *DepositTransfer {
	cloned := &DepositTransfer{
		Nonce:                 dt.Nonce,
		ToBytes:               make([]byte, len(dt.ToBytes)),
		DisplayableTo:         dt.DisplayableTo,
		FromBytes:             make([]byte, len(dt.FromBytes)),
		DisplayableFrom:       dt.DisplayableFrom,
		SourceTokenBytes:      make([]byte, len(dt.SourceTokenBytes)),
		DestinationTokenBytes: make([]byte, len(dt.DestinationTokenBytes)),
		DisplayableToken:      dt.DisplayableToken,
		Amount:                big.NewInt(0),
		Data:                  make([]byte, len(dt.Data)),
		DisplayableData:       dt.DisplayableData,
	}

	copy(cloned.ToBytes, dt.ToBytes)
	copy(cloned.FromBytes, dt.FromBytes)
	copy(cloned.SourceTokenBytes, dt.SourceTokenBytes)
	copy(cloned.DestinationTokenBytes, dt.DestinationTokenBytes)
	copy(cloned.Data, dt.Data)
	if dt.Amount != nil {
		cloned.Amount.Set(dt.Amount)
	}

	return cloned
}
