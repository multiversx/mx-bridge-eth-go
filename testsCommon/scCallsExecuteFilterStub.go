package testsCommon

import "github.com/multiversx/mx-bridge-eth-go/parsers"

// ScCallsExecuteFilterStub -
type ScCallsExecuteFilterStub struct {
	ShouldExecuteCalled func(callData parsers.ProxySCCompleteCallData) bool
}

// ShouldExecute -
func (stub *ScCallsExecuteFilterStub) ShouldExecute(callData parsers.ProxySCCompleteCallData) bool {
	if stub.ShouldExecuteCalled != nil {
		return stub.ShouldExecuteCalled(callData)
	}

	return true
}

// IsInterfaceNil -
func (stub *ScCallsExecuteFilterStub) IsInterfaceNil() bool {
	return stub == nil
}
