package ethToElrond

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/ethElrond/steps"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

type getPendingStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *getPendingStep) Execute(ctx context.Context) core.StepIdentifier {
	step.bridge.ResetRetriesCountOnElrond()
	lastEthBatchExecuted, err := step.bridge.GetLastExecutedEthBatchIDFromElrond(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error fetching last executed eth batch ID", "error", err)
		return step.Identifier()
	}

	err = step.bridge.GetAndStoreBatchFromEthereum(ctx, lastEthBatchExecuted+1)
	if err != nil {
		step.bridge.PrintInfo(logger.LogDebug, "error fetching eth batch", "batch ID", lastEthBatchExecuted+1, "error", err)
		return step.Identifier()
	}

	batch := step.bridge.GetStoredBatch()
	if batch == nil {
		step.bridge.PrintInfo(logger.LogDebug, "no new batch found on eth", "last executed on Elrond", lastEthBatchExecuted)
		return step.Identifier()
	}

	isValid, err := step.bridge.ValidateEthereumBatch(batch.Json())
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error validating Ethereum batch", "error", err)
		return step.Identifier()
	}

	if !isValid {
		step.bridge.PrintInfo(logger.LogError, "batch not valid"+batch.String())
		return step.Identifier()
	}

	step.bridge.PrintInfo(logger.LogInfo, "fetched new batch from Ethereum "+batch.String())

	err = step.bridge.VerifyLastDepositNonceExecutedOnEthereumBatch(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "verification failed on the new batch from Ethereum", "batch ID", lastEthBatchExecuted+1, "error", err)
		return step.Identifier()
	}

	return ProposingTransferOnElrond
}

// Identifier returns the step's identifier
func (step *getPendingStep) Identifier() core.StepIdentifier {
	return GettingPendingBatchFromEthereum
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *getPendingStep) IsInterfaceNil() bool {
	return step == nil
}
