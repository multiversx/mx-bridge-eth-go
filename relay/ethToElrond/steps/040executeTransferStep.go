package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond"
)

type executeTransferStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (step *executeTransferStep) Execute(ctx context.Context) relay.StepIdentifier {
	if step.bridge.IsLeader() {
		step.bridge.ExecuteTransferOnDestination(ctx)
	}

	step.bridge.WaitStepToFinish(step.Identifier(), ctx)
	if step.bridge.WasTransferExecutedOnDestination() {
		step.bridge.CleanTopology()
		step.bridge.SetStatusExecutedOnAllTransactions()

		return ethToElrond.ProposeSetStatus
	}

	// remain in this step
	return step.Identifier()
}

// Identifier returns the step's identifier
func (step *executeTransferStep) Identifier() relay.StepIdentifier {
	return ethToElrond.ExecuteTransfer
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *executeTransferStep) IsInterfaceNil() bool {
	return step == nil
}
