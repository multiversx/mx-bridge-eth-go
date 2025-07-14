package dtos

import "math/big"

// Batch represents a batch of deposits from safe contract
type Batch struct {
	Nonce                  uint64
	BlockNumber            uint64
	LastUpdatedTimestampMs uint64
	DepositsCount          uint16
}

// Deposit represents a single deposit in the batch
type Deposit struct {
	Nonce        uint64
	TokenAddress string
	Amount       *big.Int
	Depositor    [32]byte
	Recipient    [32]byte
	Status       uint8
}

// DepositStatus represents the status of a deposit
type DepositStatus byte

const (
	None DepositStatus = iota
	Pending
	InProgress
	Executed
	Rejected
)
