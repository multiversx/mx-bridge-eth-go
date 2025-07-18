package multiversxtoeth

import (
	"context"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/core"
	bridgeTests "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	"github.com/stretchr/testify/assert"
)

func TestExecute_WaitForQuorumOnSetStatus(t *testing.T) {
	t.Parallel()

	t.Run("error on ProcessQuorumReachedOnMultiversX", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorWaitForQuorumOnSetStatus()
		bridgeStub.ProcessQuorumReachedOnMultiversXCalled = func(ctx context.Context) (bool, error) {
			return false, expectedError
		}

		step := waitForQuorumOnSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("max retries reached", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorWaitForQuorumOnSetStatus()
		bridgeStub.ProcessMaxQuorumRetriesOnMultiversXCalled = func() bool {
			return true
		}

		step := waitForQuorumOnSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("quorum not reached", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorWaitForQuorumOnSetStatus()
		bridgeStub.ProcessQuorumReachedOnMultiversXCalled = func(ctx context.Context) (bool, error) {
			return false, nil
		}

		step := waitForQuorumOnSetStatusStep{
			bridge: bridgeStub,
		}

		assert.False(t, step.IsInterfaceNil())

		expectedStepIdentifier := step.Identifier()
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})

	t.Run("quorum reached", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorWaitForQuorumOnSetStatus()
		bridgeStub.ProcessQuorumReachedOnMultiversXCalled = func(ctx context.Context) (bool, error) {
			return true, nil
		}

		step := waitForQuorumOnSetStatusStep{
			bridge: bridgeStub,
		}

		assert.False(t, step.IsInterfaceNil())

		expectedStepIdentifier := core.StepIdentifier(PerformingSetStatus)
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})
}

func createStubExecutorWaitForQuorumOnSetStatus() *bridgeTests.BridgeExecutorStub {
	stub := bridgeTests.NewBridgeExecutorStub()
	stub.ProcessMaxQuorumRetriesOnMultiversXCalled = func() bool {
		return false
	}
	return stub
}
