package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	v2 "github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/elrondToEth"
)

type signProposedSetStatusStep struct {
	bridge elrondToEth.ElrondToEthBridge
}

func (step *signProposedSetStatusStep) Execute(ctx context.Context) (core.StepIdentifier, error) {
	batch := step.bridge.GetStoredBatch()
	if batch == nil {
		step.bridge.GetLogger().Debug("no batch found")
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	actionID, err := step.bridge.GetAndStoreActionIDForSetStatusFromElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error fetching action ID", "batch ID", batch.ID, "error", err)
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}
	if actionID == v2.InvalidActionID {
		step.bridge.GetLogger().Error("contract error, got invalid action ID",
			"batch ID", batch.ID, "error", err, "action ID", actionID)
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	step.bridge.GetLogger().Info("fetched action ID", "action ID", actionID, "batch ID", batch.ID)

	wasSigned, err := step.bridge.WasProposedSetStatusSignedOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error determining if the proposed transfer was signed or not",
			"batch ID", batch.ID, "error", err)
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	if wasSigned {
		return elrondToEth.WaitingForQuorumOnSetStatus, nil
	}

	err = step.bridge.SignProposedSetStatusOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error signing the proposed transfer",
			"batch ID", batch.ID, "error", err)
		return elrondToEth.GettingPendingBatchFromElrond, nil
	}

	return elrondToEth.WaitingForQuorumOnSetStatus, nil
}

func (step *signProposedSetStatusStep) Identifier() core.StepIdentifier {
	return elrondToEth.SigningProposedSetStatusOnElrond
}

func (step *signProposedSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
