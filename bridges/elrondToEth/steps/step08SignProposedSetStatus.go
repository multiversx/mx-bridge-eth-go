package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/elrondToEth"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
)

type signProposedSetStatusStep struct {
	bridge bridge.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *signProposedSetStatusStep) Execute(ctx context.Context) core.StepIdentifier {
	storedBatch := step.bridge.GetStoredBatch()
	if storedBatch == nil {
		step.bridge.GetLogger().Debug("nil stored batch")
		return elrondToEth.GettingPendingBatchFromElrond
	}

	actionID, err := step.bridge.GetAndStoreActionIDForProposeSetStatusFromElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error fetching action ID", "batch ID", storedBatch.ID, "error", err)
		return elrondToEth.GettingPendingBatchFromElrond
	}
	if actionID == bridges.InvalidActionID {
		step.bridge.GetLogger().Error("contract error, got invalid action ID",
			"batch ID", storedBatch.ID, "error", err, "action ID", actionID)
		return elrondToEth.GettingPendingBatchFromElrond
	}

	step.bridge.GetLogger().Info("fetched action ID", "action ID", actionID, "batch ID", storedBatch.ID)

	wasSigned, err := step.bridge.WasActionSignedOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error determining if the proposed transfer was signed or not",
			"batch ID", storedBatch.ID, "error", err)
		return elrondToEth.GettingPendingBatchFromElrond
	}

	if wasSigned {
		return elrondToEth.WaitingForQuorumOnSetStatus
	}

	err = step.bridge.SignActionOnElrond(ctx)
	if err != nil {
		step.bridge.GetLogger().Error("error signing the proposed transfer",
			"batch ID", storedBatch.ID, "error", err)
		return elrondToEth.GettingPendingBatchFromElrond
	}

	return elrondToEth.WaitingForQuorumOnSetStatus
}

// Identifier returns the step's identifier
func (step *signProposedSetStatusStep) Identifier() core.StepIdentifier {
	return elrondToEth.SigningProposedSetStatusOnElrond
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *signProposedSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
