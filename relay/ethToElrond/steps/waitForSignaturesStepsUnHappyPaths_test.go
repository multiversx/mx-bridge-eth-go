package steps

import (
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond/steps/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlowAsLeaderWaitSigsForTransferQuorumNotReachedWithStubChecking(t *testing.T) {
	t.Parallel()

	bem := mock.NewBridgeExecutorMock()
	setAllDecisionHandlersToTrue(bem)
	bem.IsQuorumReachedForProposeTransferCalled = falseHandler

	steps, err := CreateSteps(bem)
	require.Nil(t, err)
	smm := mock.NewStateMachineMock(steps, ethToElrond.GetPending)
	err = smm.Initialize()
	require.Nil(t, err)

	numSteps := 7
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	expectedSteps := []relay.StepIdentifier{
		ethToElrond.GetPending,
		ethToElrond.ProposeTransfer,
		ethToElrond.WaitForSignaturesForProposeTransfer,
		ethToElrond.ProposeSetStatus,
		ethToElrond.WaitForSignaturesForProposeSetStatus,
		ethToElrond.ExecuteSetStatus,
		ethToElrond.GetPending,
	}

	assert.Equal(t, expectedSteps, smm.ExecutedSteps)
	assert.Equal(t, 2, bem.GetFunctionCounter("GetPendingBatch"))
	assert.Equal(t, 2, bem.GetFunctionCounter("HasPendingBatch"))
	assert.Equal(t, 3, bem.GetFunctionCounter("IsLeader"))
	assert.Equal(t, 1, bem.GetFunctionCounter("ProposeTransferOnDestination"))
	assert.Equal(t, 5, bem.GetFunctionCounter("WaitStepToFinish"))
	assert.Equal(t, 2, bem.GetFunctionCounter("WasProposeTransferExecutedOnDestination"))
	assert.Equal(t, 1, bem.GetFunctionCounter("SignProposeTransferOnDestination"))
	assert.Equal(t, 1, bem.GetFunctionCounter("IsQuorumReachedForProposeTransfer"))
	assert.Equal(t, 0, bem.GetFunctionCounter("ExecuteTransferOnDestination"))
	assert.Equal(t, 0, bem.GetFunctionCounter("WasTransferExecutedOnDestination"))
	assert.Equal(t, 2, bem.GetFunctionCounter("CleanTopology"))
	assert.Equal(t, 2, bem.GetFunctionCounter("SetStatusExecutedOnAllTransactions"))
	assert.Equal(t, 1, bem.GetFunctionCounter("ProposeSetStatusOnSource"))
	assert.Equal(t, 1, bem.GetFunctionCounter("WasProposeSetStatusExecutedOnSource"))
	assert.Equal(t, 1, bem.GetFunctionCounter("SignProposeSetStatusOnDestination"))
	assert.Equal(t, 1, bem.GetFunctionCounter("IsQuorumReachedForProposeSetStatus"))
	assert.Equal(t, 1, bem.GetFunctionCounter("ExecuteSetStatusOnSource"))
	assert.Equal(t, 1, bem.GetFunctionCounter("WasSetStatusExecutedOnSource"))
}

func TestFlowAsLeaderWaitSigsTransferWasNotProposedWithStubChecking(t *testing.T) {
	t.Parallel()

	bem := mock.NewBridgeExecutorMock()
	setAllDecisionHandlersToTrue(bem)
	bem.IsQuorumReachedForProposeTransferCalled = falseHandler
	counter := 0
	bem.WasProposeTransferExecutedOnDestinationCalled = func() bool {
		counter++
		return counter <= 1
	}

	steps, err := CreateSteps(bem)
	require.Nil(t, err)
	smm := mock.NewStateMachineMock(steps, ethToElrond.GetPending)
	err = smm.Initialize()
	require.Nil(t, err)

	numSteps := 7
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	expectedSteps := []relay.StepIdentifier{
		ethToElrond.GetPending,
		ethToElrond.ProposeTransfer,
		ethToElrond.WaitForSignaturesForProposeTransfer,
		ethToElrond.WaitForSignaturesForProposeTransfer,
		ethToElrond.WaitForSignaturesForProposeTransfer,
		ethToElrond.WaitForSignaturesForProposeTransfer,
		ethToElrond.WaitForSignaturesForProposeTransfer,
	}

	assert.Equal(t, expectedSteps, smm.ExecutedSteps)
	assert.Equal(t, 1, bem.GetFunctionCounter("GetPendingBatch"))
	assert.Equal(t, 1, bem.GetFunctionCounter("HasPendingBatch"))
	assert.Equal(t, 1, bem.GetFunctionCounter("IsLeader"))
	assert.Equal(t, 1, bem.GetFunctionCounter("ProposeTransferOnDestination"))
	assert.Equal(t, 6, bem.GetFunctionCounter("WaitStepToFinish"))
	assert.Equal(t, 6, bem.GetFunctionCounter("WasProposeTransferExecutedOnDestination"))
	assert.Equal(t, 1, bem.GetFunctionCounter("SignProposeTransferOnDestination"))
	assert.Equal(t, 5, bem.GetFunctionCounter("IsQuorumReachedForProposeTransfer"))
	assert.Equal(t, 0, bem.GetFunctionCounter("ExecuteTransferOnDestination"))
}

func TestFlowAsLeaderWaitSigsForSetStatusQuorumNotReachedWithStubChecking(t *testing.T) {
	t.Parallel()

	bem := mock.NewBridgeExecutorMock()
	setAllDecisionHandlersToTrue(bem)
	bem.IsQuorumReachedForProposeSetStatusCalled = falseHandler

	steps, err := CreateSteps(bem)
	require.Nil(t, err)
	smm := mock.NewStateMachineMock(steps, ethToElrond.WaitForSignaturesForProposeSetStatus)
	err = smm.Initialize()
	require.Nil(t, err)

	numSteps := 2
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	expectedSteps := []relay.StepIdentifier{
		ethToElrond.WaitForSignaturesForProposeSetStatus,
		ethToElrond.GetPending,
	}

	assert.Equal(t, expectedSteps, smm.ExecutedSteps)
	assert.Equal(t, 1, bem.GetFunctionCounter("GetPendingBatch"))
	assert.Equal(t, 1, bem.GetFunctionCounter("HasPendingBatch"))
	assert.Equal(t, 1, bem.GetFunctionCounter("IsQuorumReachedForProposeSetStatus"))
	assert.Equal(t, 1, bem.GetFunctionCounter("WaitStepToFinish"))
	assert.Equal(t, 1, bem.GetFunctionCounter("WasProposeSetStatusExecutedOnSource"))
	assert.Equal(t, 1, bem.GetFunctionCounter("CleanTopology"))
	assert.Equal(t, 1, bem.GetFunctionCounter("SetStatusExecutedOnAllTransactions"))
}
