package testsCommon

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
)

var fullPathBridgeStub = "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon.(*BridgeStub)."

// BridgeStub -
type BridgeStub struct {
	functionCalledCounter map[string]int
	mutBridge             sync.RWMutex

	WasProposedTransferCalled  func(ctx context.Context, batch *bridge.Batch) bool
	WasProposedSetStatusCalled func(ctx context.Context, batch *bridge.Batch) bool
	WasExecutedCalled          func(ctx context.Context, id bridge.ActionId, id2 bridge.BatchId) bool

	GetPendingCalled                               func(ctx context.Context) *bridge.Batch
	ProposeSetStatusCalled                         func(ctx context.Context, batch *bridge.Batch)
	ProposeTransferCalled                          func(ctx context.Context, batch *bridge.Batch) (string, error)
	GetActionIdForProposeTransferCalled            func(ctx context.Context, batch *bridge.Batch) bridge.ActionId
	GetActionIdForSetStatusOnPendingTransferCalled func(ctx context.Context, batch *bridge.Batch) bridge.ActionId
	SignCalled                                     func(ctx context.Context, id bridge.ActionId) (string, error)
	ExecuteCalled                                  func(ctx context.Context, id bridge.ActionId, batch *bridge.Batch, sigHolder bridge.SignaturesHolder) (string, error)
	SignersCountCalled                             func(batch *bridge.Batch, id bridge.ActionId, sigHolder bridge.SignaturesHolder) uint
	GetTransactionsStatusesCalled                  func(ctx context.Context, batchID bridge.BatchId) ([]uint8, error)

	ProposeTransferError error
	SignError            error
	ExecuteError         error
}

// NewBridgeStub creates a new BridgeStub instance
func NewBridgeStub() *BridgeStub {
	return &BridgeStub{
		functionCalledCounter: make(map[string]int),
	}
}

// -------- decision functions

// WasProposedTransfer -
func (b *BridgeStub) WasProposedTransfer(ctx context.Context, batch *bridge.Batch) bool {
	b.incrementFunctionCounter()
	if b.WasProposedTransferCalled != nil {
		return b.WasProposedTransferCalled(ctx, batch)
	}
	return false
}

// WasProposedSetStatus -
func (b *BridgeStub) WasProposedSetStatus(ctx context.Context, batch *bridge.Batch) bool {
	b.incrementFunctionCounter()
	if b.WasProposedSetStatusCalled != nil {
		return b.WasProposedSetStatusCalled(ctx, batch)
	}
	return false
}

// WasExecuted -
func (b *BridgeStub) WasExecuted(ctx context.Context, id bridge.ActionId, id2 bridge.BatchId) bool {
	b.incrementFunctionCounter()
	if b.WasExecutedCalled != nil {
		return b.WasExecutedCalled(ctx, id, id2)
	}
	return false
}

// -------- action functions

// GetPending -
func (b *BridgeStub) GetPending(ctx context.Context) *bridge.Batch {
	b.incrementFunctionCounter()
	if b.GetPendingCalled != nil {
		return b.GetPendingCalled(ctx)
	}
	return nil
}

// ProposeSetStatus -
func (b *BridgeStub) ProposeSetStatus(ctx context.Context, batch *bridge.Batch) {
	b.incrementFunctionCounter()
	if b.ProposeSetStatusCalled != nil {
		b.ProposeSetStatusCalled(ctx, batch)
	}
}

// ProposeTransfer -
func (b *BridgeStub) ProposeTransfer(ctx context.Context, batch *bridge.Batch) (string, error) {
	b.incrementFunctionCounter()
	if b.ProposeTransferCalled != nil {
		called, err := b.ProposeTransferCalled(ctx, batch)
		if err != nil {
			return "", err
		}
		return called, nil
	}
	return "propose_tx_hash", b.ProposeTransferError
}

// GetActionIdForProposeTransfer -
func (b *BridgeStub) GetActionIdForProposeTransfer(ctx context.Context, batch *bridge.Batch) bridge.ActionId {
	b.incrementFunctionCounter()
	return b.GetActionIdForProposeTransferCalled(ctx, batch)
}

// GetActionIdForSetStatusOnPendingTransfer -
func (b *BridgeStub) GetActionIdForSetStatusOnPendingTransfer(ctx context.Context, batch *bridge.Batch) bridge.ActionId {
	b.incrementFunctionCounter()
	return b.GetActionIdForSetStatusOnPendingTransferCalled(ctx, batch)
}

// Sign -
func (b *BridgeStub) Sign(ctx context.Context, id bridge.ActionId, _ *bridge.Batch) (string, error) {
	b.incrementFunctionCounter()
	if b.SignCalled != nil {
		return b.SignCalled(ctx, id)
	}
	return "sign_tx_hash", b.SignError
}

// Execute -
func (b *BridgeStub) Execute(ctx context.Context, id bridge.ActionId, batch *bridge.Batch, sigHolder bridge.SignaturesHolder) (string, error) {
	b.incrementFunctionCounter()
	if b.ExecuteCalled != nil {
		return b.ExecuteCalled(ctx, id, batch, sigHolder)
	}
	return "execute_tx_hash", b.ExecuteError
}

// SignersCount -
func (b *BridgeStub) SignersCount(batch *bridge.Batch, id bridge.ActionId, sigHolder bridge.SignaturesHolder) uint {
	b.incrementFunctionCounter()
	if b.SignersCountCalled != nil {
		return b.SignersCountCalled(batch, id, sigHolder)
	}
	return 0
}

// GetTransactionsStatuses -
func (b *BridgeStub) GetTransactionsStatuses(ctx context.Context, batchID bridge.BatchId) ([]uint8, error) {
	b.incrementFunctionCounter()
	if b.GetTransactionsStatusesCalled != nil {
		return b.GetTransactionsStatusesCalled(ctx, batchID)
	}

	return make([]byte, 0), nil
}

// -------- helper functions

// incrementFunctionCounter increments the counter for the function that called it
func (b *BridgeStub) incrementFunctionCounter() {
	b.mutBridge.Lock()
	defer b.mutBridge.Unlock()

	pc, _, _, _ := runtime.Caller(1)
	fmt.Printf("BridgeExecutorMock: called %s\n", runtime.FuncForPC(pc).Name())
	b.functionCalledCounter[runtime.FuncForPC(pc).Name()]++
}

// GetFunctionCounter returns the called counter of a given function
func (b *BridgeStub) GetFunctionCounter(function string) int {
	b.mutBridge.Lock()
	defer b.mutBridge.Unlock()

	return b.functionCalledCounter[fullPathBridgeStub+function]
}

// IsInterfaceNil returns true if there is no value under the interface
func (b *BridgeStub) IsInterfaceNil() bool {
	return b == nil
}
