package elrondToEth

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	bridgeTests "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridge"
	"github.com/stretchr/testify/assert"
)

func TestExecute_ResolveSetStatus(t *testing.T) {
	t.Parallel()
	t.Run("nil batch on GetStoredBatch", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorResolveSetStatus()
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return nil
		}
		clearWasCalled := false
		bridgeStub.ClearStoredP2PSignaturesForEthereumCalled = func() {
			clearWasCalled = true
		}

		step := resolveSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
		assert.True(t, clearWasCalled)
	})

	t.Run("error on GetStoredBatch", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorResolveSetStatus()
		bridgeStub.GetBatchFromElrondCalled = func(ctx context.Context) (*clients.TransferBatch, error) {
			return nil, expectedError
		}
		clearWasCalled := false
		bridgeStub.ClearStoredP2PSignaturesForEthereumCalled = func() {
			clearWasCalled = true
		}

		step := resolveSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
		assert.True(t, clearWasCalled)
	})

	t.Run("nil batch on GetBatchFromElrond", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorResolveSetStatus()
		bridgeStub.GetBatchFromElrondCalled = func(ctx context.Context) (*clients.TransferBatch, error) {
			return nil, nil
		}
		clearWasCalled := false
		bridgeStub.ClearStoredP2PSignaturesForEthereumCalled = func() {
			clearWasCalled = true
		}

		step := resolveSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
		assert.True(t, clearWasCalled)
	})

	t.Run("WaitForFinalBatchStatusesCalled returns nil, should go to GettingPendingBatchFromElrond", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorResolveSetStatus()

		step := resolveSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})
	t.Run("WaitForFinalBatchStatusesCalled should finish with success and go to ProposingSetStatusOnElrond", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorResolveSetStatus()
		bridgeStub.WaitForFinalBatchStatusesCalled = func(ctx context.Context) []byte {
			return []byte{1, 2, 3}
		}

		wasCalled := false
		bridgeStub.ResolveNewDepositsStatusesCalled = func(numDeposits uint64) {
			wasCalled = true
		}
		clearWasCalled := false
		bridgeStub.ClearStoredP2PSignaturesForEthereumCalled = func() {
			clearWasCalled = true
		}

		step := resolveSetStatusStep{
			bridge: bridgeStub,
		}

		assert.False(t, step.IsInterfaceNil())

		expectedStep := core.StepIdentifier(ProposingSetStatusOnElrond)
		stepIdentifier := step.Execute(context.Background())
		assert.True(t, wasCalled)
		assert.NotEqual(t, step.Identifier(), stepIdentifier)
		assert.Equal(t, expectedStep, stepIdentifier)
		assert.True(t, clearWasCalled)
	})
}

func createStubExecutorResolveSetStatus() *bridgeTests.BridgeExecutorStub {
	stub := bridgeTests.NewBridgeExecutorStub()
	stub.GetStoredBatchCalled = func() *clients.TransferBatch {
		return testBatch
	}
	stub.GetBatchFromElrondCalled = func(ctx context.Context) (*clients.TransferBatch, error) {
		return testBatch, nil
	}
	return stub
}
