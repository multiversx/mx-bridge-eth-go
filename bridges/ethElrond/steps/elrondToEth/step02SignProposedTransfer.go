package elrondToEth

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/ethElrond/steps"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

type signProposedTransferStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *signProposedTransferStep) Execute(_ context.Context) core.StepIdentifier {
	storedBatch := step.bridge.GetStoredBatch()
	if storedBatch == nil {
		step.bridge.PrintInfo(logger.LogDebug, "nil batch stored")
		return GettingPendingBatchFromElrond
	}

	err := step.bridge.SignTransferOnEthereum()
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error signing", "batch ID", storedBatch.ID, "error", err)
		return GettingPendingBatchFromElrond
	}

	return WaitingForQuorumOnTransfer
}

// Identifier returns the step's identifier
func (step *signProposedTransferStep) Identifier() core.StepIdentifier {
	return SigningProposedTransferOnEthereum
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *signProposedTransferStep) IsInterfaceNil() bool {
	return step == nil
}
