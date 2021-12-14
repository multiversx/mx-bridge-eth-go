package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/elrondToEth"
)

type performSetStatusStep struct {
	bridge elrondToEth.ElrondToEthBridge
}

func (step *performSetStatusStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	wasPerformed, err := step.bridge.WasSetStatusPerformedOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error determining if the set status was proposed or not",
			"action ID", step.bridge.GetStoredActionIDForSetStatus(), "error", err)
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	if wasPerformed {
		step.bridge.GetLogger().Info("action ID performed",
			"action ID", step.bridge.GetStoredActionIDForSetStatus())
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	if !step.bridge.MyTurnAsLeader() {
		step.bridge.GetLogger().Debug("not my turn as leader in this round")

		return step.Identifier(), nil
	}

	err = step.bridge.PerformSetStatusOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Info("error performing action ID",
			"action ID", step.bridge.GetStoredActionIDForSetStatus(), "error", err)
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	return step.Identifier(), nil
}

func (step *performSetStatusStep) Identifier() core.StepIdentifier {
	return elrondToEth.PerformingSetStatus
}

func (step *performSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
