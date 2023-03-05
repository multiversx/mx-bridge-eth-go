package multiversxtoeth

import (
	"context"

	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX/steps"
	"github.com/multiversx/mx-bridge-eth-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type signProposedTransferStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *signProposedTransferStep) Execute(_ context.Context) core.StepIdentifier {
	storedBatch := step.bridge.GetStoredBatch()
	if storedBatch == nil {
		step.bridge.PrintInfo(logger.LogDebug, "nil batch stored")
		return GettingPendingBatchFromMultiversX
	}

	err := step.bridge.SignTransferOnEthereum()
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error signing", "batch ID", storedBatch.ID, "error", err)
		return GettingPendingBatchFromMultiversX
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
