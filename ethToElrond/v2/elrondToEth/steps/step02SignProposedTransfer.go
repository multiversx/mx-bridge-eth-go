package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/elrondToEth"
)

type signProposedTransferStep struct {
	bridge elrondToEth.ElrondToEthBridge
}

func (step *signProposedTransferStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	batch := step.bridge.GetStoredBatch()
	if batch == nil {
		step.bridge.GetLogger().Debug("no new batch found on Elrond")
		return step.Identifier(), nil
	}

	err := step.bridge.SignTransferOnEthereum(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error signing", "batch ID", batch.ID, "error", err)
		return step.Identifier(), nil
	}

	return elrondToEth.WaitingForQuorumOnTransfer, nil
}

func (step *signProposedTransferStep) Identifier() core.StepIdentifier {
	return elrondToEth.SigningProposedTransferOnEthereum
}

func (step *signProposedTransferStep) IsInterfaceNil() bool {
	return step == nil
}
