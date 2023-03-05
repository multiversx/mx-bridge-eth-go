package multiversxtoeth

import (
	"context"

	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX/steps"
	"github.com/multiversx/mx-bridge-eth-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type proposeSetStatusStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *proposeSetStatusStep) Execute(ctx context.Context) core.StepIdentifier {
	batch := step.bridge.GetStoredBatch()
	if batch == nil {
		step.bridge.PrintInfo(logger.LogDebug, "nil batch stored")
		return GettingPendingBatchFromMultiversX
	}

	if step.bridge.ProcessMaxRetriesOnWasTransferProposedOnMultiversX() {
		step.bridge.PrintInfo(logger.LogDebug, "max number of retries reached, resetting counter")
		return GettingPendingBatchFromMultiversX
	}

	wasSetStatusProposed, err := step.bridge.WasSetStatusProposedOnMultiversX(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error determining if the set status action was proposed or not on MultiversX",
			"batch ID", batch.ID, "error", err)
		return GettingPendingBatchFromMultiversX
	}

	if wasSetStatusProposed {
		return SigningProposedSetStatusOnMultiversX
	}

	if !step.bridge.MyTurnAsLeader() {
		step.bridge.PrintInfo(logger.LogDebug, "not my turn as leader in this round")
		return step.Identifier()
	}

	err = step.bridge.ProposeSetStatusOnMultiversX(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error proposing transfer on MultiversX",
			"batch ID", batch.ID, "error", err)
		return GettingPendingBatchFromMultiversX
	}

	return SigningProposedSetStatusOnMultiversX
}

// Identifier returns the step's identifier
func (step *proposeSetStatusStep) Identifier() core.StepIdentifier {
	return ProposingSetStatusOnMultiversX
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *proposeSetStatusStep) IsInterfaceNil() bool {
	return step == nil
}
