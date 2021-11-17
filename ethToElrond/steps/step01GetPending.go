package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

type getPendingStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (step *getPendingStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	err := step.bridge.GetPendingBatch(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error while getting the batch", "error", err)
		return step.Identifier(), nil
	}

	if step.bridge.HasPendingBatch() {
		step.bridge.PrintInfo(logger.LogDebug, "found pending batch")

		return ethToElrond.ProposingTransfer, nil
	}

	// remain in this step
	return step.Identifier(), nil
}

// Identifier returns the step's identifier
func (step *getPendingStep) Identifier() core.StepIdentifier {
	return ethToElrond.GettingPending
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *getPendingStep) IsInterfaceNil() bool {
	return step == nil
}
