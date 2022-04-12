package ethToElrond

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/ethElrond/steps"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	bridgeTests "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/stateMachine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	getAndStoreBatchFromEthereum                  = "GetAndStoreBatchFromEthereum"
	getLastExecutedEthBatchIDFromElrond           = "GetLastExecutedEthBatchIDFromElrond"
	verifyLastDepositNonceExecutedOnEthereumBatch = "VerifyLastDepositNonceExecutedOnEthereumBatch"
	validateBatch                                 = "ValidateBatch"
	wasTransferProposedOnElrond                   = "WasTransferProposedOnElrond"
	wasActionSignedOnElrond                       = "WasActionSignedOnElrond"
	signActionOnElrond                            = "SignActionOnElrond"
	getAndStoreActionIDForProposeTransferOnElrond = "GetAndStoreActionIDForProposeTransferOnElrond"
	ProcessMaxQuorumRetriesOnElrond               = "ProcessMaxQuorumRetriesOnElrond"
	resetRetriesCountOnElrond                     = "ResetRetriesCountOnElrond"
	processQuorumReachedOnElrond                  = "ProcessQuorumReachedOnElrond"
	wasActionPerformedOnElrond                    = "WasActionPerformedOnElrond"
	proposeTransferOnElrond                       = "ProposeTransferOnElrond"
	performActionOnElrond                         = "PerformActionOnElrond"
)

var trueHandler = func() bool { return true }
var falseHandler = func() bool { return false }

type errorHandler struct {
	lastError error
}

func (eh *errorHandler) storeAndReturnError(err error) error {
	eh.lastError = err
	return err
}

type argsBridgeStub struct {
	failingStep                      string
	myTurnHandler                    func() bool
	wasTransferProposedHandler       func() bool
	wasProposedTransferSignedHandler func() bool
	wasActionSigned                  func() bool
	isQuorumReachedHandler           func() bool
	wasActionIDPerformedHandler      func() bool
	maxRetriesReachedHandler         func() bool
	validateBatchHandler             func() bool
}

func createMockBridge(args argsBridgeStub) (*bridgeTests.BridgeExecutorStub, *errorHandler) {
	errHandler := &errorHandler{}
	stub := bridgeTests.NewBridgeExecutorStub()
	expectedErr := errors.New("expected error")
	stub.MyTurnAsLeaderCalled = func() bool {
		return args.myTurnHandler()
	}
	stub.GetAndStoreActionIDForProposeTransferOnElrondCalled = func(ctx context.Context) (uint64, error) {
		if args.failingStep == getAndStoreActionIDForProposeTransferOnElrond {
			return 0, errHandler.storeAndReturnError(expectedErr)
		}

		return 2, errHandler.storeAndReturnError(nil)
	}
	stub.GetStoredActionIDCalled = func() uint64 {
		return 2
	}
	stub.GetAndStoreBatchFromEthereumCalled = func(ctx context.Context, nonce uint64) error {
		if args.failingStep == getAndStoreBatchFromEthereum {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.GetStoredBatchCalled = func() *clients.TransferBatch {
		return &clients.TransferBatch{}
	}
	stub.GetLastExecutedEthBatchIDFromElrondCalled = func(ctx context.Context) (uint64, error) {
		if args.failingStep == getLastExecutedEthBatchIDFromElrond {
			return 0, errHandler.storeAndReturnError(expectedErr)
		}

		return 3, errHandler.storeAndReturnError(nil)
	}
	stub.VerifyLastDepositNonceExecutedOnEthereumBatchCalled = func(ctx context.Context) error {
		if args.failingStep == verifyLastDepositNonceExecutedOnEthereumBatch {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.WasTransferProposedOnElrondCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasTransferProposedOnElrond {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasTransferProposedHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.ProposeTransferOnElrondCalled = func(ctx context.Context) error {
		if args.failingStep == proposeTransferOnElrond {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.WasActionSignedOnElrondCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasActionSignedOnElrond {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasActionSigned(), errHandler.storeAndReturnError(nil)
	}
	stub.SignActionOnElrondCalled = func(ctx context.Context) error {
		if args.failingStep == signActionOnElrond {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.ProcessQuorumReachedOnElrondCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == processQuorumReachedOnElrond {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.isQuorumReachedHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.WasActionPerformedOnElrondCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasActionPerformedOnElrond {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasActionIDPerformedHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.PerformActionOnElrondCalled = func(ctx context.Context) error {
		if args.failingStep == performActionOnElrond {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.ProcessMaxQuorumRetriesOnElrondCalled = func() bool {
		return args.maxRetriesReachedHandler()
	}
	stub.ValidateBatchCalled = func(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
		if args.failingStep == validateBatch {
			return false, errHandler.storeAndReturnError(expectedErr)
		}
		return args.validateBatchHandler(), errHandler.storeAndReturnError(nil)
	}

	return stub, errHandler
}

func createStateMachine(t *testing.T, executor steps.Executor, initialStep core.StepIdentifier) *stateMachine.StateMachineMock {
	stepsSlice, err := CreateSteps(executor)
	require.Nil(t, err)

	sm := stateMachine.NewStateMachineMock(stepsSlice, initialStep)
	err = sm.Initialize()
	require.Nil(t, err)

	return sm
}

func TestHappyCaseWhenLeader(t *testing.T) {
	t.Parallel()

	args := argsBridgeStub{
		myTurnHandler:                    trueHandler,
		isQuorumReachedHandler:           trueHandler,
		wasActionIDPerformedHandler:      trueHandler,
		validateBatchHandler:             trueHandler,
		maxRetriesReachedHandler:         falseHandler,
		wasProposedTransferSignedHandler: falseHandler,
		wasTransferProposedHandler:       falseHandler,
		wasActionSigned:                  falseHandler,
	}
	executor, eh := createMockBridge(args)
	sm := createStateMachine(t, executor, GettingPendingBatchFromEthereum)
	numSteps := 20
	for i := 0; i < numSteps; i++ {
		err := sm.Execute(context.Background())
		require.Nil(t, err)
	}

	assert.Equal(t, 4, executor.GetFunctionCounter(resetRetriesCountOnElrond))
	assert.Equal(t, 4, executor.GetFunctionCounter(getLastExecutedEthBatchIDFromElrond))
	assert.Equal(t, 4, executor.GetFunctionCounter(getAndStoreBatchFromEthereum))
	assert.Equal(t, 4, executor.GetFunctionCounter(verifyLastDepositNonceExecutedOnEthereumBatch))

	assert.Equal(t, 4, executor.GetFunctionCounter(wasTransferProposedOnElrond))
	assert.Equal(t, 4, executor.GetFunctionCounter(proposeTransferOnElrond))

	assert.Equal(t, 4, executor.GetFunctionCounter(getAndStoreActionIDForProposeTransferOnElrond))
	assert.Equal(t, 4, executor.GetFunctionCounter(wasActionSignedOnElrond))
	assert.Equal(t, 4, executor.GetFunctionCounter(signActionOnElrond))

	assert.Equal(t, 4, executor.GetFunctionCounter(ProcessMaxQuorumRetriesOnElrond))
	assert.Equal(t, 4, executor.GetFunctionCounter(processQuorumReachedOnElrond))

	assert.Equal(t, 4, executor.GetFunctionCounter(wasActionPerformedOnElrond))
	assert.Equal(t, 0, executor.GetFunctionCounter(performActionOnElrond))

	assert.Nil(t, eh.lastError)
}

func TestHappyCaseWhenLeaderAndActionIdNotPerformed(t *testing.T) {
	t.Parallel()

	numCalled := 0
	args := argsBridgeStub{
		myTurnHandler:          trueHandler,
		isQuorumReachedHandler: trueHandler,
		validateBatchHandler:   trueHandler,
		wasActionIDPerformedHandler: func() bool {
			numCalled++
			return numCalled > 1
		},
		maxRetriesReachedHandler:         falseHandler,
		wasProposedTransferSignedHandler: falseHandler,
		wasTransferProposedHandler:       falseHandler,
		wasActionSigned:                  falseHandler,
	}
	executor, eh := createMockBridge(args)
	sm := createStateMachine(t, executor, GettingPendingBatchFromEthereum)
	numSteps := 20
	for i := 0; i < numSteps; i++ {
		err := sm.Execute(context.Background())
		require.Nil(t, err)
	}

	assert.Equal(t, 4, executor.GetFunctionCounter(resetRetriesCountOnElrond))
	assert.Equal(t, 4, executor.GetFunctionCounter(getLastExecutedEthBatchIDFromElrond))
	assert.Equal(t, 4, executor.GetFunctionCounter(getAndStoreBatchFromEthereum))
	assert.Equal(t, 4, executor.GetFunctionCounter(verifyLastDepositNonceExecutedOnEthereumBatch))

	assert.Equal(t, 4, executor.GetFunctionCounter(wasTransferProposedOnElrond))
	assert.Equal(t, 4, executor.GetFunctionCounter(proposeTransferOnElrond))

	assert.Equal(t, 4, executor.GetFunctionCounter(getAndStoreActionIDForProposeTransferOnElrond))
	assert.Equal(t, 4, executor.GetFunctionCounter(wasActionSignedOnElrond))
	assert.Equal(t, 4, executor.GetFunctionCounter(signActionOnElrond))

	assert.Equal(t, 4, executor.GetFunctionCounter(ProcessMaxQuorumRetriesOnElrond))
	assert.Equal(t, 4, executor.GetFunctionCounter(processQuorumReachedOnElrond))

	assert.Equal(t, 4, executor.GetFunctionCounter(wasActionPerformedOnElrond))
	assert.Equal(t, 1, executor.GetFunctionCounter(performActionOnElrond))

	assert.Nil(t, eh.lastError)
}

func TestOneStepErrors_ShouldReturnToPendingBatch(t *testing.T) {
	stepsThatCanError := []core.StepIdentifier{
		getAndStoreActionIDForProposeTransferOnElrond,
		getAndStoreBatchFromEthereum,
		getLastExecutedEthBatchIDFromElrond,
		verifyLastDepositNonceExecutedOnEthereumBatch,
		validateBatch,
		wasTransferProposedOnElrond,
		proposeTransferOnElrond,
		wasTransferProposedOnElrond,
		signActionOnElrond,
		processQuorumReachedOnElrond,
		wasActionPerformedOnElrond,
		performActionOnElrond,
	}

	for _, stepThatError := range stepsThatCanError {
		testErrorFlow(t, stepThatError)
	}
}

func testErrorFlow(t *testing.T, stepThatErrors core.StepIdentifier) {
	numCalled := 0
	args := argsBridgeStub{
		failingStep:            string(stepThatErrors),
		myTurnHandler:          trueHandler,
		isQuorumReachedHandler: trueHandler,
		validateBatchHandler:   trueHandler,
		wasActionIDPerformedHandler: func() bool {
			numCalled++
			return numCalled > 1
		},
		maxRetriesReachedHandler:         falseHandler,
		wasProposedTransferSignedHandler: falseHandler,
		wasTransferProposedHandler:       falseHandler,
		wasActionSigned:                  falseHandler,
	}

	executor, eh := createMockBridge(args)
	sm := createStateMachine(t, executor, GettingPendingBatchFromEthereum)

	maxNumSteps := 10
	for i := 0; i < maxNumSteps; i++ {
		err := sm.Execute(context.Background())
		assert.Nil(t, err)

		if eh.lastError != nil {
			if sm.CurrentStep.Identifier() == GettingPendingBatchFromEthereum {
				return
			}

			require.Fail(t, fmt.Sprintf("should have jumped to initial step, got next step %s, stepThatErrors %s",
				sm.CurrentStep.Identifier(), stepThatErrors))
		}
	}

	require.Fail(t, fmt.Sprintf("max number of steps reached but not jumped to initial step, stepThatErrors %s", stepThatErrors))
}
