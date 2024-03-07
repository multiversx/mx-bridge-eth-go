package multiversxtoeth

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	bridgeTests "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
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

	t.Run("error on GetBatchFromMultiversX", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorGetPending()
		bridgeStub.GetBatchFromMultiversXCalled = func(ctx context.Context) (*clients.TransferBatch, error) {
			return nil, expectedError
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})
	t.Run("nil batch on GetBatchFromMultiversX", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorGetPending()
		bridgeStub.GetBatchFromMultiversXCalled = func(ctx context.Context) (*clients.TransferBatch, error) {
			return nil, nil
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})
	t.Run("error on StoreBatchFromMultiversX", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorGetPending()
		bridgeStub.StoreBatchFromMultiversXCalled = func(batch *clients.TransferBatch) error {
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
		bridgeStub.ValidateBatchCalled = func(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
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
		bridgeStub.ValidateBatchCalled = func(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
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
	t.Run("error on WasTransferPerformedOnEthereum", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorGetPending()
		bridgeStub.WasTransferPerformedOnEthereumCalled = func(ctx context.Context) (bool, error) {
			return true, nil
		}
		bridgeStub.CheckAvailableTokensCalled = func(ctx context.Context, ethTokens []common.Address, mvxTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error {
			return expectedError
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
		t.Run("if transfer already performed next step should be ResolvingSetStatusOnMultiversX", func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorGetPending()
			bridgeStub.WasTransferPerformedOnEthereumCalled = func(ctx context.Context) (bool, error) {
				return true, nil
			}
			checkAvailableTokensCalled := false
			bridgeStub.CheckAvailableTokensCalled = func(ctx context.Context, ethTokens []common.Address, mvxTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error {
				checkAvailableTokensCalled = true
				return nil
			}

			step := getPendingStep{
				bridge: bridgeStub,
			}

			assert.False(t, step.IsInterfaceNil())

			expectedStepIdentifier := core.StepIdentifier(ResolvingSetStatusOnMultiversX)
			stepIdentifier := step.Execute(context.Background())
			assert.Equal(t, expectedStepIdentifier, stepIdentifier)
			assert.True(t, checkAvailableTokensCalled)
		})
		t.Run("if transfer was not performed next step should be SigningProposedTransferOnEthereum", func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorGetPending()
			bridgeStub.WasTransferPerformedOnEthereumCalled = func(ctx context.Context) (bool, error) {
				return false, nil
			}
			checkAvailableTokensCalled := false
			bridgeStub.CheckAvailableTokensCalled = func(ctx context.Context, ethTokens []common.Address, mvxTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error {
				checkAvailableTokensCalled = true
				return nil
			}

			step := getPendingStep{
				bridge: bridgeStub,
			}

			expectedStepIdentifier := core.StepIdentifier(SigningProposedTransferOnEthereum)
			stepIdentifier := step.Execute(context.Background())
			assert.Equal(t, expectedStepIdentifier, stepIdentifier)
			assert.True(t, checkAvailableTokensCalled)
		})
	})
}

func createStubExecutorGetPending() *bridgeTests.BridgeExecutorStub {
	stub := bridgeTests.NewBridgeExecutorStub()
	stub.GetBatchFromMultiversXCalled = func(ctx context.Context) (*clients.TransferBatch, error) {
		return testBatch, nil
	}
	stub.StoreBatchFromMultiversXCalled = func(batch *clients.TransferBatch) error {
		return nil
	}
	stub.ValidateBatchCalled = func(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
		return true, nil
	}
	return stub
}
