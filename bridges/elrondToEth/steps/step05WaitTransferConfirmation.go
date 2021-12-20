package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/elrondToEth"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
)

type waitTransferConfirmationStep struct {
	bridge bridge.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *waitTransferConfirmationStep) Execute(ctx context.Context) core.StepIdentifier {
	step.bridge.WaitForTransferConfirmation(ctx)
	return elrondToEth.PerformingTransfer
}

// Identifier returns the step's identifier
func (step *waitTransferConfirmationStep) Identifier() core.StepIdentifier {
	return elrondToEth.WaitingTransferConfirmation
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *waitTransferConfirmationStep) IsInterfaceNil() bool {
	return step == nil
}
