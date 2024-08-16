package multiversxtoeth

import (
	"context"
	"testing"

	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	bridgeTests "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	"github.com/stretchr/testify/assert"
)

func TestExecute_ResolveSetStatus(t *testing.T) {
	t.Parallel()

	t.Run("nil batch on GetStoredBatch", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorResolveSetStatus()
		bridgeStub.GetStoredBatchCalled = func() *bridgeCore.TransferBatch {
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
		bridgeStub.GetBatchFromMultiversXCalled = func(ctx context.Context) (*bridgeCore.TransferBatch, error) {
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
	t.Run("nil batch on GetBatchFromMultiversX", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorResolveSetStatus()
		bridgeStub.GetBatchFromMultiversXCalled = func(ctx context.Context) (*bridgeCore.TransferBatch, error) {
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
	t.Run("WaitAndReturnFinalBatchStatusesCalled returns nil, should go to GettingPendingBatchFromMultiversX", func(t *testing.T) {
		t.Parallel()

		bridgeStub := createStubExecutorResolveSetStatus()

		step := resolveSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})
	t.Run("WaitAndReturnFinalBatchStatusesCalled returns empty slice, should go to GettingPendingBatchFromMultiversX", func(t *testing.T) {
		t.Parallel()

		bridgeStub := createStubExecutorResolveSetStatus()
		bridgeStub.WaitAndReturnFinalBatchStatusesCalled = func(ctx context.Context) []byte {
			return make([]byte, 0)
		}

		step := resolveSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})
	t.Run("WaitAndReturnFinalBatchStatusesCalled should finish with success and go to ProposingSetStatusOnMultiversX", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorResolveSetStatus()
		bridgeStub.WaitAndReturnFinalBatchStatusesCalled = func(ctx context.Context) []byte {
			return []byte{bridgeCore.Executed, bridgeCore.Rejected}
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

		expectedStep := bridgeCore.StepIdentifier(ProposingSetStatusOnMultiversX)
		stepIdentifier := step.Execute(context.Background())
		assert.True(t, wasCalled)
		assert.NotEqual(t, step.Identifier(), stepIdentifier)
		assert.Equal(t, expectedStep, stepIdentifier)
		assert.True(t, clearWasCalled)
	})
}

func createStubExecutorResolveSetStatus() *bridgeTests.BridgeExecutorStub {
	stub := bridgeTests.NewBridgeExecutorStub()
	stub.GetStoredBatchCalled = func() *bridgeCore.TransferBatch {
		return testBatch
	}
	stub.GetBatchFromMultiversXCalled = func(ctx context.Context) (*bridgeCore.TransferBatch, error) {
		return testBatch, nil
	}
	return stub
}
