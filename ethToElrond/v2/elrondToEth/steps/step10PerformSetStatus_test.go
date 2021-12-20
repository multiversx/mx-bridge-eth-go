package steps

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"
)

func TestExecute_PerformSetStatus(t *testing.T) {
	t.Parallel()

	t.Run("error on WasActionPerformedOnElrondCalled", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorPerformSetStatus()
		bridgeStub.WasActionPerformedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return false, expectedError
		}

		step := performSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("error on PerformActionOnElrondCalled", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorPerformSetStatus()
		bridgeStub.MyTurnAsLeaderCalled = func() bool {
			return true
		}
		bridgeStub.PerformActionOnElrondCalled = func(ctx context.Context) error {
			return expectedError
		}

		step := performSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		t.Run("if transfer was performed we should go to initial step", func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorPerformSetStatus()
			bridgeStub.WasActionPerformedOnElrondCalled = func(ctx context.Context) (bool, error) {
				return true, nil
			}

			step := performSetStatusStep{
				bridge: bridgeStub,
			}

			assert.False(t, step.IsInterfaceNil())
			stepIdentifier := step.Execute(context.Background())
			assert.Equal(t, initialStep, stepIdentifier)
		})
		t.Run("if not leader, wait in this step", func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorPerformSetStatus()
			wasCalled := false
			bridgeStub.PerformActionOnElrondCalled = func(ctx context.Context) error {
				wasCalled = true
				return nil
			}

			step := performSetStatusStep{
				bridge: bridgeStub,
			}

			stepIdentifier := step.Execute(context.Background())
			assert.False(t, wasCalled)
			assert.Equal(t, step.Identifier(), stepIdentifier)
		})
		t.Run("if leader, first perform Set Status and then check again WasSetStatusPerformedOnElrond", func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorPerformSetStatus()
			bridgeStub.MyTurnAsLeaderCalled = func() bool {
				return true
			}
			wasCalled := false
			bridgeStub.PerformActionOnElrondCalled = func(ctx context.Context) error {
				wasCalled = true
				return nil
			}
			step := performSetStatusStep{
				bridge: bridgeStub,
			}

			stepIdentifier := step.Execute(context.Background())
			assert.True(t, wasCalled)
			assert.Equal(t, step.Identifier(), stepIdentifier)
		})
	})
}

func createStubExecutorPerformSetStatus() *bridgeV2.BridgeExecutorStub {
	stub := bridgeV2.NewBridgeExecutorStub()
	stub.GetLoggerCalled = func() logger.Logger {
		return testLogger
	}
	stub.WasActionPerformedOnElrondCalled = func(ctx context.Context) (bool, error) {
		return false, nil
	}
	stub.MyTurnAsLeaderCalled = func() bool {
		return false
	}
	return stub
}
