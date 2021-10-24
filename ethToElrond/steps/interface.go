package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

// BridgeExecutor defines the behavior of the component that handles the operations to be done on the bridge
type BridgeExecutor interface {
	HasPendingBatch() bool
	IsLeader() bool
	WasProposeTransferExecutedOnDestination(ctx context.Context) bool
	WasProposeSetStatusExecutedOnSource(ctx context.Context) bool
	WasTransferExecutedOnDestination(ctx context.Context) bool
	WasSetStatusExecutedOnSource(ctx context.Context) bool
	IsQuorumReachedForProposeTransfer(ctx context.Context) bool
	IsQuorumReachedForProposeSetStatus(ctx context.Context) bool

	PrintInfo(logLevel logger.LogLevel, message string, extras ...interface{})
	GetPendingBatch(ctx context.Context)
	ProposeTransferOnDestination(ctx context.Context) error
	ProposeSetStatusOnSource(ctx context.Context)
	CleanTopology()
	ExecuteTransferOnDestination(ctx context.Context)
	ExecuteSetStatusOnSource(ctx context.Context)
	SetStatusRejectedOnAllTransactions(err error)
	SetTransactionsStatusesIfNeeded(ctx context.Context) error
	SignProposeTransferOnDestination(ctx context.Context)
	SignProposeSetStatusOnSource(ctx context.Context)
	WaitStepToFinish(step core.StepIdentifier, ctx context.Context) error

	IsInterfaceNil() bool
}
