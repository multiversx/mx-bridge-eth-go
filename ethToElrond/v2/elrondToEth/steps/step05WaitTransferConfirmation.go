package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/elrondToEth"
)

type waitTransferConfirmationStep struct {
	bridge elrondToEth.ElrondToEthBridge
}

func (step *waitTransferConfirmationStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	return elrondToEth.PerformingTransfer, nil
}

func (step *waitTransferConfirmationStep) Identifier() core.StepIdentifier {
	return elrondToEth.WaitingTransferConfirmation
}

func (step *waitTransferConfirmationStep) IsInterfaceNil() bool {
	return step == nil
}
