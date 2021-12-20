package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/elrondToEth"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
)

type proposeSetStatusStep struct {
	bridge bridge.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *proposeSetStatusStep) Execute(ctx context.Context) core.StepIdentifier {
	batch := step.bridge.GetStoredBatch()
	if batch == nil {
		step.bridge.GetLogger().Debug("nil batch stored")
		return elrondToEth.GettingPendingBatchFromElrond
	}

	wasSetStatusProposed, err := step.bridge.WasSetStatusProposedOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error determining if the set status action was proposed or not on Elrond",
			"batch ID", batch.ID, "error", err)
		return elrondToEth.GettingPendingBatchFromElrond
	}

	if wasSetStatusProposed {
		return elrondToEth.SigningProposedSetStatusOnElrond
	}

	if !step.bridge.MyTurnAsLeader() {
		return step.Identifier()
	}

	err = step.bridge.ProposeSetStatusOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error proposing transfer on Elrond",
			"batch ID", batch.ID, "error", err)
		return elrondToEth.GettingPendingBatchFromElrond
	}

	return elrondToEth.SigningProposedSetStatusOnElrond
}

// Identifier returns the step's identifier
func (step *proposeSetStatusStep) Identifier() core.StepIdentifier {
	return elrondToEth.ProposingSetStatusOnElrond
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *proposeSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
