package mock

import (
	"runtime"
	"sync"
)

var fullPath = "github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond/steps/mock.(*BridgeExecutorMock)."

// BridgeExecutorMock -
type BridgeExecutorMock struct {
	FunctionCalledCounter map[string]int
	mutExecutor           sync.RWMutex

	PrintDebugInfoCalled  func(message string, extras ...interface{})
	GetPendingBatchCalled func()
	HasPendingBatchCalled func() bool
	IsLeaderCalled        func() bool
	ProposeTransferCalled func() error
}

// PrintDebugInfo -
func (bem *BridgeExecutorMock) PrintDebugInfo(message string, extras ...interface{}) {
	bem.IncrementFunctionCounter()
	if bem.PrintDebugInfoCalled != nil {
		bem.PrintDebugInfoCalled(message, extras...)
	}
}

// GetPendingBatch -
func (bem *BridgeExecutorMock) GetPendingBatch() {
	bem.IncrementFunctionCounter()
	if bem.GetPendingBatchCalled != nil {
		bem.GetPendingBatchCalled()
	}
}

// HasPendingBatch -
func (bem *BridgeExecutorMock) HasPendingBatch() bool {
	bem.IncrementFunctionCounter()
	if bem.HasPendingBatchCalled != nil {
		return bem.HasPendingBatchCalled()
	}

	return false
}

// IsLeader -
func (bem *BridgeExecutorMock) IsLeader() bool {
	bem.IncrementFunctionCounter()
	if bem.IsLeaderCalled != nil {
		return bem.IsLeaderCalled()
	}

	return false
}

// ProposeTransfer -
func (bem *BridgeExecutorMock) ProposeTransfer() error {
	bem.IncrementFunctionCounter()
	if bem.ProposeTransferCalled != nil {
		return bem.ProposeTransferCalled()
	}

	return nil
}

// IsInterfaceNil -
func (bem *BridgeExecutorMock) IsInterfaceNil() bool {
	return bem == nil
}

// IncrementFunctionCounter increments the counter for the function that called it
func (bem *BridgeExecutorMock) IncrementFunctionCounter() {
	bem.mutExecutor.Lock()
	defer bem.mutExecutor.Unlock()

	pc, _, _, _ := runtime.Caller(1)
	bem.FunctionCalledCounter[runtime.FuncForPC(pc).Name()]++
}

// GetFunctionCounter returns the called counter of a given function
func (bem *BridgeExecutorMock) GetFunctionCounter(function string) int {
	bem.mutExecutor.Lock()
	defer bem.mutExecutor.Unlock()

	return bem.FunctionCalledCounter[fullPath+function]
}
