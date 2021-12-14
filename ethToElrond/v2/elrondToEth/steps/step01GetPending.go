package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/elrondToEth"
)

type getPendingStep struct {
	bridge elrondToEth.ElrondToEthBridge
}

func (step *getPendingStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	step.bridge.ResetRetriesCountOnEthereum()
	step.bridge.ResetRetriesCountOnElrond()

	err := step.bridge.GetAndStoreBatchFromElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error fetching Elrond batch", "error", err)
		return step.Identifier(), nil
	}

	batch := step.bridge.GetStoredBatch()
	if batch == nil {
		step.bridge.GetLogger().Debug("no new batch found on Elrond")
		return step.Identifier(), nil
	}

	step.bridge.GetLogger().Info("fetched new batch from Elrond " + batch.String())

	wasPerformed, err := step.bridge.WasTransferPerformedOnEthereum(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error determining if transfer was performed or not", "error", err)
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}
	if wasPerformed {
		step.bridge.GetLogger().Info("transfer performed")
		return elrondToEth.ProposingSetStatusOnElrond, nil
	}

	return elrondToEth.SigningProposedTransferOnEthereum, nil
}

func (step *getPendingStep) Identifier() core.StepIdentifier {
	return elrondToEth.GettingPendingBatchFromElrond
}

func (step *getPendingStep) IsInterfaceNil() bool {
	return step == nil
}
