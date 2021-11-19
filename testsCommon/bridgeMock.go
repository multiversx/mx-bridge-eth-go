package testsCommon

import (
	"context"
	"reflect"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests"
)

// BridgeMock -
type BridgeMock struct {
	sync.RWMutex
	pendingBatch                  *bridge.Batch
	proposedTransferBatch         *bridge.Batch
	actionID                      bridge.ActionId
	signedActionIDMap             map[string]int
	executedActionID              bridge.ActionId
	executedBatchID               bridge.BatchId
	proposedStatusBatch           *bridge.Batch
	GetPendingCalled              func()
	GetTransactionsStatusesCalled func(ctx context.Context, batchId bridge.BatchId) ([]uint8, error)
	ProposeTransferCalled         func(_ context.Context, batch *bridge.Batch) (string, error)
}

// GetPending -
func (bm *BridgeMock) GetPending(_ context.Context) (*bridge.Batch, error) {
	if bm.GetPendingCalled != nil {
		bm.GetPendingCalled()
	}

	bm.RLock()
	defer bm.RUnlock()

	return bm.pendingBatch, nil
}

// SetPending -
func (bm *BridgeMock) SetPending(pendingBatch *bridge.Batch) {
	bm.Lock()
	defer bm.Unlock()

	bm.pendingBatch = integrationTests.CloneBatch(pendingBatch)
}

// ProposeSetStatus -
func (bm *BridgeMock) ProposeSetStatus(_ context.Context, batch *bridge.Batch) {
	bm.Lock()
	defer bm.Unlock()

	bm.proposedStatusBatch = integrationTests.CloneBatch(batch)
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

	bm.proposedTransferBatch = integrationTests.CloneBatch(batch)

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
func (bm *BridgeMock) GetActionIdForProposeTransfer(_ context.Context, _ *bridge.Batch) bridge.ActionId {
	bm.RLock()
	defer bm.RUnlock()

	return bm.actionID
}

// SetActionID -
func (bm *BridgeMock) SetActionID(actionID bridge.ActionId) {
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
func (bm *BridgeMock) GetActionIdForSetStatusOnPendingTransfer(_ context.Context, _ *bridge.Batch) bridge.ActionId {
	bm.RLock()
	defer bm.RUnlock()

	return bm.actionID
}

// WasExecuted -
func (bm *BridgeMock) WasExecuted(_ context.Context, id bridge.ActionId, id2 bridge.BatchId) bool {
	bm.RLock()
	defer bm.RUnlock()

	return reflect.DeepEqual(id, bm.executedActionID) && reflect.DeepEqual(id2, bm.executedBatchID)
}

// Sign -
func (bm *BridgeMock) Sign(_ context.Context, id bridge.ActionId, _ *bridge.Batch) (string, error) {
	bm.Lock()
	defer bm.Unlock()

	if bm.signedActionIDMap == nil {
		bm.signedActionIDMap = make(map[string]int)
	}
	idString := integrationTests.ActionIdToString(id)
	bm.signedActionIDMap[idString]++

	return "", nil
}

// SignedActionIDMap -
func (bm *BridgeMock) SignedActionIDMap() map[string]int {
	bm.RLock()
	defer bm.RUnlock()

	return bm.signedActionIDMap
}

// Execute -
func (bm *BridgeMock) Execute(_ context.Context, id bridge.ActionId, batch *bridge.Batch, _ bridge.SignaturesHolder) (string, error) {
	bm.Lock()
	defer bm.Unlock()

	bm.executedActionID = id
	bm.executedBatchID = batch.Id

	return "", nil
}

// GetExecuted -
func (bm *BridgeMock) GetExecuted() (bridge.ActionId, bridge.BatchId) {
	bm.RLock()
	defer bm.RUnlock()

	return bm.executedActionID, bm.executedBatchID
}

// SignersCount -
func (bm *BridgeMock) SignersCount(_ context.Context, _ *bridge.Batch, id bridge.ActionId, _ bridge.SignaturesHolder) uint {
	bm.RLock()
	defer bm.RUnlock()

	if bm.signedActionIDMap == nil {
		return 0
	}

	idString := integrationTests.ActionIdToString(id)
	return uint(bm.signedActionIDMap[idString])
}

// GetTransactionsStatuses -
func (bm *BridgeMock) GetTransactionsStatuses(ctx context.Context, batchId bridge.BatchId) ([]uint8, error) {
	if bm.GetTransactionsStatusesCalled != nil {
		return bm.GetTransactionsStatusesCalled(ctx, batchId)
	}

	return make([]byte, 0), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bm *BridgeMock) IsInterfaceNil() bool {
	return bm == nil
}
