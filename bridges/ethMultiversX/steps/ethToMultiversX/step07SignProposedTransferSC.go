package ethtomultiversx

import (
	"context"

	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX"
	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX/steps"
	"github.com/multiversx/mx-bridge-eth-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type signProposedSCTransferStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *signProposedSCTransferStep) Execute(ctx context.Context) core.StepIdentifier {
	batch := step.bridge.GetSCExecStoredBatch()
	if batch == nil {
		step.bridge.PrintInfo(logger.LogDebug, "no batch found")
		return GettingPendingBatchFromEthereum
	}

	actionID, err := step.bridge.GetAndStoreActionIDForProposeSCTransferOnMultiversX(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error fetching action ID", "batch ID", batch.ID, "error", err)
		return GettingPendingBatchFromEthereum
	}
	if actionID == ethmultiversx.InvalidActionID {
		step.bridge.PrintInfo(logger.LogError, "contract error, got invalid action ID",
			"batch ID", batch.ID, "error", err, "action ID", actionID)
		return GettingPendingBatchFromEthereum
	}

	step.bridge.PrintInfo(logger.LogInfo, "fetched action ID", "action ID", actionID, "batch ID", batch.ID)

	wasSigned, err := step.bridge.WasActionSignedOnMultiversX(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error determining if the proposed sc transfer was signed or not",
			"batch ID", batch.ID, "error", err)
		return GettingPendingBatchFromEthereum
	}

	if wasSigned {
		return WaitingForQuorum
	}

	err = step.bridge.SignActionOnMultiversX(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error signing the proposed transfer",
			"batch ID", batch.ID, "error", err)
		return GettingPendingBatchFromEthereum
	}

	return WaitingForQuorum
}

// Identifier returns the step's identifier
func (step *signProposedSCTransferStep) Identifier() core.StepIdentifier {
	return SigningProposedTransferOnMultiversX
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *signProposedSCTransferStep) IsInterfaceNil() bool {
	return step == nil
}
