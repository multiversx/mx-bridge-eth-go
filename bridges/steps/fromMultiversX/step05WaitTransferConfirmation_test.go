package multiversxtoeth

import (
	"context"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/core"
	bridgeTests "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
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
