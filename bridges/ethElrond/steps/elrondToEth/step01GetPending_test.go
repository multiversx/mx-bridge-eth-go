package elrondToEth

import (
	"context"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	bridgeTests "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridge"
	"github.com/stretchr/testify/assert"
)

var expectedError = errors.New("expected error")
var testBatch = &clients.TransferBatch{
	ID:       112233,
	Deposits: nil,
	Statuses: nil,
}

func TestExecute_GetPending(t *testing.T) {
	t.Parallel()

	t.Run("error on GetBatchFromElrond", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorGetPending()
		bridgeStub.GetBatchFromElrondCalled = func(ctx context.Context) (*clients.TransferBatch, error) {
			return nil, expectedError
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("nil batch on GetBatchFromElrond", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorGetPending()
		bridgeStub.GetBatchFromElrondCalled = func(ctx context.Context) (*clients.TransferBatch, error) {
			return nil, nil
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("error on StoreBatchFromElrond", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorGetPending()
		bridgeStub.StoreBatchFromElrondCalled = func(batch *clients.TransferBatch) error {
			return expectedError
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("error on ValidateBatch", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorGetPending()
		bridgeStub.ValidateBatchCalled = func(batch *clients.TransferBatch) (bool, error) {
			return false, expectedError
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("batch not validated on ValidateBatch", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorGetPending()
		bridgeStub.ValidateBatchCalled = func(batch *clients.TransferBatch) (bool, error) {
			return false, nil
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("error on WasTransferPerformedOnEthereum", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorGetPending()
		bridgeStub.WasTransferPerformedOnEthereumCalled = func(ctx context.Context) (bool, error) {
			return false, expectedError
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		t.Run("if transfer already performed next step should be ResolvingSetStatusOnElrond", func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorGetPending()
			bridgeStub.WasTransferPerformedOnEthereumCalled = func(ctx context.Context) (bool, error) {
				return true, nil
			}

			step := getPendingStep{
				bridge: bridgeStub,
			}

			assert.False(t, step.IsInterfaceNil())

			expectedStepIdentifier := core.StepIdentifier(ResolvingSetStatusOnElrond)
			stepIdentifier := step.Execute(context.Background())
			assert.Equal(t, expectedStepIdentifier, stepIdentifier)
		})
		t.Run("if transfer was not performed next step should be SigningProposedTransferOnEthereum", func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorGetPending()
			bridgeStub.WasTransferPerformedOnEthereumCalled = func(ctx context.Context) (bool, error) {
				return false, nil
			}

			step := getPendingStep{
				bridge: bridgeStub,
			}

			expectedStepIdentifier := core.StepIdentifier(SigningProposedTransferOnEthereum)
			stepIdentifier := step.Execute(context.Background())
			assert.Equal(t, expectedStepIdentifier, stepIdentifier)
		})
	})
}

func createStubExecutorGetPending() *bridgeTests.BridgeExecutorStub {
	stub := bridgeTests.NewBridgeExecutorStub()
	stub.GetBatchFromElrondCalled = func(ctx context.Context) (*clients.TransferBatch, error) {
		return testBatch, nil
	}
	stub.StoreBatchFromElrondCalled = func(batch *clients.TransferBatch) error {
		return nil
	}
	stub.ValidateBatchCalled = func(batch *clients.TransferBatch) (bool, error) {
		return true, nil
	}
	return stub
}
