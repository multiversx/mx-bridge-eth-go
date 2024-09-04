package ethereum

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// DepositInfo is the deposit info list
type DepositInfo struct {
	DepositNonce          uint64         `json:"DepositNonce"`
	Token                 string         `json:"Token"`
	ContractAddressString string         `json:"ContractAddress"`
	ContractAddress       common.Address `json:"-"`
	Amount                *big.Int       `json:"-"`
	AmountString          string         `json:"Amount"`
}

// BatchInfo is the batch info list
type BatchInfo struct {
	OldSafeContractAddress string         `json:"OldSafeContractAddress"`
	NewSafeContractAddress string         `json:"NewSafeContractAddress"`
	BatchID                uint64         `json:"BatchID"`
	DepositsInfo           []*DepositInfo `json:"DepositsInfo"`
}
