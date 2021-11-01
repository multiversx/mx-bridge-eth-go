package steps

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/stateMachine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlowAsLeaderWaitSigsForTransferQuorumNotReachedWithStubChecking(t *testing.T) {
	t.Parallel()

	bem := stateMachine.NewBridgeExecutorMock()
	setAllDecisionHandlersToTrue(bem)
	bem.IsQuorumReachedForProposeTransferCalled = falseHandlerWithContext

	steps, err := CreateSteps(bem)
	require.Nil(t, err)
	smm := stateMachine.NewStateMachineMock(steps, ethToElrond.GettingPending)
	err = smm.Initialize()
	require.Nil(t, err)

	numSteps := 7
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	expectedSteps := []core.StepIdentifier{
		ethToElrond.GettingPending,
		ethToElrond.ProposingTransfer,
		ethToElrond.WaitingSignaturesForProposeTransfer,
		ethToElrond.ProposingSetStatus,
		ethToElrond.WaitingSignaturesForProposeSetStatus,
		ethToElrond.ExecutingSetStatus,
		ethToElrond.GettingPending,
	}

	assert.Equal(t, expectedSteps, smm.ExecutedSteps)
	assert.Equal(t, 2, bem.GetFunctionCounter(getPendingBatch))
	assert.Equal(t, 2, bem.GetFunctionCounter(hasPendingBatch))
	assert.Equal(t, 3, bem.GetFunctionCounter(isLeader))
	assert.Equal(t, 1, bem.GetFunctionCounter(proposeTransferOnDestination))
	assert.Equal(t, 5, bem.GetFunctionCounter(waitStepToFinish))
	assert.Equal(t, 2, bem.GetFunctionCounter(wasProposeTransferExecutedOnDestination))
	assert.Equal(t, 1, bem.GetFunctionCounter(signProposeTransferOnDestination))
	assert.Equal(t, 1, bem.GetFunctionCounter(isQuorumReachedForProposeTransfer))
	assert.Equal(t, 0, bem.GetFunctionCounter(executeTransferOnDestination))
	assert.Equal(t, 0, bem.GetFunctionCounter(wasTransferExecutedOnDestination))
	assert.Equal(t, 2, bem.GetFunctionCounter(cleanTopology))
	assert.Equal(t, 1, bem.GetFunctionCounter(updateTransactionsStatusesIfNeeded))
	assert.Equal(t, 1, bem.GetFunctionCounter(proposeSetStatusOnSource))
	assert.Equal(t, 1, bem.GetFunctionCounter(wasProposeSetStatusExecutedOnSource))
	assert.Equal(t, 1, bem.GetFunctionCounter(signProposeSetStatusOnSource))
	assert.Equal(t, 1, bem.GetFunctionCounter(isQuorumReachedForProposeSetStatus))
	assert.Equal(t, 1, bem.GetFunctionCounter(executeSetStatusOnSource))
	assert.Equal(t, 1, bem.GetFunctionCounter(wasSetStatusExecutedOnSource))
}

func TestFlowAsLeaderWaitSigsTransferWasNotProposedWithStubChecking(t *testing.T) {
	t.Parallel()

	bem := stateMachine.NewBridgeExecutorMock()
	setAllDecisionHandlersToTrue(bem)
	bem.IsQuorumReachedForProposeTransferCalled = falseHandlerWithContext
	counter := 0
	bem.WasProposeTransferExecutedOnDestinationCalled = func(ctx context.Context) bool {
		counter++
		return counter <= 1
	}

	steps, err := CreateSteps(bem)
	require.Nil(t, err)
	smm := stateMachine.NewStateMachineMock(steps, ethToElrond.GettingPending)
	err = smm.Initialize()
	require.Nil(t, err)

	numSteps := 7
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	expectedSteps := []core.StepIdentifier{
		ethToElrond.GettingPending,
		ethToElrond.ProposingTransfer,
		ethToElrond.WaitingSignaturesForProposeTransfer,
		ethToElrond.WaitingSignaturesForProposeTransfer,
		ethToElrond.WaitingSignaturesForProposeTransfer,
		ethToElrond.WaitingSignaturesForProposeTransfer,
		ethToElrond.WaitingSignaturesForProposeTransfer,
	}

	assert.Equal(t, expectedSteps, smm.ExecutedSteps)
	assert.Equal(t, 1, bem.GetFunctionCounter(getPendingBatch))
	assert.Equal(t, 1, bem.GetFunctionCounter(hasPendingBatch))
	assert.Equal(t, 1, bem.GetFunctionCounter(isLeader))
	assert.Equal(t, 1, bem.GetFunctionCounter(proposeTransferOnDestination))
	assert.Equal(t, 6, bem.GetFunctionCounter(waitStepToFinish))
	assert.Equal(t, 6, bem.GetFunctionCounter(wasProposeTransferExecutedOnDestination))
	assert.Equal(t, 1, bem.GetFunctionCounter(signProposeTransferOnDestination))
	assert.Equal(t, 5, bem.GetFunctionCounter(isQuorumReachedForProposeTransfer))
	assert.Equal(t, 0, bem.GetFunctionCounter(executeTransferOnDestination))
}

func TestFlowAsLeaderWaitSigsForSetStatusQuorumNotReachedWithStubChecking(t *testing.T) {
	t.Parallel()

	bem := stateMachine.NewBridgeExecutorMock()
	setAllDecisionHandlersToTrue(bem)
	bem.IsQuorumReachedForProposeSetStatusCalled = falseHandlerWithContext

	steps, err := CreateSteps(bem)
	require.Nil(t, err)
	smm := stateMachine.NewStateMachineMock(steps, ethToElrond.WaitingSignaturesForProposeSetStatus)
	err = smm.Initialize()
	require.Nil(t, err)

	numSteps := 2
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	expectedSteps := []core.StepIdentifier{
		ethToElrond.WaitingSignaturesForProposeSetStatus,
		ethToElrond.GettingPending,
	}

	assert.Equal(t, expectedSteps, smm.ExecutedSteps)
	assert.Equal(t, 1, bem.GetFunctionCounter(getPendingBatch))
	assert.Equal(t, 1, bem.GetFunctionCounter(hasPendingBatch))
	assert.Equal(t, 1, bem.GetFunctionCounter(isQuorumReachedForProposeSetStatus))
	assert.Equal(t, 1, bem.GetFunctionCounter(waitStepToFinish))
	assert.Equal(t, 1, bem.GetFunctionCounter(wasProposeSetStatusExecutedOnSource))
	assert.Equal(t, 1, bem.GetFunctionCounter(cleanTopology))
	assert.Equal(t, 0, bem.GetFunctionCounter(updateTransactionsStatusesIfNeeded))
}
