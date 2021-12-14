package steps

import (
	"context"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond"
	"github.com/stretchr/testify/assert"
)

func TestExecutesignProposedTransferStep(t *testing.T) {
	t.Parallel()

	t.Run("nil batch", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return nil
		}

		step := signProposedTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(ethToElrond.GettingPendingBatchFromEthereum)
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("error on WasProposedTransferSigned", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return testBatch
		}
		bridgeStub.WasProposedTransferSignedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return false, expectedError
		}

		step := signProposedTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(ethToElrond.GettingPendingBatchFromEthereum)
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("error on SignProposedTransfer", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return testBatch
		}
		bridgeStub.WasProposedTransferSignedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return false, nil
		}
		bridgeStub.SignProposedTransferOnElrondCalled = func(ctx context.Context) error {
			return expectedError
		}

		step := signProposedTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(ethToElrond.GettingPendingBatchFromEthereum)
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("get action ID errors", func(t *testing.T) {
		expectedErr := errors.New("expected error")
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return testBatch
		}
		bridgeStub.WasProposedTransferSignedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return true, nil
		}
		bridgeStub.GetAndStoreActionIDFromElrondCalled = func(ctx context.Context) (uint64, error) {
			return 0, expectedErr
		}

		step := signProposedTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(ethToElrond.GettingPendingBatchFromEthereum)
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("invalid action ID", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return testBatch
		}
		bridgeStub.WasProposedTransferSignedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return true, nil
		}
		bridgeStub.GetAndStoreActionIDFromElrondCalled = func(ctx context.Context) (uint64, error) {
			return invalidActionID, nil
		}

		step := signProposedTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(ethToElrond.GettingPendingBatchFromEthereum)
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("should work - transfer was already signed", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return testBatch
		}
		bridgeStub.WasProposedTransferSignedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return true, nil
		}
		bridgeStub.GetAndStoreActionIDFromElrondCalled = func(ctx context.Context) (uint64, error) {
			return 2, nil
		}

		step := signProposedTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(ethToElrond.WaitingForQuorum)
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("should work", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return testBatch
		}
		bridgeStub.WasProposedTransferSignedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return false, nil
		}
		bridgeStub.SignProposedTransferOnElrondCalled = func(ctx context.Context) error {
			return nil
		}
		bridgeStub.GetAndStoreActionIDFromElrondCalled = func(ctx context.Context) (uint64, error) {
			return 2, nil
		}

		step := signProposedTransferStep{
			bridge: bridgeStub,
		}
		// Test Identifier()
		expectedStepIdentifier := core.StepIdentifier(ethToElrond.SigningProposedTransferOnElrond)
		assert.Equal(t, expectedStepIdentifier, step.Identifier())
		// Test IsInterfaceNil
		assert.NotNil(t, step.IsInterfaceNil())

		expectedStepIdentifier = ethToElrond.WaitingForQuorum
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})
}
