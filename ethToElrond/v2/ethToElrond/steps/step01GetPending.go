package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond"
)

type getPendingStep struct {
	bridge bridge.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *getPendingStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	step.bridge.ResetRetriesCountOnElrond()
	lastEthBatchExecuted, err := step.bridge.GetLastExecutedEthBatchIDFromElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error fetching last executed eth batch ID", "error", err)
		return step.Identifier(), nil
	}

	err = step.bridge.GetAndStoreBatchFromEthereum(ctx, lastEthBatchExecuted+1)
	if err != nil {
		step.bridge.GetLogger().Error("error fetching eth batch", "batch ID", lastEthBatchExecuted+1, "error", err)
		return step.Identifier(), nil
	}

	batch := step.bridge.GetStoredBatch()
	if batch == nil {
		step.bridge.GetLogger().Debug("no new batch found on eth", "last executed on Elrond", lastEthBatchExecuted)
		return step.Identifier(), nil
	}

	step.bridge.GetLogger().Info("fetched new batch from Ethereum " + batch.String())

	err = step.bridge.VerifyLastDepositNonceExecutedOnEthereumBatch(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("verification failed on the new batch from Ethereum", "batch ID", lastEthBatchExecuted+1, "error", err)
		return step.Identifier(), nil
	}

	return ethToElrond.ProposingTransferOnElrond, nil
}

// Identifier returns the step's identifier
func (step *getPendingStep) Identifier() core.StepIdentifier {
	return ethToElrond.GettingPendingBatchFromEthereum
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *getPendingStep) IsInterfaceNil() bool {
	return step == nil
}
