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

	t.Run("error on WasTransferPerformedOnElrond", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorPerformSetStatus()
		bridgeStub.WasSetStatusPerformedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return false, expectedError
		}

		step := performSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("error on PerformTransferOnElrond", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorPerformSetStatus()
		bridgeStub.MyTurnAsLeaderCalled = func() bool {
			return true
		}
		bridgeStub.PerformSetStatusOnElrondCalled = func(ctx context.Context) error {
			return expectedError
		}

		step := performSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		t.Run("if transfer was performed we should go to initial step", func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorPerformSetStatus()
			bridgeStub.WasSetStatusPerformedOnElrondCalled = func(ctx context.Context) (bool, error) {
				return true, nil
			}

			step := performSetStatusStep{
				bridge: bridgeStub,
			}

			assert.False(t, step.IsInterfaceNil())
			stepIdentifier, err := step.Execute(context.Background())
			assert.Nil(t, err)
			assert.Equal(t, initialStep, stepIdentifier)
		})
		t.Run("if not leader, wait in this step", func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorPerformSetStatus()
			wasCalled := false
			bridgeStub.PerformSetStatusOnElrondCalled = func(ctx context.Context) error {
				wasCalled = true
				return nil
			}

			step := performSetStatusStep{
				bridge: bridgeStub,
			}

			stepIdentifier, err := step.Execute(context.Background())
			assert.Nil(t, err)
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
			bridgeStub.PerformSetStatusOnElrondCalled = func(ctx context.Context) error {
				wasCalled = true
				return nil
			}
			step := performSetStatusStep{
				bridge: bridgeStub,
			}

			stepIdentifier, err := step.Execute(context.Background())
			assert.Nil(t, err)
			assert.True(t, wasCalled)
			assert.Equal(t, step.Identifier(), stepIdentifier)
		})
	})
}

func createStubExecutorPerformSetStatus() *bridgeV2.ElrondToEthBridgeStub {
	stub := bridgeV2.NewElrondToEthBridgeStub()
	stub.GetLoggerCalled = func() logger.Logger {
		return testLogger
	}
	stub.WasSetStatusPerformedOnElrondCalled = func(ctx context.Context) (bool, error) {
		return false, nil
	}
	stub.MyTurnAsLeaderCalled = func() bool {
		return false
	}
	return stub
}
