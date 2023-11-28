package ethtomultiversx

import (
	"context"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExecuteProposeSCTransfer(t *testing.T) {
	t.Parallel()

	t.Run("nil batch", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutor()
		bridgeStub.GetSCExecStoredBatchCalled = func() *clients.TransferBatch {
			return nil
		}

		step := proposeSCTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(GettingPendingBatchFromEthereum)
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("no sc calls in this batch", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutor()
		bridgeStub.GetSCExecStoredBatchCalled = func() *clients.TransferBatch {
			return &clients.TransferBatch{}
		}

		step := proposeSCTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(SigningProposedTransferOnMultiversX)
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("error on WasTransferProposedOnMultiversX", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutor()
		bridgeStub.GetSCExecStoredBatchCalled = func() *clients.TransferBatch {
			return testBatch
		}
		bridgeStub.WasSCTransferProposedOnMultiversXCalled = func(ctx context.Context) (bool, error) {
			return false, expectedError
		}

		step := proposeSCTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(GettingPendingBatchFromEthereum)
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("not leader", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutor()
		bridgeStub.GetSCExecStoredBatchCalled = func() *clients.TransferBatch {
			return testBatch
		}
		bridgeStub.WasSCTransferProposedOnMultiversXCalled = func(ctx context.Context) (bool, error) {
			return false, nil
		}
		bridgeStub.MyTurnAsLeaderCalled = func() bool {
			return false
		}

		step := proposeSCTransferStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})
}
