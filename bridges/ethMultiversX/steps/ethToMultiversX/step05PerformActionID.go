package ethtomultiversx

import (
	"context"

	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX/steps"
	"github.com/multiversx/mx-bridge-eth-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type performActionIDStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *performActionIDStep) Execute(ctx context.Context) core.StepIdentifier {
	wasPerformed, err := step.bridge.WasActionPerformedOnMultiversX(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error determining if the action ID was proposed or not",
			"action ID", step.bridge.GetStoredActionID(), "error", err)
		return GettingPendingBatchFromEthereum
	}

	if wasPerformed {
		step.bridge.PrintInfo(logger.LogInfo, "action ID performed",
			"action ID", step.bridge.GetStoredActionID())
		return step.computeNextStep()
	}

	if !step.bridge.MyTurnAsLeader() {
		step.bridge.PrintInfo(logger.LogDebug, "not my turn as leader in this round")
		return step.Identifier()
	}

	err = step.bridge.PerformActionOnMultiversX(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error performing action ID",
			"action ID", step.bridge.GetStoredActionID(), "error", err)
		return GettingPendingBatchFromEthereum
	}

	return step.Identifier()
}

func (step *performActionIDStep) computeNextStep() core.StepIdentifier {
	if step.bridge.GetBatchTypeExecutionStep() == ProposingSCTransfersOnMultiversX {
		return GettingPendingBatchFromEthereum
	}

	return ProposingSCTransfersOnMultiversX
}

// Identifier returns the step's identifier
func (step *performActionIDStep) Identifier() core.StepIdentifier {
	return PerformingActionID
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *performActionIDStep) IsInterfaceNil() bool {
	return step == nil
}
