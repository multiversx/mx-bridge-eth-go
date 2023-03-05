package multiversxtoeth

import (
	"context"

	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX/steps"
	"github.com/multiversx/mx-bridge-eth-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type waitForQuorumOnSetStatusStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *waitForQuorumOnSetStatusStep) Execute(ctx context.Context) core.StepIdentifier {
	if step.bridge.ProcessMaxQuorumRetriesOnMultiversX() {
		step.bridge.PrintInfo(logger.LogDebug, "max number of retries reached, resetting counter")
		return GettingPendingBatchFromMultiversX
	}

	isQuorumReached, err := step.bridge.ProcessQuorumReachedOnMultiversX(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error while checking the quorum", "error", err)
		return GettingPendingBatchFromMultiversX
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
