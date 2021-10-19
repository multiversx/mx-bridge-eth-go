package steps

import "github.com/ElrondNetwork/elrond-eth-bridge/relay"

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
	GetPendingBatch()
	ProposeTransferOnDestination() error
	ProposeSetStatusOnSource()
	CleanTopology()
	ExecuteTransferOnDestination()
	ExecuteSetStatusOnSource()
	SetStatusRejectedOnAllTransactions()
	SetStatusExecutedOnAllTransactions()
	SignProposeTransferOnDestination()
	SignProposeSetStatusOnDestination()
	WaitStepToFinish(step relay.StepIdentifier)

	IsInterfaceNil() bool
}
