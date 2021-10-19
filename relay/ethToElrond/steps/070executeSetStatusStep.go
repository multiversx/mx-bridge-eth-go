package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond"
)

type executeSetStatusStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (step *executeSetStatusStep) Execute(ctx context.Context) (relay.StepIdentifier, error) {
	if step.bridge.IsLeader() {
		step.bridge.ExecuteSetStatusOnSource(ctx)
	}

	step.bridge.WaitStepToFinish(step.Identifier(), ctx)
	if step.bridge.WasSetStatusExecutedOnSource() {
		step.bridge.CleanTopology()
		step.bridge.SetStatusExecutedOnAllTransactions()

		return ethToElrond.GetPending, nil
	}

	// remain in this step
	return step.Identifier(), nil
}

// Identifier returns the step's identifier
func (step *executeSetStatusStep) Identifier() relay.StepIdentifier {
	return ethToElrond.ExecuteSetStatus
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *executeSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
