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

func TestExecute_ResolveSetStatus(t *testing.T) {
	t.Parallel()
	t.Run("nil batch on GetStoredBatch", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorResolveSetStatus()
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return nil
		}

		step := resolveSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("error on GetStoredBatch", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorResolveSetStatus()
		bridgeStub.GetBatchFromElrondCalled = func(ctx context.Context) (*clients.TransferBatch, error) {
			return nil, expectedError
		}

		step := resolveSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("nil batch on GetBatchFromElrond", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorResolveSetStatus()
		bridgeStub.GetBatchFromElrondCalled = func(ctx context.Context) (*clients.TransferBatch, error) {
			return nil, nil
		}

		step := resolveSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("error on GetBatchStatusesFromEthereum", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorResolveSetStatus()
		bridgeStub.GetBatchStatusesFromEthereumCalled = func(ctx context.Context) ([]byte, error) {
			return nil, expectedError
		}

		step := resolveSetStatusStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})
	t.Run("should call ResolveNewDepositsStatuses and go to ProposingSetStatusOnElrond", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorResolveSetStatus()
		bridgeStub.GetBatchStatusesFromEthereumCalled = func(ctx context.Context) ([]byte, error) {
			return []byte{clients.Executed, clients.Rejected}, nil
		}
		wasCalled := false
		bridgeStub.ResolveNewDepositsStatusesCalled = func(numDeposits uint64) {
			wasCalled = true
		}

		step := resolveSetStatusStep{
			bridge: bridgeStub,
		}

		assert.False(t, step.IsInterfaceNil())

		expectedStep := core.StepIdentifier(elrondToEth.ProposingSetStatusOnElrond)
		stepIdentifier := step.Execute(context.Background())
		assert.True(t, wasCalled)
		assert.NotEqual(t, step.Identifier(), stepIdentifier)
		assert.Equal(t, expectedStep, stepIdentifier)
	})
}

func createStubExecutorResolveSetStatus() *bridgeV2.BridgeExecutorStub {
	stub := bridgeV2.NewBridgeExecutorStub()
	stub.GetLoggerCalled = func() logger.Logger {
		return testLogger
	}
	stub.GetStoredBatchCalled = func() *clients.TransferBatch {
		return testBatch
	}
	stub.GetBatchFromElrondCalled = func(ctx context.Context) (*clients.TransferBatch, error) {
		return testBatch, nil
	}
	return stub
}
