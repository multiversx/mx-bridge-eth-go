package mock

import (
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
	GetPendingBatchCalled                    func()
	ProposeTransferOnDestinationCalled       func() error
	ProposeSetStatusOnSourceCalled           func()
	CleanTopologyCalled                      func()
	ExecuteTransferOnDestinationCalled       func()
	ExecuteSetStatusOnSourceCalled           func()
	SetStatusRejectedOnAllTransactionsCalled func()
	SetStatusExecutedOnAllTransactionsCalled func()
	SignProposeTransferOnDestinationCalled   func()
	SignProposeSetStatusOnDestinationCalled  func()
	WaitStepToFinishCalled                   func(step relay.StepIdentifier)
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
func (bem *BridgeExecutorMock) GetPendingBatch() {
	bem.incrementFunctionCounter()
	if bem.GetPendingBatchCalled != nil {
		bem.GetPendingBatchCalled()
	}
}

// ProposeTransferOnDestination -
func (bem *BridgeExecutorMock) ProposeTransferOnDestination() error {
	bem.incrementFunctionCounter()
	if bem.ProposeTransferOnDestinationCalled != nil {
		return bem.ProposeTransferOnDestinationCalled()
	}

	return nil
}

// ProposeSetStatusOnSource -
func (bem *BridgeExecutorMock) ProposeSetStatusOnSource() {
	bem.incrementFunctionCounter()
	if bem.ProposeSetStatusOnSourceCalled != nil {
		bem.ProposeSetStatusOnSourceCalled()
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
func (bem *BridgeExecutorMock) ExecuteTransferOnDestination() {
	bem.incrementFunctionCounter()
	if bem.ExecuteTransferOnDestinationCalled != nil {
		bem.ExecuteTransferOnDestinationCalled()
	}
}

// ExecuteSetStatusOnSource -
func (bem *BridgeExecutorMock) ExecuteSetStatusOnSource() {
	bem.incrementFunctionCounter()
	if bem.ExecuteSetStatusOnSourceCalled != nil {
		bem.ExecuteSetStatusOnSourceCalled()
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
func (bem *BridgeExecutorMock) SignProposeTransferOnDestination() {
	bem.incrementFunctionCounter()
	if bem.SignProposeTransferOnDestinationCalled != nil {
		bem.SignProposeTransferOnDestinationCalled()
	}
}

// SignProposeSetStatusOnDestination -
func (bem *BridgeExecutorMock) SignProposeSetStatusOnDestination() {
	bem.incrementFunctionCounter()
	if bem.SignProposeSetStatusOnDestinationCalled != nil {
		bem.SignProposeSetStatusOnDestinationCalled()
	}
}

// WaitStepToFinish -
func (bem *BridgeExecutorMock) WaitStepToFinish(step relay.StepIdentifier) {
	bem.incrementFunctionCounter()
	if bem.WaitStepToFinishCalled != nil {
		bem.WaitStepToFinishCalled(step)
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
