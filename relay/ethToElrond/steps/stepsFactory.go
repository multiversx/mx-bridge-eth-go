package steps

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
)

// CreateSteps creates all machine states providing the bridge executor
func CreateSteps(executor BridgeExecutor) (relay.MachineStates, error) {
	if check.IfNil(executor) {
		return nil, ErrNilBridgeExecutor
	}

	return createMachineStates(executor), nil
}

func createMachineStates(executor BridgeExecutor) relay.MachineStates {
	machineStates := make(relay.MachineStates)

	machineStates[ethToElrond.GetPending] = &getPendingStep{
		bridge: executor,
	}
	machineStates[ethToElrond.ProposeTransfer] = &proposeTransferStep{
		bridge: executor,
	}

	return machineStates
}
