package relay

import (
	"context"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
)

// TODO remove this after the relay.go full refactoring

type bridgeStub struct {
	pendingBatchCallIndex          int
	pendingBatches                 []*bridge.Batch
	wasProposedTransfer            bool
	lastProposedBatch              *bridge.Batch
	lastWasProposedTransferBatchId bridge.BatchId
	lastSignedActionId             bridge.ActionId
	signersCount                   uint
	lastExecutedActionId           bridge.ActionId
	wasExecuted                    bool
	proposeTransferActionId        bridge.ActionId
	proposeTransferError           error
	proposedStatusBatch            *bridge.Batch
	proposeSetStatusActionId       bridge.ActionId

	proposeSetStatusMutex sync.Mutex
	proposeTransferMutex  sync.Mutex
	signMutex             sync.Mutex
	executeMutex          sync.Mutex
}

func (b *bridgeStub) lock() {
	b.proposeSetStatusMutex.Lock()
	b.proposeTransferMutex.Lock()
	b.signMutex.Lock()
	b.executeMutex.Lock()
}

// GetPending -
func (b *bridgeStub) GetPending(_ context.Context) *bridge.Batch {
	defer func() { b.pendingBatchCallIndex++ }()

	if b.pendingBatchCallIndex >= len(b.pendingBatches) {
		return nil
	} else {
		return b.pendingBatches[b.pendingBatchCallIndex]
	}
}

// ProposeSetStatus -
func (b *bridgeStub) ProposeSetStatus(_ context.Context, batch *bridge.Batch) {
	b.proposeSetStatusMutex.Lock()
	b.proposedStatusBatch = batch
}

// ProposeTransfer -
func (b *bridgeStub) ProposeTransfer(_ context.Context, batch *bridge.Batch) (string, error) {
	b.proposeTransferMutex.Lock()
	b.wasProposedTransfer = true
	b.lastProposedBatch = batch

	return "propose_tx_hash", b.proposeTransferError
}

// WasProposedTransfer -
func (b *bridgeStub) WasProposedTransfer(_ context.Context, batch *bridge.Batch) bool {
	b.lastWasProposedTransferBatchId = batch.Id
	return b.wasProposedTransfer
}

// GetActionIdForProposeTransfer -
func (b *bridgeStub) GetActionIdForProposeTransfer(_ context.Context, _ *bridge.Batch) bridge.ActionId {
	return b.proposeTransferActionId
}

// WasProposedSetStatus -
func (b *bridgeStub) WasProposedSetStatus(_ context.Context, _ *bridge.Batch) bool {
	return true
}

// GetActionIdForSetStatusOnPendingTransfer -
func (b *bridgeStub) GetActionIdForSetStatusOnPendingTransfer(_ context.Context, _ *bridge.Batch) bridge.ActionId {
	return b.proposeSetStatusActionId
}

// WasExecuted -
func (b *bridgeStub) WasExecuted(_ context.Context, _ bridge.ActionId, _ bridge.BatchId) bool {
	return b.wasExecuted
}

// Sign -
func (b *bridgeStub) Sign(_ context.Context, actionId bridge.ActionId) (string, error) {
	b.signMutex.Lock()
	b.lastSignedActionId = actionId

	return "sign_tx_hash", nil
}

// Execute -
func (b *bridgeStub) Execute(_ context.Context, actionId bridge.ActionId, _ *bridge.Batch) (string, error) {
	b.executeMutex.Lock()
	b.lastExecutedActionId = actionId

	return "execution hash", nil
}

// SignersCount -
func (b *bridgeStub) SignersCount(_ context.Context, _ bridge.ActionId) uint {
	return b.signersCount
}
