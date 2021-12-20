package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/elrondToEth"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
)

type waitForQuorumOnTransferStep struct {
	bridge bridge.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *waitForQuorumOnTransferStep) Execute(ctx context.Context) core.StepIdentifier {
	if step.bridge.ProcessMaxRetriesOnEthereum() {
		step.bridge.GetLogger().Debug("max number of retries reached, resetting counter")
		return elrondToEth.GettingPendingBatchFromElrond
	}

	isQuorumReached, err := step.bridge.ProcessQuorumReachedOnEthereum(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error while checking the quorum on Ethereum", "error", err)
		return elrondToEth.GettingPendingBatchFromElrond
	}

	step.bridge.GetLogger().Debug("quorum reached check", "is reached", isQuorumReached)

	if !isQuorumReached {
		return step.Identifier()
	}

	return elrondToEth.PerformingTransfer
}

// Identifier returns the step's identifier
func (step *waitForQuorumOnTransferStep) Identifier() core.StepIdentifier {
	return elrondToEth.WaitingForQuorumOnTransfer
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *waitForQuorumOnTransferStep) IsInterfaceNil() bool {
	return step == nil
}
