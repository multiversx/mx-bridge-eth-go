package ethToElrond

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/bridgeExecutors"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/steps"
	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests"
	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-eth-bridge/stateMachine"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBridgeExecutorWithStateMachineOnCompleteExecutionFlow(t *testing.T) {
	sourceBridge := &mock.BridgeMock{}
	destinationBridge := &mock.BridgeMock{}

	batchID := bridge.NewBatchId(12345)
	sourceActionID := bridge.NewActionId(663725)
	pendingBatch := &bridge.Batch{
		Id: batchID,
		Transactions: []*bridge.DepositTransaction{
			{
				To:           "to1",
				From:         "from1",
				TokenAddress: "token address 1",
				Amount:       big.NewInt(1000),
				DepositNonce: big.NewInt(2),
				BlockNonce:   big.NewInt(2000000),
				Status:       0,
				Error:        nil,
			},
			{
				To:           "to2",
				From:         "from2",
				TokenAddress: "token address 2",
				Amount:       big.NewInt(1001),
				DepositNonce: big.NewInt(3),
				BlockNonce:   big.NewInt(2000001),
				Status:       0,
				Error:        nil,
			},
		},
	}
	sourceBridge.SetPending(pendingBatch)
	sourceBridge.SetActionID(sourceActionID)
	numGetPendingCalled := 0
	chDone := make(chan struct{})
	sourceBridge.GetPendingCalled = func() {
		numGetPendingCalled++
		if numGetPendingCalled == 2 {
			close(chDone)
		}
	}

	destinationActionID := bridge.NewActionId(343553)
	destinationBridge.SetActionID(destinationActionID)
	destinationBridge.GetTransactionsStatusesCalled = func(ctx context.Context, batchId bridge.BatchId) ([]uint8, error) {
		return makeMockStatuses(len(pendingBatch.Transactions)), nil
	}

	sm, err := createAndStartBridge(sourceBridge, destinationBridge, 1, 1, true, "test")
	require.Nil(t, err)

	select {
	case <-chDone:
		_ = sm.Close()
	case <-time.After(time.Second * 5):
		_ = sm.Close()
		require.Fail(t, "timeout while executing the whole process")
	}

	expectedStatuses := makeMockStatuses(len(pendingBatch.Transactions))
	checkStatusWhenExecutedOnSource(t, sourceBridge, pendingBatch, sourceActionID, expectedStatuses)
	checkStatusWhenExecutedOnDestination(t, destinationBridge, pendingBatch, destinationActionID)
}

func TestBridgeExecutorWithStateMachineFailedToProposeTransfer(t *testing.T) {
	proposeErr := errors.New("propose error")
	sourceBridge := &mock.BridgeMock{}
	destinationBridge := &mock.BridgeMock{
		ProposeTransferCalled: func(_ context.Context, batch *bridge.Batch) (string, error) {
			return "", proposeErr
		},
	}

	batchID := bridge.NewBatchId(12345)
	sourceActionID := bridge.NewActionId(663725)
	pendingBatch := &bridge.Batch{
		Id: batchID,
		Transactions: []*bridge.DepositTransaction{
			{
				To:           "to1",
				From:         "from1",
				TokenAddress: "token address 1",
				Amount:       big.NewInt(1000),
				DepositNonce: big.NewInt(2),
				BlockNonce:   big.NewInt(2000000),
				Status:       0,
				Error:        nil,
			},
			{
				To:           "to2",
				From:         "from2",
				TokenAddress: "token address 2",
				Amount:       big.NewInt(1001),
				DepositNonce: big.NewInt(3),
				BlockNonce:   big.NewInt(2000001),
				Status:       0,
				Error:        nil,
			},
		},
	}
	sourceBridge.SetPending(pendingBatch)
	sourceBridge.SetActionID(sourceActionID)
	numGetPendingCalled := 0
	chDone := make(chan struct{})
	sourceBridge.GetPendingCalled = func() {
		numGetPendingCalled++
		if numGetPendingCalled == 2 {
			close(chDone)
		}
	}

	destinationActionID := bridge.NewActionId(343553)
	destinationBridge.SetActionID(destinationActionID)
	destinationBridge.GetTransactionsStatusesCalled = func(ctx context.Context, batchId bridge.BatchId) ([]uint8, error) {
		require.Fail(t, "should have not checked the destination bridge for transactions statuses")
		return nil, nil
	}

	sm, err := createAndStartBridge(sourceBridge, destinationBridge, 1, 1, true, "test")
	require.Nil(t, err)

	select {
	case <-chDone:
		_ = sm.Close()
	case <-time.After(time.Second * 5):
		_ = sm.Close()
		require.Fail(t, "timeout while executing the whole process")
	}

	expectedStatuses := bytes.Repeat([]byte{bridge.Rejected}, len(pendingBatch.Transactions))
	checkStatusWhenExecutedOnSource(t, sourceBridge, pendingBatch, sourceActionID, expectedStatuses)
	checkStatusWhenExecutedOnDestination(t, destinationBridge, nil, destinationActionID)
}

func makeMockStatuses(numTxs int) []byte {
	statuses := make([]byte, numTxs)
	for i := 0; i < numTxs; i++ {
		if i%2 == 0 {
			statuses[i] = bridge.Rejected
		} else {
			statuses[i] = bridge.Executed
		}
	}

	return statuses
}

func createAndStartBridge(
	sourceBridge bridge.Bridge,
	destinationBridge bridge.Bridge,
	quorum uint,
	numPeers int,
	isLeader bool,
	name string,
) (io.Closer, error) {
	quorumProvider := &mock.QuorumProviderStub{
		GetQuorumCalled: func(ctx context.Context) (uint, error) {
			return quorum, nil
		},
	}

	topologyProvider := &mock.TopologyProviderStub{
		PeerCountCalled: func() int {
			return numPeers
		},
		AmITheLeaderCalled: func() bool {
			return isLeader
		},
	}

	logExecutor := logger.GetOrCreate(name + "/executor")
	argsExecutor := bridgeExecutors.ArgsEthElrondBridgeExecutor{
		ExecutorName:      name,
		Logger:            logExecutor,
		SourceBridge:      sourceBridge,
		DestinationBridge: destinationBridge,
		TopologyProvider:  topologyProvider,
		QuorumProvider:    quorumProvider,
		Timer:             &mock.TimerMock{},
		DurationsMap:      createMockDurationsMap(),
	}

	bridgeExecutor, err := bridgeExecutors.NewEthElrondBridgeExecutor(argsExecutor)
	if err != nil {
		return nil, err
	}

	stepsMap, err := steps.CreateSteps(bridgeExecutor)
	if err != nil {
		return nil, err
	}

	logStateMachine := logger.GetOrCreate(name + "/statement")
	argsStateMachine := stateMachine.ArgsStateMachine{
		StateMachineName:     name,
		Steps:                stepsMap,
		StartStateIdentifier: ethToElrond.GettingPending,
		DurationBetweenSteps: time.Millisecond,
		Log:                  logStateMachine,
		Timer:                &mock.TimerMock{},
	}

	return stateMachine.NewStateMachine(argsStateMachine)
}

func createMockDurationsMap() map[core.StepIdentifier]time.Duration {
	return map[core.StepIdentifier]time.Duration{
		ethToElrond.GettingPending:                       time.Millisecond,
		ethToElrond.ProposingTransfer:                    time.Millisecond,
		ethToElrond.WaitingSignaturesForProposeTransfer:  time.Millisecond,
		ethToElrond.ExecutingTransfer:                    time.Millisecond,
		ethToElrond.ProposingSetStatus:                   time.Millisecond,
		ethToElrond.WaitingSignaturesForProposeSetStatus: time.Millisecond,
		ethToElrond.ExecutingSetStatus:                   time.Millisecond,
	}
}

func checkStatusWhenExecutedOnSource(
	t *testing.T,
	sourceBridge *mock.BridgeMock,
	pendingBatch *bridge.Batch,
	sourceActionID bridge.ActionId,
	expectedStatuses []byte,
) {
	assert.Nil(t, sourceBridge.GetProposedTransferBatch())

	expectedSignedMapOnSource := map[string]int{
		integrationTests.ActionIdToString(sourceActionID): 1,
	}
	assert.Equal(t, expectedSignedMapOnSource, sourceBridge.SignedActionIDMap())

	executedActionID, executedBatchID := sourceBridge.GetExecuted()
	assert.Equal(t, sourceActionID, executedActionID)
	assert.Equal(t, pendingBatch.Id, executedBatchID)

	proposedStatusBatch := sourceBridge.GetProposedSetStatusBatch()
	require.Equal(t, len(pendingBatch.Transactions), len(proposedStatusBatch.Transactions))
	for i, tx := range proposedStatusBatch.Transactions {
		assert.Equal(t, expectedStatuses[i], tx.Status)
	}

	assert.Nil(t, sourceBridge.GetProposedTransferBatch())
}

func checkStatusWhenExecutedOnDestination(
	t *testing.T,
	destinationBridge *mock.BridgeMock,
	pendingBatch *bridge.Batch,
	destinationActionID bridge.ActionId,
) {
	proposedBatch := integrationTests.CloneBatch(pendingBatch)
	assert.Equal(t, proposedBatch, destinationBridge.GetProposedTransferBatch())

	expectedSignedMapOnSource := map[string]int{
		integrationTests.ActionIdToString(destinationActionID): 1,
	}
	if pendingBatch == nil {
		expectedSignedMapOnSource = nil
	}
	assert.Equal(t, expectedSignedMapOnSource, destinationBridge.SignedActionIDMap())

	executedActionID, executedBatchID := destinationBridge.GetExecuted()
	if pendingBatch != nil {
		assert.Equal(t, destinationActionID, executedActionID)
		assert.Equal(t, pendingBatch.Id, executedBatchID)
	}

	assert.Nil(t, destinationBridge.GetProposedSetStatusBatch())
}
