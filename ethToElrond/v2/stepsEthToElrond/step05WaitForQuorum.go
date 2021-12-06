package stepsEthToElrond

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
)

type waitForQuorumStep struct {
	bridge EthToElrondBridge
}

// Execute will execute this step returning the next step to be executed
func (step *waitForQuorumStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	isQuorumReached, err := step.bridge.IsQuorumReached(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error while checking the quorum", "error", err)
		return GetPendingBatchFromEthereum, nil
	}

	step.bridge.GetLogger().Debug("quorum reached check", "is reached", isQuorumReached)

	if !isQuorumReached {
		return step.Identifier(), nil
	}

	return PerformActionID, nil
}

// Identifier returns the step's identifier
func (step *waitForQuorumStep) Identifier() core.StepIdentifier {
	return WaitForQuorum
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *waitForQuorumStep) IsInterfaceNil() bool {
	return step == nil
}
