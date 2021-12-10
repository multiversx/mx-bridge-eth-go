package bridgeExecutors

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

// ArgsEthElrondBridgeExecutor is the DTO used in the NewEthElrondBridgeExecutor constructor function
type ArgsEthElrondBridgeExecutor struct {
	ExecutorName      string
	Logger            logger.Logger
	SourceBridge      bridge.Bridge
	DestinationBridge bridge.Bridge
	TopologyProvider  TopologyProvider
	QuorumProvider    bridge.QuorumProvider
	Timer             core.Timer
	DurationsMap      map[core.StepIdentifier]time.Duration
	StatusHandler     core.StatusHandler
}

// ethElrondBridgeExecutor represents the eth-elrond bridge executor adapter
// this implementation is not concurrent safe. Should be called from a single go routine
type ethElrondBridgeExecutor struct {
	*signaturesHolder
	executorName      string
	logger            logger.Logger
	sourceBridge      bridge.Bridge
	destinationBridge bridge.Bridge
	pendingBatch      *bridge.Batch
	actionID          bridge.ActionId
	topologyProvider  TopologyProvider
	quorumProvider    bridge.QuorumProvider
	timer             core.Timer
	durationsMap      map[core.StepIdentifier]time.Duration
	statusHandler     core.StatusHandler
}

// NewEthElrondBridgeExecutor will return a new instance of the ethElrondBridgeExecutor struct
func NewEthElrondBridgeExecutor(args ArgsEthElrondBridgeExecutor) (*ethElrondBridgeExecutor, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &ethElrondBridgeExecutor{
		signaturesHolder:  newSignatureHolder(),
		executorName:      args.ExecutorName,
		logger:            args.Logger,
		sourceBridge:      args.SourceBridge,
		destinationBridge: args.DestinationBridge,
		topologyProvider:  args.TopologyProvider,
		quorumProvider:    args.QuorumProvider,
		timer:             args.Timer,
		durationsMap:      args.DurationsMap,
		statusHandler:     args.StatusHandler,
	}, nil
}

func checkArgs(args ArgsEthElrondBridgeExecutor) error {
	if check.IfNil(args.Logger) {
		return ErrNilLogger
	}
	if check.IfNil(args.SourceBridge) {
		return fmt.Errorf("%w for the source bridge", ErrNilBridge)
	}
	if check.IfNil(args.DestinationBridge) {
		return fmt.Errorf("%w for the destination bridge", ErrNilBridge)
	}
	if check.IfNil(args.TopologyProvider) {
		return ErrNilTopologyProvider
	}
	if check.IfNil(args.QuorumProvider) {
		return ErrNilQuorumProvider
	}
	if check.IfNil(args.Timer) {
		return ErrNilTimer
	}
	if args.DurationsMap == nil {
		return ErrNilDurationsMap
	}
	if check.IfNil(args.StatusHandler) {
		return ErrNilStatusHandler
	}

	return nil
}

// HasPendingBatch returns true if the pending batch is not nil
func (executor *ethElrondBridgeExecutor) HasPendingBatch() bool {
	return executor.pendingBatch != nil
}

// IsLeader returns true if the current instance is the leader in this round
func (executor *ethElrondBridgeExecutor) IsLeader() bool {
	return executor.topologyProvider.AmITheLeader()
}

// WasProposeTransferExecutedOnDestination returns true if the propose transfer was done on the destination bridge
func (executor *ethElrondBridgeExecutor) WasProposeTransferExecutedOnDestination(ctx context.Context) bool {
	return executor.destinationBridge.WasProposedTransfer(ctx, executor.pendingBatch)
}

// WasProposeSetStatusExecutedOnSource returns true if the propose set status was done on the source bridge
func (executor *ethElrondBridgeExecutor) WasProposeSetStatusExecutedOnSource(ctx context.Context) bool {
	return executor.sourceBridge.WasProposedSetStatus(ctx, executor.pendingBatch)
}

// WasTransferExecutedOnDestination returns true if the action ID was executed on the destination bridge
func (executor *ethElrondBridgeExecutor) WasTransferExecutedOnDestination(ctx context.Context) bool {
	return executor.destinationBridge.WasExecuted(ctx, executor.actionID, executor.pendingBatch.Id)
}

// WasSetStatusExecutedOnSource returns true if the action ID was executed on the source bridge
func (executor *ethElrondBridgeExecutor) WasSetStatusExecutedOnSource(ctx context.Context) bool {
	return executor.sourceBridge.WasExecuted(ctx, executor.actionID, executor.pendingBatch.Id)
}

// IsQuorumReachedForProposeTransfer returns true if the quorum has been reached for the propose transfer operation
func (executor *ethElrondBridgeExecutor) IsQuorumReachedForProposeTransfer(ctx context.Context) bool {
	return executor.isQuorumReachedOnBridge(ctx, executor.destinationBridge)
}

func (executor *ethElrondBridgeExecutor) isQuorumReachedOnBridge(ctx context.Context, bridge bridge.Bridge) bool {
	count := bridge.SignersCount(ctx, executor.pendingBatch, executor.actionID, executor)
	quorum, err := executor.quorumProvider.GetQuorum(ctx)
	if err != nil {
		executor.logger.Error(executor.appendMessageToName(err.Error()))

		return false
	}

	executor.logger.Info(executor.appendMessageToName("got signatures"),
		"got", count, "quorum", quorum)

	return executor.isQuorumReached(quorum, count)
}

func (executor *ethElrondBridgeExecutor) isQuorumReached(quorum uint, count uint) bool {
	return count >= quorum
}

// IsQuorumReachedForProposeSetStatus returns true if the quorum has been reached for the propose set status operation
func (executor *ethElrondBridgeExecutor) IsQuorumReachedForProposeSetStatus(ctx context.Context) bool {
	return executor.isQuorumReachedOnBridge(ctx, executor.sourceBridge)
}

// PrintInfo will print the provided data through the inner logger instance
func (executor *ethElrondBridgeExecutor) PrintInfo(logLevel logger.LogLevel, message string, extras ...interface{}) {
	message = executor.appendMessageToName(message)

	// TODO add a new method in the logger repo to print with a desired level, directly
	switch logLevel {
	case logger.LogTrace:
		executor.logger.Trace(message, extras...)
	case logger.LogDebug:
		executor.logger.Debug(message, extras...)
	case logger.LogInfo:
		executor.logger.Info(message, extras...)
	case logger.LogWarning:
		executor.logger.Warn(message, extras...)
		executor.setExecutionMessageInStatusHandler(logLevel, message, extras...)
	case logger.LogError:
		executor.logger.Error(message, extras...)
		executor.setExecutionMessageInStatusHandler(logLevel, message, extras...)
	case logger.LogNone:
	}
}

func (executor *ethElrondBridgeExecutor) appendMessageToName(message string) string {
	return fmt.Sprintf("%s: %s", executor.executorName, message)
}

func (executor *ethElrondBridgeExecutor) setExecutionMessageInStatusHandler(level logger.LogLevel, message string, extras ...interface{}) {
	msg := fmt.Sprintf("%s: %s", level, message)
	for i := 0; i < len(extras)-1; i++ {
		msg += fmt.Sprintf(" %s = %s", convertObjectToString(extras[i]), convertObjectToString(extras[i+1]))
	}

	executor.statusHandler.SetStringMetric(core.MetricLastError, msg)
}

// GetPendingBatch will fetch the pending batch from the source bridge
func (executor *ethElrondBridgeExecutor) GetPendingBatch(ctx context.Context) error {
	executor.statusHandler.SetStringMetric(core.MetricLastError, "")

	pendingBatch, err := executor.sourceBridge.GetPending(ctx)
	if err != nil {
		return err
	}

	executor.pendingBatch = pendingBatch

	return nil
}

// IsPendingBatchReady returns true if a pending batch is ready
func (executor *ethElrondBridgeExecutor) IsPendingBatchReady(ctx context.Context) (bool, error) {
	pendingBatch, err := executor.sourceBridge.GetPending(ctx)
	if err != nil {
		return false, err
	}

	return pendingBatch != nil, nil
}

// ProposeTransferOnDestination will propose the transfer for the existing pending batch on the destination bridge
func (executor *ethElrondBridgeExecutor) ProposeTransferOnDestination(ctx context.Context) error {
	_, err := executor.destinationBridge.ProposeTransfer(ctx, executor.pendingBatch)

	return err
}

// ProposeSetStatusOnSource will propose the status on the source bridge
func (executor *ethElrondBridgeExecutor) ProposeSetStatusOnSource(ctx context.Context) {
	executor.sourceBridge.ProposeSetStatus(ctx, executor.pendingBatch)
}

// CleanStoredSignatures will clean any stored messages & signatures
func (executor *ethElrondBridgeExecutor) CleanStoredSignatures() {
	executor.logger.Debug("removing stored signatures")
	executor.ClearStoredSignatures()
}

// ExecuteTransferOnDestination will execute the action ID on the destination bridge
func (executor *ethElrondBridgeExecutor) ExecuteTransferOnDestination(ctx context.Context) {
	_, err := executor.destinationBridge.Execute(ctx, executor.actionID, executor.pendingBatch, executor)
	if err != nil {
		executor.logger.Error(executor.appendMessageToName(err.Error()))
	}
}

// ExecuteSetStatusOnSource will execute the action ID on the source bridge
func (executor *ethElrondBridgeExecutor) ExecuteSetStatusOnSource(ctx context.Context) {
	_, err := executor.sourceBridge.Execute(ctx, executor.actionID, executor.pendingBatch, executor)
	if err != nil {
		executor.logger.Error(executor.appendMessageToName(err.Error()))
	}
}

// SetStatusRejectedOnAllTransactions will set the status on all transactions to rejected, providing also the error
func (executor *ethElrondBridgeExecutor) SetStatusRejectedOnAllTransactions(err error) {
	executor.pendingBatch.SetStatusOnAllTransactions(bridge.Rejected, err)
}

// UpdateTransactionsStatusesIfNeeded will update all transactions to the status got from the destination bridge, if
// the transactions statuses are not set to Rejected
func (executor *ethElrondBridgeExecutor) UpdateTransactionsStatusesIfNeeded(ctx context.Context) error {
	if !executor.isStatusesCheckOnDestinationNeeded() {
		return nil
	}

	batchId := executor.pendingBatch.Id
	statuses, err := executor.destinationBridge.GetTransactionsStatuses(ctx, executor.pendingBatch.Id)
	if err != nil {
		return err
	}

	if len(statuses) != len(executor.pendingBatch.Transactions) {
		executor.logger.Warn("pending transaction len mismatch in UpdateTransactionsStatusesIfNeeded",
			"batchID", batchId, "statuses", len(statuses),
			"num pending batch transactions", len(executor.pendingBatch.Transactions))
	}

	for i := 0; i < len(executor.pendingBatch.Transactions) && i < len(statuses); i++ {
		executor.pendingBatch.Statuses[i] = statuses[i]
		executor.updateStatusInStatusHandler(statuses[i])
	}
	for i := len(statuses); i < len(executor.pendingBatch.Statuses); i++ {
		executor.pendingBatch.Statuses[i] = bridge.Rejected
		executor.updateStatusInStatusHandler(bridge.Rejected)
	}
	executor.statusHandler.AddIntMetric(core.MetricNumBatches, 1)

	return nil
}

func (executor *ethElrondBridgeExecutor) updateStatusInStatusHandler(status byte) {
	switch status {
	case bridge.Executed:
		executor.statusHandler.AddIntMetric(core.MetricNumTransactionsSucceeded, 1)
	case bridge.Rejected:
		executor.statusHandler.AddIntMetric(core.MetricNumTransactionsRejected, 1)
	default:
		executor.logger.Warn("error saving transaction metric", "unknown status", status)
	}
}

// isStatusesCheckOnDestinationNeeded will return true if at least one transaction status is different from the Rejected value
func (executor *ethElrondBridgeExecutor) isStatusesCheckOnDestinationNeeded() bool {
	if executor.pendingBatch == nil {
		executor.logger.Error("nil pending batch on ethElrondBridgeExecutor.isStatusesCheckOnDestinationNeeded")
		return false
	}
	// if all statuses are rejected, there was an error, so we do not need to check the statuses on destination
	for _, stat := range executor.pendingBatch.Statuses {
		if stat != bridge.Rejected {
			return true
		}
	}

	return false
}

// SignProposeTransferOnDestination will fetch and sign the action ID for the propose transfer operation
func (executor *ethElrondBridgeExecutor) SignProposeTransferOnDestination(ctx context.Context) {
	executor.logger.Info(executor.appendMessageToName("signing propose transfer"), "batch ID", executor.getBatchID())
	executor.actionID = executor.destinationBridge.GetActionIdForProposeTransfer(ctx, executor.pendingBatch)
	_, err := executor.destinationBridge.Sign(ctx, executor.actionID, executor.pendingBatch)
	if err != nil {
		executor.logger.Error(executor.appendMessageToName(err.Error()))
	}
}

// SignProposeSetStatusOnSource will fetch and sign the batch ID for the set status operation
func (executor *ethElrondBridgeExecutor) SignProposeSetStatusOnSource(ctx context.Context) {
	executor.logger.Info(executor.appendMessageToName("signing set status"), "batch ID", executor.getBatchID())
	executor.actionID = executor.sourceBridge.GetActionIdForSetStatusOnPendingTransfer(ctx, executor.pendingBatch)
	_, err := executor.sourceBridge.Sign(ctx, executor.actionID, executor.pendingBatch)
	if err != nil {
		executor.logger.Error(executor.appendMessageToName(err.Error()))
	}
}

// WaitStepToFinish will wait a predefined time and then will return. Returns the error if the provided context
// signals the `Done` event
func (executor *ethElrondBridgeExecutor) WaitStepToFinish(step core.StepIdentifier, ctx context.Context) error {
	duration, found := executor.durationsMap[step]
	if !found {
		return fmt.Errorf("%w for step %s", ErrDurationForStepNotFound, step)
	}

	executor.logger.Info(executor.appendMessageToName("waiting for transfer proposal"),
		"step", step, "batch ID", executor.getBatchID(), "duration", duration)

	select {
	case <-time.After(duration):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (executor *ethElrondBridgeExecutor) getBatchID() *big.Int {
	if executor.pendingBatch == nil {
		return nil
	}

	return executor.pendingBatch.Id
}

// IsInterfaceNil returns true if there is no value under the interface
func (executor *ethElrondBridgeExecutor) IsInterfaceNil() bool {
	return executor == nil
}
