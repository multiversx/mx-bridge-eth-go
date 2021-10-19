package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond"
)

type proposeSetStatusStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (step *proposeSetStatusStep) Execute(ctx context.Context) (relay.StepIdentifier, error) {
	if step.bridge.IsLeader() {
		step.bridge.ProposeSetStatusOnSource(ctx)
	}

	step.bridge.WaitStepToFinish(step.Identifier(), ctx)
	if !step.bridge.WasProposeSetStatusExecutedOnSource() {
		// remain in this step
		return step.Identifier(), nil
	}

	step.bridge.SignProposeSetStatusOnDestination(ctx)

	return ethToElrond.WaitForSignaturesForProposeSetStatus, nil
}

// Identifier returns the step's identifier
func (step *proposeSetStatusStep) Identifier() relay.StepIdentifier {
	return ethToElrond.ProposeSetStatus
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *proposeSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
