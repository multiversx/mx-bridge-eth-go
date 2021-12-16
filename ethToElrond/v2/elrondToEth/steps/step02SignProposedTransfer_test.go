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

var initialStep = core.StepIdentifier(elrondToEth.GettingPendingBatchFromElrond)

func TestExecute_SignProposedTransfer(t *testing.T) {
	t.Parallel()

	t.Run("nil batch on GetStoredBatchFromElrond", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubeEecutorSignproposedtransfer()
		bridgeStub.GetStoredBatchFromElrondCalled = func() *clients.TransferBatch {
			return nil
		}

		step := signProposedTransferStep{
			bridge: bridgeStub,
		}

		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("nil batch on SignTransferOnEthereum", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubeEecutorSignproposedtransfer()
		bridgeStub.SignTransferOnEthereumCalled = func(ctx context.Context) error {
			return expectedError
		}

		step := signProposedTransferStep{
			bridge: bridgeStub,
		}

		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, initialStep, stepIdentifier)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		bridgeStub := createStubeEecutorSignproposedtransfer()

		step := signProposedTransferStep{
			bridge: bridgeStub,
		}

		assert.False(t, step.IsInterfaceNil())

		expectedStepIdentifier := core.StepIdentifier(elrondToEth.WaitingForQuorumOnTransfer)
		stepIdentifier, err := step.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedStepIdentifier, stepIdentifier)
	})
}

func createStubeEecutorSignproposedtransfer() *bridgeV2.ElrondToEthBridgeStub {
	stub := bridgeV2.NewElrondToEthBridgeStub()
	stub.GetLoggerCalled = func() logger.Logger {
		return testLogger
	}
	stub.GetStoredBatchFromElrondCalled = func() *clients.TransferBatch {
		return testBatch
	}
	stub.SignTransferOnEthereumCalled = func(ctx context.Context) error {
		return nil
	}
	return stub
}
