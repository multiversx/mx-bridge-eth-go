package steps

import (
	"context"
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
	GetLogger                                     = "GetLogger"
	MyTurnAsLeader                                = "MyTurnAsLeader"
	GetAndStoreActionID                           = "GetAndStoreActionID"
	GetAndStoreBatchFromEthereum                  = "GetAndStoreBatchFromEthereum"
	GetStoredBatch                                = "GetStoredBatch"
	GetLastExecutedEthBatchIDFromElrond           = "GetLastExecutedEthBatchIDFromElrond"
	VerifyLastDepositNonceExecutedOnEthereumBatch = "VerifyLastDepositNonceExecutedOnEthereumBatch"
	WasTransferProposedOnElrond                   = "WasTransferProposedOnElrond"
	WasProposedTransferSigned                     = "WasProposedTransferSigned"
	SignProposedTransfer                          = "SignProposedTransfer"
	IsQuorumReached                               = "IsQuorumReached"
	WasActionIDPerformed                          = "WasActionIDPerformed"
	ProposeTransferOnElrond                       = "ProposeTransferOnElrond"
	PerformActionID                               = "PerformActionID"
)

type testMode byte

const (
	transferProposed       testMode = 1
	asLeader               testMode = 2
	proposedTransferSigned testMode = 4
)

func myTurnAsLeader(mode testMode) bool {
	return mode&asLeader > 0
}

func wasTransferProposedOnElrond(mode testMode) bool {
	return mode&transferProposed > 0
}

func wasProposedTransferSigned(mode testMode) bool {
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

func d1(t *testing.T, bridgeStub *bridgeV2.EthToElrondBridgeStub, numIterations int, numSteps int) {
	assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetLastExecutedEthBatchIDFromElrond))
	assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetAndStoreBatchFromEthereum))
	assert.Equal(t, numSteps*numIterations, bridgeStub.GetFunctionCounter(GetStoredBatch))
	assert.Equal(t, 2, bridgeStub.GetFunctionCounter(GetLogger))
	assert.Equal(t, 1, bridgeStub.GetFunctionCounter(VerifyLastDepositNonceExecutedOnEthereumBatch))
	assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetAndStoreActionID))
	assert.Equal(t, numSteps*numIterations-2, bridgeStub.GetFunctionCounter(WasTransferProposedOnElrond))
	assert.Equal(t, numSteps*numIterations-2, bridgeStub.GetFunctionCounter(MyTurnAsLeader))
}

func d1p(t *testing.T, bridgeStub *bridgeV2.EthToElrondBridgeStub, numIterations int, numSteps int, mode testMode) {

	// wasTransferProposedOnElrondValue := wasTransferProposedOnElrond(mode)
	// myTurnAsLeaderReturnValue := myTurnAsLeader(mode)
	wasProposedTransferSignedValue := wasProposedTransferSigned(mode)

	assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetLastExecutedEthBatchIDFromElrond))
	assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetAndStoreBatchFromEthereum))
	assert.Equal(t, 4, bridgeStub.GetFunctionCounter(GetStoredBatch))
	assert.Equal(t, 3, bridgeStub.GetFunctionCounter(GetLogger))
	assert.Equal(t, 1, bridgeStub.GetFunctionCounter(VerifyLastDepositNonceExecutedOnEthereumBatch))
	assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetAndStoreActionID))
	assert.Equal(t, numSteps*numIterations-5, bridgeStub.GetFunctionCounter(WasActionIDPerformed))
	assert.Equal(t, numSteps*numIterations-5, bridgeStub.GetFunctionCounter(MyTurnAsLeader))
	assert.Equal(t, 1, bridgeStub.GetFunctionCounter(WasProposedTransferSigned))
	assert.Equal(t, 1, bridgeStub.GetFunctionCounter(WasTransferProposedOnElrond))
	if wasProposedTransferSignedValue == false {
		assert.Equal(t, 1, bridgeStub.GetFunctionCounter(SignProposedTransfer))
	}

	assert.Equal(t, 1, bridgeStub.GetFunctionCounter(IsQuorumReached))
}

func d2(t *testing.T, bridgeStub *bridgeV2.EthToElrondBridgeStub, failingStep string, mode testMode, numIterations int, numSteps int) {

	wasTransferProposedOnElrondValue := wasTransferProposedOnElrond(mode)
	myTurnAsLeaderReturnValue := myTurnAsLeader(mode)
	wasProposedTransferSignedValue := wasProposedTransferSigned(mode)

	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(GetLastExecutedEthBatchIDFromElrond))
	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(GetAndStoreBatchFromEthereum))
	// assert.Equal(t, int(math.Min(3, float64(numSteps)))*numIterations, bridgeStub.GetFunctionCounter(GetLogger))
	if failingStep == ethToElrond.GettingPendingBatchFromEthereum {
		return
	}
	assert.Equal(t, int(math.Min(4, float64(numSteps)))*numIterations, bridgeStub.GetFunctionCounter(GetStoredBatch))
	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(VerifyLastDepositNonceExecutedOnEthereumBatch))
	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(GetAndStoreActionID))
	if failingStep == ethToElrond.GettingActionIdForProposeTransfer {
		return
	}
	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(WasTransferProposedOnElrond))

	if wasTransferProposedOnElrondValue == false {
		if failingStep == ethToElrond.PerformingActionID {
			assert.Equal(t, 2*numIterations, bridgeStub.GetFunctionCounter(MyTurnAsLeader))
		} else {
			assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(MyTurnAsLeader))
		}

		assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(ProposeTransferOnElrond))
	}

	if failingStep == ethToElrond.ProposingTransferOnElrond {
		return
	}

	if myTurnAsLeaderReturnValue == false && wasProposedTransferSignedValue == false {
		assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(SignProposedTransfer))
	}

	if failingStep == ethToElrond.SigningProposedTransferOnElrond {
		return
	}

	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(IsQuorumReached))

	if failingStep == ethToElrond.WaitingForQuorum {
		return
	}

	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(WasActionIDPerformed))
	assert.Equal(t, numIterations, bridgeStub.GetFunctionCounter(PerformActionID))

}

func checkAfterExecution(t *testing.T, bridgeStub *bridgeV2.EthToElrondBridgeStub, failingStep string, mode testMode, numIterations int, numSteps int) {

	wasTransferProposedOnElrondValue := wasTransferProposedOnElrond(mode)
	myTurnAsLeaderReturnValue := myTurnAsLeader(mode)
	wasProposedTransferSignedValue := wasProposedTransferSigned(mode)

	switch failingStep {
	case ethToElrond.GettingPendingBatchFromEthereum:
		d2(t, bridgeStub, failingStep, mode, numIterations, numSteps)
	case ethToElrond.GettingActionIdForProposeTransfer:
		d2(t, bridgeStub, failingStep, mode, numIterations, numSteps)
	case ethToElrond.ProposingTransferOnElrond:
		if wasTransferProposedOnElrondValue == true {
			return
		}
		if myTurnAsLeaderReturnValue == true {
			d2(t, bridgeStub, failingStep, mode, numIterations, numSteps)
		} else {
			d1(t, bridgeStub, numIterations, numSteps)
		}
	case ethToElrond.SigningProposedTransferOnElrond:
		if wasTransferProposedOnElrondValue == true {
			if myTurnAsLeaderReturnValue == true {
				if wasProposedTransferSignedValue == true {
					return
				} else {
					d2(t, bridgeStub, failingStep, mode, numIterations, numSteps)
				}
			} else {
				if wasProposedTransferSignedValue == true {
					return
				} else {
					d2(t, bridgeStub, failingStep, mode, numIterations, numSteps)
				}
			}
		} else {
			if myTurnAsLeaderReturnValue == true {
				if wasProposedTransferSignedValue == true {
					return
				} else {
					d2(t, bridgeStub, failingStep, mode, numIterations, numSteps)
				}
			} else {
				d1(t, bridgeStub, numIterations, numSteps)
			}
		}
	case ethToElrond.WaitingForQuorum:
		if wasTransferProposedOnElrondValue == true {
			d2(t, bridgeStub, failingStep, mode, numIterations, numSteps)
		} else {
			if myTurnAsLeaderReturnValue == true {
				d2(t, bridgeStub, failingStep, mode, numIterations, numSteps)
			} else {
				d1(t, bridgeStub, numIterations, numSteps)
			}
		}
	case ethToElrond.PerformingActionID:
		if wasTransferProposedOnElrondValue == true {
			if myTurnAsLeaderReturnValue == true {
				d2(t, bridgeStub, failingStep, mode, numIterations, numSteps)
			} else {
				d1p(t, bridgeStub, numIterations, numSteps, mode)
			}
		} else {
			if myTurnAsLeaderReturnValue == true {
				d2(t, bridgeStub, failingStep, mode, numIterations, numSteps)
			} else {
				d1(t, bridgeStub, numIterations, numSteps)
			}
		}
	}
}
func testFlow(t *testing.T, mode testMode) {
	for numSteps, failingStep := range ethToElrond.StepList {
		t.Run(fmt.Sprintf("at %s, mode: %d", failingStep, mode), func(t *testing.T) {
			bridgeStub := createStubExecutorFailingAt(failingStep, mode)

			smm := createStateMachineMock(t, bridgeStub)

			numIterations := 100
			for i := 0; i < numIterations; i++ {
				for stepNo := 0; stepNo <= numSteps; stepNo++ {
					err := smm.ExecuteOneStep()
					require.Nil(t, err)
				}
			}

			checkAfterExecution(t, bridgeStub, failingStep, mode, numIterations, numSteps+1)
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

	if failingStep == ethToElrond.GettingActionIdForProposeTransfer {
		bridgeStub.GetAndStoreActionIDCalled = func(ctx context.Context) (uint64, error) {
			return 1122, expectedError
		}
		return bridgeStub
	} else {
		bridgeStub.GetAndStoreActionIDCalled = func(ctx context.Context) (uint64, error) {
			return 0, nil
		}
	}

	bridgeStub.WasTransferProposedOnElrondCalled = func(ctx context.Context) (bool, error) {
		return wasTransferProposedOnElrond(mode), nil
	}
	bridgeStub.MyTurnAsLeaderCalled = func() bool {
		return myTurnAsLeader(mode)
	}

	bridgeStub.WasProposedTransferSignedCalled = func(ctx context.Context) (bool, error) {
		return wasProposedTransferSigned(mode), nil
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
		bridgeStub.SignProposedTransferCalled = func(ctx context.Context) error {
			return expectedError
		}
		return bridgeStub
	} else {
		bridgeStub.SignProposedTransferCalled = func(ctx context.Context) error {
			return nil
		}
	}

	if failingStep == ethToElrond.WaitingForQuorum {
		bridgeStub.IsQuorumReachedCalled = func(ctx context.Context) (bool, error) {
			return false, expectedError
		}
		return bridgeStub
	} else {
		bridgeStub.IsQuorumReachedCalled = func(ctx context.Context) (bool, error) {
			return true, nil
		}
	}

	bridgeStub.WasActionIDPerformedCalled = func(ctx context.Context) (bool, error) {
		return false, nil
	}

	if failingStep == ethToElrond.PerformingActionID {
		bridgeStub.PerformActionIDCalled = func(ctx context.Context) error {
			return expectedError
		}
		return bridgeStub
	} else {
		bridgeStub.PerformActionIDCalled = func(ctx context.Context) error {
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
