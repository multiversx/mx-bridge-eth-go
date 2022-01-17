package elrondToEth

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/ethElrond/steps"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

type performTransferStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *performTransferStep) Execute(ctx context.Context) core.StepIdentifier {
	wasPerformed, err := step.bridge.WasTransferPerformedOnEthereum(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error determining if transfer was performed or not", "error", err)
		return GettingPendingBatchFromElrond
	}

	if wasPerformed {
		step.bridge.PrintInfo(logger.LogInfo, "transfer performed")
		return ResolvingSetStatusOnElrond
	}

	if step.bridge.MyTurnAsLeader() {
		err = step.bridge.PerformTransferOnEthereum(ctx)
		if err != nil {
			step.bridge.PrintInfo(logger.LogError, "error performing action ID",
				"action ID", step.bridge.GetStoredActionID(), "error", err)
			return GettingPendingBatchFromElrond
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
