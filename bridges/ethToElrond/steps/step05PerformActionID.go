package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/ethToElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
)

type performActionIDStep struct {
	bridge bridge.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *performActionIDStep) Execute(ctx context.Context) core.StepIdentifier {
	wasPerformed, err := step.bridge.WasActionPerformedOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error determining if the action ID was proposed or not",
			"action ID", step.bridge.GetStoredActionID(), "error", err)
		return ethToElrond.GettingPendingBatchFromEthereum
	}

	if wasPerformed {
		step.bridge.GetLogger().Info("action ID performed",
			"action ID", step.bridge.GetStoredActionID())
		return ethToElrond.GettingPendingBatchFromEthereum
	}

	if !step.bridge.MyTurnAsLeader() {
		step.bridge.GetLogger().Debug("not my turn as leader in this round")

		return step.Identifier()
	}

	err = step.bridge.PerformActionOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Info("error performing action ID",
			"action ID", step.bridge.GetStoredActionID(), "error", err)
		return ethToElrond.GettingPendingBatchFromEthereum
	}

	return step.Identifier()
}

// Identifier returns the step's identifier
func (step *performActionIDStep) Identifier() core.StepIdentifier {
	return ethToElrond.PerformingActionID
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *performActionIDStep) IsInterfaceNil() bool {
	return step == nil
}
