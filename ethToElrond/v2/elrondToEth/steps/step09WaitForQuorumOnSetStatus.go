package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/elrondToEth"
)

type waitForQuorumOnSetStatus struct {
	bridge elrondToEth.ElrondToEthBridge
}

func (step *waitForQuorumOnSetStatus) Execute(ctx context.Context) (core.StepIdentifier, error) {
	if step.bridge.ProcessMaxRetriesOnElrond() {
		step.bridge.GetLogger().Debug("max number of retries reached, resetting counter")
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	ProcessQuorumReachedOnElrond, err := step.bridge.IsQuorumReachedOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error while checking the quorum", "error", err)
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	step.bridge.GetLogger().Debug("quorum reached check", "is reached", ProcessQuorumReachedOnElrond)

	if !ProcessQuorumReachedOnElrond {
		return step.Identifier(), nil
	}

	return elrondToEth.PerformingSetStatus, nil
}

func (step *waitForQuorumOnSetStatus) Identifier() core.StepIdentifier {
	return elrondToEth.WaitingForQuorumOnSetStatus
}

func (step *waitForQuorumOnSetStatus) IsInterfaceNil() bool {
	return step == nil
}
