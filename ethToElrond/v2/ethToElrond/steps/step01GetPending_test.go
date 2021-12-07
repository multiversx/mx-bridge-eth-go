package steps

import (
	"context"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"
)

var testLogger = logger.GetOrCreate("test")
var expectedError = errors.New("expected error")
var testBatch = &clients.TransferBatch {
	ID:       112233,
	Deposits: nil,
	Statuses: nil,
}
func TestExecuteGetPending(t *testing.T) {
	t.Parallel()

	t.Run("error on GetLastExecutedEthBatchIDFromElrond", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetLastExecutedEthBatchIDFromElrondCalled = func(ctx context.Context) (uint64, error) {
			return 1122, expectedError
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("error on GetAndStoreBatchFromEthereum", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetLastExecutedEthBatchIDFromElrondCalled = func(ctx context.Context) (uint64, error) {
			return 1122, nil
		}
		bridgeStub.GetAndStoreBatchFromEthereumCalled = func(ctx context.Context, nonce uint64) error {
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

	t.Run("nil on GetStoredBatch", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetLastExecutedEthBatchIDFromElrondCalled = func(ctx context.Context) (uint64, error) {
			return 1122, nil
		}
		bridgeStub.GetAndStoreBatchFromEthereumCalled = func(ctx context.Context, nonce uint64) error {
			return nil
		}
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return nil
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("error on VerifyLastDepositNonceExecutedOnEthereumBatch", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetLastExecutedEthBatchIDFromElrondCalled = func(ctx context.Context) (uint64, error) {
			return 1122, nil
		}
		bridgeStub.GetAndStoreBatchFromEthereumCalled = func(ctx context.Context, nonce uint64) error {
			return nil
		}
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return testBatch
		}
		bridgeStub.VerifyLastDepositNonceExecutedOnEthereumBatchCalled = func(ctx context.Context) error {
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

	t.Run("should work", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetLastExecutedEthBatchIDFromElrondCalled = func(ctx context.Context) (uint64, error) {
			return 1122, nil
		}
		bridgeStub.GetAndStoreBatchFromEthereumCalled = func(ctx context.Context, nonce uint64) error {
			return nil
		}
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return testBatch
		}
		bridgeStub.VerifyLastDepositNonceExecutedOnEthereumBatchCalled = func(ctx context.Context) error {
			return nil
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}
		// Test Identifier()
		expectedStepIdentifier := core.StepIdentifier(ethToElrond.GettingPendingBatchFromEthereum)
		assert.Equal(t, expectedStepIdentifier, step.Identifier())
		// Test IsInterfaceNil()
		assert.False(t, step.IsInterfaceNil())

		// Test next step
		expectedStepIdentifier = ethToElrond.GettingActionIdForProposeTransfer
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
		assert.Equal(t, testBatch, step.bridge.GetStoredBatch())
	})
}

func createStubExecutor() *bridgeV2.EthToElrondBridgeStub {
	stub := bridgeV2.NewEthToElrondBridgeStub()
	stub.GetLoggerCalled = func() logger.Logger {
		return testLogger
	}
	return stub
}
