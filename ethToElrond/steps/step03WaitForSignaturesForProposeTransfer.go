package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

type waitForSignaturesForProposeTransferStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (step *waitForSignaturesForProposeTransferStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	step.bridge.PrintInfo(logger.LogDebug, "waiting for signatures for propose transfer")
	err := step.bridge.WaitStepToFinish(step.Identifier(), ctx)
	if err != nil {
		return step.Identifier(), err
	}

	if step.bridge.IsQuorumReachedForProposeTransfer(ctx) {
		step.bridge.PrintInfo(logger.LogDebug, "quorum reached for propose transfer")
		return ethToElrond.ExecutingTransfer, nil
	}

	if step.bridge.WasProposeTransferExecutedOnDestination(ctx) {
		step.bridge.PrintInfo(logger.LogDebug, "transfer propose was executed on destination")
		step.bridge.CleanStoredSignatures()

		return ethToElrond.ProposingSetStatus, nil
	}

	// remain in this step
	return step.Identifier(), nil
}

// Identifier returns the step's identifier
func (step *waitForSignaturesForProposeTransferStep) Identifier() core.StepIdentifier {
	return ethToElrond.WaitingSignaturesForProposeTransfer
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *waitForSignaturesForProposeTransferStep) IsInterfaceNil() bool {
	return step == nil
}
