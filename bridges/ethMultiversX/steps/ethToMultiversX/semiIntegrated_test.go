package ethtomultiversx

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX/steps"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/core"
	bridgeTests "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/stateMachine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	getAndStoreBatchFromEthereum                      = "GetAndStoreBatchFromEthereum"
	getLastExecutedEthBatchIDFromMultiversX           = "GetLastExecutedEthBatchIDFromMultiversX"
	verifyLastDepositNonceExecutedOnEthereumBatch     = "VerifyLastDepositNonceExecutedOnEthereumBatch"
	wasTransferProposedOnMultiversX                   = "WasTransferProposedOnMultiversX"
	wasActionSignedOnMultiversX                       = "WasActionSignedOnMultiversX"
	signActionOnMultiversX                            = "SignActionOnMultiversX"
	getAndStoreActionIDForProposeTransferOnMultiversX = "GetAndStoreActionIDForProposeTransferOnMultiversX"
	ProcessMaxQuorumRetriesOnMultiversX               = "ProcessMaxQuorumRetriesOnMultiversX"
	resetRetriesCountOnMultiversX                     = "ResetRetriesCountOnMultiversX"
	processQuorumReachedOnMultiversX                  = "ProcessQuorumReachedOnMultiversX"
	wasActionPerformedOnMultiversX                    = "WasActionPerformedOnMultiversX"
	proposeTransferOnMultiversX                       = "ProposeTransferOnMultiversX"
	performActionOnMultiversX                         = "PerformActionOnMultiversX"
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
	stub.GetAndStoreActionIDForProposeTransferOnMultiversXCalled = func(ctx context.Context) (uint64, error) {
		if args.failingStep == getAndStoreActionIDForProposeTransferOnMultiversX {
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
	stub.GetLastExecutedEthBatchIDFromMultiversXCalled = func(ctx context.Context) (uint64, error) {
		if args.failingStep == getLastExecutedEthBatchIDFromMultiversX {
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
	stub.WasTransferProposedOnMultiversXCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasTransferProposedOnMultiversX {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasTransferProposedHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.ProposeTransferOnMultiversXCalled = func(ctx context.Context) error {
		if args.failingStep == proposeTransferOnMultiversX {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.WasActionSignedOnMultiversXCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasActionSignedOnMultiversX {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasActionSigned(), errHandler.storeAndReturnError(nil)
	}
	stub.SignActionOnMultiversXCalled = func(ctx context.Context) error {
		if args.failingStep == signActionOnMultiversX {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.ProcessQuorumReachedOnMultiversXCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == processQuorumReachedOnMultiversX {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.isQuorumReachedHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.WasActionPerformedOnMultiversXCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasActionPerformedOnMultiversX {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasActionIDPerformedHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.PerformActionOnMultiversXCalled = func(ctx context.Context) error {
		if args.failingStep == performActionOnMultiversX {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.ProcessMaxQuorumRetriesOnMultiversXCalled = func() bool {
		return args.maxRetriesReachedHandler()
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

	assert.Equal(t, 4, executor.GetFunctionCounter(resetRetriesCountOnMultiversX))
	assert.Equal(t, 4, executor.GetFunctionCounter(getLastExecutedEthBatchIDFromMultiversX))
	assert.Equal(t, 4, executor.GetFunctionCounter(getAndStoreBatchFromEthereum))
	assert.Equal(t, 4, executor.GetFunctionCounter(verifyLastDepositNonceExecutedOnEthereumBatch))

	assert.Equal(t, 4, executor.GetFunctionCounter(wasTransferProposedOnMultiversX))
	assert.Equal(t, 4, executor.GetFunctionCounter(proposeTransferOnMultiversX))

	assert.Equal(t, 4, executor.GetFunctionCounter(getAndStoreActionIDForProposeTransferOnMultiversX))
	assert.Equal(t, 4, executor.GetFunctionCounter(wasActionSignedOnMultiversX))
	assert.Equal(t, 4, executor.GetFunctionCounter(signActionOnMultiversX))

	assert.Equal(t, 4, executor.GetFunctionCounter(ProcessMaxQuorumRetriesOnMultiversX))
	assert.Equal(t, 4, executor.GetFunctionCounter(processQuorumReachedOnMultiversX))

	assert.Equal(t, 4, executor.GetFunctionCounter(wasActionPerformedOnMultiversX))
	assert.Equal(t, 0, executor.GetFunctionCounter(performActionOnMultiversX))

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

	assert.Equal(t, 4, executor.GetFunctionCounter(resetRetriesCountOnMultiversX))
	assert.Equal(t, 4, executor.GetFunctionCounter(getLastExecutedEthBatchIDFromMultiversX))
	assert.Equal(t, 4, executor.GetFunctionCounter(getAndStoreBatchFromEthereum))
	assert.Equal(t, 4, executor.GetFunctionCounter(verifyLastDepositNonceExecutedOnEthereumBatch))

	assert.Equal(t, 4, executor.GetFunctionCounter(wasTransferProposedOnMultiversX))
	assert.Equal(t, 4, executor.GetFunctionCounter(proposeTransferOnMultiversX))

	assert.Equal(t, 4, executor.GetFunctionCounter(getAndStoreActionIDForProposeTransferOnMultiversX))
	assert.Equal(t, 4, executor.GetFunctionCounter(wasActionSignedOnMultiversX))
	assert.Equal(t, 4, executor.GetFunctionCounter(signActionOnMultiversX))

	assert.Equal(t, 4, executor.GetFunctionCounter(ProcessMaxQuorumRetriesOnMultiversX))
	assert.Equal(t, 4, executor.GetFunctionCounter(processQuorumReachedOnMultiversX))

	assert.Equal(t, 4, executor.GetFunctionCounter(wasActionPerformedOnMultiversX))
	assert.Equal(t, 1, executor.GetFunctionCounter(performActionOnMultiversX))

	assert.Nil(t, eh.lastError)
}

func TestOneStepErrors_ShouldReturnToPendingBatch(t *testing.T) {
	stepsThatCanError := []core.StepIdentifier{
		getAndStoreActionIDForProposeTransferOnMultiversX,
		getAndStoreBatchFromEthereum,
		getLastExecutedEthBatchIDFromMultiversX,
		verifyLastDepositNonceExecutedOnEthereumBatch,
		wasTransferProposedOnMultiversX,
		proposeTransferOnMultiversX,
		wasTransferProposedOnMultiversX,
		signActionOnMultiversX,
		processQuorumReachedOnMultiversX,
		wasActionPerformedOnMultiversX,
		performActionOnMultiversX,
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
