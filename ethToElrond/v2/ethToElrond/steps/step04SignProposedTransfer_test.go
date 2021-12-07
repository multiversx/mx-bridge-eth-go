package steps

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond"
	"github.com/stretchr/testify/assert"
)

func TestExecutesignProposedTransferStep(t *testing.T) {
	t.Parallel()

	t.Run("error on WasProposedTransferSigned", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return testBatch
		}
		bridgeStub.WasProposedTransferSignedCalled = func(ctx context.Context) (bool, error) {
			return false, expectedError
		}

		step := signProposedTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(ethToElrond.GetPendingBatchFromEthereum)
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("error on SignProposedTransfer", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return testBatch
		}
		bridgeStub.WasProposedTransferSignedCalled = func(ctx context.Context) (bool, error) {
			return false, nil
		}
		bridgeStub.SignProposedTransferCalled = func(ctx context.Context) error {
			return expectedError
		}

		step := signProposedTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(ethToElrond.GetPendingBatchFromEthereum)
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("should work - transfer was already signed", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return testBatch
		}
		bridgeStub.WasProposedTransferSignedCalled = func(ctx context.Context) (bool, error) {
			return true, nil
		}

		step := signProposedTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(ethToElrond.WaitForQuorum)
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("should work", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return testBatch
		}
		bridgeStub.WasProposedTransferSignedCalled = func(ctx context.Context) (bool, error) {
			return false, nil
		}
		bridgeStub.SignProposedTransferCalled = func(ctx context.Context) error {
			return nil
		}

		step := signProposedTransferStep{
			bridge: bridgeStub,
		}
		// Test Identifier()
		expectedStepIdentifier := core.StepIdentifier(ethToElrond.SignProposedTransferOnElrond)
		assert.Equal(t, expectedStepIdentifier, step.Identifier())
		// Test IsInterfaceNil
		assert.NotNil(t, step.IsInterfaceNil())

		expectedStepIdentifier = ethToElrond.WaitForQuorum
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})
}
