package ethereum

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// DepositInfo is the deposit info list
type DepositInfo struct {
	DepositNonce    uint64 `json:"DepositNonce"`
	Token           string `json:"Token"`
	ContractAddress string `json:"ContractAddress"`
	contractAddress common.Address
	amount          *big.Int
	Amount          string `json:"Amount"`
}

// BatchInfo is the batch info list
type BatchInfo struct {
	OldSafeContractAddress string         `json:"OldSafeContractAddress"`
	NewSafeContractAddress string         `json:"NewSafeContractAddress"`
	BatchID                uint64         `json:"BatchID"`
	MessageHash            common.Hash    `json:"MessageHash"`
	DepositsInfo           []*DepositInfo `json:"DepositsInfo"`
}

// SignatureInfo is the struct holding signature info
type SignatureInfo struct {
	PublicKey   string `json:"PublicKey"`
	MessageHash string `json:"MessageHash"`
	Signature   string `json:"Signature"`
}
