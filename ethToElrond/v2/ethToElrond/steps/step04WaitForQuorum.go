package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond"
)

type waitForQuorumStep struct {
	bridge bridge.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *waitForQuorumStep) Execute(ctx context.Context) core.StepIdentifier {
	if step.bridge.ProcessMaxRetriesOnElrond() {
		step.bridge.GetLogger().Debug("max number of retries reached, resetting counter")
		return ethToElrond.GettingPendingBatchFromEthereum
	}

	isQuorumReached, err := step.bridge.ProcessQuorumReachedOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error while checking the quorum", "error", err)
		return ethToElrond.GettingPendingBatchFromEthereum
	}

	step.bridge.GetLogger().Debug("quorum reached check", "is reached", isQuorumReached)

	if !isQuorumReached {
		return step.Identifier()
	}

	return ethToElrond.PerformingActionID
}

// Identifier returns the step's identifier
func (step *waitForQuorumStep) Identifier() core.StepIdentifier {
	return ethToElrond.WaitingForQuorum
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *waitForQuorumStep) IsInterfaceNil() bool {
	return step == nil
}
