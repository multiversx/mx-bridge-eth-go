package testsCommon

import (
	"github.com/multiversx/mx-bridge-eth-go/core"
)

// ScCallsExecuteFilterStub -
type ScCallsExecuteFilterStub struct {
	ShouldExecuteCalled func(callData core.ProxySCCompleteCallData) bool
}

// ShouldExecute -
func (stub *ScCallsExecuteFilterStub) ShouldExecute(callData core.ProxySCCompleteCallData) bool {
	if stub.ShouldExecuteCalled != nil {
		return stub.ShouldExecuteCalled(callData)
	}

	return true
}

// IsInterfaceNil -
func (stub *ScCallsExecuteFilterStub) IsInterfaceNil() bool {
	return stub == nil
}
