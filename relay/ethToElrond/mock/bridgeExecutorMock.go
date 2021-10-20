package mock

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var fullPath = "github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond/mock.(*BridgeExecutorMock)."

// BridgeExecutorMock -
type BridgeExecutorMock struct {
	functionCalledCounter map[string]int
	mutExecutor           sync.RWMutex

	HasPendingBatchCalled                         func() bool
	IsLeaderCalled                                func() bool
	WasProposeTransferExecutedOnDestinationCalled func(ctx context.Context) bool
	WasProposeSetStatusExecutedOnSourceCalled     func(ctx context.Context) bool
	WasExecutedOnDestinationCalled                func(ctx context.Context) bool
	WasExecutedOnSourceCalled                     func(ctx context.Context) bool
	IsQuorumReachedForProposeTransferCalled       func(ctx context.Context) bool
	IsQuorumReachedForProposeSetStatusCalled      func(ctx context.Context) bool

	PrintInfoCalled                          func(logLevel logger.LogLevel, message string, extras ...interface{})
	GetPendingBatchCalled                    func(ctx context.Context)
	ProposeTransferOnDestinationCalled       func(ctx context.Context) error
	ProposeSetStatusOnSourceCalled           func(ctx context.Context)
	CleanTopologyCalled                      func()
	ExecuteTransferOnDestinationCalled       func(ctx context.Context)
	ExecuteSetStatusOnSourceCalled           func(ctx context.Context)
	SetStatusRejectedOnAllTransactionsCalled func(err error)
	SetStatusExecutedOnAllTransactionsCalled func()
	SignProposeTransferOnDestinationCalled   func(ctx context.Context)
	SignProposeSetStatusOnSourceCalled       func(ctx context.Context)
	WaitStepToFinishCalled                   func(step relay.StepIdentifier, ctx context.Context) error
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
func (bem *BridgeExecutorMock) WasProposeTransferExecutedOnDestination(ctx context.Context) bool {
	bem.incrementFunctionCounter()
	if bem.WasProposeTransferExecutedOnDestinationCalled != nil {
		return bem.WasProposeTransferExecutedOnDestinationCalled(ctx)
	}

	return false
}

// WasProposeSetStatusExecutedOnSource -
func (bem *BridgeExecutorMock) WasProposeSetStatusExecutedOnSource(ctx context.Context) bool {
	bem.incrementFunctionCounter()
	if bem.WasProposeSetStatusExecutedOnSourceCalled != nil {
		return bem.WasProposeSetStatusExecutedOnSourceCalled(ctx)
	}

	return false
}

// WasExecutedOnDestination -
func (bem *BridgeExecutorMock) WasExecutedOnDestination(ctx context.Context) bool {
	bem.incrementFunctionCounter()
	if bem.WasExecutedOnDestinationCalled != nil {
		return bem.WasExecutedOnDestinationCalled(ctx)
	}

	return false
}

// WasExecutedOnSource -
func (bem *BridgeExecutorMock) WasExecutedOnSource(ctx context.Context) bool {
	bem.incrementFunctionCounter()
	if bem.WasExecutedOnSourceCalled != nil {
		return bem.WasExecutedOnSourceCalled(ctx)
	}

	return false
}

// IsQuorumReachedForProposeTransfer -
func (bem *BridgeExecutorMock) IsQuorumReachedForProposeTransfer(ctx context.Context) bool {
	bem.incrementFunctionCounter()
	if bem.IsQuorumReachedForProposeTransferCalled != nil {
		return bem.IsQuorumReachedForProposeTransferCalled(ctx)
	}

	return false
}

// IsQuorumReachedForProposeSetStatus -
func (bem *BridgeExecutorMock) IsQuorumReachedForProposeSetStatus(ctx context.Context) bool {
	bem.incrementFunctionCounter()
	if bem.IsQuorumReachedForProposeSetStatusCalled != nil {
		return bem.IsQuorumReachedForProposeSetStatusCalled(ctx)
	}

	return false
}

// -------- action functions

// PrintInfo -
func (bem *BridgeExecutorMock) PrintInfo(logLevel logger.LogLevel, message string, extras ...interface{}) {
	bem.incrementFunctionCounter()
	if bem.PrintInfoCalled != nil {
		bem.PrintInfoCalled(logLevel, message, extras...)
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
func (bem *BridgeExecutorMock) SetStatusRejectedOnAllTransactions(err error) {
	bem.incrementFunctionCounter()
	if bem.SetStatusRejectedOnAllTransactionsCalled != nil {
		bem.SetStatusRejectedOnAllTransactionsCalled(err)
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

// SignProposeSetStatusOnSource -
func (bem *BridgeExecutorMock) SignProposeSetStatusOnSource(ctx context.Context) {
	bem.incrementFunctionCounter()
	if bem.SignProposeSetStatusOnSourceCalled != nil {
		bem.SignProposeSetStatusOnSourceCalled(ctx)
	}
}

// WaitStepToFinish -
func (bem *BridgeExecutorMock) WaitStepToFinish(step relay.StepIdentifier, ctx context.Context) error {
	bem.incrementFunctionCounter()
	if bem.WaitStepToFinishCalled != nil {
		return bem.WaitStepToFinishCalled(step, ctx)
	}

	return nil
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
