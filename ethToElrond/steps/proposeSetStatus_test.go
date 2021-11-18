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

func TestFlowProposeSetStatusTheBatchIsNotReady(t *testing.T) {
	t.Parallel()

	bem := stateMachine.NewBridgeExecutorMock()
	setAllDecisionHandlersToTrue(bem)
	bem.IsPendingBatchReadyCalled = func(ctx context.Context) (bool, error) {
		return false, nil
	}

	steps, err := CreateSteps(bem)
	require.Nil(t, err)
	smm := stateMachine.NewStateMachineMock(steps, ethToElrond.GettingPending)
	err = smm.Initialize()
	require.Nil(t, err)

	numSteps := 6
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	expectedSteps := []core.StepIdentifier{
		ethToElrond.GettingPending,
		ethToElrond.ProposingTransfer,
		ethToElrond.WaitingSignaturesForProposeTransfer,
		ethToElrond.ExecutingTransfer,
		ethToElrond.ProposingSetStatus,
		ethToElrond.GettingPending,
	}

	assert.Equal(t, expectedSteps, smm.ExecutedSteps)
}

func TestFlowProposeSetStatusTheBatchCanNotBeRead(t *testing.T) {
	t.Parallel()

	bem := stateMachine.NewBridgeExecutorMock()
	setAllDecisionHandlersToTrue(bem)
	bem.IsPendingBatchReadyCalled = func(ctx context.Context) (bool, error) {
		return false, errors.New("error reading batch")
	}

	steps, err := CreateSteps(bem)
	require.Nil(t, err)
	smm := stateMachine.NewStateMachineMock(steps, ethToElrond.GettingPending)
	err = smm.Initialize()
	require.Nil(t, err)

	numSteps := 12
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	expectedSteps := []core.StepIdentifier{
		ethToElrond.GettingPending,
		ethToElrond.ProposingTransfer,
		ethToElrond.WaitingSignaturesForProposeTransfer,
		ethToElrond.ExecutingTransfer,
		ethToElrond.ProposingSetStatus,
		ethToElrond.ProposingSetStatus,
		ethToElrond.ProposingSetStatus,
		ethToElrond.ProposingSetStatus,
		ethToElrond.ProposingSetStatus,
		ethToElrond.ProposingSetStatus,
		ethToElrond.ProposingSetStatus,
		ethToElrond.ProposingSetStatus,
	}

	assert.Equal(t, expectedSteps, smm.ExecutedSteps)
}
