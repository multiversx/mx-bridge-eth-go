package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/elrondToEth"
)

type performSetStatusStep struct {
	bridge bridge.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *performSetStatusStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	wasPerformed, err := step.bridge.WasActionPerformedOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error determining if the set status was proposed or not",
			"action ID", step.bridge.GetStoredActionID(), "error", err)
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	if wasPerformed {
		step.bridge.GetLogger().Info("action ID performed",
			"action ID", step.bridge.GetStoredActionID())
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	if !step.bridge.MyTurnAsLeader() {
		step.bridge.GetLogger().Debug("not my turn as leader in this round")

		return step.Identifier(), nil
	}

	err = step.bridge.PerformActionOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Info("error performing action ID",
			"action ID", step.bridge.GetStoredActionID(), "error", err)
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	return step.Identifier(), nil
}

// Identifier returns the step's identifier
func (step *performSetStatusStep) Identifier() core.StepIdentifier {
	return elrondToEth.PerformingSetStatus
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *performSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
