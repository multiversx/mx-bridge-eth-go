package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/elrondToEth"
)

type resolveSetStatusStep struct {
	bridge elrondToEth.ElrondToEthBridge
}

func (step *resolveSetStatusStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	storedBatch := step.bridge.GetStoredBatchFromElrond()
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
	if statuses == nil {
		step.bridge.GetLogger().Error("nil transaction statuses")
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	for i, transactionStatus := range statuses {
		batch.Statuses[i] = transactionStatus
	}

	numStoredBatchDeposits := len(storedBatch.Statuses)
	numNewFetchedBatchDeposits := len(batch.Statuses)
	if numStoredBatchDeposits != numNewFetchedBatchDeposits {
		err = step.bridge.ResolveNewDpositsStatuses(ctx, uint64(numStoredBatchDeposits))
		if err != nil {
			step.bridge.GetLogger().Error("error while updating transaction statuses", "error", err)
			return elrondToEth.GettingPendingBatchFromElrond, nil
		}
	}

	return elrondToEth.ProposingSetStatusOnElrond, nil
}

func (step *resolveSetStatusStep) Identifier() core.StepIdentifier {
	return elrondToEth.ResolvingSetStatusOnElrond
}

func (step *resolveSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
