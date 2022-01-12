package elrondToEth

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/ethElrond/steps"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

type proposeSetStatusStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *proposeSetStatusStep) Execute(ctx context.Context) core.StepIdentifier {
	batch := step.bridge.GetStoredBatch()
	if batch == nil {
		step.bridge.PrintInfo(logger.LogDebug, "nil batch stored")
		return GettingPendingBatchFromElrond
	}

	wasSetStatusProposed, err := step.bridge.WasSetStatusProposedOnElrond(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error determining if the set status action was proposed or not on Elrond",
			"batch ID", batch.ID, "error", err)
		return GettingPendingBatchFromElrond
	}

	if wasSetStatusProposed {
		return SigningProposedSetStatusOnElrond
	}

	if !step.bridge.MyTurnAsLeader() {
		return step.Identifier()
	}

	err = step.bridge.ProposeSetStatusOnElrond(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error proposing transfer on Elrond",
			"batch ID", batch.ID, "error", err)
		return GettingPendingBatchFromElrond
	}

	return SigningProposedSetStatusOnElrond
}

// Identifier returns the step's identifier
func (step *proposeSetStatusStep) Identifier() core.StepIdentifier {
	return ProposingSetStatusOnElrond
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *proposeSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
