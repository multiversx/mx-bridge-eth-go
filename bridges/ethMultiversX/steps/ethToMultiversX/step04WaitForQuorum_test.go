package ethtomultiversx

import (
	"context"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/stretchr/testify/assert"
)

func TestExecuteWaitForQuorumStep(t *testing.T) {
	t.Parallel()

	t.Run("error on IsQuorumReached", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutor()
		bridgeStub.ProcessQuorumReachedOnMultiversXCalled = func(ctx context.Context) (bool, error) {
			return false, expectedError
		}

		step := waitForQuorumStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(GettingPendingBatchFromEthereum)
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("should work - quorum not reached", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutor()
		bridgeStub.ProcessQuorumReachedOnMultiversXCalled = func(ctx context.Context) (bool, error) {
			return false, nil
		}

		step := waitForQuorumStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(WaitingForQuorum)
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutor()
		bridgeStub.ProcessQuorumReachedOnMultiversXCalled = func(ctx context.Context) (bool, error) {
			return true, nil
		}

		step := waitForQuorumStep{
			bridge: bridgeStub,
		}
		// Test Identifier()
		expectedStepIdentifier := core.StepIdentifier(WaitingForQuorum)
		assert.Equal(t, expectedStepIdentifier, step.Identifier())
		// Test IsInterfaceNil
		assert.NotNil(t, step.IsInterfaceNil())

		expectedStepIdentifier = PerformingActionID
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("max retries reached", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutor()
		bridgeStub.ProcessMaxQuorumRetriesOnMultiversXCalled = func() bool {
			return true
		}

		step := waitForQuorumStep{
			bridge: bridgeStub,
		}

		expectedStepIdentifier := core.StepIdentifier(GettingPendingBatchFromEthereum)
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})
}
