package steps

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/elrondToEth"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"
)

func TestExecute_ProposeSetStatus(t *testing.T) {
	t.Parallel()
	t.Run("nil batch on GetStoredBatchFromElrond", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorProposeSetStatus()
		bridgeStub.GetStoredBatchFromElrondCalled = func() *clients.TransferBatch {
			return nil
		}

		step := proposeSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("error on WasSetStatusProposedOnElrond", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorProposeSetStatus()
		bridgeStub.WasSetStatusProposedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return false, expectedError
		}

		step := proposeSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("error on ProposeSetStatusOnElrond", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorProposeSetStatus()
		bridgeStub.ProposeSetStatusOnElrondCalled = func(ctx context.Context) error {
			return expectedError
		}

		step := proposeSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		t.Run("if SetStatus was proposed it should go to SigningProposedSetStatusOnElrond", func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorProposeSetStatus()
			bridgeStub.WasSetStatusProposedOnElrondCalled = func(ctx context.Context) (bool, error) {
				return true, nil
			}

			step := proposeSetStatusStep{
				bridge: bridgeStub,
			}

			assert.False(t, step.IsInterfaceNil())
			expectedStep := core.StepIdentifier(elrondToEth.SigningProposedSetStatusOnElrond)
			stepIdentifier, err := step.Execute(context.Background())
			assert.Nil(t, err)
			assert.Equal(t, expectedStep, stepIdentifier)

		})
		t.Run("if SetStatus was not proposed", func(t *testing.T) {
			t.Parallel()
			t.Run("if not leader, should stay in current step", func(t *testing.T) {
				t.Parallel()
				bridgeStub := createStubExecutorProposeSetStatus()
				bridgeStub.MyTurnAsLeaderCalled = func() bool {
					return false
				}
				step := proposeSetStatusStep{
					bridge: bridgeStub,
				}

				stepIdentifier, err := step.Execute(context.Background())
				assert.Nil(t, err)
				assert.Equal(t, step.Identifier(), stepIdentifier)

			})
			t.Run("if leader, should go to SigningProposedTransferOnElrond", func(t *testing.T) {
				t.Parallel()
				bridgeStub := createStubExecutorProposeSetStatus()

				step := proposeSetStatusStep{
					bridge: bridgeStub,
				}

				expectedStep := core.StepIdentifier(elrondToEth.SigningProposedSetStatusOnElrond)
				stepIdentifier, err := step.Execute(context.Background())
				assert.Nil(t, err)
				assert.Equal(t, expectedStep, stepIdentifier)

			})
		})

	})
}

func createStubExecutorProposeSetStatus() *bridgeV2.ElrondToEthBridgeStub {
	stub := bridgeV2.NewElrondToEthBridgeStub()
	stub.GetLoggerCalled = func() logger.Logger {
		return testLogger
	}
	stub.GetStoredBatchFromElrondCalled = func() *clients.TransferBatch {
		return testBatch
	}
	stub.WasSetStatusProposedOnElrondCalled = func(ctx context.Context) (bool, error) {
		return false, nil
	}
	stub.MyTurnAsLeaderCalled = func() bool {
		return true
	}
	stub.ProposeSetStatusOnElrondCalled = func(ctx context.Context) error {
		return nil
	}
	return stub
}
