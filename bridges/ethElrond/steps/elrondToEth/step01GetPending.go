package elrondToEth

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/ethElrond/steps"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

type getPendingStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *getPendingStep) Execute(ctx context.Context) core.StepIdentifier {
	step.bridge.ResetRetriesCountOnEthereum()
	step.resetCountersOnElrond()

	batch, err := step.bridge.GetBatchFromElrond(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogDebug, "error fetching Elrond batch", "error", err)
		return step.Identifier()
	}
	if batch == nil {
		step.bridge.PrintInfo(logger.LogDebug, "no new batch found on Elrond")
		return step.Identifier()
	}

	err = step.bridge.StoreBatchFromElrond(batch)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error storing Elrond batch", "error", err)
		return step.Identifier()
	}

	isValid, err := step.bridge.ValidateBatch(batch)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error validating Elrond batch", "error", err)
		return step.Identifier()
	}

	if !isValid {
		step.bridge.PrintInfo(logger.LogError, "batch not valid "+batch.String())
		return step.Identifier()
	}

	step.bridge.PrintInfo(logger.LogInfo, "fetched new batch from Elrond "+batch.String())

	wasPerformed, err := step.bridge.WasTransferPerformedOnEthereum(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogError, "error determining if transfer was performed or not", "error", err)
		return step.Identifier()
	}
	if wasPerformed {
		step.bridge.PrintInfo(logger.LogInfo, "transfer performed")
		return ResolvingSetStatusOnElrond
	}

	return SigningProposedTransferOnEthereum
}

// Identifier returns the step's identifier
func (step *getPendingStep) Identifier() core.StepIdentifier {
	return GettingPendingBatchFromElrond
}

// IsInterfaceNil returns true if there is no value under the interface
func (step *getPendingStep) IsInterfaceNil() bool {
	return step == nil
}

func (step *getPendingStep) resetCountersOnElrond() {
	step.bridge.ResetRetriesCountOnElrond()
	step.bridge.ResetRetriesOnWasTransferProposedOnElrond()
}
