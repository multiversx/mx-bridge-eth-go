package multiversxtoeth

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
	getBatchFromMultiversX                               = "GetBatchFromMultiversX"
	storeBatchFromMultiversX                             = "StoreBatchFromMultiversX"
	wasTransferPerformedOnEthereum                       = "WasTransferPerformedOnEthereum"
	signTransferOnEthereum                               = "SignTransferOnEthereum"
	ProcessMaxQuorumRetriesOnEthereum                    = "ProcessMaxQuorumRetriesOnEthereum"
	processQuorumReachedOnEthereum                       = "ProcessQuorumReachedOnEthereum"
	performTransferOnEthereum                            = "PerformTransferOnEthereum"
	getBatchStatusesFromEthereum                         = "GetBatchStatusesFromEthereum"
	wasSetStatusProposedOnMultiversX                     = "WasSetStatusProposedOnMultiversX"
	proposeSetStatusOnMultiversX                         = "ProposeSetStatusOnMultiversX"
	getAndStoreActionIDForProposeSetStatusFromMultiversX = "GetAndStoreActionIDForProposeSetStatusFromMultiversX"
	wasActionSignedOnMultiversX                          = "WasActionSignedOnMultiversX"
	signActionOnMultiversX                               = "SignActionOnMultiversX"
	ProcessMaxQuorumRetriesOnMultiversX                  = "ProcessMaxQuorumRetriesOnMultiversX"
	processQuorumReachedOnMultiversX                     = "ProcessQuorumReachedOnMultiversX"
	wasActionPerformedOnMultiversX                       = "WasActionPerformedOnMultiversX"
	performActionOnMultiversX                            = "PerformActionOnMultiversX"
	resetRetriesCountOnEthereum                          = "ResetRetriesCountOnEthereum"
	resetRetriesCountOnMultiversX                        = "ResetRetriesCountOnMultiversX"
	getStoredBatch                                       = "GetStoredBatch"
	myTurnAsLeader                                       = "MyTurnAsLeader"
	waitForTransferConfirmation                          = "WaitForTransferConfirmation"
	WaitAndReturnFinalBatchStatuses                      = "WaitAndReturnFinalBatchStatuses"
	resolveNewDepositsStatuses                           = "ResolveNewDepositsStatuses"
	getStoredActionID                                    = "GetStoredActionID"
)

type argsBridgeStub struct {
	failingStep                             string
	wasTransferPerformedOnEthereumHandler   func() bool
	processQuorumReachedOnEthereumHandler   func() bool
	processQuorumReachedOnMultiversXHandler func() bool
	myTurnHandler                           func() bool
	wasSetStatusProposedOnMultiversXHandler func() bool
	wasActionSignedOnMultiversXHandler      func() bool
	wasActionPerformedOnMultiversXHandler   func() bool
	maxRetriesReachedEthereumHandler        func() bool
	maxRetriesReachedMultiversXHandler      func() bool
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

func createStateMachine(t *testing.T, executor steps.Executor, initialStep core.StepIdentifier) *stateMachine.StateMachineMock {
	stepsSlice, err := CreateSteps(executor)
	require.Nil(t, err)

	sm := stateMachine.NewStateMachineMock(stepsSlice, initialStep)
	err = sm.Initialize()
	require.Nil(t, err)

	return sm
}

func createMockBridge(args argsBridgeStub) (*bridgeTests.BridgeExecutorStub, *errorHandler) {
	errHandler := &errorHandler{}
	stub := bridgeTests.NewBridgeExecutorStub()
	expectedErr := errors.New("expected error")
	stub.MyTurnAsLeaderCalled = func() bool {
		return args.myTurnHandler()
	}
	stub.GetAndStoreActionIDForProposeSetStatusFromMultiversXCalled = func(ctx context.Context) (uint64, error) {
		if args.failingStep == getAndStoreActionIDForProposeSetStatusFromMultiversX {
			return 0, errHandler.storeAndReturnError(expectedErr)
		}

		return 2, errHandler.storeAndReturnError(nil)
	}
	stub.GetStoredActionIDCalled = func() uint64 {
		return 2
	}
	stub.GetBatchFromMultiversXCalled = func(ctx context.Context) (*clients.TransferBatch, error) {
		if args.failingStep == getBatchFromMultiversX {
			return &clients.TransferBatch{}, errHandler.storeAndReturnError(expectedErr)
		}
		return &clients.TransferBatch{}, errHandler.storeAndReturnError(nil)
	}
	stub.StoreBatchFromMultiversXCalled = func(batch *clients.TransferBatch) error {
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
	stub.WaitAndReturnFinalBatchStatusesCalled = func(ctx context.Context) []byte {
		if args.failingStep == getBatchStatusesFromEthereum {
			return nil
		}
		return []byte{0x3}
	}
	stub.GetBatchStatusesFromEthereumCalled = func(ctx context.Context) ([]byte, error) {
		if args.failingStep == getBatchStatusesFromEthereum {
			return nil, errHandler.storeAndReturnError(expectedErr)
		}
		return []byte{}, errHandler.storeAndReturnError(nil)
	}
	stub.ResolveNewDepositsStatusesCalled = func(numDeposits uint64) {

	}
	stub.WasSetStatusProposedOnMultiversXCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasSetStatusProposedOnMultiversX {
			return false, errHandler.storeAndReturnError(expectedErr)
		}
		return args.wasSetStatusProposedOnMultiversXHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.ProposeSetStatusOnMultiversXCalled = func(ctx context.Context) error {
		if args.failingStep == proposeSetStatusOnMultiversX {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.WasActionSignedOnMultiversXCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasActionSignedOnMultiversX {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasActionSignedOnMultiversXHandler(), errHandler.storeAndReturnError(nil)
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

		return args.processQuorumReachedOnMultiversXHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.WasActionPerformedOnMultiversXCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasActionPerformedOnMultiversX {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasActionPerformedOnMultiversXHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.PerformActionOnMultiversXCalled = func(ctx context.Context) error {
		if args.failingStep == performActionOnMultiversX {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.ProcessMaxQuorumRetriesOnMultiversXCalled = func() bool {
		return args.maxRetriesReachedEthereumHandler()
	}
	stub.ProcessMaxQuorumRetriesOnEthereumCalled = func() bool {
		return args.maxRetriesReachedMultiversXHandler()
	}
	stub.ValidateBatchCalled = func(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
		return true, nil
	}

	return stub, errHandler
}

func TestHappyCaseWhenLeaderSetStatusAlreadySigned(t *testing.T) {
	t.Parallel()

	numCalled := 0
	args := argsBridgeStub{
		myTurnHandler:                           trueHandler,
		processQuorumReachedOnEthereumHandler:   trueHandler,
		processQuorumReachedOnMultiversXHandler: trueHandler,
		wasActionSignedOnMultiversXHandler:      trueHandler,
		wasActionPerformedOnMultiversXHandler: func() bool {
			numCalled++
			return numCalled > 1
		},
		wasTransferPerformedOnEthereumHandler:   falseHandler,
		maxRetriesReachedEthereumHandler:        falseHandler,
		maxRetriesReachedMultiversXHandler:      falseHandler,
		wasSetStatusProposedOnMultiversXHandler: falseHandler,
	}
	executor, eh := createMockBridge(args)
	sm := createStateMachine(t, executor, GettingPendingBatchFromMultiversX)
	numSteps := 12
	for i := 0; i < numSteps; i++ {
		err := sm.Execute(context.Background())
		require.Nil(t, err)
	}

	assert.Equal(t, 1, executor.GetFunctionCounter(resetRetriesCountOnEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(resetRetriesCountOnMultiversX))
	assert.Equal(t, 2, executor.GetFunctionCounter(getBatchFromMultiversX))
	assert.Equal(t, 1, executor.GetFunctionCounter(storeBatchFromMultiversX))
	assert.Equal(t, 3, executor.GetFunctionCounter(wasTransferPerformedOnEthereum))
	assert.Equal(t, 4, executor.GetFunctionCounter(getStoredBatch))
	assert.Equal(t, 1, executor.GetFunctionCounter(signTransferOnEthereum))
	assert.Equal(t, 3, executor.GetFunctionCounter(wasTransferPerformedOnEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(ProcessMaxQuorumRetriesOnEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(processQuorumReachedOnEthereum))
	assert.Equal(t, 3, executor.GetFunctionCounter(myTurnAsLeader))
	assert.Equal(t, 1, executor.GetFunctionCounter(ProcessMaxQuorumRetriesOnMultiversX))
	assert.Equal(t, 1, executor.GetFunctionCounter(processQuorumReachedOnMultiversX))
	assert.Equal(t, 1, executor.GetFunctionCounter(waitForTransferConfirmation))
	assert.Equal(t, 1, executor.GetFunctionCounter(resolveNewDepositsStatuses))
	assert.Equal(t, 1, executor.GetFunctionCounter(wasSetStatusProposedOnMultiversX))
	assert.Equal(t, 1, executor.GetFunctionCounter(performTransferOnEthereum))
	assert.Equal(t, 1, executor.GetFunctionCounter(WaitAndReturnFinalBatchStatuses))
	assert.Equal(t, 1, executor.GetFunctionCounter(proposeSetStatusOnMultiversX))
	assert.Equal(t, 1, executor.GetFunctionCounter(getAndStoreActionIDForProposeSetStatusFromMultiversX))
	assert.Equal(t, 2, executor.GetFunctionCounter(wasActionPerformedOnMultiversX))
	assert.Equal(t, 1, executor.GetFunctionCounter(performActionOnMultiversX))

	assert.Equal(t, 1, executor.GetFunctionCounter(wasActionSignedOnMultiversX))
	assert.Equal(t, 1, executor.GetFunctionCounter(getStoredActionID))

	assert.Nil(t, eh.lastError)
}

func TestOneStepErrors_ShouldReturnToPendingBatch(t *testing.T) {
	stepsThatCanError := []core.StepIdentifier{
		getBatchFromMultiversX,
		wasTransferPerformedOnEthereum,
		signTransferOnEthereum,
		processQuorumReachedOnEthereum,
		performTransferOnEthereum,
		wasSetStatusProposedOnMultiversX,
		proposeSetStatusOnMultiversX,
		getAndStoreActionIDForProposeSetStatusFromMultiversX,
		wasActionSignedOnMultiversX,
		processQuorumReachedOnMultiversX,
		wasActionPerformedOnMultiversX,
		performActionOnMultiversX,
		signActionOnMultiversX,
	}

	for _, stepThatError := range stepsThatCanError {
		testErrorFlow(t, stepThatError)
	}
}

func testErrorFlow(t *testing.T, stepThatErrors core.StepIdentifier) {
	t.Logf("\n\n\nnew test for stepThatError: %s", stepThatErrors)
	numCalled := 0
	args := argsBridgeStub{
		failingStep:                             string(stepThatErrors),
		myTurnHandler:                           trueHandler,
		processQuorumReachedOnEthereumHandler:   trueHandler,
		processQuorumReachedOnMultiversXHandler: trueHandler,
		wasActionSignedOnMultiversXHandler:      trueHandler,
		wasActionPerformedOnMultiversXHandler: func() bool {
			numCalled++
			return numCalled > 1
		},
		wasTransferPerformedOnEthereumHandler:   falseHandler,
		maxRetriesReachedEthereumHandler:        falseHandler,
		maxRetriesReachedMultiversXHandler:      falseHandler,
		wasSetStatusProposedOnMultiversXHandler: falseHandler,
	}

	if stepThatErrors == "SignActionOnMultiversX" {
		args.wasActionSignedOnMultiversXHandler = falseHandler
	}

	executor, eh := createMockBridge(args)
	sm := createStateMachine(t, executor, GettingPendingBatchFromMultiversX)

	maxNumSteps := 12
	for i := 0; i < maxNumSteps; i++ {
		err := sm.Execute(context.Background())
		assert.Nil(t, err)

		if eh.lastError != nil {
			if sm.CurrentStep.Identifier() == GettingPendingBatchFromMultiversX {
				return
			}

			require.Fail(t, fmt.Sprintf("should have jumped to initial step, got next step %s, stepThatErrors %s",
				sm.CurrentStep.Identifier(), stepThatErrors))
		}
	}

	require.Fail(t, fmt.Sprintf("max number of steps reached but not jumped to initial step, stepThatErrors %s", stepThatErrors))
}
