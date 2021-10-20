package mock

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
)

var fullPath = "github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond/bridgeExecutors/mock.(*bridgeStub)."

// BridgeStub -
type BridgeStub struct {
	functionCalledCounter map[string]int
	mutExecutor           sync.RWMutex

	GetPendingCalled                               func(ctx context.Context) *bridge.Batch
	ProposeSetStatusCalled                         func(ctx context.Context, batch *bridge.Batch)
	ProposeTransferCalled                          func(ctx context.Context, batch *bridge.Batch) (string, error)
	WasProposedTransferCalled                      func(ctx context.Context, batch *bridge.Batch) bool
	GetActionIdForProposeTransferCalled            func(ctx context.Context, batch *bridge.Batch) bridge.ActionId
	WasProposedSetStatusCalled                     func(ctx context.Context, batch *bridge.Batch) bool
	GetActionIdForSetStatusOnPendingTransferCalled func(ctx context.Context, batch *bridge.Batch) bridge.ActionId
	WasExecutedCalled                              func(ctx context.Context, id bridge.ActionId, id2 bridge.BatchId) bool
	SignCalled                                     func(ctx context.Context, id bridge.ActionId) (string, error)
	ExecuteCalled                                  func(ctx context.Context, id bridge.ActionId, batch *bridge.Batch) (string, error)
	SignersCountCalled                             func(ctx context.Context, id bridge.ActionId) uint

	ProposeTransferError error
	SignError            error
	ExecuteError         error
}

// GetPending -
func (b *BridgeStub) GetPending(ctx context.Context) *bridge.Batch {
	if b.GetPendingCalled != nil {
		return b.GetPendingCalled(ctx)
	}
	return nil
}

// ProposeSetStatus -
func (b *BridgeStub) ProposeSetStatus(ctx context.Context, batch *bridge.Batch) {
	if b.ProposeSetStatusCalled != nil {
		b.ProposeSetStatusCalled(ctx, batch)
	}
}

// ProposeTransfer -
func (b *BridgeStub) ProposeTransfer(ctx context.Context, batch *bridge.Batch) (string, error) {
	if b.ProposeTransferCalled != nil {
		called, err := b.ProposeTransferCalled(ctx, batch)
		if err != nil {
			return "", err
		}
		return called, nil
	}
	return "propose_tx_hash", b.ProposeTransferError
}

// WasProposedTransfer -
func (b *BridgeStub) WasProposedTransfer(ctx context.Context, batch *bridge.Batch) bool {
	if b.WasProposedTransferCalled != nil {
		return b.WasProposedTransferCalled(ctx, batch)
	}
	return false
}

// GetActionIdForProposeTransfer -
func (b *BridgeStub) GetActionIdForProposeTransfer(ctx context.Context, batch *bridge.Batch) bridge.ActionId {
	return b.GetActionIdForProposeTransferCalled(ctx, batch)
}

// WasProposedSetStatus -
func (b *BridgeStub) WasProposedSetStatus(ctx context.Context, batch *bridge.Batch) bool {
	if b.WasProposedSetStatusCalled != nil {
		return b.WasProposedSetStatusCalled(ctx, batch)
	}
	return false
}

// GetActionIdForSetStatusOnPendingTransfer -
func (b *BridgeStub) GetActionIdForSetStatusOnPendingTransfer(ctx context.Context, batch *bridge.Batch) bridge.ActionId {
	return b.GetActionIdForSetStatusOnPendingTransferCalled(ctx, batch)
}

// WasExecuted -
func (b *BridgeStub) WasExecuted(ctx context.Context, id bridge.ActionId, id2 bridge.BatchId) bool {
	if b.WasExecutedCalled != nil {
		return b.WasExecutedCalled(ctx, id, id2)
	}
	return false
}

// Sign -
func (b *BridgeStub) Sign(ctx context.Context, id bridge.ActionId) (string, error) {
	if b.SignCalled != nil {
		return b.SignCalled(ctx, id)
	}
	return "sign_tx_hash", b.SignError
}

// Execute -
func (b *BridgeStub) Execute(ctx context.Context, id bridge.ActionId, batch *bridge.Batch) (string, error) {
	if b.ExecuteCalled != nil {
		return b.ExecuteCalled(ctx, id, batch)
	}
	return "execute_tx_hash", b.ExecuteError
}

// SignersCount -
func (b *BridgeStub) SignersCount(ctx context.Context, id bridge.ActionId) uint {
	if b.SignersCountCalled != nil {
		return b.SignersCountCalled(ctx, id)
	}
	return 0
}

// -------- helper functions

// incrementFunctionCounter increments the counter for the function that called it
func (b *BridgeStub) incrementFunctionCounter() {
	b.mutExecutor.Lock()
	defer b.mutExecutor.Unlock()

	pc, _, _, _ := runtime.Caller(1)
	fmt.Printf("BridgeExecutorMock: called %s\n", runtime.FuncForPC(pc).Name())
	b.functionCalledCounter[runtime.FuncForPC(pc).Name()]++
}

// GetFunctionCounter returns the called counter of a given function
func (b *BridgeStub) GetFunctionCounter(function string) int {
	b.mutExecutor.Lock()
	defer b.mutExecutor.Unlock()

	return b.functionCalledCounter[fullPath+function]
}
