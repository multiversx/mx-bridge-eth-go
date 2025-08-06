package ethtomultiversx

import (
	"context"
	"github.com/multiversx/mx-bridge-eth-go/bridges/steps"

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
	err = step.bridge.CheckPeerClientAvailability(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogDebug, "Peer client unavailable", "message", err)
	}
	step.bridge.ResetRetriesCountOnMultiversX()
	lastEthBatchExecuted, err := step.bridge.GetLastExecutedPeerBatchIDFromMultiversX(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error fetching last executed eth batch ID", "error", err)
		return step.Identifier()
	}

	err = step.bridge.GetAndStoreBatchFromPeerChain(ctx, lastEthBatchExecuted+1)
	if err != nil {
		step.bridge.PrintInfo(logger.LogDebug, "cannot fetch eth batch", "batch ID", lastEthBatchExecuted+1, "message", err)
		return step.Identifier()
	}

	batch := step.bridge.GetStoredBatch()
	if batch == nil {
		step.bridge.PrintInfo(logger.LogDebug, "no new batch found on eth", "last executed on MultiversX", lastEthBatchExecuted)
		return step.Identifier()
	}

	step.bridge.PrintInfo(logger.LogInfo, "fetched new batch from PeerChain "+batch.String())

	err = step.bridge.VerifyLastDepositNonceExecutedOnPeerBatch(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "verification failed on the new batch from PeerChain", "batch ID", lastEthBatchExecuted+1, "error", err)
		return step.Identifier()
	}

	argLists := batchProcessor.ExtractListToMvx(batch)
	err = step.bridge.CheckAvailableTokens(ctx, argLists.PeerTokens, argLists.MvxTokenBytes, argLists.Amounts, argLists.Direction)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error checking available tokens", "error", err, "batch", batch.String())
		return step.Identifier()
	}

	return ProposingTransferOnMultiversX
}

// Identifier returns the step's identifier
func (step *getPendingStep) Identifier() core.StepIdentifier {
	return GettingPendingBatchFromPeerChain
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *getPendingStep) IsInterfaceNil() bool {
	return step == nil
}
