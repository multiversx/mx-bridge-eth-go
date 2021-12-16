package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/elrondToEth"
)

type getPendingStep struct {
	bridge elrondToEth.ElrondToEthBridge
}

// Execute will execute this step returning the next step to be executed
func (step *getPendingStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	step.bridge.ResetRetriesCountOnEthereum()
	step.bridge.ResetRetriesCountOnElrond()

	batch, err := step.bridge.GetBatchFromElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error fetching Elrond batch", "error", err)
		return step.Identifier(), nil
	}
	if batch == nil {
		step.bridge.GetLogger().Debug("no new batch found on Elrond")
		return step.Identifier(), nil
	}

	err = step.bridge.StoreBatchFromElrond(ctx, batch)
	if err != nil {
		step.bridge.GetLogger().Error("error storing Elrond batch", "error", err)
		return step.Identifier(), nil
	}

	step.bridge.GetLogger().Info("fetched new batch from Elrond " + batch.String())

	wasPerformed, err := step.bridge.WasTransferPerformedOnEthereum(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error determining if transfer was performed or not", "error", err)
		return step.Identifier(), nil
	}
	if wasPerformed {
		step.bridge.GetLogger().Info("transfer performed")
		return elrondToEth.ResolvingSetStatusOnElrond, nil
	}

	return elrondToEth.SigningProposedTransferOnEthereum, nil
}

// Identifier returns the step's identifier
func (step *getPendingStep) Identifier() core.StepIdentifier {
	return elrondToEth.GettingPendingBatchFromElrond
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *getPendingStep) IsInterfaceNil() bool {
	return step == nil
}
