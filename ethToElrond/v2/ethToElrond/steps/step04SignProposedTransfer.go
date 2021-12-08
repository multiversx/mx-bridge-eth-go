package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond"
)

type signProposedTransferStep struct {
	bridge ethToElrond.EthToElrondBridge
}

// Execute will execute this step returning the next step to be executed
func (step *signProposedTransferStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	batch := step.bridge.GetStoredBatch()

	if batch == nil {
		step.bridge.GetLogger().Debug("no batch found")
		return ethToElrond.GettingPendingBatchFromEthereum, nil
	}

	wasSigned, err := step.bridge.WasProposedTransferSignedOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error determining if the proposed transfer was signed or not",
			"batch ID", batch.ID, "error", err)
		return ethToElrond.GettingPendingBatchFromEthereum, nil
	}

	if wasSigned {
		return ethToElrond.WaitingForQuorum, nil
	}

	err = step.bridge.SignProposedTransferOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error signing the proposed transfer",
			"batch ID", batch.ID, "error", err)
		return ethToElrond.GettingPendingBatchFromEthereum, nil
	}

	return ethToElrond.WaitingForQuorum, nil
}

// Identifier returns the step's identifier
func (step *signProposedTransferStep) Identifier() core.StepIdentifier {
	return ethToElrond.SigningProposedTransferOnElrond
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *signProposedTransferStep) IsInterfaceNil() bool {
	return step == nil
}
