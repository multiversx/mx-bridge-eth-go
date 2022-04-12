package ethToElrond

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

func TestExecuteGetPending(t *testing.T) {
	t.Parallel()

	t.Run("error on GetLastExecutedEthBatchIDFromElrond", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutor()
		bridgeStub.GetLastExecutedEthBatchIDFromElrondCalled = func(ctx context.Context) (uint64, error) {
			return 1122, expectedError
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("error on GetAndStoreBatchFromEthereum", func(t *testing.T) {
		t.Parallel()
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
		stepIdentifier := step.Execute(context.Background())
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
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("error on ValidateBatch", func(t *testing.T) {
		t.Parallel()
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

	t.Run("error on VerifyLastDepositNonceExecutedOnEthereumBatch", func(t *testing.T) {
		t.Parallel()
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
		bridgeStub.ValidateBatchCalled = func(batch *clients.TransferBatch) (bool, error) {
			return true, nil
		}
		bridgeStub.VerifyLastDepositNonceExecutedOnEthereumBatchCalled = func(ctx context.Context) error {
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
		bridgeStub.ValidateBatchCalled = func(batch *clients.TransferBatch) (bool, error) {
			return true, nil
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}
		// Test Identifier()
		expectedStepIdentifier := core.StepIdentifier(GettingPendingBatchFromEthereum)
		assert.Equal(t, expectedStepIdentifier, step.Identifier())
		// Test IsInterfaceNil()
		assert.False(t, step.IsInterfaceNil())

		// Test next step
		expectedStepIdentifier = ProposingTransferOnElrond
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
		assert.Equal(t, testBatch, step.bridge.GetStoredBatch())
	})
}

func createStubExecutor() *bridgeTests.BridgeExecutorStub {
	stub := bridgeTests.NewBridgeExecutorStub()

	return stub
}
