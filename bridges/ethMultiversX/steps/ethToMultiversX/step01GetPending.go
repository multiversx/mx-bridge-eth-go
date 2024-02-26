package ethtomultiversx

import (
	"context"
	"encoding/json"

	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX/steps"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type getPendingStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *getPendingStep) Execute(ctx context.Context) core.StepIdentifier {
	err := step.bridge.CheckMultiversXClientAvailability(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogDebug, "MultiversX client unavailable", "message", err)
	}
	err = step.bridge.CheckEthereumClientAvailability(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogDebug, "Ethereum client unavailable", "message", err)
	}
	step.bridge.ResetRetriesCountOnMultiversX()
	lastEthBatchExecuted, err := step.bridge.GetLastExecutedEthBatchIDFromMultiversX(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error fetching last executed eth batch ID", "error", err)
		return step.Identifier()
	}

	err = step.bridge.GetAndStoreBatchFromEthereum(ctx, lastEthBatchExecuted+1)
	if err != nil {
		step.bridge.PrintInfo(logger.LogDebug, "cannot fetch eth batch", "batch ID", lastEthBatchExecuted+1, "message", err)
		return step.Identifier()
	}

	batch := step.bridge.GetStoredBatch()
	if batch == nil {
		step.bridge.PrintInfo(logger.LogDebug, "no new batch found on eth", "last executed on MultiversX", lastEthBatchExecuted)
		return step.Identifier()
	}

	isValid, err := step.bridge.ValidateBatch(ctx, batch)
	if err != nil {
		body, _ := json.Marshal(batch)
		step.bridge.PrintInfo(logger.LogError, "error validating Ethereum batch", "error", err, "batch", string(body))
		return step.Identifier()
	}

	if !isValid {
		step.bridge.PrintInfo(logger.LogError, "batch not valid "+batch.String())
		return step.Identifier()
	}

	argLists := batchProcessor.ExtractListEthToMvx(batch)
	err = step.bridge.CheckAvailableTokens(ctx, argLists.EthTokens, argLists.MvxTokenBytes, argLists.Amounts)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error checking available tokens", "error", err, "batch", batch.String())
		return step.Identifier()
	}

	step.bridge.PrintInfo(logger.LogInfo, "fetched new batch from Ethereum "+batch.String())

	err = step.bridge.VerifyLastDepositNonceExecutedOnEthereumBatch(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "verification failed on the new batch from Ethereum", "batch ID", lastEthBatchExecuted+1, "error", err)
		return step.Identifier()
	}

	return ProposingTransferOnMultiversX
}

// Identifier returns the step's identifier
func (step *getPendingStep) Identifier() core.StepIdentifier {
	return GettingPendingBatchFromEthereum
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *getPendingStep) IsInterfaceNil() bool {
	return step == nil
}
