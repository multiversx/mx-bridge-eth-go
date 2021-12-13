package steps

import (
	"context"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/stateMachine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	myTurnAsLeader                                = "MyTurnAsLeader"
	getAndStoreActionID                           = "GetAndStoreActionIDFromElrond"
	getAndStoreBatchFromEthereum                  = "GetAndStoreBatchFromEthereum"
	getLastExecutedEthBatchIDFromElrond           = "GetLastExecutedEthBatchIDFromElrond"
	verifyLastDepositNonceExecutedOnEthereumBatch = "VerifyLastDepositNonceExecutedOnEthereumBatch"
	wasTransferProposedOnElrond                   = "WasTransferProposedOnElrond"
	wasProposedTransferSigned                     = "WasProposedTransferSignedOnElrond"
	signProposedTransfer                          = "SignProposedTransferOnElrond"
	processMaxRetriesOnElrond                     = "ProcessMaxRetriesOnElrond"
	resetRetriesCountOnElrond                     = "ResetRetriesCountOnElrond"
	isQuorumReached                               = "IsQuorumReachedOnElrond"
	wasActionIDPerformed                          = "WasActionIDPerformedOnElrond"
	proposeTransferOnElrond                       = "ProposeTransferOnElrond"
	performActionID                               = "PerformActionIDOnElrond"
)

type testMode byte

const (
	transferProposed       testMode = 1
	asLeader               testMode = 2
	proposedTransferSigned testMode = 4
)

const maxRetriesAllowed = 3

var retriesNumber = 0

func getmyTurnAsLeaderValue(mode testMode) bool {
	return mode&asLeader > 0
}

func getWasTransferProposedOnElrondValue(mode testMode) bool {
	return mode&transferProposed > 0
}

func getWasProposedTransferSignedValue(mode testMode) bool {
	return mode&proposedTransferSigned > 0
}

func createStateMachineMock(t *testing.T, bridgeStub *bridgeV2.EthToElrondBridgeStub) *stateMachine.StateMachineMock {
	steps, err := CreateSteps(bridgeStub)
	require.Nil(t, err)
	smm := stateMachine.NewStateMachineMock(steps, ethToElrond.GettingPendingBatchFromEthereum)
	err = smm.Initialize()
	require.Nil(t, err)
	return smm
}

func testSubFlow(t *testing.T, bridgeStub *bridgeV2.EthToElrondBridgeStub, failingStep string, mode testMode, numIterations int) {
	wasTransferProposedOnElrondValue := getWasTransferProposedOnElrondValue(mode)
	wasProposedTransferSignedValue := getWasProposedTransferSignedValue(mode)

	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(getLastExecutedEthBatchIDFromElrond))
	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(getAndStoreBatchFromEthereum))
	if failingStep == ethToElrond.GettingPendingBatchFromEthereum {
		return
	}
	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(verifyLastDepositNonceExecutedOnEthereumBatch))
	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(wasTransferProposedOnElrond))
	if wasTransferProposedOnElrondValue == false {
		if failingStep == ethToElrond.PerformingActionID {
			assert.Equal(t, 2*numIterations, bridgeStub.GetFunctionCounter(myTurnAsLeader))
		} else {
			assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(myTurnAsLeader))
		}

		assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(proposeTransferOnElrond))
	} else if failingStep == ethToElrond.PerformingActionID {
		assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(myTurnAsLeader))
	}
	if failingStep == ethToElrond.ProposingTransferOnElrond {
		return
	}
	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(wasProposedTransferSigned))
	if wasProposedTransferSignedValue == false {
		assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(signProposedTransfer))
	}
	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(getAndStoreActionID))
	if failingStep == ethToElrond.SigningProposedTransferOnElrond {
		return
	}
	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(isQuorumReached))
	if failingStep == ethToElrond.WaitingForQuorum {
		return
	}
	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(wasActionIDPerformed))
	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(performActionID))
}

func testFlow(t *testing.T, mode testMode) {
	for numSteps, failingStep := range ethToElrond.FailingStepList {
		currentNumStep, currentFailingStep := numSteps, failingStep
		// Avoid invalid test cases
		testSigningFailureWhileTransferAlreadySigned := getWasProposedTransferSignedValue(mode) && currentFailingStep == ethToElrond.SigningProposedTransferOnElrond
		testProposingFailureWhileTransferAlreadyProposed := getWasTransferProposedOnElrondValue(mode) && currentFailingStep == ethToElrond.ProposingTransferOnElrond
		testProposingWhileNotLeader := !getmyTurnAsLeaderValue(mode) && currentFailingStep == ethToElrond.ProposingTransferOnElrond
		testPerformingActionWhileNotLeader := !getmyTurnAsLeaderValue(mode) && currentFailingStep == ethToElrond.PerformingActionID
		testWaitingQuorumWhileNotLeaderAndNotAlreadyProposed := !getmyTurnAsLeaderValue(mode) && !getWasTransferProposedOnElrondValue(mode) && currentFailingStep == ethToElrond.WaitingForQuorum
		if testSigningFailureWhileTransferAlreadySigned ||
			testProposingFailureWhileTransferAlreadyProposed ||
			testProposingWhileNotLeader ||
			testPerformingActionWhileNotLeader ||
			testWaitingQuorumWhileNotLeaderAndNotAlreadyProposed {
			continue
		}
		if currentFailingStep == ethToElrond.NoFailing {
			// TODO implement this
			continue
		}
		t.Run(fmt.Sprintf("at %s, mode: %d", failingStep, mode), func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorFailingAt(currentFailingStep, mode)
			smm := createStateMachineMock(t, bridgeStub)
			numIterations := 100
			for i := 0; i < numIterations; i++ {
				retriesNumber = 0
				for stepNo := 0; stepNo <= currentNumStep; stepNo++ {
					err := smm.ExecuteOneStep()
					require.Nil(t, err)
				}
			}
			testSubFlow(t, bridgeStub, currentFailingStep, mode, numIterations)
		})
	}
}

func createStubExecutorFailingAt(failingStep string, mode testMode) *bridgeV2.EthToElrondBridgeStub {
	bridgeStub := createStubExecutor()
	bridgeStub.GetLastExecutedEthBatchIDFromElrondCalled = func(ctx context.Context) (uint64, error) {
		return 1122, nil
	}
	if failingStep == ethToElrond.GettingPendingBatchFromEthereum {
		bridgeStub.GetAndStoreBatchFromEthereumCalled = func(ctx context.Context, nonce uint64) error {
			return expectedError
		}
		return bridgeStub
	} else {
		bridgeStub.GetAndStoreBatchFromEthereumCalled = func(ctx context.Context, nonce uint64) error {
			return nil
		}
		bridgeStub.GetStoredBatchCalled = func() *clients.TransferBatch {
			return testBatch
		}
		bridgeStub.VerifyLastDepositNonceExecutedOnEthereumBatchCalled = func(ctx context.Context) error {
			return nil
		}
	}

	bridgeStub.GetAndStoreActionIDFromElrondCalled = func(ctx context.Context) (uint64, error) {
		return 2, nil
	}
	bridgeStub.WasTransferProposedOnElrondCalled = func(ctx context.Context) (bool, error) {
		return getWasTransferProposedOnElrondValue(mode), nil
	}
	bridgeStub.MyTurnAsLeaderCalled = func() bool {
		return getmyTurnAsLeaderValue(mode)
	}
	bridgeStub.WasProposedTransferSignedOnElrondCalled = func(ctx context.Context) (bool, error) {
		return getWasProposedTransferSignedValue(mode), nil
	}
	if failingStep == ethToElrond.ProposingTransferOnElrond {
		bridgeStub.ProposeTransferOnElrondCalled = func(ctx context.Context) error {
			return expectedError
		}
		return bridgeStub
	} else {
		bridgeStub.ProposeTransferOnElrondCalled = func(ctx context.Context) error {
			return nil
		}
	}
	if failingStep == ethToElrond.SigningProposedTransferOnElrond {
		bridgeStub.SignProposedTransferOnElrondCalled = func(ctx context.Context) error {
			return expectedError
		}
		return bridgeStub
	} else {
		bridgeStub.SignProposedTransferOnElrondCalled = func(ctx context.Context) error {
			return nil
		}
	}
	bridgeStub.ProcessMaxRetriesOnElrondCalled = func() bool {
		if retriesNumber < maxRetriesAllowed {
			retriesNumber++
			return false
		}
		return true
	}
	bridgeStub.ResetRetriesCountOnElrondCalled = func() {
		retriesNumber = 0
	}
	if failingStep == ethToElrond.WaitingForQuorum {
		bridgeStub.IsQuorumReachedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return false, expectedError
		}
		return bridgeStub
	} else {
		bridgeStub.IsQuorumReachedOnElrondCalled = func(ctx context.Context) (bool, error) {
			return true, nil
		}
	}
	bridgeStub.WasActionIDPerformedOnElrondCalled = func(ctx context.Context) (bool, error) {
		return false, nil
	}
	if failingStep == ethToElrond.PerformingActionID {
		bridgeStub.PerformActionIDOnElrondCalled = func(ctx context.Context) error {
			return expectedError
		}
		return bridgeStub
	} else {
		bridgeStub.PerformActionIDOnElrondCalled = func(ctx context.Context) error {
			return nil
		}
	}
	return bridgeStub
}

func TestEndlessLoop(t *testing.T) {
	t.Parallel()
	numModes := testMode(8)
	for mode := testMode(1); mode < numModes; mode++ {
		testFlow(t, mode)
	}
}
