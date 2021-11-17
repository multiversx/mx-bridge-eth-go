package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

type proposeTransferStep struct {
	bridge BridgeExecutor
}

// Execute will execute this step returning the next step to be executed
func (step *proposeTransferStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	if step.bridge.IsLeader() {
		step.bridge.PrintInfo(logger.LogDebug, "propose transfer (my turn)")

		err := step.bridge.ProposeTransferOnDestination(ctx)
		if err != nil {
			step.bridge.PrintInfo(logger.LogError, "bridge.ProposeTransfer", "error", err)
			step.bridge.SetStatusRejectedOnAllTransactions(err)

			return ethToElrond.ProposingSetStatus, nil
		}
	}

	step.bridge.PrintInfo(logger.LogDebug, "waiting propose transfer step to finish")
	err := step.bridge.WaitStepToFinish(step.Identifier(), ctx)
	if err != nil {
		return step.Identifier(), err
	}

	if !step.bridge.WasProposeTransferExecutedOnDestination(ctx) {
		step.bridge.PrintInfo(logger.LogDebug, "waiting propose transfer: was not proposed, will remain in this step")
		return step.Identifier(), nil
	}

	step.bridge.SignProposeTransferOnDestination(ctx)
	step.bridge.PrintInfo(logger.LogDebug, "signed propose transfer on destination")

	return ethToElrond.WaitingSignaturesForProposeTransfer, nil
}

// Identifier returns the step's identifier
func (step *proposeTransferStep) Identifier() core.StepIdentifier {
	return ethToElrond.ProposingTransfer
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *proposeTransferStep) IsInterfaceNil() bool {
	return step == nil
}
