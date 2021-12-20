package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/elrondToEth"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
)

type signProposedTransferStep struct {
	bridge bridge.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *signProposedTransferStep) Execute(ctx context.Context) core.StepIdentifier {
	storedBatch := step.bridge.GetStoredBatch()
	if storedBatch == nil {
		step.bridge.GetLogger().Debug("nil batch stored")
		return elrondToEth.GettingPendingBatchFromElrond
	}

	err := step.bridge.SignTransferOnEthereum()
	if err != nil {
		step.bridge.GetLogger().Error("error signing", "batch ID", storedBatch.ID, "error", err)
		return elrondToEth.GettingPendingBatchFromElrond
	}

	return elrondToEth.WaitingForQuorumOnTransfer
}

// Identifier returns the step's identifier
func (step *signProposedTransferStep) Identifier() core.StepIdentifier {
	return elrondToEth.SigningProposedTransferOnEthereum
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *signProposedTransferStep) IsInterfaceNil() bool {
	return step == nil
}
