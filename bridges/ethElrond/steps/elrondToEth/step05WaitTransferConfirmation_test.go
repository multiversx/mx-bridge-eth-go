package elrondToEth

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	bridgeTests "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridge"
	"github.com/stretchr/testify/assert"
)

func TestExecute_WaitTransferConfirmation(t *testing.T) {
	t.Parallel()
	t.Run("should call WaitForTransferConfirmation and go to PerformingTransfer", func(t *testing.T) {
		bridgeStub := bridgeTests.NewBridgeExecutorStub()

		step := waitTransferConfirmationStep{
			bridge: bridgeStub,
		}

		assert.False(t, step.IsInterfaceNil())

		stepIdentifier := step.Execute(context.Background())
		expectedStep := core.StepIdentifier(PerformingTransfer)
		assert.Equal(t, expectedStep, stepIdentifier)
	})
}
