package bridgeExecutors

import "github.com/ElrondNetwork/elrond-eth-bridge/bridge"

// SetPendingBatch -
func (executor *ethElrondBridgeExecutor) SetPendingBatch(batch *bridge.Batch) {
	executor.pendingBatch = batch
}
