package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond"
)

type waitForSignaturesForProposeTransferStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (step *waitForSignaturesForProposeTransferStep) Execute(ctx context.Context) (relay.StepIdentifier, error) {
	err := step.bridge.WaitStepToFinish(step.Identifier(), ctx)
	if err != nil {
		return step.Identifier(), err
	}

	if step.bridge.IsQuorumReachedForProposeTransfer(ctx) {
		return ethToElrond.ExecutingTransfer, nil
	}

	if step.bridge.WasProposeTransferExecutedOnDestination(ctx) {
		step.bridge.CleanTopology()
		step.bridge.SetStatusExecutedOnAllTransactions()

		return ethToElrond.ProposingSetStatus, nil
	}

	// remain in this step
	return step.Identifier(), nil
}

// Identifier returns the step's identifier
func (step *waitForSignaturesForProposeTransferStep) Identifier() relay.StepIdentifier {
	return ethToElrond.WaitingSignaturesForProposeTransfer
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *waitForSignaturesForProposeTransferStep) IsInterfaceNil() bool {
	return step == nil
}
