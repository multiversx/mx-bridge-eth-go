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
func (step *waitForSignaturesForProposeTransferStep) Execute(ctx context.Context) relay.StepIdentifier {
	step.bridge.WaitStepToFinish(step.Identifier(), ctx)
	if step.bridge.IsQuorumReachedForProposeTransfer() {
		return ethToElrond.ExecuteTransfer
	}

	if step.bridge.WasProposeTransferExecutedOnDestination() {
		step.bridge.CleanTopology()
		step.bridge.SetStatusExecutedOnAllTransactions()

		return ethToElrond.ProposeSetStatus
	}

	// remain in this step
	return step.Identifier()
}

// Identifier returns the step's identifier
func (step *waitForSignaturesForProposeTransferStep) Identifier() relay.StepIdentifier {
	return ethToElrond.WaitForSignaturesForProposeTransfer
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *waitForSignaturesForProposeTransferStep) IsInterfaceNil() bool {
	return step == nil
}
