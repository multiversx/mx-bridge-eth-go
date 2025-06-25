package dtos

import "math/big"

type Batch struct {
	Nonce                  uint64
	BlockNumber            uint64
	LastUpdatedBlockNumber uint64
	DepositsCount          uint16
}

type Deposit struct {
	Nonce        *big.Int
	TokenAddress string
	Amount       *big.Int
	Depositor    [32]byte
	Recipient    [32]byte
	Status       uint8
}
