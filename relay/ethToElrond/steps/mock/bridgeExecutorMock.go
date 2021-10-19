package mock

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
)

var fullPath = "github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond/steps/mock.(*BridgeExecutorMock)."

// BridgeExecutorMock -
type BridgeExecutorMock struct {
	functionCalledCounter map[string]int
	mutExecutor           sync.RWMutex

	HasPendingBatchCalled                         func() bool
	IsLeaderCalled                                func() bool
	WasProposeTransferExecutedOnDestinationCalled func() bool
	WasProposeSetStatusExecutedOnSourceCalled     func() bool
	WasTransferExecutedOnDestinationCalled        func() bool
	WasSetStatusExecutedOnSourceCalled            func() bool
	IsQuorumReachedForProposeTransferCalled       func() bool
	IsQuorumReachedForProposeSetStatusCalled      func() bool

	PrintDebugInfoCalled                     func(message string, extras ...interface{})
	GetPendingBatchCalled                    func(ctx context.Context)
	ProposeTransferOnDestinationCalled       func(ctx context.Context) error
	ProposeSetStatusOnSourceCalled           func(ctx context.Context)
	CleanTopologyCalled                      func()
	ExecuteTransferOnDestinationCalled       func(ctx context.Context)
	ExecuteSetStatusOnSourceCalled           func(ctx context.Context)
	SetStatusRejectedOnAllTransactionsCalled func()
	SetStatusExecutedOnAllTransactionsCalled func()
	SignProposeTransferOnDestinationCalled   func(ctx context.Context)
	SignProposeSetStatusOnDestinationCalled  func(ctx context.Context)
	WaitStepToFinishCalled                   func(step relay.StepIdentifier, ctx context.Context)
}

// NewBridgeExecutorMock creates a new BridgeExecutorMock instance
func NewBridgeExecutorMock() *BridgeExecutorMock {
	return &BridgeExecutorMock{
		functionCalledCounter: make(map[string]int),
	}
}

// -------- decision functions

// HasPendingBatch -
func (bem *BridgeExecutorMock) HasPendingBatch() bool {
	bem.incrementFunctionCounter()
	if bem.HasPendingBatchCalled != nil {
		return bem.HasPendingBatchCalled()
	}

	return false
}

// IsLeader -
func (bem *BridgeExecutorMock) IsLeader() bool {
	bem.incrementFunctionCounter()
	if bem.IsLeaderCalled != nil {
		return bem.IsLeaderCalled()
	}

	return false
}

// WasProposeTransferExecutedOnDestination -
func (bem *BridgeExecutorMock) WasProposeTransferExecutedOnDestination() bool {
	bem.incrementFunctionCounter()
	if bem.WasProposeTransferExecutedOnDestinationCalled != nil {
		return bem.WasProposeTransferExecutedOnDestinationCalled()
	}

	return false
}

// WasProposeSetStatusExecutedOnSource -
func (bem *BridgeExecutorMock) WasProposeSetStatusExecutedOnSource() bool {
	bem.incrementFunctionCounter()
	if bem.WasProposeSetStatusExecutedOnSourceCalled != nil {
		return bem.WasProposeSetStatusExecutedOnSourceCalled()
	}

	return false
}

// WasTransferExecutedOnDestination -
func (bem *BridgeExecutorMock) WasTransferExecutedOnDestination() bool {
	bem.incrementFunctionCounter()
	if bem.WasTransferExecutedOnDestinationCalled != nil {
		return bem.WasTransferExecutedOnDestinationCalled()
	}

	return false
}

// WasSetStatusExecutedOnSource -
func (bem *BridgeExecutorMock) WasSetStatusExecutedOnSource() bool {
	bem.incrementFunctionCounter()
	if bem.WasSetStatusExecutedOnSourceCalled != nil {
		return bem.WasSetStatusExecutedOnSourceCalled()
	}

	return false
}

// IsQuorumReachedForProposeTransfer -
func (bem *BridgeExecutorMock) IsQuorumReachedForProposeTransfer() bool {
	bem.incrementFunctionCounter()
	if bem.IsQuorumReachedForProposeTransferCalled != nil {
		return bem.IsQuorumReachedForProposeTransferCalled()
	}

	return false
}

// IsQuorumReachedForProposeSetStatus -
func (bem *BridgeExecutorMock) IsQuorumReachedForProposeSetStatus() bool {
	bem.incrementFunctionCounter()
	if bem.IsQuorumReachedForProposeSetStatusCalled != nil {
		return bem.IsQuorumReachedForProposeSetStatusCalled()
	}

	return false
}

// -------- action functions

// PrintDebugInfo -
func (bem *BridgeExecutorMock) PrintDebugInfo(message string, extras ...interface{}) {
	bem.incrementFunctionCounter()
	if bem.PrintDebugInfoCalled != nil {
		bem.PrintDebugInfoCalled(message, extras...)
	}
}

// GetPendingBatch -
func (bem *BridgeExecutorMock) GetPendingBatch(ctx context.Context) {
	bem.incrementFunctionCounter()
	if bem.GetPendingBatchCalled != nil {
		bem.GetPendingBatchCalled(ctx)
	}
}

// ProposeTransferOnDestination -
func (bem *BridgeExecutorMock) ProposeTransferOnDestination(ctx context.Context) error {
	bem.incrementFunctionCounter()
	if bem.ProposeTransferOnDestinationCalled != nil {
		return bem.ProposeTransferOnDestinationCalled(ctx)
	}

	return nil
}

// ProposeSetStatusOnSource -
func (bem *BridgeExecutorMock) ProposeSetStatusOnSource(ctx context.Context) {
	bem.incrementFunctionCounter()
	if bem.ProposeSetStatusOnSourceCalled != nil {
		bem.ProposeSetStatusOnSourceCalled(ctx)
	}
}

// CleanTopology -
func (bem *BridgeExecutorMock) CleanTopology() {
	bem.incrementFunctionCounter()
	if bem.CleanTopologyCalled != nil {
		bem.CleanTopologyCalled()
	}
}

// ExecuteTransferOnDestination -
func (bem *BridgeExecutorMock) ExecuteTransferOnDestination(ctx context.Context) {
	bem.incrementFunctionCounter()
	if bem.ExecuteTransferOnDestinationCalled != nil {
		bem.ExecuteTransferOnDestinationCalled(ctx)
	}
}

// ExecuteSetStatusOnSource -
func (bem *BridgeExecutorMock) ExecuteSetStatusOnSource(ctx context.Context) {
	bem.incrementFunctionCounter()
	if bem.ExecuteSetStatusOnSourceCalled != nil {
		bem.ExecuteSetStatusOnSourceCalled(ctx)
	}
}

// SetStatusRejectedOnAllTransactions -
func (bem *BridgeExecutorMock) SetStatusRejectedOnAllTransactions() {
	bem.incrementFunctionCounter()
	if bem.SetStatusRejectedOnAllTransactionsCalled != nil {
		bem.SetStatusRejectedOnAllTransactionsCalled()
	}
}

// SetStatusExecutedOnAllTransactions -
func (bem *BridgeExecutorMock) SetStatusExecutedOnAllTransactions() {
	bem.incrementFunctionCounter()
	if bem.SetStatusExecutedOnAllTransactionsCalled != nil {
		bem.SetStatusExecutedOnAllTransactionsCalled()
	}
}

// SignProposeTransferOnDestination -
func (bem *BridgeExecutorMock) SignProposeTransferOnDestination(ctx context.Context) {
	bem.incrementFunctionCounter()
	if bem.SignProposeTransferOnDestinationCalled != nil {
		bem.SignProposeTransferOnDestinationCalled(ctx)
	}
}

// SignProposeSetStatusOnDestination -
func (bem *BridgeExecutorMock) SignProposeSetStatusOnDestination(ctx context.Context) {
	bem.incrementFunctionCounter()
	if bem.SignProposeSetStatusOnDestinationCalled != nil {
		bem.SignProposeSetStatusOnDestinationCalled(ctx)
	}
}

// WaitStepToFinish -
func (bem *BridgeExecutorMock) WaitStepToFinish(step relay.StepIdentifier, ctx context.Context) {
	bem.incrementFunctionCounter()
	if bem.WaitStepToFinishCalled != nil {
		bem.WaitStepToFinishCalled(step, ctx)
	}
}

// -------- helper functions

// incrementFunctionCounter increments the counter for the function that called it
func (bem *BridgeExecutorMock) incrementFunctionCounter() {
	bem.mutExecutor.Lock()
	defer bem.mutExecutor.Unlock()

	pc, _, _, _ := runtime.Caller(1)
	fmt.Printf("BridgeExecutorMock: called %s\n", runtime.FuncForPC(pc).Name())
	bem.functionCalledCounter[runtime.FuncForPC(pc).Name()]++
}

// GetFunctionCounter returns the called counter of a given function
func (bem *BridgeExecutorMock) GetFunctionCounter(function string) int {
	bem.mutExecutor.Lock()
	defer bem.mutExecutor.Unlock()

	return bem.functionCalledCounter[fullPath+function]
}

// IsInterfaceNil -
func (bem *BridgeExecutorMock) IsInterfaceNil() bool {
	return bem == nil
}
