package elrondToEth

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/ethElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
)

type resolveSetStatusStep struct {
	bridge ethElrond.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *resolveSetStatusStep) Execute(ctx context.Context) core.StepIdentifier {
	storedBatch := step.bridge.GetStoredBatch()
	if storedBatch == nil {
		step.bridge.GetLogger().Debug("nil batch stored")
		return GettingPendingBatchFromElrond
	}

	batch, err := step.bridge.GetBatchFromElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error while fetching batch", "error", err)
		return GettingPendingBatchFromElrond
	}
	if batch == nil {
		step.bridge.GetLogger().Debug("nil batch fetched")
		return GettingPendingBatchFromElrond
	}

	statuses, err := step.bridge.GetBatchStatusesFromEthereum(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error while fetching transaction statuses", "error", err)
		return GettingPendingBatchFromElrond
	}

	storedBatch.Statuses = statuses

	step.bridge.ResolveNewDepositsStatuses(uint64(len(batch.Statuses)))

	return ProposingSetStatusOnElrond
}

// Identifier returns the step's identifier
func (step *resolveSetStatusStep) Identifier() core.StepIdentifier {
	return ResolvingSetStatusOnElrond
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *resolveSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
