package multiversxtoeth

import (
	"context"
	"fmt"
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
		step.bridge.PrintInfo(logger.LogDebug, "PeerChain client unavailable", "message", err)
	}
	step.bridge.ResetRetriesCountOnPeerChain()
	step.resetCountersOnMultiversX()

	batch, err := step.bridge.GetBatchFromMultiversX(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogDebug, "cannot fetch MultiversX batch", "message", err)
		return step.Identifier()
	}
	if batch == nil {
		step.bridge.PrintInfo(logger.LogDebug, "no new batch found on MultiversX")
		return step.Identifier()
	}

	err = step.bridge.StoreBatchFromMultiversX(batch)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error storing MultiversX batch", "error", err)
		return step.Identifier()
	}

	step.bridge.PrintInfo(logger.LogInfo, "fetched new batch from MultiversX "+batch.String())

	wasPerformed, err := step.bridge.WasTransferPerformedOnPeerChain(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error determining if transfer was performed or not", "error", err)
		return step.Identifier()
	}
	fmt.Println("WAS PERFORMED", wasPerformed)
	if wasPerformed {
		step.bridge.PrintInfo(logger.LogInfo, "transfer performed")
		return ResolvingSetStatusOnMultiversX
	}

	argLists := batchProcessor.ExtractListFromMvx(batch)
	err = step.bridge.CheckAvailableTokens(ctx, argLists.PeerTokens, argLists.MvxTokenBytes, argLists.Amounts, argLists.Direction)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error checking available tokens", "error", err, "batch", batch.String())
		return step.Identifier()
	}

	return SigningProposedTransferOnPeerChain
}

// Identifier returns the step's identifier
func (step *getPendingStep) Identifier() core.StepIdentifier {
	return GettingPendingBatchFromMultiversX
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *getPendingStep) IsInterfaceNil() bool {
	return step == nil
}

func (step *getPendingStep) resetCountersOnMultiversX() {
	step.bridge.ResetRetriesCountOnMultiversX()
	step.bridge.ResetRetriesOnWasTransferProposedOnMultiversX()
}
