package dtos

// Batch represents a batch of deposits from safe contract
type Batch struct {
	Nonce                  uint64
	TimestampMs            uint64
	LastUpdatedTimestampMs uint64
	DepositsCount          uint16
}

// Deposit represents a single deposit in the batch
type Deposit struct {
	Nonce          uint64
	TokenTypeBytes []byte
	Amount         uint64
	Sender         [32]byte
	Recipient      []byte
}
