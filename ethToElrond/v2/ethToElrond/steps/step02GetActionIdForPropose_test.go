package steps

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond"
	"github.com/stretchr/testify/assert"
)

func TestExecuteGetActionIdForProposeStep(t *testing.T) {
	t.Parallel()

	t.Run("error on GetAndStoreActionID", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetAndStoreActionIDCalled = func(ctx context.Context) (uint64, error) {
			return 1122, expectedError
		}
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return &clients.TransferBatch{
				ID:	   112233,
				Deposits: nil,
				Statuses: nil,
			}
		}

		step := getActionIdForProposeStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(ethToElrond.GetPendingBatchFromEthereum)
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("should work", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.GetAndStoreActionIDCalled = func(ctx context.Context) (uint64, error) {
			return 0, nil
		}
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return &clients.TransferBatch{
				ID:	   112233,
				Deposits: nil,
				Statuses: nil,
			}
		}

		step := getActionIdForProposeStep{
			bridge: bridgeStub,
		}
		// Test Identifier()
		expectedStepIdentifier := core.StepIdentifier(ethToElrond.GetActionIdForProposeStep)
		assert.Equal(t, expectedStepIdentifier, step.Identifier())
		// Test IsInterfaceNil()
		assert.False(t, step.IsInterfaceNil())

		// Test next step
		expectedStepIdentifier = ethToElrond.ProposeTransferOnElrond
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})
}
