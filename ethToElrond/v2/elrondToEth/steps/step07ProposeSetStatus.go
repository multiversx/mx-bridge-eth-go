package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/elrondToEth"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond"
)

type proposeSetStatusStep struct {
	bridge elrondToEth.ElrondToEthBridge
}

// Execute will execute this step returning the next step to be executed
func (step *proposeSetStatusStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	batch := step.bridge.GetStoredBatchFromElrond()
	if batch == nil {
		step.bridge.GetLogger().Debug("nil batch stored")
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	wasSetStatusProposed, err := step.bridge.WasSetStatusProposedOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error determining if the set status action was proposed or not on Elrond",
			"batch ID", batch.ID, "error", err)
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	if wasSetStatusProposed {
		return elrondToEth.SigningProposedSetStatusOnElrond, nil
	}

	if !step.bridge.MyTurnAsLeader() {
		return step.Identifier(), nil
	}

	err = step.bridge.ProposeSetStatusOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error proposing transfer on Elrond",
			"batch ID", batch.ID, "error", err)
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	return ethToElrond.SigningProposedTransferOnElrond, nil
}

// Identifier returns the step's identifier
func (step *proposeSetStatusStep) Identifier() core.StepIdentifier {
	return elrondToEth.ProposingSetStatusOnElrond
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *proposeSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
