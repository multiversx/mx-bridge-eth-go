package ethtomultiversx

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	bridgeTests "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	"github.com/stretchr/testify/assert"
)

var expectedError = errors.New("expected error")
var testBatch = &core.TransferBatch{
	ID: 112233,
	Deposits: []*core.DepositTransfer{
		{
			Nonce:                 0,
			ToBytes:               []byte("to"),
			FromBytes:             []byte("from"),
			SourceTokenBytes:      []byte("source token"),
			DestinationTokenBytes: []byte("destination token"),
			Amount:                big.NewInt(37),
		},
	},
	Statuses: []byte{0},
}

func TestExecuteGetPending(t *testing.T) {
	t.Parallel()

	t.Run("error on GetLastExecutedPeerBatchIDFromMultiversX", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutor()
		bridgeStub.GetLastExecutedPeerBatchIDFromMultiversXCalled = func(ctx context.Context) (uint64, error) {
			return 1122, expectedError
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})
	t.Run("error on GetAndStoreBatchFromPeerChain", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutor()
		bridgeStub.GetLastExecutedPeerBatchIDFromMultiversXCalled = func(ctx context.Context) (uint64, error) {
			return 1122, nil
		}
		bridgeStub.GetAndStoreBatchFromPeerChainCalled = func(ctx context.Context, nonce uint64) error {
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
		bridgeStub.GetLastExecutedPeerBatchIDFromMultiversXCalled = func(ctx context.Context) (uint64, error) {
			return 1122, nil
		}
		bridgeStub.GetAndStoreBatchFromPeerChainCalled = func(ctx context.Context, nonce uint64) error {
			return nil
		}
		bridgeStub.GetStoredBatchCalled = func() *core.TransferBatch {
			return nil
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})
	t.Run("error on VerifyLastDepositNonceExecutedOnPeerBatch", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutor()
		bridgeStub.GetLastExecutedPeerBatchIDFromMultiversXCalled = func(ctx context.Context) (uint64, error) {
			return 1122, nil
		}
		bridgeStub.GetAndStoreBatchFromPeerChainCalled = func(ctx context.Context, nonce uint64) error {
			return nil
		}
		bridgeStub.GetStoredBatchCalled = func() *core.TransferBatch {
			return testBatch
		}
		bridgeStub.VerifyLastDepositNonceExecutedOnPeerBatchCalled = func(ctx context.Context) error {
			return expectedError
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})
	t.Run("error on CheckAvailableTokens", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutor()
		bridgeStub.CheckAvailableTokensCalled = func(ctx context.Context, peerTokens [][]byte, mvxTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error {
			return expectedError
		}
		bridgeStub.GetLastExecutedPeerBatchIDFromMultiversXCalled = func(ctx context.Context) (uint64, error) {
			return 1122, nil
		}
		bridgeStub.GetAndStoreBatchFromPeerChainCalled = func(ctx context.Context, nonce uint64) error {
			return nil
		}
		bridgeStub.GetStoredBatchCalled = func() *core.TransferBatch {
			return testBatch
		}
		bridgeStub.VerifyLastDepositNonceExecutedOnPeerBatchCalled = func(ctx context.Context) error {
			return nil
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
		bridgeStub.GetLastExecutedPeerBatchIDFromMultiversXCalled = func(ctx context.Context) (uint64, error) {
			return 1122, nil
		}
		bridgeStub.GetAndStoreBatchFromPeerChainCalled = func(ctx context.Context, nonce uint64) error {
			return nil
		}
		bridgeStub.GetStoredBatchCalled = func() *core.TransferBatch {
			return testBatch
		}
		bridgeStub.VerifyLastDepositNonceExecutedOnPeerBatchCalled = func(ctx context.Context) error {
			return nil
		}
		checkAvailableTokensCalled := false
		bridgeStub.CheckAvailableTokensCalled = func(ctx context.Context, peerTokens [][]byte, mvxTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error {
			checkAvailableTokensCalled = true
			return nil
		}

		step := getPendingStep{
			bridge: bridgeStub,
		}
		// Test Identifier()
		expectedStepIdentifier := core.StepIdentifier(GettingPendingBatchFromPeerChain)
		assert.Equal(t, expectedStepIdentifier, step.Identifier())
		// Test IsInterfaceNil()
		assert.False(t, step.IsInterfaceNil())

		// Test next step
		expectedStepIdentifier = ProposingTransferOnMultiversX
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
		assert.Equal(t, testBatch, step.bridge.GetStoredBatch())
		assert.True(t, checkAvailableTokensCalled)
	})
}

func createStubExecutor() *bridgeTests.BridgeExecutorStub {
	stub := bridgeTests.NewBridgeExecutorStub()

	return stub
}
