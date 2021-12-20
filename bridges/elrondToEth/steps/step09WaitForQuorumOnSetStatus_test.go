package steps

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/elrondToEth"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	bridgeTests "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridge"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"
)

func TestExecute_WaitForQuorumOnSetStatus(t *testing.T) {
	t.Parallel()

	t.Run("error on ProcessQuorumReachedOnElrond", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorWaitForQuorumOnSetStatus()
		bridgeStub.ProcessQuorumReachedOnElrondCalled = func(ctx context.Context) (bool, error) {
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
		bridgeStub.ProcessMaxRetriesOnElrondCalled = func() bool {
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
		bridgeStub.ProcessQuorumReachedOnElrondCalled = func(ctx context.Context) (bool, error) {
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
		bridgeStub.ProcessQuorumReachedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return true, nil
		}

		step := waitForQuorumOnSetStatusStep{
			bridge: bridgeStub,
		}

		assert.False(t, step.IsInterfaceNil())

		expectedStepIdentifier := core.StepIdentifier(elrondToEth.PerformingSetStatus)
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})
}

func createStubExecutorWaitForQuorumOnSetStatus() *bridgeTests.BridgeExecutorStub {
	stub := bridgeTests.NewBridgeExecutorStub()
	stub.GetLoggerCalled = func() logger.Logger {
		return testLogger
	}
	stub.ProcessMaxRetriesOnElrondCalled = func() bool {
		return false
	}
	return stub
}
