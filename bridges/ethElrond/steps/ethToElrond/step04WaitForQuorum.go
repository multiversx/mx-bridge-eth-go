package ethToElrond

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/ethElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
)

type waitForQuorumStep struct {
	bridge ethElrond.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *waitForQuorumStep) Execute(ctx context.Context) core.StepIdentifier {
	if step.bridge.ProcessMaxRetriesOnElrond() {
		step.bridge.GetLogger().Debug("max number of retries reached, resetting counter")
		return GettingPendingBatchFromEthereum
	}

	isQuorumReached, err := step.bridge.ProcessQuorumReachedOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error while checking the quorum", "error", err)
		return GettingPendingBatchFromEthereum
	}

	step.bridge.GetLogger().Debug("quorum reached check", "is reached", isQuorumReached)

	if !isQuorumReached {
		return step.Identifier()
	}

	return PerformingActionID
}

// Identifier returns the step's identifier
func (step *waitForQuorumStep) Identifier() core.StepIdentifier {
	return WaitingForQuorum
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *waitForQuorumStep) IsInterfaceNil() bool {
	return step == nil
}
