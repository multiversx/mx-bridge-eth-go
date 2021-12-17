package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/elrondToEth"
)

type resolveSetStatusStep struct {
	bridge bridge.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *resolveSetStatusStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	storedBatch := step.bridge.GetStoredBatch()
	if storedBatch == nil {
		step.bridge.GetLogger().Debug("nil batch stored")
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	batch, err := step.bridge.GetBatchFromElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error while fetching batch", "error", err)
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}
	if batch == nil {
		step.bridge.GetLogger().Debug("nil batch fetched")
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	statuses, err := step.bridge.GetBatchStatusesFromEthereum(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error while fetching transaction statuses", "error", err)
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	storedBatch.Statuses = statuses

	step.bridge.ResolveNewDepositsStatuses(uint64(len(batch.Statuses)))

	return elrondToEth.ProposingSetStatusOnElrond, nil
}

// Identifier returns the step's identifier
func (step *resolveSetStatusStep) Identifier() core.StepIdentifier {
	return elrondToEth.ResolvingSetStatusOnElrond
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *resolveSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
