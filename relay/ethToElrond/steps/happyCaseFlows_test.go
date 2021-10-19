package steps

import (
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond/steps/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var trueHandler = func() bool { return true }
var falseHandler = func() bool { return false }

func setAllDecisionHandlersToTrue(bem *mock.BridgeExecutorMock) {
	bem.HasPendingBatchCalled = trueHandler
	bem.IsLeaderCalled = trueHandler
	bem.WasProposeTransferExecutedOnDestinationCalled = trueHandler
	bem.WasProposeSetStatusExecutedOnSourceCalled = trueHandler
	bem.WasTransferExecutedOnDestinationCalled = trueHandler
	bem.WasSetStatusExecutedOnSourceCalled = trueHandler
	bem.IsQuorumReachedForProposeTransferCalled = trueHandler
	bem.IsQuorumReachedForProposeSetStatusCalled = trueHandler
}

func TestGetPendingEndlessLoop(t *testing.T) {
	t.Parallel()

	bem := mock.NewBridgeExecutorMock()
	bem.HasPendingBatchCalled = func() bool {
		return false
	}

	steps, err := CreateSteps(bem)
	require.Nil(t, err)
	smm := mock.NewStateMachineMock(steps, ethToElrond.GetPending)
	err = smm.Initialize()
	require.Nil(t, err)

	numSteps := 100
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	assert.Equal(t, numSteps, bem.GetFunctionCounter("GetPendingBatch"))
}

func TestFlowAsLeaderForTwoCompleteFlows(t *testing.T) {
	t.Parallel()

	bem := mock.NewBridgeExecutorMock()
	setAllDecisionHandlersToTrue(bem)

	steps, err := CreateSteps(bem)
	require.Nil(t, err)
	smm := mock.NewStateMachineMock(steps, ethToElrond.GetPending)
	err = smm.Initialize()
	require.Nil(t, err)

	numSteps := 14
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	expectedSteps := []relay.StepIdentifier{
		ethToElrond.GetPending,
		ethToElrond.ProposeTransfer,
		ethToElrond.WaitForSignaturesForProposeTransfer,
		ethToElrond.ExecuteTransfer,
		ethToElrond.ProposeSetStatus,
		ethToElrond.WaitForSignaturesForProposeSetStatus,
		ethToElrond.ExecuteSetStatus,
		ethToElrond.GetPending,
		ethToElrond.ProposeTransfer,
		ethToElrond.WaitForSignaturesForProposeTransfer,
		ethToElrond.ExecuteTransfer,
		ethToElrond.ProposeSetStatus,
		ethToElrond.WaitForSignaturesForProposeSetStatus,
		ethToElrond.ExecuteSetStatus,
	}

	assert.Equal(t, expectedSteps, smm.ExecutedSteps)
}

func TestFlowAsSignerForTwoCompleteFlows(t *testing.T) {
	t.Parallel()

	bem := mock.NewBridgeExecutorMock()
	setAllDecisionHandlersToTrue(bem)
	bem.IsLeaderCalled = falseHandler

	steps, err := CreateSteps(bem)
	require.Nil(t, err)
	smm := mock.NewStateMachineMock(steps, ethToElrond.GetPending)
	err = smm.Initialize()
	require.Nil(t, err)

	numSteps := 14
	for i := 0; i < numSteps; i++ {
		err = smm.ExecuteOneStep()
		require.Nil(t, err)
	}

	expectedSteps := []relay.StepIdentifier{
		ethToElrond.GetPending,
		ethToElrond.ProposeTransfer,
		ethToElrond.WaitForSignaturesForProposeTransfer,
		ethToElrond.ExecuteTransfer,
		ethToElrond.ProposeSetStatus,
		ethToElrond.WaitForSignaturesForProposeSetStatus,
		ethToElrond.ExecuteSetStatus,
		ethToElrond.GetPending,
		ethToElrond.ProposeTransfer,
		ethToElrond.WaitForSignaturesForProposeTransfer,
		ethToElrond.ExecuteTransfer,
		ethToElrond.ProposeSetStatus,
		ethToElrond.WaitForSignaturesForProposeSetStatus,
		ethToElrond.ExecuteSetStatus,
	}

	assert.Equal(t, expectedSteps, smm.ExecutedSteps)
}

func TestFlowAsLeaderForOneCompleteFlowWithStubChecking(t *testing.T) {
	t.Parallel()

	bem := mock.NewBridgeExecutorMock()
	setAllDecisionHandlersToTrue(bem)

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
		ethToElrond.ExecuteTransfer,
		ethToElrond.ProposeSetStatus,
		ethToElrond.WaitForSignaturesForProposeSetStatus,
		ethToElrond.ExecuteSetStatus,
	}

	assert.Equal(t, expectedSteps, smm.ExecutedSteps)
	assert.Equal(t, 1, bem.GetFunctionCounter("GetPendingBatch"))
	assert.Equal(t, 1, bem.GetFunctionCounter("HasPendingBatch"))
	assert.Equal(t, 4, bem.GetFunctionCounter("IsLeader"))
	assert.Equal(t, 1, bem.GetFunctionCounter("ProposeTransferOnDestination"))
	assert.Equal(t, 6, bem.GetFunctionCounter("WaitStepToFinish"))
	assert.Equal(t, 1, bem.GetFunctionCounter("WasProposeTransferExecutedOnDestination"))
	assert.Equal(t, 1, bem.GetFunctionCounter("SignProposeTransferOnDestination"))
	assert.Equal(t, 1, bem.GetFunctionCounter("IsQuorumReachedForProposeTransfer"))
	assert.Equal(t, 1, bem.GetFunctionCounter("ExecuteTransferOnDestination"))
	assert.Equal(t, 1, bem.GetFunctionCounter("WasTransferExecutedOnDestination"))
	assert.Equal(t, 2, bem.GetFunctionCounter("CleanTopology"))
	assert.Equal(t, 2, bem.GetFunctionCounter("SetStatusExecutedOnAllTransactions"))
	assert.Equal(t, 1, bem.GetFunctionCounter("ProposeSetStatusOnSource"))
	assert.Equal(t, 1, bem.GetFunctionCounter("WasProposeSetStatusExecutedOnSource"))
	assert.Equal(t, 1, bem.GetFunctionCounter("SignProposeSetStatusOnDestination"))
	assert.Equal(t, 1, bem.GetFunctionCounter("IsQuorumReachedForProposeSetStatus"))
	assert.Equal(t, 1, bem.GetFunctionCounter("ExecuteSetStatusOnSource"))
	assert.Equal(t, 1, bem.GetFunctionCounter("WasSetStatusExecutedOnSource"))
}

func TestFlowAsSignerForOneCompleteFlowWithStubChecking(t *testing.T) {
	t.Parallel()

	bem := mock.NewBridgeExecutorMock()
	setAllDecisionHandlersToTrue(bem)
	bem.IsLeaderCalled = falseHandler

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
		ethToElrond.ExecuteTransfer,
		ethToElrond.ProposeSetStatus,
		ethToElrond.WaitForSignaturesForProposeSetStatus,
		ethToElrond.ExecuteSetStatus,
	}

	assert.Equal(t, expectedSteps, smm.ExecutedSteps)
	assert.Equal(t, 1, bem.GetFunctionCounter("GetPendingBatch"))
	assert.Equal(t, 1, bem.GetFunctionCounter("HasPendingBatch"))
	assert.Equal(t, 4, bem.GetFunctionCounter("IsLeader"))
	assert.Equal(t, 0, bem.GetFunctionCounter("ProposeTransferOnDestination"))
	assert.Equal(t, 6, bem.GetFunctionCounter("WaitStepToFinish"))
	assert.Equal(t, 1, bem.GetFunctionCounter("WasProposeTransferExecutedOnDestination"))
	assert.Equal(t, 1, bem.GetFunctionCounter("SignProposeTransferOnDestination"))
	assert.Equal(t, 1, bem.GetFunctionCounter("IsQuorumReachedForProposeTransfer"))
	assert.Equal(t, 0, bem.GetFunctionCounter("ExecuteTransferOnDestination"))
	assert.Equal(t, 1, bem.GetFunctionCounter("WasTransferExecutedOnDestination"))
	assert.Equal(t, 2, bem.GetFunctionCounter("CleanTopology"))
	assert.Equal(t, 2, bem.GetFunctionCounter("SetStatusExecutedOnAllTransactions"))
	assert.Equal(t, 0, bem.GetFunctionCounter("ProposeSetStatusOnSource"))
	assert.Equal(t, 1, bem.GetFunctionCounter("WasProposeSetStatusExecutedOnSource"))
	assert.Equal(t, 1, bem.GetFunctionCounter("SignProposeSetStatusOnDestination"))
	assert.Equal(t, 1, bem.GetFunctionCounter("IsQuorumReachedForProposeSetStatus"))
	assert.Equal(t, 0, bem.GetFunctionCounter("ExecuteSetStatusOnSource"))
	assert.Equal(t, 1, bem.GetFunctionCounter("WasSetStatusExecutedOnSource"))
}
