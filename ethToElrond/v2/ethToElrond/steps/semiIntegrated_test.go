package steps

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/stateMachine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	myTurnAsLeader                                = "myTurnAsLeader"
	getAndStoreActionID                           = "getAndStoreActionID"
	getAndStoreBatchFromEthereum                  = "getAndStoreBatchFromEthereum"
	getStoredBatch                                = "getStoredBatch"
	getLastExecutedEthBatchIDFromElrond           = "getLastExecutedEthBatchIDFromElrond"
	verifyLastDepositNonceExecutedOnEthereumBatch = "verifyLastDepositNonceExecutedOnEthereumBatch"
	wasTransferProposedOnElrond                   = "wasTransferProposedOnElrond"
	wasProposedTransferSigned                     = "wasProposedTransferSigned"
	signProposedTransfer                          = "signProposedTransfer"
	isQuorumReached                               = "isQuorumReached"
	wasActionIDPerformed                          = "wasActionIDPerformed"
	proposeTransferOnElrond                       = "proposeTransferOnElrond"
	performActionID                               = "performActionID"
)

type testMode byte

const (
	transferProposed       testMode = 1
	asLeader               testMode = 2
	proposedTransferSigned testMode = 4
)

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

func testSubFlow1(t *testing.T, bridgeStub *bridgeV2.EthToElrondBridgeStub, failingStep string, mode testMode, numIterations int, numSteps int) {
	wasProposedTransferSignedValue := getWasProposedTransferSignedValue(mode)

	assert.Equal(t, 1, bridgeStub.GetFunctionCounter(getLastExecutedEthBatchIDFromElrond))
	assert.Equal(t, 1, bridgeStub.GetFunctionCounter(getAndStoreBatchFromEthereum))
	assert.Equal(t, 1, bridgeStub.GetFunctionCounter(verifyLastDepositNonceExecutedOnEthereumBatch))
	assert.Equal(t, 1, bridgeStub.GetFunctionCounter(getAndStoreActionID))
	if failingStep == ethToElrond.PerformingActionID || failingStep == ethToElrond.NoFailing {
		assert.Equal(t, 4, bridgeStub.GetFunctionCounter(getStoredBatch))
		assert.Equal(t, 1, bridgeStub.GetFunctionCounter(wasTransferProposedOnElrond))
		assert.Equal(t, numSteps*numIterations-5, bridgeStub.GetFunctionCounter(wasActionIDPerformed))
		assert.Equal(t, numSteps*numIterations-5, bridgeStub.GetFunctionCounter(myTurnAsLeader))
		assert.Equal(t, 1, bridgeStub.GetFunctionCounter(wasProposedTransferSigned))
		if wasProposedTransferSignedValue == false {
			assert.Equal(t, 1, bridgeStub.GetFunctionCounter(signProposedTransfer))
		}

		assert.Equal(t, 1, bridgeStub.GetFunctionCounter(isQuorumReached))
		return
	}
	assert.Equal(t, numSteps*numIterations, bridgeStub.GetFunctionCounter(getStoredBatch))
	assert.Equal(t, numSteps*numIterations-2, bridgeStub.GetFunctionCounter(wasTransferProposedOnElrond))
	assert.Equal(t, numSteps*numIterations-2, bridgeStub.GetFunctionCounter(myTurnAsLeader))
}

func testSubFlow2(t *testing.T, bridgeStub *bridgeV2.EthToElrondBridgeStub, failingStep string, mode testMode, numIterations int, numSteps int) {
	wasTransferProposedOnElrondValue := getWasTransferProposedOnElrondValue(mode)
	myTurnAsLeaderReturnValue := getmyTurnAsLeaderValue(mode)
	wasProposedTransferSignedValue := getWasProposedTransferSignedValue(mode)

	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(getLastExecutedEthBatchIDFromElrond))
	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(getAndStoreBatchFromEthereum))
	if failingStep == ethToElrond.GettingPendingBatchFromEthereum {
		return
	}
	assert.Equal(t, int(math.Min(4, float64(numSteps)))*numIterations, bridgeStub.GetFunctionCounter(getStoredBatch))
	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(verifyLastDepositNonceExecutedOnEthereumBatch))
	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(getAndStoreActionID))
	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(wasTransferProposedOnElrond))
	if wasTransferProposedOnElrondValue == false {
		if failingStep == ethToElrond.PerformingActionID {
			assert.Equal(t, 2*numIterations, bridgeStub.GetFunctionCounter(myTurnAsLeader))
		} else {
			assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(myTurnAsLeader))
		}

		assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(proposeTransferOnElrond))
	}
	if failingStep == ethToElrond.ProposingTransferOnElrond {
		return
	}
	if myTurnAsLeaderReturnValue == false && wasProposedTransferSignedValue == false {
		assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(signProposedTransfer))
	}
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

func checkAfterExecution(t *testing.T, bridgeStub *bridgeV2.EthToElrondBridgeStub, failingStep string, mode testMode, numIterations int, numSteps int) {
	myTurnAsLeaderReturnValue := getmyTurnAsLeaderValue(mode)

	switch failingStep {
	case ethToElrond.GettingPendingBatchFromEthereum:
		testSubFlow2(t, bridgeStub, failingStep, mode, numIterations, numSteps)
	case ethToElrond.ProposingTransferOnElrond, ethToElrond.PerformingActionID:
		if myTurnAsLeaderReturnValue {
			testSubFlow2(t, bridgeStub, failingStep, mode, numIterations, numSteps)
		} else {
			testSubFlow1(t, bridgeStub, failingStep, mode, numIterations, numSteps)
		}
	case ethToElrond.SigningProposedTransferOnElrond:
		transferNotProposedOrAsleaderNotsigned := ^transferProposed | (asLeader & proposedTransferSigned)
		if mode == transferNotProposedOrAsleaderNotsigned {
			testSubFlow2(t, bridgeStub, failingStep, mode, numIterations, numSteps)
		} else {
			testSubFlow1(t, bridgeStub, failingStep, mode, numIterations, numSteps)
		}
	case ethToElrond.WaitingForQuorum:
		transferProposedOrAsLeader := transferProposed | asLeader
		if mode == transferProposedOrAsLeader {
			testSubFlow2(t, bridgeStub, failingStep, mode, numIterations, numSteps)
		} else {
			testSubFlow1(t, bridgeStub, failingStep, mode, numIterations, numSteps)
		}
	case ethToElrond.NoFailing:
		// TODO: finish also the case when no error
		return
	default:
		assert.Error(t, errors.New("unexpected step"))
	}
}

func testFlow(t *testing.T, mode testMode) {
	for numSteps, failingStep := range ethToElrond.FailingStepList {
		currentNumStep, currentFailingStep := numSteps, failingStep
		t.Run(fmt.Sprintf("at %s, mode: %d", failingStep, mode), func(t *testing.T) {
			t.Parallel()
			bridgeStub := createStubExecutorFailingAt(currentFailingStep, mode)
			smm := createStateMachineMock(t, bridgeStub)
			numIterations := 100
			for i := 0; i < numIterations; i++ {
				for stepNo := 0; stepNo <= currentNumStep; stepNo++ {
					err := smm.ExecuteOneStep()
					require.Nil(t, err)
				}
			}
			checkAfterExecution(t, bridgeStub, failingStep, mode, numIterations, currentNumStep+1)
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
		return 0, nil
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
