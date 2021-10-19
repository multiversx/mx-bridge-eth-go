package steps

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond"
)

type getPendingStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (step *getPendingStep) Execute() relay.StepIdentifier {
	step.bridge.GetPendingBatch()
	if step.bridge.HasPendingBatch() {
		return ethToElrond.ProposingTransfer
	}

	// remain in this step
	return step.Identifier()
}

// Identifier returns the step's identifier
func (step *getPendingStep) Identifier() relay.StepIdentifier {
	return ethToElrond.GettingPending
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *getPendingStep) IsInterfaceNil() bool {
	return step == nil
}
