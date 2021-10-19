package steps

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond"
)

type executeSetStatusStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (step *executeSetStatusStep) Execute() relay.StepIdentifier {
	if step.bridge.IsLeader() {
		step.bridge.ExecuteSetStatusOnSource()
	}

	step.bridge.WaitStepToFinish(step.Identifier())
	if step.bridge.WasSetStatusExecutedOnSource() {
		step.bridge.CleanTopology()
		step.bridge.SetStatusExecutedOnAllTransactions()

		return ethToElrond.GettingPending
	}

	// remain in this step
	return step.Identifier()
}

// Identifier returns the step's identifier
func (step *executeSetStatusStep) Identifier() relay.StepIdentifier {
	return ethToElrond.ExecutingSetStatus
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *executeSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
