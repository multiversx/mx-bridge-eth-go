package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond"
)

type proposeSetStatusStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (step *proposeSetStatusStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	err := step.bridge.SetTransactionsStatusesAccordingToDestination(ctx)
	if err != nil {
		return step.Identifier(), nil
	}

	if step.bridge.IsLeader() {
		step.bridge.ProposeSetStatusOnSource(ctx)
	}

	err = step.bridge.WaitStepToFinish(step.Identifier(), ctx)
	if err != nil {
		return step.Identifier(), err
	}

	if !step.bridge.WasProposeSetStatusExecutedOnSource(ctx) {
		// remain in this step
		return step.Identifier(), nil
	}

	step.bridge.SignProposeSetStatusOnSource(ctx)

	return ethToElrond.WaitingSignaturesForProposeSetStatus, nil
}

// Identifier returns the step's identifier
func (step *proposeSetStatusStep) Identifier() core.StepIdentifier {
	return ethToElrond.ProposingSetStatus
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *proposeSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
