package testsCommon

import (
	"context"
	"reflect"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
)

// BridgeMock -
type BridgeMock struct {
	sync.RWMutex
	pendingBatch                  *bridge.Batch
	proposedTransferBatch         *bridge.Batch
	actionID                      bridge.ActionID
	signedActionIDMap             map[string]int
	executedActionID              bridge.ActionID
	executedBatchID               bridge.BatchID
	proposedStatusBatch           *bridge.Batch
	GetPendingCalled              func()
	GetTransactionsStatusesCalled func(ctx context.Context, batchID bridge.BatchID) ([]uint8, error)
	ProposeTransferCalled         func(_ context.Context, batch *bridge.Batch) (string, error)
}

// GetPending -
func (bm *BridgeMock) GetPending(_ context.Context) *bridge.Batch {
	if bm.GetPendingCalled != nil {
		bm.GetPendingCalled()
	}

	bm.RLock()
	defer bm.RUnlock()

	return bm.pendingBatch
}

// SetPending -
func (bm *BridgeMock) SetPending(pendingBatch *bridge.Batch) {
	bm.Lock()
	defer bm.Unlock()

	bm.pendingBatch = pendingBatch.Clone()
}

// ProposeSetStatus -
func (bm *BridgeMock) ProposeSetStatus(_ context.Context, batch *bridge.Batch) {
	bm.Lock()
	defer bm.Unlock()

	bm.proposedStatusBatch = batch.Clone()
}

// GetProposedSetStatusBatch -
func (bm *BridgeMock) GetProposedSetStatusBatch() *bridge.Batch {
	bm.RLock()
	defer bm.RUnlock()

	return bm.proposedStatusBatch
}

// ProposeTransfer -
func (bm *BridgeMock) ProposeTransfer(ctx context.Context, batch *bridge.Batch) (string, error) {
	if bm.ProposeTransferCalled != nil {
		return bm.ProposeTransferCalled(ctx, batch)
	}

	bm.Lock()
	defer bm.Unlock()

	bm.proposedTransferBatch = batch.Clone()

	return "", nil
}

// GetProposedTransferBatch -
func (bm *BridgeMock) GetProposedTransferBatch() *bridge.Batch {
	bm.RLock()
	defer bm.RUnlock()

	return bm.proposedTransferBatch
}

// WasProposedTransfer -
func (bm *BridgeMock) WasProposedTransfer(_ context.Context, batch *bridge.Batch) bool {
	bm.RLock()
	defer bm.RUnlock()

	return reflect.DeepEqual(batch, bm.proposedTransferBatch)
}

// GetActionIdForProposeTransfer -
func (bm *BridgeMock) GetActionIdForProposeTransfer(_ context.Context, _ *bridge.Batch) bridge.ActionID {
	bm.RLock()
	defer bm.RUnlock()

	return bm.actionID
}

// SetActionID -
func (bm *BridgeMock) SetActionID(actionID bridge.ActionID) {
	bm.Lock()
	defer bm.Unlock()

	bm.actionID = actionID
}

// WasProposedSetStatus -
func (bm *BridgeMock) WasProposedSetStatus(_ context.Context, batch *bridge.Batch) bool {
	bm.RLock()
	defer bm.RUnlock()

	return reflect.DeepEqual(batch, bm.proposedStatusBatch)
}

// GetActionIdForSetStatusOnPendingTransfer -
func (bm *BridgeMock) GetActionIdForSetStatusOnPendingTransfer(_ context.Context, _ *bridge.Batch) bridge.ActionID {
	bm.RLock()
	defer bm.RUnlock()

	return bm.actionID
}

// WasExecuted -
func (bm *BridgeMock) WasExecuted(_ context.Context, id bridge.ActionID, id2 bridge.BatchID) bool {
	bm.RLock()
	defer bm.RUnlock()

	return reflect.DeepEqual(id, bm.executedActionID) && reflect.DeepEqual(id2, bm.executedBatchID)
}

// Sign -
func (bm *BridgeMock) Sign(_ context.Context, id bridge.ActionID, _ *bridge.Batch) (string, error) {
	bm.Lock()
	defer bm.Unlock()

	if bm.signedActionIDMap == nil {
		bm.signedActionIDMap = make(map[string]int)
	}
	bm.signedActionIDMap[id.String()]++

	return "", nil
}

// SignedActionIDMap -
func (bm *BridgeMock) SignedActionIDMap() map[string]int {
	bm.RLock()
	defer bm.RUnlock()

	return bm.signedActionIDMap
}

// Execute -
func (bm *BridgeMock) Execute(_ context.Context, id bridge.ActionID, batch *bridge.Batch, _ bridge.SignaturesHolder) (string, error) {
	bm.Lock()
	defer bm.Unlock()

	bm.executedActionID = id
	bm.executedBatchID = batch.ID

	return "", nil
}

// GetExecuted -
func (bm *BridgeMock) GetExecuted() (bridge.ActionID, bridge.BatchID) {
	bm.RLock()
	defer bm.RUnlock()

	return bm.executedActionID, bm.executedBatchID
}

// SignersCount -
func (bm *BridgeMock) SignersCount(_ *bridge.Batch, id bridge.ActionID, _ bridge.SignaturesHolder) uint {
	bm.RLock()
	defer bm.RUnlock()

	if bm.signedActionIDMap == nil {
		return 0
	}

	return uint(bm.signedActionIDMap[id.String()])
}

// GetTransactionsStatuses -
func (bm *BridgeMock) GetTransactionsStatuses(ctx context.Context, batchID bridge.BatchID) ([]uint8, error) {
	if bm.GetTransactionsStatusesCalled != nil {
		return bm.GetTransactionsStatusesCalled(ctx, batchID)
	}

	return make([]byte, 0), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bm *BridgeMock) IsInterfaceNil() bool {
	return bm == nil
}
