package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	v2 "github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond"
)

type signProposedTransferStep struct {
	bridge bridge.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *signProposedTransferStep) Execute(ctx context.Context) core.StepIdentifier {
	batch := step.bridge.GetStoredBatch()
	if batch == nil {
		step.bridge.GetLogger().Debug("no batch found")
		return ethToElrond.GettingPendingBatchFromEthereum
	}

	actionID, err := step.bridge.GetAndStoreActionIDForProposeTransferOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error fetching action ID", "batch ID", batch.ID, "error", err)
		return ethToElrond.GettingPendingBatchFromEthereum
	}
	if actionID == v2.InvalidActionID {
		step.bridge.GetLogger().Error("contract error, got invalid action ID",
			"batch ID", batch.ID, "error", err, "action ID", actionID)
		return ethToElrond.GettingPendingBatchFromEthereum
	}

	step.bridge.GetLogger().Info("fetched action ID", "action ID", actionID, "batch ID", batch.ID)

	wasSigned, err := step.bridge.WasActionSignedOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error determining if the proposed transfer was signed or not",
			"batch ID", batch.ID, "error", err)
		return ethToElrond.GettingPendingBatchFromEthereum
	}

	if wasSigned {
		return ethToElrond.WaitingForQuorum
	}

	err = step.bridge.SignActionOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error signing the proposed transfer",
			"batch ID", batch.ID, "error", err)
		return ethToElrond.GettingPendingBatchFromEthereum
	}

	return ethToElrond.WaitingForQuorum
}

// Identifier returns the step's identifier
func (step *signProposedTransferStep) Identifier() core.StepIdentifier {
	return ethToElrond.SigningProposedTransferOnElrond
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *signProposedTransferStep) IsInterfaceNil() bool {
	return step == nil
}
