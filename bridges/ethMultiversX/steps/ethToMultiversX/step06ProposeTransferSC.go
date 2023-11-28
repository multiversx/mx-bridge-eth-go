package ethtomultiversx

import (
	"context"
	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX/steps"
	"github.com/multiversx/mx-bridge-eth-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type proposeSCTransferStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *proposeSCTransferStep) Execute(ctx context.Context) core.StepIdentifier {
	batch := step.bridge.GetSCExecStoredBatch()

	if batch == nil {
		step.bridge.PrintInfo(logger.LogDebug, "no batch found")
		return GettingPendingBatchFromEthereum
	}

	if len(batch.Deposits) == 0 {
		step.bridge.PrintInfo(logger.LogDebug, "no sc transfers found")
		return GettingPendingBatchFromEthereum
	}

	wasTransferProposed, err := step.bridge.WasSCTransferProposedOnMultiversX(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error determining if the batch was proposed or not on MultiversX",
			"batch ID", batch.ID, "error", err)
		return GettingPendingBatchFromEthereum
	}

	if wasTransferProposed {
		return SigningProposedSCTransferOnMultiversX
	}

	batchSCMetadata, err := step.bridge.GetBatchSCMetadata(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error fetching sc events for current batch",
			"batch ID", batch.ID, "error", err)
		return GettingPendingBatchFromEthereum
	}
	if len(batchSCMetadata.Deposits) != len(batch.Deposits) {
		step.bridge.PrintInfo(logger.LogError, "invalid number of sc events for batch",
			"batch ID", batch.ID, "needed", len(batch.Deposits), "having", len(batchSCMetadata.Deposits))
		return GettingPendingBatchFromEthereum
	}

	if !step.bridge.MyTurnAsLeader() {
		step.bridge.PrintInfo(logger.LogDebug, "not my turn as leader in this round")
		return step.Identifier()
	}

	err = step.bridge.ProposeSCTransferOnMultiversX(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error proposing SC transfer on MultiversX",
			"batch ID", batch.ID, "error", err)
		return GettingPendingBatchFromEthereum
	}

	return SigningProposedSCTransferOnMultiversX
}

// Identifier returns the step's identifier
func (step *proposeSCTransferStep) Identifier() core.StepIdentifier {
	return ProposingSCTransfersOnMultiversX
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *proposeSCTransferStep) IsInterfaceNil() bool {
	return step == nil
}
