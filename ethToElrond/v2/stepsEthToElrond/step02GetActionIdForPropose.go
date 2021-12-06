package stepsEthToElrond

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
)

type getActionIdForProposeStep struct {
	bridge EthToElrondBridge
}

// Execute will execute this step returning the next step to be executed
func (step *getActionIdForProposeStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	batch := step.bridge.GetStoredBatch()

	actionID, err := step.bridge.GetAndStoreActionID(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error fetching action ID", "batch ID", batch.ID, "error", err)
		return GetPendingBatchFromEthereum, nil
	}

	step.bridge.GetLogger().Info("fetched action ID", "action ID", actionID, "batch ID", batch.ID)

	return ProposeTransferOnElrond, nil
}

// Identifier returns the step's identifier
func (step *getActionIdForProposeStep) Identifier() core.StepIdentifier {
	return GetActionIdForProposeStep
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *getActionIdForProposeStep) IsInterfaceNil() bool {
	return step == nil
}
