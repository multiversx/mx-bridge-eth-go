package relay

import (
	"context"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
)

// TODO remove this after the relay.go full refactoring

type bridgeMock struct {
	sync.RWMutex
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
}

// GetPending -
func (b *bridgeMock) GetPending(_ context.Context) *bridge.Batch {
	b.Lock()
	defer b.Unlock()

	defer func() {
		b.pendingBatchCallIndex++
	}()

	if b.pendingBatchCallIndex >= len(b.pendingBatches) {
		return nil
	} else {
		return b.pendingBatches[b.pendingBatchCallIndex]
	}
}

// ProposeSetStatus -
func (b *bridgeMock) ProposeSetStatus(_ context.Context, batch *bridge.Batch) {
	b.Lock()
	defer b.Unlock()

	b.proposedStatusBatch = batch
}

// ProposeTransfer -
func (b *bridgeMock) ProposeTransfer(_ context.Context, batch *bridge.Batch) (string, error) {
	b.Lock()
	defer b.Unlock()

	b.wasProposedTransfer = true
	b.lastProposedBatch = batch

	return "propose_tx_hash", b.proposeTransferError
}

// WasProposedTransfer -
func (b *bridgeMock) WasProposedTransfer(_ context.Context, batch *bridge.Batch) bool {
	b.Lock()
	defer b.Unlock()

	b.lastWasProposedTransferBatchId = batch.Id

	return b.wasProposedTransfer
}

// GetActionIdForProposeTransfer -
func (b *bridgeMock) GetActionIdForProposeTransfer(_ context.Context, _ *bridge.Batch) bridge.ActionId {
	b.RLock()
	defer b.RUnlock()

	return b.proposeTransferActionId
}

// WasProposedSetStatus -
func (b *bridgeMock) WasProposedSetStatus(_ context.Context, _ *bridge.Batch) bool {
	return true
}

// GetActionIdForSetStatusOnPendingTransfer -
func (b *bridgeMock) GetActionIdForSetStatusOnPendingTransfer(_ context.Context, _ *bridge.Batch) bridge.ActionId {
	b.RLock()
	defer b.RUnlock()

	return b.proposeSetStatusActionId
}

// WasExecuted -
func (b *bridgeMock) WasExecuted(_ context.Context, _ bridge.ActionId, _ bridge.BatchId) bool {
	return b.wasExecuted
}

// Sign -
func (b *bridgeMock) Sign(_ context.Context, actionId bridge.ActionId) (string, error) {
	b.Lock()
	defer b.Unlock()

	b.lastSignedActionId = actionId

	return "sign_tx_hash", nil
}

// Execute -
func (b *bridgeMock) Execute(_ context.Context, actionId bridge.ActionId, _ *bridge.Batch) (string, error) {
	b.Lock()
	defer b.Unlock()

	b.lastExecutedActionId = actionId

	return "execution hash", nil
}

// SignersCount -
func (b *bridgeMock) SignersCount(_ context.Context, _ bridge.ActionId) uint {
	return b.signersCount
}
