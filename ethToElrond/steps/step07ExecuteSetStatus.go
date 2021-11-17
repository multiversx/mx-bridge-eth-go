package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

type executeSetStatusStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (step *executeSetStatusStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	if step.bridge.IsLeader() {
		step.bridge.PrintInfo(logger.LogDebug, "executing set status (my turn)")
		step.bridge.ExecuteSetStatusOnSource(ctx)
	}

	step.bridge.PrintInfo(logger.LogDebug, "waiting set status step to finish")
	err := step.bridge.WaitStepToFinish(step.Identifier(), ctx)
	if err != nil {
		return step.Identifier(), err
	}

	if step.bridge.WasSetStatusExecutedOnSource(ctx) {
		step.bridge.PrintInfo(logger.LogDebug, "set status was executed on source")
		step.bridge.CleanStoredSignatures()

		return ethToElrond.GettingPending, nil
	}

	// remain in this step
	return step.Identifier(), nil
}

// Identifier returns the step's identifier
func (step *executeSetStatusStep) Identifier() core.StepIdentifier {
	return ethToElrond.ExecutingSetStatus
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *executeSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
