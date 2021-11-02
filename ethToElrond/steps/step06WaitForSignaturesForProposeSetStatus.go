package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond"
)

type waitForSignaturesForProposeSetStatusStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (step *waitForSignaturesForProposeSetStatusStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	err := step.bridge.WaitStepToFinish(step.Identifier(), ctx)
	if err != nil {
		return step.Identifier(), err
	}

	if step.bridge.IsQuorumReachedForProposeSetStatus(ctx) {
		return ethToElrond.ExecutingSetStatus, nil
	}

	if step.bridge.WasProposeSetStatusExecutedOnSource(ctx) {
		step.bridge.CleanStoredSignatures()

		return ethToElrond.GettingPending, nil
	}

	// remain in this step
	return step.Identifier(), nil
}

// Identifier returns the step's identifier
func (step *waitForSignaturesForProposeSetStatusStep) Identifier() core.StepIdentifier {
	return ethToElrond.WaitingSignaturesForProposeSetStatus
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *waitForSignaturesForProposeSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
