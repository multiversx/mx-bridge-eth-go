package multiversxtoeth

import (
	"context"
	"github.com/multiversx/mx-bridge-eth-go/bridges/steps"

	"github.com/multiversx/mx-bridge-eth-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type performTransferStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *performTransferStep) Execute(ctx context.Context) core.StepIdentifier {
	wasPerformed, err := step.bridge.WasTransferPerformedOnPeerChain(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error determining if transfer was performed or not", "error", err)
		return GettingPendingBatchFromMultiversX
	}

	if wasPerformed {
		step.bridge.PrintInfo(logger.LogInfo, "transfer performed")
		return ResolvingSetStatusOnMultiversX
	}

	if step.bridge.MyTurnAsLeader() {
		err = step.bridge.PerformTransferOnPeerChain(ctx)
		if err != nil {
			step.bridge.PrintInfo(logger.LogError, "error performing transfer on PeerChain", "error", err)
			return GettingPendingBatchFromMultiversX
		}
	} else {
		step.bridge.PrintInfo(logger.LogDebug, "not my turn as leader in this round")
	}

	return WaitingTransferConfirmation
}

// Identifier returns the step's identifier
func (step *performTransferStep) Identifier() core.StepIdentifier {
	return PerformingTransfer
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *performTransferStep) IsInterfaceNil() bool {
	return step == nil
}
