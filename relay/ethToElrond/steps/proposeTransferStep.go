package steps

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond"
)

type proposeTransferStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (pts *proposeTransferStep) Execute() relay.StepIdentifier {
	if !pts.bridge.IsLeader() {
		// TODO go to the correct next step
		return ethToElrond.GetPending
	}

	err := pts.bridge.ProposeTransfer()
	pts.bridge.PrintDebugInfo("bridge.ProposeTransfer", "error", err)
	if err != nil {
		// TODO go to the correct next step
		return ethToElrond.GetPending
	}

	// TODO go to the correct next step
	return ethToElrond.GetPending
}

// IsInterfaceNil returns true if there is no value under the interface
func (pts *proposeTransferStep) IsInterfaceNil() bool {
	return pts == nil
}
