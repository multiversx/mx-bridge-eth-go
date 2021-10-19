package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
)

// BridgeExecutor defines the behavior of the component that handles the operations to be done on the bridge
type BridgeExecutor interface {
	HasPendingBatch() bool
	IsLeader() bool
	WasProposeTransferExecutedOnDestination() bool
	WasProposeSetStatusExecutedOnSource() bool
	WasTransferExecutedOnDestination() bool
	WasSetStatusExecutedOnSource() bool
	IsQuorumReachedForProposeTransfer() bool
	IsQuorumReachedForProposeSetStatus() bool

	PrintDebugInfo(message string, extras ...interface{})
	GetPendingBatch(ctx context.Context)
	ProposeTransferOnDestination(ctx context.Context) error
	ProposeSetStatusOnSource(ctx context.Context)
	CleanTopology()
	ExecuteTransferOnDestination(ctx context.Context)
	ExecuteSetStatusOnSource(ctx context.Context)
	SetStatusRejectedOnAllTransactions()
	SetStatusExecutedOnAllTransactions()
	SignProposeTransferOnDestination(ctx context.Context)
	SignProposeSetStatusOnDestination(ctx context.Context)
	WaitStepToFinish(step relay.StepIdentifier, ctx context.Context)

	IsInterfaceNil() bool
}
