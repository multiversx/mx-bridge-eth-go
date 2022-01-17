package elrondToEth

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/ethElrond/steps"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

type waitForQuorumOnSetStatusStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *waitForQuorumOnSetStatusStep) Execute(ctx context.Context) core.StepIdentifier {
	if step.bridge.ProcessMaxRetriesOnElrond() {
		step.bridge.PrintInfo(logger.LogDebug, "max number of retries reached, resetting counter")
		return GettingPendingBatchFromElrond
	}

	isQuorumReached, err := step.bridge.ProcessQuorumReachedOnElrond(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error while checking the quorum", "error", err)
		return GettingPendingBatchFromElrond
	}

	step.bridge.PrintInfo(logger.LogDebug, "quorum reached check", "is reached", isQuorumReached)

	if !isQuorumReached {
		return step.Identifier()
	}

	return PerformingSetStatus
}

// Identifier returns the step's identifier
func (step *waitForQuorumOnSetStatusStep) Identifier() core.StepIdentifier {
	return WaitingForQuorumOnSetStatus
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *waitForQuorumOnSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
