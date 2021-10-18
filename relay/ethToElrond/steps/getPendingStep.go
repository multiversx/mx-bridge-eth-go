package steps

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond"
)

type getPendingStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (gps *getPendingStep) Execute() relay.StepIdentifier {
	gps.bridge.GetPendingBatch()
	if gps.bridge.HasPendingBatch() {
		return ethToElrond.ProposeTransfer
	}

	return ethToElrond.GetPending
}

// IsInterfaceNil returns true if there is no value under the interface
func (gps *getPendingStep) IsInterfaceNil() bool {
	return gps == nil
}
