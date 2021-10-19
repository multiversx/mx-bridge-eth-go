package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond"
)

type getPendingStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (step *getPendingStep) Execute(ctx context.Context) relay.StepIdentifier {
	step.bridge.GetPendingBatch(ctx)
	if step.bridge.HasPendingBatch() {
		return ethToElrond.ProposeTransfer
	}

	// remain in this step
	return step.Identifier()
}

// Identifier returns the step's identifier
func (step *getPendingStep) Identifier() relay.StepIdentifier {
	return ethToElrond.GetPending
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *getPendingStep) IsInterfaceNil() bool {
	return step == nil
}
