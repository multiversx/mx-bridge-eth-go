package steps

import (
	"context"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/stateMachine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlowGetPendingContinuouslyErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	bem := stateMachine.NewBridgeExecutorMock()
	setAllDecisionHandlersToTrue(bem)
	bem.GetPendingBatchCalled = func(ctx context.Context) error {
		return expectedErr
	}

	steps, err := CreateSteps(bem)
	require.Nil(t, err)
	smm := stateMachine.NewStateMachineMock(steps, ethToElrond.GettingPending)
	err = smm.Initialize()
	require.Nil(t, err)

	numSteps := 100
	expectedSteps := make([]core.StepIdentifier, 0)
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)

		expectedSteps = append(expectedSteps, ethToElrond.GettingPending)
	}

	assert.Equal(t, expectedSteps, smm.ExecutedSteps)
	assert.Equal(t, numSteps, bem.GetFunctionCounter(getPendingBatch))
}
