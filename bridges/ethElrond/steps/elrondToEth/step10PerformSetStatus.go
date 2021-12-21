package elrondToEth

import (
	"context"
	logger "github.com/ElrondNetwork/elrond-go-logger"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/ethElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
)

type performSetStatusStep struct {
	bridge ethElrond.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *performSetStatusStep) Execute(ctx context.Context) core.StepIdentifier {
	wasPerformed, err := step.bridge.WasActionPerformedOnElrond(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error determining if the set status was proposed or not",
			"action ID", step.bridge.GetStoredActionID(), "error", err)
		return GettingPendingBatchFromElrond
	}

	if wasPerformed {
		step.bridge.PrintInfo(logger.LogInfo, "action ID performed",
			"action ID", step.bridge.GetStoredActionID())
		return GettingPendingBatchFromElrond
	}

	if !step.bridge.MyTurnAsLeader() {
		step.bridge.PrintInfo(logger.LogDebug, "not my turn as leader in this round")

		return step.Identifier()
	}

	err = step.bridge.PerformActionOnElrond(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogInfo, "error performing action ID",
			"action ID", step.bridge.GetStoredActionID(), "error", err)
		return GettingPendingBatchFromElrond
	}

	return step.Identifier()
}

// Identifier returns the step's identifier
func (step *performSetStatusStep) Identifier() core.StepIdentifier {
	return PerformingSetStatus
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *performSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
