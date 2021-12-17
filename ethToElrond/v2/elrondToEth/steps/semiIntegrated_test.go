package steps

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/elrondToEth"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/stateMachine"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	getBatchFromElrond                               = "GetBatchFromElrond"
	storeBatchFromElrond                             = "StoreBatchFromElrond"
	wasTransferPerformedOnEthereum                   = "WasTransferPerformedOnEthereum"
	signTransferOnEthereum                           = "SignTransferOnEthereum"
	processMaxRetriesOnEthereum                      = "ProcessMaxRetriesOnEthereum"
	processQuorumReachedOnEthereum                   = "ProcessQuorumReachedOnEthereum"
	performTransferOnEthereum                        = "PerformTransferOnEthereum"
	getBatchStatusesFromEthereum                     = "GetBatchStatusesFromEthereum"
	wasSetStatusProposedOnElrond                     = "WasSetStatusProposedOnElrond"
	proposeSetStatusOnElrond                         = "ProposeSetStatusOnElrond"
	getAndStoreActionIDForProposeSetStatusFromElrond = "GetAndStoreActionIDForProposeSetStatusFromElrond"
	wasActionSignedOnElrond                          = "WasActionSignedOnElrond"
	signActionOnElrond                               = "SignActionOnElrond"
	processMaxRetriesOnElrond                        = "ProcessMaxRetriesOnElrond"
	processQuorumReachedOnElrond                     = "ProcessQuorumReachedOnElrond"
	wasActionPerformedOnElrond                       = "WasActionPerformedOnElrond"
	performActionOnElrond                            = "PerformActionOnElrond"
	resetRetriesCountOnEthereum                      = "ResetRetriesCountOnEthereum"
	resetRetriesCountOnElrond                        = "ResetRetriesCountOnElrond"
	getStoredBatch                                   = "GetStoredBatch"
	myTurnAsLeader                                   = "MyTurnAsLeader"
	waitForTransferConfirmation                      = "WaitForTransferConfirmation"
	resolveNewDepositsStatuses                       = "ResolveNewDepositsStatuses"
	getStoredActionID                                = "GetStoredActionID"
)

type argsBridgeStub struct {
	failingStep                           string
	wasTransferPerformedOnEthereumHandler func() bool
	processQuorumReachedOnEthereumHandler func() bool
	processQuorumReachedOnElrondHandler   func() bool
	myTurnHandler                         func() bool
	wasSetStatusProposedOnElrondHandler   func() bool
	wasActionSignedOnElrondHandler        func() bool
	wasActionPerformedOnElrondHandler     func() bool
	maxRetriesReachedEthereumHandler      func() bool
	maxRetriesReachedElrondHandler        func() bool
	resolveNewDepositsStatusesHandler     func() bool
}

var trueHandler = func() bool { return true }
var falseHandler = func() bool { return false }

type errorHandler struct {
	lastError error
}

func (eh *errorHandler) storeAndReturnError(err error) error {
	eh.lastError = err
	return err
}

func createStateMachine(t *testing.T, executor bridge.Executor, initialStep core.StepIdentifier) *stateMachine.StateMachineMock {
	steps, err := CreateSteps(executor)
	require.Nil(t, err)

	sm := stateMachine.NewStateMachineMock(steps, initialStep)
	err = sm.Initialize()
	require.Nil(t, err)

	return sm
}

func createMockBridge(args argsBridgeStub) (*bridgeV2.BridgeExecutorStub, *errorHandler) {
	errHandler := &errorHandler{}
	stub := bridgeV2.NewBridgeExecutorStub()
	expectedErr := errors.New("expected error")
	stub.GetLoggerCalled = func() logger.Logger {
		return logger.GetOrCreate("test")
	}
	stub.MyTurnAsLeaderCalled = func() bool {
		return args.myTurnHandler()
	}
	stub.GetAndStoreActionIDForProposeSetStatusFromElrondCalled = func(ctx context.Context) (uint64, error) {
		if args.failingStep == getAndStoreActionIDForProposeSetStatusFromElrond {
			return 0, errHandler.storeAndReturnError(expectedErr)
		}

		return 2, errHandler.storeAndReturnError(nil)
	}
	stub.GetStoredActionIDCalled = func() uint64 {
		return 2
	}
	stub.GetBatchFromElrondCalled = func(ctx context.Context) (*clients.TransferBatch, error) {
		if args.failingStep == getBatchFromElrond {
			return &clients.TransferBatch{}, errHandler.storeAndReturnError(expectedErr)
		}
		return &clients.TransferBatch{}, errHandler.storeAndReturnError(nil)
	}
	stub.StoreBatchFromElrondCalled = func(batch *clients.TransferBatch) error {
		return nil
	}
	stub.GetStoredBatchCalled = func() *clients.TransferBatch {
		return &clients.TransferBatch{}
	}
	stub.WasTransferPerformedOnEthereumCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasTransferPerformedOnEthereum {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasTransferPerformedOnEthereumHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.SignTransferOnEthereumCalled = func() error {
		if args.failingStep == signTransferOnEthereum {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.ProcessQuorumReachedOnEthereumCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == processQuorumReachedOnEthereum {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.processQuorumReachedOnEthereumHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.PerformTransferOnEthereumCalled = func(ctx context.Context) error {
		if args.failingStep == performTransferOnEthereum {
			return errHandler.storeAndReturnError(expectedErr)
		}
		return errHandler.storeAndReturnError(nil)
	}
	stub.WaitForTransferConfirmationCalled = func(ctx context.Context) {
		stub.WasTransferPerformedOnEthereumCalled = func(ctx context.Context) (bool, error) {
			return true, errHandler.storeAndReturnError(nil)
		}
	}
	stub.GetBatchStatusesFromEthereumCalled = func(ctx context.Context) ([]byte, error) {
		if args.failingStep == getBatchStatusesFromEthereum {
			return nil, errHandler.storeAndReturnError(expectedErr)
		}
		return []byte{}, errHandler.storeAndReturnError(nil)
	}
	stub.ResolveNewDepositsStatusesCalled = func(numDeposits uint64) {

	}
	stub.WasSetStatusProposedOnElrondCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasSetStatusProposedOnElrond {
			return false, errHandler.storeAndReturnError(expectedErr)
		}
		return args.wasSetStatusProposedOnElrondHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.ProposeSetStatusOnElrondCalled = func(ctx context.Context) error {
		if args.failingStep == proposeSetStatusOnElrond {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.WasActionSignedOnElrondCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasActionSignedOnElrond {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasActionSignedOnElrondHandler(), errHandler.storeAndReturnError(nil)
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

		return args.processQuorumReachedOnElrondHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.WasActionPerformedOnElrondCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasActionPerformedOnElrond {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasActionPerformedOnElrondHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.PerformActionOnElrondCalled = func(ctx context.Context) error {
		if args.failingStep == performActionOnElrond {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.ProcessMaxRetriesOnElrondCalled = func() bool {
		return args.maxRetriesReachedEthereumHandler()
	}
	stub.ProcessMaxRetriesOnEthereumCalled = func() bool {
		return args.maxRetriesReachedElrondHandler()
	}

	return stub, errHandler
}

func TestHappyCaseWhenLeaderSetStatusAlreadySigned(t *testing.T) {
	t.Parallel()

	numCalled := 0
	args := argsBridgeStub{
		myTurnHandler:                         trueHandler,
		processQuorumReachedOnEthereumHandler: trueHandler,
		processQuorumReachedOnElrondHandler:   trueHandler,
		wasActionSignedOnElrondHandler:        trueHandler,
		wasActionPerformedOnElrondHandler: func() bool {
			numCalled++
			return numCalled > 1
		},
		wasTransferPerformedOnEthereumHandler: falseHandler,
		maxRetriesReachedEthereumHandler:      falseHandler,
		maxRetriesReachedElrondHandler:        falseHandler,
		wasSetStatusProposedOnElrondHandler:   falseHandler,
	}
	executor, eh := createMockBridge(args)
	sm := createStateMachine(t, executor, elrondToEth.GettingPendingBatchFromElrond)
	numSteps := 12
	for i := 0; i < numSteps; i++ {
		err := sm.ExecuteOneStep()
		require.Nil(t, err)
	}

	assert.Equal(t, 1, executor.GetFunctionCounter(resetRetriesCountOnEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(resetRetriesCountOnElrond))
	assert.Equal(t, 2, executor.GetFunctionCounter(getBatchFromElrond))
	assert.Equal(t, 1, executor.GetFunctionCounter(storeBatchFromElrond))
	assert.Equal(t, 3, executor.GetFunctionCounter(wasTransferPerformedOnEthereum))
	assert.Equal(t, 4, executor.GetFunctionCounter(getStoredBatch))
	assert.Equal(t, 1, executor.GetFunctionCounter(signTransferOnEthereum))
	assert.Equal(t, 3, executor.GetFunctionCounter(wasTransferPerformedOnEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(processMaxRetriesOnEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(processQuorumReachedOnEthereum))
	assert.Equal(t, 3, executor.GetFunctionCounter(myTurnAsLeader))
	assert.Equal(t, 1, executor.GetFunctionCounter(processMaxRetriesOnElrond))
	assert.Equal(t, 1, executor.GetFunctionCounter(processQuorumReachedOnElrond))
	assert.Equal(t, 1, executor.GetFunctionCounter(waitForTransferConfirmation))
	assert.Equal(t, 1, executor.GetFunctionCounter(resolveNewDepositsStatuses))
	assert.Equal(t, 1, executor.GetFunctionCounter(wasSetStatusProposedOnElrond))
	assert.Equal(t, 1, executor.GetFunctionCounter(performTransferOnEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(getBatchStatusesFromEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(proposeSetStatusOnElrond))
	assert.Equal(t, 1, executor.GetFunctionCounter(getAndStoreActionIDForProposeSetStatusFromElrond))
	assert.Equal(t, 2, executor.GetFunctionCounter(wasActionPerformedOnElrond))
	assert.Equal(t, 1, executor.GetFunctionCounter(performActionOnElrond))

	assert.Equal(t, 1, executor.GetFunctionCounter(wasActionSignedOnElrond))
	assert.Equal(t, 1, executor.GetFunctionCounter(getStoredActionID))

	assert.Nil(t, eh.lastError)
}

func TestOneStepErrors_ShouldReturnToPendingBatch(t *testing.T) {
	stepsThatCanError := []core.StepIdentifier{
		getBatchFromElrond,
		wasTransferPerformedOnEthereum,
		signTransferOnEthereum,
		processQuorumReachedOnEthereum,
		performTransferOnEthereum,
		getBatchStatusesFromEthereum,
		wasSetStatusProposedOnElrond,
		proposeSetStatusOnElrond,
		getAndStoreActionIDForProposeSetStatusFromElrond,
		wasActionSignedOnElrond,
		processQuorumReachedOnElrond,
		wasActionPerformedOnElrond,
		performActionOnElrond,
		signActionOnElrond,
	}

	for _, stepThatError := range stepsThatCanError {
		testErrorFlow(t, stepThatError)
	}
}

func testErrorFlow(t *testing.T, stepThatErrors core.StepIdentifier) {
	t.Logf("\n\n\nnew test for stepThatError: %s", stepThatErrors)
	numCalled := 0
	args := argsBridgeStub{
		failingStep:                           string(stepThatErrors),
		myTurnHandler:                         trueHandler,
		processQuorumReachedOnEthereumHandler: trueHandler,
		processQuorumReachedOnElrondHandler:   trueHandler,
		wasActionSignedOnElrondHandler:        trueHandler,
		wasActionPerformedOnElrondHandler: func() bool {
			numCalled++
			return numCalled > 1
		},
		wasTransferPerformedOnEthereumHandler: falseHandler,
		maxRetriesReachedEthereumHandler:      falseHandler,
		maxRetriesReachedElrondHandler:        falseHandler,
		wasSetStatusProposedOnElrondHandler:   falseHandler,
	}

	if stepThatErrors == "SignActionOnElrond" {
		args.wasActionSignedOnElrondHandler = falseHandler
	}

	executor, eh := createMockBridge(args)
	sm := createStateMachine(t, executor, elrondToEth.GettingPendingBatchFromElrond)

	maxNumSteps := 12
	for i := 0; i < maxNumSteps; i++ {
		err := sm.ExecuteOneStep()
		assert.Nil(t, err)

		if eh.lastError != nil {
			if sm.CurrentStep.Identifier() == elrondToEth.GettingPendingBatchFromElrond {
				return
			}

			require.Fail(t, fmt.Sprintf("should have jumped to initial step, got next step %s, stepThatErrors %s",
				sm.CurrentStep.Identifier(), stepThatErrors))
		}
	}

	require.Fail(t, fmt.Sprintf("max number of steps reached but not jumped to initial step, stepThatErrors %s", stepThatErrors))
}
