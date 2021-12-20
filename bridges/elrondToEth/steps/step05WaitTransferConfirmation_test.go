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

func TestExecute_WaitTransferConfirmation(t *testing.T) {
	t.Parallel()
	t.Run("should call WaitForTransferConfirmation and go to PerformingTransfer", func(t *testing.T) {
		bridgeStub := bridgeTests.NewBridgeExecutorStub()
		bridgeStub.GetLoggerCalled = func() logger.Logger {
			return testLogger
		}

		step := waitTransferConfirmationStep{
			bridge: bridgeStub,
		}

		assert.False(t, step.IsInterfaceNil())

		stepIdentifier := step.Execute(context.Background())
		expectedStep := core.StepIdentifier(elrondToEth.PerformingTransfer)
		assert.Equal(t, expectedStep, stepIdentifier)
	})
}
