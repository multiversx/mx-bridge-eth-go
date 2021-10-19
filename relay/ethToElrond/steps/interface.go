package steps

// BridgeExecutor defines the behavior of the component that handles the operations to be done on the bridge
type BridgeExecutor interface {
	PrintDebugInfo(message string, extras ...interface{})
	GetPendingBatch()
	HasPendingBatch() bool
	IsLeader() bool
	ProposeTransfer() error

	IsInterfaceNil() bool
}
