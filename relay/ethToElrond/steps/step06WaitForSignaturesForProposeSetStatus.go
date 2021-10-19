package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond"
)

type waitForSignaturesForProposeSetStatusStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (step *waitForSignaturesForProposeSetStatusStep) Execute(ctx context.Context) (relay.StepIdentifier, error) {
	step.bridge.WaitStepToFinish(step.Identifier(), ctx)
	if step.bridge.IsQuorumReachedForProposeSetStatus() {
		return ethToElrond.ExecutingSetStatus, nil
	}

	if step.bridge.WasProposeSetStatusExecutedOnSource() {
		step.bridge.CleanTopology()
		step.bridge.SetStatusExecutedOnAllTransactions()

		return ethToElrond.GettingPending, nil
	}

	// remain in this step
	return step.Identifier(), nil
}

// Identifier returns the step's identifier
func (step *waitForSignaturesForProposeSetStatusStep) Identifier() relay.StepIdentifier {
	return ethToElrond.WaitingSignaturesForProposeSetStatus
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *waitForSignaturesForProposeSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
