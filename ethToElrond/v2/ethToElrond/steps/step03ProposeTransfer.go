package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond"
)

type proposeTransferStep struct {
	bridge ethToElrond.EthToElrondBridge
}

// Execute will execute this step returning the next step to be executed
func (step *proposeTransferStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	batch := step.bridge.GetStoredBatch()

	if batch == nil {
		step.bridge.GetLogger().Debug("no batch found")
		return ethToElrond.GettingPendingBatchFromEthereum, nil
	}

	wasTransferProposed, err := step.bridge.WasTransferProposedOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error determining if the batch was proposed or not on Elrond",
			"batch ID", batch.ID, "error", err)
		return ethToElrond.GettingPendingBatchFromEthereum, nil
	}

	if wasTransferProposed {
		return ethToElrond.SigningProposedTransferOnElrond, nil
	}

	if !step.bridge.MyTurnAsLeader() {
		return step.Identifier(), nil
	}

	err = step.bridge.ProposeTransferOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error proposing transfer on Elrond",
			"batch ID", batch.ID, "error", err)
		return ethToElrond.GettingPendingBatchFromEthereum, nil
	}

	return ethToElrond.SigningProposedTransferOnElrond, nil
}

// Identifier returns the step's identifier
func (step *proposeTransferStep) Identifier() core.StepIdentifier {
	return ethToElrond.ProposingTransferOnElrond
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *proposeTransferStep) IsInterfaceNil() bool {
	return step == nil
}
