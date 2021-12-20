package steps

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/elrondToEth"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	bridgeTests "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridge"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"
)

var initialStep = core.StepIdentifier(elrondToEth.GettingPendingBatchFromElrond)

func TestExecute_SignProposedTransfer(t *testing.T) {
	t.Parallel()

	t.Run("nil batch on GetStoredBatch", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorSignProposedTransfer()
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return nil
		}

		step := signProposedTransferStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("nil batch on SignTransferOnEthereum", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorSignProposedTransfer()
		bridgeStub.SignTransferOnEthereumCalled = func() error {
			return expectedError
		}

		step := signProposedTransferStep{
			bridge: bridgeStub,
		}

		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubExecutorSignProposedTransfer()

		step := signProposedTransferStep{
			bridge: bridgeStub,
		}

		assert.False(t, step.IsInterfaceNil())

		expectedStepIdentifier := core.StepIdentifier(elrondToEth.WaitingForQuorumOnTransfer)
		stepIdentifier := step.Execute(context.Background())
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})
}

func createStubExecutorSignProposedTransfer() *bridgeTests.BridgeExecutorStub {
	stub := bridgeTests.NewBridgeExecutorStub()
	stub.GetLoggerCalled = func() logger.Logger {
		return testLogger
	}
	stub.GetStoredBatchCalled = func() *clients.TransferBatch {
		return testBatch
	}
	stub.SignTransferOnEthereumCalled = func() error {
		return nil
	}
	return stub
}
