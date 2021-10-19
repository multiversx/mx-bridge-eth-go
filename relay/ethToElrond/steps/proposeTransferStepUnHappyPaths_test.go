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

func TestFlowAsLeaderProposeTransferOnDestinationErrorsWithStubChecking(t *testing.T) {
	t.Parallel()

	bem := mock.NewBridgeExecutorMock()
	setAllDecisionHandlersToTrue(bem)
	bem.ProposeTransferOnDestinationCalled = func() error {
		return errors.New("expected error")
	}

	steps, err := CreateSteps(bem)
	require.Nil(t, err)
	smm := mock.NewStateMachineMock(steps, ethToElrond.GetPending)
	err = smm.Initialize()
	require.Nil(t, err)

	numSteps := 6
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	expectedSteps := []relay.StepIdentifier{
		ethToElrond.GetPending,
		ethToElrond.ProposeTransfer,
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
	assert.Equal(t, 1, bem.GetFunctionCounter("PrintDebugInfo"))
	assert.Equal(t, 1, bem.GetFunctionCounter("SetStatusRejectedOnAllTransactions"))
	assert.Equal(t, 3, bem.GetFunctionCounter("WaitStepToFinish"))
	assert.Equal(t, 0, bem.GetFunctionCounter("WasProposeTransferExecutedOnDestination"))
	assert.Equal(t, 0, bem.GetFunctionCounter("SignProposeTransferOnDestination"))
	assert.Equal(t, 0, bem.GetFunctionCounter("IsQuorumReachedForProposeTransfer"))
	assert.Equal(t, 0, bem.GetFunctionCounter("ExecuteTransferOnDestination"))
	assert.Equal(t, 0, bem.GetFunctionCounter("WasTransferExecutedOnDestination"))
	assert.Equal(t, 1, bem.GetFunctionCounter("CleanTopology"))
	assert.Equal(t, 1, bem.GetFunctionCounter("SetStatusExecutedOnAllTransactions"))
	assert.Equal(t, 1, bem.GetFunctionCounter("ProposeSetStatusOnSource"))
	assert.Equal(t, 1, bem.GetFunctionCounter("WasProposeSetStatusExecutedOnSource"))
	assert.Equal(t, 1, bem.GetFunctionCounter("SignProposeSetStatusOnDestination"))
	assert.Equal(t, 1, bem.GetFunctionCounter("IsQuorumReachedForProposeSetStatus"))
	assert.Equal(t, 1, bem.GetFunctionCounter("ExecuteSetStatusOnSource"))
	assert.Equal(t, 1, bem.GetFunctionCounter("WasSetStatusExecutedOnSource"))
}

func TestFlowAsLeaderWasNotProposedTransferWithStubChecking(t *testing.T) {
	t.Parallel()

	bem := mock.NewBridgeExecutorMock()
	setAllDecisionHandlersToTrue(bem)
	bem.WasProposeTransferExecutedOnDestinationCalled = falseHandler

	steps, err := CreateSteps(bem)
	require.Nil(t, err)
	smm := mock.NewStateMachineMock(steps, ethToElrond.GetPending)
	err = smm.Initialize()
	require.Nil(t, err)

	numSteps := 6
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	expectedSteps := []relay.StepIdentifier{
		ethToElrond.GetPending,
		ethToElrond.ProposeTransfer,
		ethToElrond.ProposeTransfer,
		ethToElrond.ProposeTransfer,
		ethToElrond.ProposeTransfer,
		ethToElrond.ProposeTransfer,
	}

	assert.Equal(t, expectedSteps, smm.ExecutedSteps)
	assert.Equal(t, 1, bem.GetFunctionCounter("GetPendingBatch"))
	assert.Equal(t, 1, bem.GetFunctionCounter("HasPendingBatch"))
	assert.Equal(t, 5, bem.GetFunctionCounter("IsLeader"))
	assert.Equal(t, 5, bem.GetFunctionCounter("ProposeTransferOnDestination"))
	assert.Equal(t, 5, bem.GetFunctionCounter("WaitStepToFinish"))
	assert.Equal(t, 5, bem.GetFunctionCounter("WasProposeTransferExecutedOnDestination"))
}
