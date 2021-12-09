package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond"
)

type getActionIdForProposeStep struct {
	bridge ethToElrond.EthToElrondBridge
}

// Execute will execute this step returning the next step to be executed
func (step *getActionIdForProposeStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	batch := step.bridge.GetStoredBatch()

	if batch == nil {
		step.bridge.GetLogger().Debug("no batch found on eth")
		return ethToElrond.GettingPendingBatchFromEthereum, nil
	}

	actionID, err := step.bridge.GetAndStoreActionIDFromElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error fetching action ID", "batch ID", batch.ID, "error", err)
		return ethToElrond.GettingPendingBatchFromEthereum, nil
	}

	step.bridge.GetLogger().Info("fetched action ID", "action ID", actionID, "batch ID", batch.ID)

	return ethToElrond.ProposingTransferOnElrond, nil
}

// Identifier returns the step's identifier
func (step *getActionIdForProposeStep) Identifier() core.StepIdentifier {
	return ethToElrond.GettingActionIdForProposeTransfer
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *getActionIdForProposeStep) IsInterfaceNil() bool {
	return step == nil
}
