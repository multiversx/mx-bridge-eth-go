package steps

import (
	"context"
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

func createStateMachineMock(t *testing.T, bridgeStub *bridgeV2.EthToElrondBridgeStub) *stateMachine.StateMachineMock {
	steps, err := CreateSteps(bridgeStub)
	require.Nil(t, err)
	smm := stateMachine.NewStateMachineMock(steps, ethToElrond.GettingPendingBatchFromEthereum)
	err = smm.Initialize()
	require.Nil(t, err)
	return smm
}

func testFlow(t *testing.T, WasTransferProposedOnElrondValue bool, MyTurnAsLeaderReturnValue bool, WasProposedTransferSignedValue bool) {
	t.Run("at GetPending", func(t *testing.T) {
		bridgeStub := createStubExecutorFailingAt(ethToElrond.GettingPendingBatchFromEthereum, WasTransferProposedOnElrondValue, MyTurnAsLeaderReturnValue, WasProposedTransferSignedValue)

		smm := createStateMachineMock(t, bridgeStub)

		numSteps := 100
		for i := 0; i < numSteps; i++ {
			err := smm.ExecuteOneStep()
			require.Nil(t, err)
		}
		assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetLastExecutedEthBatchIDFromElrond))
		assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetAndStoreBatchFromEthereum))
		assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetLogger))
	})
	t.Run("at GetActionIdForPropose", func(t *testing.T) {
		bridgeStub := createStubExecutorFailingAt(ethToElrond.GettingActionIdForProposeTransfer, WasTransferProposedOnElrondValue, MyTurnAsLeaderReturnValue, WasProposedTransferSignedValue)

		smm := createStateMachineMock(t, bridgeStub)

		numSteps := 100
		for i := 0; i < numSteps; i++ {
			err := smm.ExecuteOneStep()
			require.Nil(t, err)
			err = smm.ExecuteOneStep()
			require.Nil(t, err)
		}

		assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetLastExecutedEthBatchIDFromElrond))
		assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetAndStoreBatchFromEthereum))
		assert.Equal(t, 2*numSteps, bridgeStub.GetFunctionCounter(GetStoredBatch))
		assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(VerifyLastDepositNonceExecutedOnEthereumBatch))
		assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetAndStoreActionID))
	})

	t.Run("at ProposeTransferOnElrond", func(t *testing.T) {
		if WasTransferProposedOnElrondValue == true {
			return
		}
		bridgeStub := createStubExecutorFailingAt(ethToElrond.ProposingTransferOnElrond, WasTransferProposedOnElrondValue, MyTurnAsLeaderReturnValue, WasProposedTransferSignedValue)

		smm := createStateMachineMock(t, bridgeStub)

		numSteps := 100
		for i := 0; i < numSteps; i++ {
			err := smm.ExecuteOneStep()
			require.Nil(t, err)
			err = smm.ExecuteOneStep()
			require.Nil(t, err)
			err = smm.ExecuteOneStep()
			require.Nil(t, err)
		}

		if MyTurnAsLeaderReturnValue == true {
			assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetLastExecutedEthBatchIDFromElrond))
			assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetAndStoreBatchFromEthereum))
			assert.Equal(t, 3*numSteps, bridgeStub.GetFunctionCounter(GetStoredBatch))
			assert.Equal(t, 3*numSteps, bridgeStub.GetFunctionCounter(GetLogger))
			assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(VerifyLastDepositNonceExecutedOnEthereumBatch))
			assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetAndStoreActionID))
			assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(WasTransferProposedOnElrond))
			assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(MyTurnAsLeader))
			assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(ProposeTransferOnElrond))
		} else {
			assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetLastExecutedEthBatchIDFromElrond))
			assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetAndStoreBatchFromEthereum))
			assert.Equal(t, 3*numSteps, bridgeStub.GetFunctionCounter(GetStoredBatch))
			assert.Equal(t, 2, bridgeStub.GetFunctionCounter(GetLogger))
			assert.Equal(t, 1, bridgeStub.GetFunctionCounter(VerifyLastDepositNonceExecutedOnEthereumBatch))
			assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetAndStoreActionID))
			assert.Equal(t, 3*numSteps-2, bridgeStub.GetFunctionCounter(WasTransferProposedOnElrond))
			assert.Equal(t, 3*numSteps-2, bridgeStub.GetFunctionCounter(MyTurnAsLeader))
		}

	})
	t.Run("at SignProposedTransferOnElrond", func(t *testing.T) {
		bridgeStub := createStubExecutorFailingAt(ethToElrond.SigningProposedTransferOnElrond, WasTransferProposedOnElrondValue, MyTurnAsLeaderReturnValue, WasProposedTransferSignedValue)

		smm := createStateMachineMock(t, bridgeStub)
		numSteps := 100
		for i := 0; i < numSteps; i++ {
			err := smm.ExecuteOneStep()
			require.Nil(t, err)
			err = smm.ExecuteOneStep()
			require.Nil(t, err)
			err = smm.ExecuteOneStep()
			require.Nil(t, err)
			err = smm.ExecuteOneStep()
			require.Nil(t, err)
		}

		if WasTransferProposedOnElrondValue == true {
			if MyTurnAsLeaderReturnValue == true {
				if WasProposedTransferSignedValue == true {
					return
				} else {
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetLastExecutedEthBatchIDFromElrond))
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetAndStoreBatchFromEthereum))
					assert.Equal(t, 4*numSteps, bridgeStub.GetFunctionCounter(GetStoredBatch))
					assert.Equal(t, 3*numSteps, bridgeStub.GetFunctionCounter(GetLogger))
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(VerifyLastDepositNonceExecutedOnEthereumBatch))
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetAndStoreActionID))
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(WasTransferProposedOnElrond))
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(SignProposedTransfer))
				}
			} else {
				if WasProposedTransferSignedValue == true {
					return
				} else {
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetLastExecutedEthBatchIDFromElrond))
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetAndStoreBatchFromEthereum))
					assert.Equal(t, 4*numSteps, bridgeStub.GetFunctionCounter(GetStoredBatch))
					assert.Equal(t, 3*numSteps, bridgeStub.GetFunctionCounter(GetLogger))
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(VerifyLastDepositNonceExecutedOnEthereumBatch))
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetAndStoreActionID))
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(WasTransferProposedOnElrond))
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(SignProposedTransfer))
				}
			}
		} else {
			if MyTurnAsLeaderReturnValue == true {
				if WasProposedTransferSignedValue == true {
					return
				} else {
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetLastExecutedEthBatchIDFromElrond))
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetAndStoreBatchFromEthereum))
					assert.Equal(t, 4*numSteps, bridgeStub.GetFunctionCounter(GetStoredBatch))
					assert.Equal(t, 3*numSteps, bridgeStub.GetFunctionCounter(GetLogger))
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(VerifyLastDepositNonceExecutedOnEthereumBatch))
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetAndStoreActionID))
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(WasTransferProposedOnElrond))
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(MyTurnAsLeader))
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(ProposeTransferOnElrond))
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(SignProposedTransfer))
				}
			} else {
				assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetLastExecutedEthBatchIDFromElrond))
				assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetAndStoreBatchFromEthereum))
				assert.Equal(t, 4*numSteps, bridgeStub.GetFunctionCounter(GetStoredBatch))
				assert.Equal(t, 2, bridgeStub.GetFunctionCounter(GetLogger))
				assert.Equal(t, 1, bridgeStub.GetFunctionCounter(VerifyLastDepositNonceExecutedOnEthereumBatch))
				assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetAndStoreActionID))
				assert.Equal(t, 4*numSteps-2, bridgeStub.GetFunctionCounter(WasTransferProposedOnElrond))
				assert.Equal(t, 4*numSteps-2, bridgeStub.GetFunctionCounter(MyTurnAsLeader))
			}
		}
	})
	t.Run("at WaitForQuorum", func(t *testing.T) {
		bridgeStub := createStubExecutorFailingAt(ethToElrond.WaitingForQuorum, WasTransferProposedOnElrondValue, MyTurnAsLeaderReturnValue, WasProposedTransferSignedValue)

		smm := createStateMachineMock(t, bridgeStub)

		numSteps := 100
		for i := 0; i < numSteps; i++ {
			err := smm.ExecuteOneStep()
			require.Nil(t, err)
			err = smm.ExecuteOneStep()
			require.Nil(t, err)
			err = smm.ExecuteOneStep()
			require.Nil(t, err)
			err = smm.ExecuteOneStep()
			require.Nil(t, err)
			err = smm.ExecuteOneStep()
			require.Nil(t, err)
		}

		if WasTransferProposedOnElrondValue == true {
			assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetLastExecutedEthBatchIDFromElrond))
			assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetAndStoreBatchFromEthereum))
			assert.Equal(t, 4*numSteps, bridgeStub.GetFunctionCounter(GetStoredBatch))
			assert.Equal(t, 3*numSteps, bridgeStub.GetFunctionCounter(GetLogger))
			assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(VerifyLastDepositNonceExecutedOnEthereumBatch))
			assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetAndStoreActionID))
			assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(WasTransferProposedOnElrond))
			assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(WasProposedTransferSigned))
			if WasProposedTransferSignedValue == false {
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(SignProposedTransfer))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(IsQuorumReached))
			}
		} else {
			if MyTurnAsLeaderReturnValue == true {
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetLastExecutedEthBatchIDFromElrond))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetAndStoreBatchFromEthereum))
				assert.Equal(t, 4*numSteps, bridgeStub.GetFunctionCounter(GetStoredBatch))
				assert.Equal(t, 3*numSteps, bridgeStub.GetFunctionCounter(GetLogger))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(VerifyLastDepositNonceExecutedOnEthereumBatch))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetAndStoreActionID))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(WasTransferProposedOnElrond))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(MyTurnAsLeader))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(ProposeTransferOnElrond))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(WasProposedTransferSigned))
				if WasProposedTransferSignedValue == false {
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(SignProposedTransfer))
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(IsQuorumReached))
				}
			} else {
				assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetLastExecutedEthBatchIDFromElrond))
				assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetAndStoreBatchFromEthereum))
				assert.Equal(t, 5*numSteps, bridgeStub.GetFunctionCounter(GetStoredBatch))
				assert.Equal(t, 2, bridgeStub.GetFunctionCounter(GetLogger))
				assert.Equal(t, 1, bridgeStub.GetFunctionCounter(VerifyLastDepositNonceExecutedOnEthereumBatch))
				assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetAndStoreActionID))
				assert.Equal(t, 5*numSteps-2, bridgeStub.GetFunctionCounter(WasTransferProposedOnElrond))
				assert.Equal(t, 5*numSteps-2, bridgeStub.GetFunctionCounter(MyTurnAsLeader))
			}
		}
	})
	t.Run("at PerformActionID", func(t *testing.T) {
		bridgeStub := createStubExecutorFailingAt(ethToElrond.PerformingActionID, WasTransferProposedOnElrondValue, MyTurnAsLeaderReturnValue, WasProposedTransferSignedValue)

		smm := createStateMachineMock(t, bridgeStub)

		numSteps := 100
		for i := 0; i < numSteps; i++ {
			err := smm.ExecuteOneStep()
			require.Nil(t, err)
			err = smm.ExecuteOneStep()
			require.Nil(t, err)
			err = smm.ExecuteOneStep()
			require.Nil(t, err)
			err = smm.ExecuteOneStep()
			require.Nil(t, err)
			err = smm.ExecuteOneStep()
			require.Nil(t, err)
			err = smm.ExecuteOneStep()
			require.Nil(t, err)
		}

		if WasTransferProposedOnElrondValue == true {
			if MyTurnAsLeaderReturnValue == true {
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetLastExecutedEthBatchIDFromElrond))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetAndStoreBatchFromEthereum))
				assert.Equal(t, 4*numSteps, bridgeStub.GetFunctionCounter(GetStoredBatch))
				assert.Equal(t, 4*numSteps, bridgeStub.GetFunctionCounter(GetLogger))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(VerifyLastDepositNonceExecutedOnEthereumBatch))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetAndStoreActionID))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(WasTransferProposedOnElrond))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(MyTurnAsLeader))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(WasProposedTransferSigned))
				if WasProposedTransferSignedValue == false {
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(SignProposedTransfer))
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(IsQuorumReached))
				}
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(WasActionIDPerformed))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(PerformActionID))
			} else {
				assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetLastExecutedEthBatchIDFromElrond))
				assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetAndStoreBatchFromEthereum))
				assert.Equal(t, 4, bridgeStub.GetFunctionCounter(GetStoredBatch))
				assert.Equal(t, 3, bridgeStub.GetFunctionCounter(GetLogger))
				assert.Equal(t, 1, bridgeStub.GetFunctionCounter(VerifyLastDepositNonceExecutedOnEthereumBatch))
				assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetAndStoreActionID))
				assert.Equal(t, 1, bridgeStub.GetFunctionCounter(WasTransferProposedOnElrond))
				assert.Equal(t, 6*numSteps-5, bridgeStub.GetFunctionCounter(MyTurnAsLeader))
				assert.Equal(t, 6*numSteps-5, bridgeStub.GetFunctionCounter(WasActionIDPerformed))
				if WasProposedTransferSignedValue == false {
					assert.Equal(t, 1, bridgeStub.GetFunctionCounter(SignProposedTransfer))
					assert.Equal(t, 1, bridgeStub.GetFunctionCounter(IsQuorumReached))
				}
			}
		} else {
			if MyTurnAsLeaderReturnValue == true {
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetLastExecutedEthBatchIDFromElrond))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetAndStoreBatchFromEthereum))
				assert.Equal(t, 4*numSteps, bridgeStub.GetFunctionCounter(GetStoredBatch))
				assert.Equal(t, 4*numSteps, bridgeStub.GetFunctionCounter(GetLogger))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(VerifyLastDepositNonceExecutedOnEthereumBatch))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(GetAndStoreActionID))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(WasTransferProposedOnElrond))
				assert.Equal(t, 2*numSteps, bridgeStub.GetFunctionCounter(MyTurnAsLeader))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(ProposeTransferOnElrond))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(WasProposedTransferSigned))
				if WasProposedTransferSignedValue == false {
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(SignProposedTransfer))
					assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(IsQuorumReached))
				}
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(WasActionIDPerformed))
				assert.Equal(t, numSteps, bridgeStub.GetFunctionCounter(PerformActionID))
			} else {
				assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetLastExecutedEthBatchIDFromElrond))
				assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetAndStoreBatchFromEthereum))
				assert.Equal(t, 6*numSteps, bridgeStub.GetFunctionCounter(GetStoredBatch))
				assert.Equal(t, 2, bridgeStub.GetFunctionCounter(GetLogger))
				assert.Equal(t, 1, bridgeStub.GetFunctionCounter(VerifyLastDepositNonceExecutedOnEthereumBatch))
				assert.Equal(t, 1, bridgeStub.GetFunctionCounter(GetAndStoreActionID))
				assert.Equal(t, 6*numSteps-2, bridgeStub.GetFunctionCounter(WasTransferProposedOnElrond))
				assert.Equal(t, 6*numSteps-2, bridgeStub.GetFunctionCounter(MyTurnAsLeader))
			}
		}
	})
}

func createStubExecutorFailingAt(failingStep string, WasTransferProposedOnElrondValue bool, MyTurnAsLeaderValue bool, WasProposedTransferSignedValue bool) *bridgeV2.EthToElrondBridgeStub {
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
			return &clients.TransferBatch{
				ID:       112233,
				Deposits: nil,
				Statuses: nil,
			}
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
		return WasTransferProposedOnElrondValue, nil
	}
	bridgeStub.MyTurnAsLeaderCalled = func() bool {
		return MyTurnAsLeaderValue
	}

	bridgeStub.WasProposedTransferSignedCalled = func(ctx context.Context) (bool, error) {
		return WasProposedTransferSignedValue, nil
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
	t.Run("WasTransferProposedOnElrond false", func(t *testing.T) {
		WasTransferProposedOnElrondValue := false
		t.Run("MyTurnAsLeader false", func(t *testing.T) {
			WasProposedTransferSignedValue := false
			t.Run("WasProposedTransferSignedValue false", func(t *testing.T) {
				testFlow(t, WasTransferProposedOnElrondValue, WasProposedTransferSignedValue, false)
			})
			t.Run("WasProposedTransferSignedValue true", func(t *testing.T) {
				testFlow(t, WasTransferProposedOnElrondValue, WasProposedTransferSignedValue, true)
			})
		})
		t.Run("MyTurnAsLeader true", func(t *testing.T) {
			WasProposedTransferSignedValue := true
			t.Run("WasProposedTransferSignedValue false", func(t *testing.T) {
				testFlow(t, WasTransferProposedOnElrondValue, WasProposedTransferSignedValue, false)
			})
			t.Run("WasProposedTransferSignedValue true", func(t *testing.T) {
				testFlow(t, WasTransferProposedOnElrondValue, WasProposedTransferSignedValue, true)
			})
		})
	})
	t.Run("WasTransferProposedOnElrond true", func(t *testing.T) {
		WasTransferProposedOnElrondValue := true
		t.Run("MyTurnAsLeader false", func(t *testing.T) {
			WasProposedTransferSignedValue := false
			t.Run("WasProposedTransferSignedValue false", func(t *testing.T) {
				testFlow(t, WasTransferProposedOnElrondValue, WasProposedTransferSignedValue, false)
			})
			t.Run("WasProposedTransferSignedValue true", func(t *testing.T) {
				testFlow(t, WasTransferProposedOnElrondValue, WasProposedTransferSignedValue, true)
			})
		})
		t.Run("MyTurnAsLeader true", func(t *testing.T) {
			WasProposedTransferSignedValue := true
			t.Run("WasProposedTransferSignedValue false", func(t *testing.T) {
				testFlow(t, WasTransferProposedOnElrondValue, WasProposedTransferSignedValue, false)
			})
			t.Run("WasProposedTransferSignedValue true", func(t *testing.T) {
				testFlow(t, WasTransferProposedOnElrondValue, WasProposedTransferSignedValue, true)
			})
		})
	})
}
