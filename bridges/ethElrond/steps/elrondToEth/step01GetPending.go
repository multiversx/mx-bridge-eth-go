package elrondToEth

import (
	"context"
	"encoding/json"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/ethElrond/steps"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

type getPendingStep struct {
	bridge steps.Executor
}

// Execute will execute this step returning the next step to be executed
func (step *getPendingStep) Execute(ctx context.Context) core.StepIdentifier {
	err := step.bridge.CheckElrondClientAvailability(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogDebug, "elrond client unavailable", "message", err)
	}
	err = step.bridge.CheckEthereumClientAvailability(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogDebug, "ethereum client unavailable", "message", err)
	}
	step.bridge.ResetRetriesCountOnEthereum()
	step.resetCountersOnElrond()

	batch, err := step.bridge.GetBatchFromElrond(ctx)
	if err != nil {
		step.bridge.PrintInfo(logger.LogDebug, "cannot fetching Elrond batch", "message", err)
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

	isValid, err := step.bridge.ValidateBatch(ctx, batch)
	if err != nil {
		body, _ := json.Marshal(batch)
		step.bridge.PrintInfo(logger.LogError, "error validating Elrond batch", "error", err, "batch", string(body))
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
