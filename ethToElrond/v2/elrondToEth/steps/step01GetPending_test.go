package steps

import (
	"context"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/elrondToEth"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"
)

var testLogger = core.NewLoggerWithIdentifier(logger.GetOrCreate("test"), "TEST")
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
		bridgeStub := createStubExecutorGetpending()
		bridgeStub.GetBatchFromElrondCalled = func(ctx context.Context) (*clients.TransferBatch, error) {
			return nil, expectedError
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("nil batch on GetBatchFromElrond", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorGetpending()
		bridgeStub.GetBatchFromElrondCalled = func(ctx context.Context) (*clients.TransferBatch, error) {
			return nil, nil
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("error on StoreBatchFromElrond", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorGetpending()
		bridgeStub.StoreBatchFromElrondCalled = func(ctx context.Context, batch *clients.TransferBatch) error {
			return expectedError
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("error on WasTransferPerformedOnEthereum", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorGetpending()
		bridgeStub.WasTransferPerformedOnEthereumCalled = func(ctx context.Context) (bool, error) {
			return false, expectedError
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		t.Run("if transfer already performed next step should be ResolvingSetStatusOnElrond", func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorGetpending()
			bridgeStub.WasTransferPerformedOnEthereumCalled = func(ctx context.Context) (bool, error) {
				return true, nil
			}

			step := getPendingStep{
				bridge: bridgeStub,
			}

			assert.False(t, step.IsInterfaceNil())

			expectedStepIdentifier := core.StepIdentifier(elrondToEth.ResolvingSetStatusOnElrond)
			stepIdentifier, err := step.Execute(context.Background())
			assert.Nil(t, err)
			assert.Equal(t, expectedStepIdentifier, stepIdentifier)
		})
		t.Run("if transfer was not performed next step should be SigningProposedTransferOnEthereum", func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorGetpending()
			bridgeStub.WasTransferPerformedOnEthereumCalled = func(ctx context.Context) (bool, error) {
				return false, nil
			}

			step := getPendingStep{
				bridge: bridgeStub,
			}

			expectedStepIdentifier := core.StepIdentifier(elrondToEth.SigningProposedTransferOnEthereum)
			stepIdentifier, err := step.Execute(context.Background())
			assert.Nil(t, err)
			assert.Equal(t, expectedStepIdentifier, stepIdentifier)
		})
	})
}

func createStubExecutorGetPending() *bridgeV2.ElrondToEthBridgeStub {
	stub := bridgeV2.NewElrondToEthBridgeStub()
	stub.GetLoggerCalled = func() logger.Logger {
		return testLogger
	}
	stub.GetBatchFromElrondCalled = func(ctx context.Context) (*clients.TransferBatch, error) {
		return testBatch, nil
	}
	stub.StoreBatchFromElrondCalled = func(ctx context.Context, batch *clients.TransferBatch) error {
		return nil
	}
	return stub
}
