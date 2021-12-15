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
	storedBatch := step.bridge.GetStoredBatchFromElrond()
	if storedBatch == nil {
		step.bridge.GetLogger().Debug("nil batch stored")
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	err := step.bridge.SignTransferOnEthereum(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error signing", "batch ID", storedBatch.ID, "error", err)
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	return elrondToEth.WaitingForQuorumOnTransfer, nil
}

func (step *signProposedTransferStep) Identifier() core.StepIdentifier {
	return elrondToEth.SigningProposedTransferOnEthereum
}

func (step *signProposedTransferStep) IsInterfaceNil() bool {
	return step == nil
}
