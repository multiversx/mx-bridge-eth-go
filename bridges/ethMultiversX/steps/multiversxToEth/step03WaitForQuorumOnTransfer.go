package multiversxtoeth

import (
	"context"

	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX/steps"
	"github.com/multiversx/mx-bridge-eth-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type waitForQuorumOnTransferStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *waitForQuorumOnTransferStep) Execute(ctx context.Context) core.StepIdentifier {
	if step.bridge.ProcessMaxQuorumRetriesOnEthereum() {
		step.bridge.PrintInfo(logger.LogDebug, "max number of retries reached, resetting counter")
		return GettingPendingBatchFromMultiversX
	}

	isQuorumReached, err := step.bridge.ProcessQuorumReachedOnEthereum(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error while checking the quorum on Ethereum", "error", err)
		return GettingPendingBatchFromMultiversX
	}

	step.bridge.PrintInfo(logger.LogDebug, "quorum reached check", "is reached", isQuorumReached)

	if !isQuorumReached {
		return step.Identifier()
	}

	return PerformingTransfer
}

// Identifier returns the step's identifier
func (step *waitForQuorumOnTransferStep) Identifier() core.StepIdentifier {
	return WaitingForQuorumOnTransfer
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *waitForQuorumOnTransferStep) IsInterfaceNil() bool {
	return step == nil
}
