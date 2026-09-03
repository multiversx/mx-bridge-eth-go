package main

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/executors/ethereum"
)

// BatchCreator defines the operations implemented by an entity that can create an Ethereum batch message that can be used
// in signing or transfer execution
type BatchCreator interface {
	CreateBatchInfo(ctx context.Context, newSafeAddress common.Address, partialMigration map[string]*ethereum.FloatWrapper) (*ethereum.BatchInfo, error)
}
