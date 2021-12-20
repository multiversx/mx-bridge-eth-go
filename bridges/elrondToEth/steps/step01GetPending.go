package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/elrondToEth"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
)

type getPendingStep struct {
	bridge bridge.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *getPendingStep) Execute(ctx context.Context) core.StepIdentifier {
	step.bridge.ResetRetriesCountOnEthereum()
	step.bridge.ResetRetriesCountOnElrond()

	batch, err := step.bridge.GetBatchFromElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error fetching Elrond batch", "error", err)
		return step.Identifier()
	}
	if batch == nil {
		step.bridge.GetLogger().Debug("no new batch found on Elrond")
		return step.Identifier()
	}

	err = step.bridge.StoreBatchFromElrond(batch)
	if err != nil {
		step.bridge.GetLogger().Error("error storing Elrond batch", "error", err)
		return step.Identifier()
	}

	step.bridge.GetLogger().Info("fetched new batch from Elrond " + batch.String())

	wasPerformed, err := step.bridge.WasTransferPerformedOnEthereum(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error determining if transfer was performed or not", "error", err)
		return step.Identifier()
	}
	if wasPerformed {
		step.bridge.GetLogger().Info("transfer performed")
		return elrondToEth.ResolvingSetStatusOnElrond
	}

	return elrondToEth.SigningProposedTransferOnEthereum
}

// Identifier returns the step's identifier
func (step *getPendingStep) Identifier() core.StepIdentifier {
	return elrondToEth.GettingPendingBatchFromElrond
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *getPendingStep) IsInterfaceNil() bool {
	return step == nil
}
