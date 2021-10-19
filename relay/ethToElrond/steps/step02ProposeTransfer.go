package steps

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond"
)

type proposeTransferStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (step *proposeTransferStep) Execute() relay.StepIdentifier {
	if step.bridge.IsLeader() {
		err := step.bridge.ProposeTransferOnDestination()
		if err != nil {
			step.bridge.PrintDebugInfo("bridge.ProposeTransfer", "error", err)
			step.bridge.SetStatusRejectedOnAllTransactions()

			return ethToElrond.ProposingSetStatus
		}
	}

	step.bridge.WaitStepToFinish(step.Identifier())
	if !step.bridge.WasProposeTransferExecutedOnDestination() {
		// remain in this step
		return step.Identifier()
	}

	step.bridge.SignProposeTransferOnDestination()

	return ethToElrond.WaitingSignaturesForProposeTransfer
}

// Identifier returns the step's identifier
func (step *proposeTransferStep) Identifier() relay.StepIdentifier {
	return ethToElrond.ProposingTransfer
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *proposeTransferStep) IsInterfaceNil() bool {
	return step == nil
}
