package steps

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond"
	"github.com/stretchr/testify/assert"
)

func TestExecutePerformActionIDStep(t *testing.T) {
	t.Parallel()

	t.Run("error on WasActionIDPerformed", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.WasActionIDPerformedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return false, expectedError
		}

		step := performActionIDStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(ethToElrond.GettingPendingBatchFromEthereum)
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("should work - actionID already performed", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.WasActionIDPerformedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return true, nil
		}

		step := performActionIDStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(ethToElrond.GettingPendingBatchFromEthereum)
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("should work - not leader", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.WasActionIDPerformedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return false, nil
		}
		bridgeStub.MyTurnAsLeaderCalled = func() bool {
			return false
		}

		step := performActionIDStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := step.Identifier()
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("error on PerformActionID", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.WasActionIDPerformedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return false, nil
		}
		bridgeStub.MyTurnAsLeaderCalled = func() bool {
			return true
		}
		bridgeStub.PerformActionIDOnElrondCalled = func(ctx context.Context) error {
			return expectedError
		}

		step := performActionIDStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(ethToElrond.GettingPendingBatchFromEthereum)
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("should work", func(t *testing.T) {
		bridgeStub := createStubExecutor()
		bridgeStub.WasActionIDPerformedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return false, nil
		}
		bridgeStub.MyTurnAsLeaderCalled = func() bool {
			return true
		}
		bridgeStub.PerformActionIDOnElrondCalled = func(ctx context.Context) error {
			return nil
		}

		step := performActionIDStep{
			bridge: bridgeStub,
		}
		// Test Identifier()
		expectedStepIdentifier := core.StepIdentifier(ethToElrond.PerformingActionID)
		assert.Equal(t, expectedStepIdentifier, step.Identifier())
		// Test IsInterfaceNil
		assert.NotNil(t, step.IsInterfaceNil())

		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})
}
