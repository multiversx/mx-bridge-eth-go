package multiversxtoeth

import (
	"context"

	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX/steps"
	"github.com/multiversx/mx-bridge-eth-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type resolveSetStatusStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *resolveSetStatusStep) Execute(ctx context.Context) core.StepIdentifier {
	step.bridge.ClearStoredP2PSignaturesForEthereum()
	storedBatch := step.bridge.GetStoredBatch()
	if storedBatch == nil {
		step.bridge.PrintInfo(logger.LogDebug, "nil batch stored")
		return GettingPendingBatchFromMultiversX
	}

	batch, err := step.bridge.GetBatchFromMultiversX(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error while fetching batch", "error", err)
		return GettingPendingBatchFromMultiversX
	}
	if batch == nil {
		step.bridge.PrintInfo(logger.LogDebug, "nil batch fetched")
		return GettingPendingBatchFromMultiversX
	}

	statuses := step.bridge.WaitAndReturnFinalBatchStatuses(ctx)
	if len(statuses) == 0 {
		return GettingPendingBatchFromMultiversX
	}

	storedBatch.Statuses = statuses

	step.bridge.ResolveNewDepositsStatuses(uint64(len(batch.Statuses)))

	return ProposingSetStatusOnMultiversX
}

// Identifier returns the step's identifier
func (step *resolveSetStatusStep) Identifier() core.StepIdentifier {
	return ResolvingSetStatusOnMultiversX
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *resolveSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
