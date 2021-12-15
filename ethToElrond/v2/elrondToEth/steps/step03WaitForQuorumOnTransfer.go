package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/elrondToEth"
)

type waitForQuorumOnTransferStep struct {
	bridge elrondToEth.ElrondToEthBridge
}

// Execute will execute this step returning the next step to be executed
func (step *waitForQuorumOnTransferStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	if step.bridge.ProcessMaxRetriesOnEthereum() {
		step.bridge.GetLogger().Debug("max number of retries reached, resetting counter")
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	isQuorumReached, err := step.bridge.IsQuorumReachedOnEthereum(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error while checking the quorum on Ethereum", "error", err)
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	step.bridge.GetLogger().Debug("quorum reached check", "is reached", isQuorumReached)

	if !isQuorumReached {
		return step.Identifier(), nil
	}

	return elrondToEth.PerformingTransfer, nil
}

// Identifier returns the step's identifier
func (step *waitForQuorumOnTransferStep) Identifier() core.StepIdentifier {
	return elrondToEth.WaitingForQuorumOnTransfer
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *waitForQuorumOnTransferStep) IsInterfaceNil() bool {
	return step == nil
}
