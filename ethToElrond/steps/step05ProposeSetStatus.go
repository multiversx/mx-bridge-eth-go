package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

type proposeSetStatusStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (step *proposeSetStatusStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	stillHavePending, err := step.bridge.IsPendingBatchReady(ctx)
	step.bridge.PrintInfo(logger.LogDebug, "checked pending batch ready",
		"stillHavePending", stillHavePending, "error", err)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error while checking for pending batch",
			"error", err)

		return step.Identifier(), nil
	}
	if !stillHavePending {
		step.bridge.PrintInfo(logger.LogDebug, "invalid state, going to get pending batch state")

		return ethToElrond.GettingPending, nil
	}

	err = step.bridge.UpdateTransactionsStatusesIfNeeded(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogDebug, "proposeSetStatus.Execute UpdateTransactionsStatusesIfNeeded", "error", err)

		return step.Identifier(), nil
	}

	if step.bridge.IsLeader() {
		step.bridge.PrintInfo(logger.LogDebug, "propose set status (my turn)")
		step.bridge.ProposeSetStatusOnSource(ctx)
	}

	step.bridge.PrintInfo(logger.LogDebug, "waiting propose set status step to finish")
	err = step.bridge.WaitStepToFinish(step.Identifier(), ctx)
	if err != nil {
		return step.Identifier(), err
	}

	if !step.bridge.WasProposeSetStatusExecutedOnSource(ctx) {
		step.bridge.PrintInfo(logger.LogDebug, "was not proposed set status executed on source")
		// remain in this step
		return step.Identifier(), nil
	}

	step.bridge.SignProposeSetStatusOnSource(ctx)
	step.bridge.PrintInfo(logger.LogDebug, "signed propose set status on source")

	return ethToElrond.WaitingSignaturesForProposeSetStatus, nil
}

// Identifier returns the step's identifier
func (step *proposeSetStatusStep) Identifier() core.StepIdentifier {
	return ethToElrond.ProposingSetStatus
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *proposeSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
