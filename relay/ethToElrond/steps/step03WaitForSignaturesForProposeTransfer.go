package steps

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond"
)

type waitForSignaturesForProposeTransferStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (step *waitForSignaturesForProposeTransferStep) Execute() relay.StepIdentifier {
	step.bridge.WaitStepToFinish(step.Identifier())
	if step.bridge.IsQuorumReachedForProposeTransfer() {
		return ethToElrond.ExecutingTransfer
	}

	if step.bridge.WasProposeTransferExecutedOnDestination() {
		step.bridge.CleanTopology()
		step.bridge.SetStatusExecutedOnAllTransactions()

		return ethToElrond.ProposingSetStatus
	}

	// remain in this step
	return step.Identifier()
}

// Identifier returns the step's identifier
func (step *waitForSignaturesForProposeTransferStep) Identifier() relay.StepIdentifier {
	return ethToElrond.WaitingSignaturesForProposeTransfer
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *waitForSignaturesForProposeTransferStep) IsInterfaceNil() bool {
	return step == nil
}
