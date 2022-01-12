package ethToElrond

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/ethElrond/steps"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

type proposeTransferStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *proposeTransferStep) Execute(ctx context.Context) core.StepIdentifier {
	batch := step.bridge.GetStoredBatch()
	if batch == nil {
		step.bridge.PrintInfo(logger.LogDebug, "no batch found")
		return GettingPendingBatchFromEthereum
	}

	wasTransferProposed, err := step.bridge.WasTransferProposedOnElrond(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error determining if the batch was proposed or not on Elrond",
			"batch ID", batch.ID, "error", err)
		return GettingPendingBatchFromEthereum
	}

	if wasTransferProposed {
		return SigningProposedTransferOnElrond
	}

	if !step.bridge.MyTurnAsLeader() {
		return step.Identifier()
	}

	err = step.bridge.ProposeTransferOnElrond(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error proposing transfer on Elrond",
			"batch ID", batch.ID, "error", err)
		return GettingPendingBatchFromEthereum
	}

	return SigningProposedTransferOnElrond
}

// Identifier returns the step's identifier
func (step *proposeTransferStep) Identifier() core.StepIdentifier {
	return ProposingTransferOnElrond
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *proposeTransferStep) IsInterfaceNil() bool {
	return step == nil
}
