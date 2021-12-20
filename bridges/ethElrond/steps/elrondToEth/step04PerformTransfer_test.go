package elrondToEth

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	bridgeTests "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridge"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"
)

func TestExecute_PerformTransfer(t *testing.T) {
	t.Parallel()

	t.Run("error on WasTransferPerformedOnEthereum", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorPerformTransfer()
		bridgeStub.WasTransferPerformedOnEthereumCalled = func(ctx context.Context) (bool, error) {
			return false, expectedError
		}

		step := performTransferStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("error on PerformTransferOnEthereum", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorPerformTransfer()
		bridgeStub.MyTurnAsLeaderCalled = func() bool {
			return true
		}
		bridgeStub.PerformTransferOnEthereumCalled = func(ctx context.Context) error {
			return expectedError
		}

		step := performTransferStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		t.Run("if transfer was performed we should go to ResolvingSetStatusOnElrond", func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorPerformTransfer()
			bridgeStub.WasTransferPerformedOnEthereumCalled = func(ctx context.Context) (bool, error) {
				return true, nil
			}

			step := performTransferStep{
				bridge: bridgeStub,
			}

			assert.False(t, step.IsInterfaceNil())
			expectedStep := core.StepIdentifier(ResolvingSetStatusOnElrond)
			stepIdentifier := step.Execute(context.Background())
			assert.Equal(t, expectedStep, stepIdentifier)
		})
		t.Run("if not leader, go to WaitingTransferConfirmation", func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorPerformTransfer()
			wasCalled := false
			bridgeStub.PerformTransferOnEthereumCalled = func(ctx context.Context) error {
				wasCalled = true
				return nil
			}

			step := performTransferStep{
				bridge: bridgeStub,
			}

			expectedStep := core.StepIdentifier(WaitingTransferConfirmation)
			stepIdentifier := step.Execute(context.Background())
			assert.False(t, wasCalled)
			assert.Equal(t, expectedStep, stepIdentifier)
		})
		t.Run("if leader, first perform Trasfer and then go to WaitingTransferConfirmation", func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorPerformTransfer()
			bridgeStub.MyTurnAsLeaderCalled = func() bool {
				return true
			}
			wasCalled := false
			bridgeStub.PerformTransferOnEthereumCalled = func(ctx context.Context) error {
				wasCalled = true
				return nil
			}
			step := performTransferStep{
				bridge: bridgeStub,
			}

			expectedStep := core.StepIdentifier(WaitingTransferConfirmation)
			stepIdentifier := step.Execute(context.Background())
			assert.True(t, wasCalled)
			assert.Equal(t, expectedStep, stepIdentifier)
		})
	})
}

func createStubExecutorPerformTransfer() *bridgeTests.BridgeExecutorStub {
	stub := bridgeTests.NewBridgeExecutorStub()
	stub.GetLoggerCalled = func() logger.Logger {
		return testLogger
	}
	stub.WasTransferPerformedOnEthereumCalled = func(ctx context.Context) (bool, error) {
		return false, nil
	}
	stub.MyTurnAsLeaderCalled = func() bool {
		return false
	}
	return stub
}
