package integrationTests

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
)

// ActionIdToString -
func ActionIdToString(id bridge.ActionId) string {
	if id == nil {
		return ""
	}

	idAsBigInt := *id

	return idAsBigInt.String()
}

// CloneBatch -
func CloneBatch(batch *bridge.Batch) *bridge.Batch {
	if batch == nil {
		return nil
	}

	newBatch := &bridge.Batch{
		Id: CloneBatchID(batch.Id),
	}

	for _, tx := range batch.Transactions {
		newBatch.Transactions = append(newBatch.Transactions, CloneDepositTransaction(tx))
	}

	return newBatch
}

// CloneBatchID -
func CloneBatchID(batchID bridge.BatchId) bridge.BatchId {
	if batchID == nil {
		return nil
	}

	batchIdAsBigInt := *batchID

	return bridge.NewBatchId(batchIdAsBigInt.Int64())
}

// CloneNonce -
func CloneNonce(nonce bridge.Nonce) bridge.Nonce {
	if nonce == nil {
		return nil
	}

	nonceAsBigInt := *nonce

	return bridge.NewNonce(nonceAsBigInt.Int64())
}

// CloneDepositTransaction -
func CloneDepositTransaction(tx *bridge.DepositTransaction) *bridge.DepositTransaction {
	return &bridge.DepositTransaction{
		To:           tx.To,
		From:         tx.From,
		TokenAddress: tx.TokenAddress,
		Amount:       big.NewInt(0).Set(tx.Amount),
		DepositNonce: CloneNonce(tx.DepositNonce),
		BlockNonce:   CloneNonce(tx.BlockNonce),
		Status:       tx.Status,
		Error:        tx.Error,
	}
}
