package mock

// BridgeExecutorMock -
type BridgeExecutorMock struct {
	NumCalledPrintDebugInfoCalled  int
	NumCalledGetPendingBatchCalled int
	NumHasPendingBatchCalled       int
	NumCalledIsLeaderCalled        int
	NumCalledProposeTransferCalled int

	PrintDebugInfoCalled  func(message string, extras ...interface{})
	GetPendingBatchCalled func()
	HasPendingBatchCalled func() bool
	IsLeaderCalled        func() bool
	ProposeTransferCalled func() error
}

// PrintDebugInfo -
func (bem *BridgeExecutorMock) PrintDebugInfo(message string, extras ...interface{}) {
	bem.NumCalledPrintDebugInfoCalled++
	if bem.PrintDebugInfoCalled != nil {
		bem.PrintDebugInfoCalled(message, extras...)
	}
}

// GetPendingBatch -
func (bem *BridgeExecutorMock) GetPendingBatch() {
	bem.NumCalledGetPendingBatchCalled++
	if bem.GetPendingBatchCalled != nil {
		bem.GetPendingBatchCalled()
	}
}

// HasPendingBatch -
func (bem *BridgeExecutorMock) HasPendingBatch() bool {
	bem.NumHasPendingBatchCalled++
	if bem.HasPendingBatchCalled != nil {
		return bem.HasPendingBatchCalled()
	}

	return false
}

// IsLeader -
func (bem *BridgeExecutorMock) IsLeader() bool {
	bem.NumCalledIsLeaderCalled++
	if bem.IsLeaderCalled != nil {
		return bem.IsLeaderCalled()
	}

	return false
}

// ProposeTransfer -
func (bem *BridgeExecutorMock) ProposeTransfer() error {
	bem.NumCalledProposeTransferCalled++
	if bem.ProposeTransferCalled != nil {
		return bem.ProposeTransferCalled()
	}

	return nil
}

// IsInterfaceNil -
func (bem *BridgeExecutorMock) IsInterfaceNil() bool {
	return bem == nil
}
