package steps

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond"
	"github.com/stretchr/testify/assert"
)

func TestExecuteProposeTransfer(t *testing.T) {
	t.Parallel()

	t.Run("error on WasTransferProposedOnElrond", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return &clients.TransferBatch{
				ID:	   112233,
				Deposits: nil,
				Statuses: nil,
			}
		}
		bridgeStub.WasTransferProposedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return false, expectedError
		}

		step := proposeTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(ethToElrond.GetPendingBatchFromEthereum)
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("not leader", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return &clients.TransferBatch{
				ID:	   112233,
				Deposits: nil,
				Statuses: nil,
			}
		}
		bridgeStub.WasTransferProposedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return false, nil
		}
		bridgeStub.MyTurnAsLeaderCalled = func() bool {
			return false
		}

		step := proposeTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("error on ProposeTransferOnElrond", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return &clients.TransferBatch{
				ID:	   112233,
				Deposits: nil,
				Statuses: nil,
			}
		}
		bridgeStub.WasTransferProposedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return false, nil
		}
		bridgeStub.MyTurnAsLeaderCalled = func() bool {
			return true
		}
		bridgeStub.ProposeTransferOnElrondCalled = func(ctx context.Context) error {
			return expectedError
		}

		step := proposeTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(ethToElrond.GetPendingBatchFromEthereum)
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("should work - transfer already proposed", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return &clients.TransferBatch{
				ID:	   112233,
				Deposits: nil,
				Statuses: nil,
			}
		}
		bridgeStub.WasTransferProposedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return true, nil
		}

		step := proposeTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(ethToElrond.SignProposedTransferOnElrond)
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("should work", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return &clients.TransferBatch{
				ID:	   112233,
				Deposits: nil,
				Statuses: nil,
			}
		}
		bridgeStub.WasTransferProposedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return false, nil
		}
		bridgeStub.MyTurnAsLeaderCalled = func() bool {
			return true
		}
		bridgeStub.ProposeTransferOnElrondCalled = func(ctx context.Context) error {
			return nil
		}

		step := proposeTransferStep{
			bridge: bridgeStub,
		}
		// Test IsInterfaceNil
		assert.NotNil(t, step.IsInterfaceNil())

		expectedStepIdentifier := core.StepIdentifier(ethToElrond.SignProposedTransferOnElrond)
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})
}