package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

type executeTransferStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (step *executeTransferStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	if step.bridge.IsLeader() {
		step.bridge.PrintInfo(logger.LogDebug, "executing transfer (my turn)")
		step.bridge.ExecuteTransferOnDestination(ctx)
	}

	step.bridge.PrintInfo(logger.LogDebug, "waiting transfer step to finish")
	err := step.bridge.WaitStepToFinish(step.Identifier(), ctx)
	if err != nil {
		return step.Identifier(), err
	}

	if step.bridge.WasTransferExecutedOnDestination(ctx) {
		step.bridge.PrintInfo(logger.LogDebug, "transfer was executed on destination")
		step.bridge.CleanStoredSignatures()

		return ethToElrond.ProposingSetStatus, nil
	}

	// remain in this step
	return step.Identifier(), nil
}

// Identifier returns the step's identifier
func (step *executeTransferStep) Identifier() core.StepIdentifier {
	return ethToElrond.ExecutingTransfer
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *executeTransferStep) IsInterfaceNil() bool {
	return step == nil
}
