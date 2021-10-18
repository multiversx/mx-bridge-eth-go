package steps

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond/steps/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlowGetPendingContinuously(t *testing.T) {
	t.Parallel()

	bem := &mock.BridgeExecutorMock{
		HasPendingBatchCalled: func() bool {
			return false
		},
	}

	steps, _ := CreateSteps(bem)
	smm := mock.NewStateMachineMock(steps, ethToElrond.GetPending)
	err := smm.Initialize()
	require.Nil(t, err)

	numSteps := 100
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	assert.Equal(t, numSteps, bem.NumCalledGetPendingBatchCalled)
}

func TestFlowProposePendingBatchNotBeingTheLeader(t *testing.T) {
	t.Parallel()

	bem := &mock.BridgeExecutorMock{
		HasPendingBatchCalled: func() bool {
			return true
		},
	}

	steps, _ := CreateSteps(bem)
	smm := mock.NewStateMachineMock(steps, ethToElrond.GetPending)
	err := smm.Initialize()
	require.Nil(t, err)

	numSteps := 3
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	expectedExecutedSteps := []relay.StepIdentifier{
		ethToElrond.GetPending,
		ethToElrond.ProposeTransfer,
		ethToElrond.GetPending,
	}

	assert.Equal(t, 2, bem.NumCalledGetPendingBatchCalled)
	assert.Equal(t, 2, bem.NumHasPendingBatchCalled)
	assert.Equal(t, 1, bem.NumCalledIsLeaderCalled)
	assert.Equal(t, 0, bem.NumCalledProposeTransferCalled)

	assert.Equal(t, expectedExecutedSteps, smm.ExecutedSteps)
}

func TestFlowProposePendingBatchBeingTheLeader(t *testing.T) {
	t.Parallel()

	bem := &mock.BridgeExecutorMock{
		HasPendingBatchCalled: func() bool {
			return true
		},
		IsLeaderCalled: func() bool {
			return true
		},
	}

	steps, _ := CreateSteps(bem)
	smm := mock.NewStateMachineMock(steps, ethToElrond.GetPending)
	err := smm.Initialize()
	require.Nil(t, err)

	numSteps := 3
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	expectedExecutedSteps := []relay.StepIdentifier{
		ethToElrond.GetPending,
		ethToElrond.ProposeTransfer,
		ethToElrond.GetPending,
	}

	assert.Equal(t, 2, bem.NumCalledGetPendingBatchCalled)
	assert.Equal(t, 2, bem.NumHasPendingBatchCalled)
	assert.Equal(t, 1, bem.NumCalledIsLeaderCalled)
	assert.Equal(t, 1, bem.NumCalledProposeTransferCalled)

	assert.Equal(t, expectedExecutedSteps, smm.ExecutedSteps)
}

func TestFlowProposePendingBatchBeingTheLeaderErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	bem := &mock.BridgeExecutorMock{
		HasPendingBatchCalled: func() bool {
			return true
		},
		IsLeaderCalled: func() bool {
			return true
		},
		ProposeTransferCalled: func() error {
			return expectedErr
		},
	}

	steps, _ := CreateSteps(bem)
	smm := mock.NewStateMachineMock(steps, ethToElrond.GetPending)
	err := smm.Initialize()
	require.Nil(t, err)

	numSteps := 3
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	expectedExecutedSteps := []relay.StepIdentifier{
		ethToElrond.GetPending,
		ethToElrond.ProposeTransfer,
		ethToElrond.GetPending,
	}

	assert.Equal(t, 2, bem.NumCalledGetPendingBatchCalled)
	assert.Equal(t, 2, bem.NumHasPendingBatchCalled)
	assert.Equal(t, 1, bem.NumCalledIsLeaderCalled)
	assert.Equal(t, 1, bem.NumCalledProposeTransferCalled)

	assert.Equal(t, expectedExecutedSteps, smm.ExecutedSteps)
}
