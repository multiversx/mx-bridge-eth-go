package ethElrond

import (
	"context"
	"time"
)

type testBridgeExecutor struct {
	*bridgeExecutor
	wasTransferPerformedOnEthereumHandle func(ctx context.Context) (bool, error)
}

// NewBridgeExecutor creates a test bridge executor
func NewTestBridgeExecutor(args ArgsBridgeExecutor) (*testBridgeExecutor, error) {
	executor, err := NewBridgeExecutor(args)
	if err != nil {
		return nil, err
	}

	return &testBridgeExecutor{
		bridgeExecutor: executor,
	}, nil
}

// WaitForTransferConfirmation -
func (executor *testBridgeExecutor) WaitForTransferConfirmation(ctx context.Context) {
	wasPerformed := false
	for i := 0; i < splits && !wasPerformed; i++ {
		timer := time.NewTimer(executor.timeForTransferExecution / splits)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			executor.log.Debug("closing due to context expiration")
			return
		case <-timer.C:
		}

		wasPerformed, _ = executor.WasTransferPerformedOnEthereum(ctx)
	}
}

// WasTransferPerformedOnEthereum -
func (executor *testBridgeExecutor) WasTransferPerformedOnEthereum(ctx context.Context) (bool, error) {
	if executor.wasTransferPerformedOnEthereumHandle != nil {
		return executor.wasTransferPerformedOnEthereumHandle(ctx)
	}
	return executor.bridgeExecutor.WasTransferPerformedOnEthereum(ctx)
}

// SetTimeSinceHandler -
func (executor *testBridgeExecutor) SetWasTransferPerformedOnEthereumHandle(handler func(ctx context.Context) (bool, error)) {
	executor.wasTransferPerformedOnEthereumHandle = handler
}

// IsInterfaceNil returns true if there is no value under the interface
func (executor *testBridgeExecutor) IsInterfaceNil() bool {
	return executor == nil
}
