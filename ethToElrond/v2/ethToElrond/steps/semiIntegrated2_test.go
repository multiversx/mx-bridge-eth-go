package steps

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/stateMachine"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	isQuorumReachedHandler           func() bool
	wasActionIDPerformedHandler      func() bool
	maxRetriesReachedHandler         func() bool
}

func createMockBridge(args argsBridgeStub) (*bridgeV2.EthToElrondBridgeStub, *errorHandler) {
	errHandler := &errorHandler{}
	stub := bridgeV2.NewEthToElrondBridgeStub()
	expectedErr := errors.New("expected error")
	stub.GetLoggerCalled = func() logger.Logger {
		return logger.GetOrCreate("test")
	}
	stub.MyTurnAsLeaderCalled = func() bool {
		return args.myTurnHandler()
	}
	stub.GetAndStoreActionIDFromElrondCalled = func(ctx context.Context) (uint64, error) {
		if args.failingStep == getAndStoreActionID {
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
	stub.WasProposedTransferSignedOnElrondCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasProposedTransferSigned {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasProposedTransferSignedHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.SignProposedTransferOnElrondCalled = func(ctx context.Context) error {
		if args.failingStep == signProposedTransfer {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.IsQuorumReachedOnElrondCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == isQuorumReached {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.isQuorumReachedHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.WasActionIDPerformedOnElrondCalled = func(ctx context.Context) (bool, error) {
		if args.failingStep == wasActionIDPerformed {
			return false, errHandler.storeAndReturnError(expectedErr)
		}

		return args.wasActionIDPerformedHandler(), errHandler.storeAndReturnError(nil)
	}
	stub.PerformActionIDOnElrondCalled = func(ctx context.Context) error {
		if args.failingStep == performActionID {
			return errHandler.storeAndReturnError(expectedErr)
		}

		return errHandler.storeAndReturnError(nil)
	}
	stub.ProcessMaxRetriesOnElrondCalled = func() bool {
		return args.maxRetriesReachedHandler()
	}

	return stub, errHandler
}

func createStateMachine(t *testing.T, executor ethToElrond.EthToElrondBridge, initialStep core.StepIdentifier) *stateMachine.StateMachineMock {
	steps, err := CreateSteps(executor)
	require.Nil(t, err)

	sm := stateMachine.NewStateMachineMock(steps, initialStep)
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
		maxRetriesReachedHandler:         falseHandler,
		wasProposedTransferSignedHandler: falseHandler,
		wasTransferProposedHandler:       falseHandler,
	}
	executor, eh := createMockBridge(args)
	sm := createStateMachine(t, executor, ethToElrond.GettingPendingBatchFromEthereum)
	numSteps := 20
	for i := 0; i < numSteps; i++ {
		err := sm.ExecuteOneStep()
		require.Nil(t, err)
	}

	assert.Equal(t, 4, executor.GetFunctionCounter(resetRetriesCountOnElrond))
	assert.Equal(t, 4, executor.GetFunctionCounter(getLastExecutedEthBatchIDFromElrond))
	assert.Equal(t, 4, executor.GetFunctionCounter(getAndStoreBatchFromEthereum))
	assert.Equal(t, 4, executor.GetFunctionCounter(verifyLastDepositNonceExecutedOnEthereumBatch))

	assert.Equal(t, 4, executor.GetFunctionCounter(wasTransferProposedOnElrond))
	assert.Equal(t, 4, executor.GetFunctionCounter(proposeTransferOnElrond))

	assert.Equal(t, 4, executor.GetFunctionCounter(getAndStoreActionID))
	assert.Equal(t, 4, executor.GetFunctionCounter(wasProposedTransferSigned))
	assert.Equal(t, 4, executor.GetFunctionCounter(signProposedTransfer))

	assert.Equal(t, 4, executor.GetFunctionCounter(processMaxRetriesOnElrond))
	assert.Equal(t, 4, executor.GetFunctionCounter(isQuorumReached))

	assert.Equal(t, 4, executor.GetFunctionCounter(wasActionIDPerformed))
	assert.Equal(t, 0, executor.GetFunctionCounter(performActionID))

	assert.Nil(t, eh.lastError)
}

func TestHappyCaseWhenLeaderAndActionIdNotPerformed(t *testing.T) {
	t.Parallel()

	numCalled := 0
	args := argsBridgeStub{
		myTurnHandler:          trueHandler,
		isQuorumReachedHandler: trueHandler,
		wasActionIDPerformedHandler: func() bool {
			numCalled++
			if numCalled > 1 {
				return true
			}

			return false
		},
		maxRetriesReachedHandler:         falseHandler,
		wasProposedTransferSignedHandler: falseHandler,
		wasTransferProposedHandler:       falseHandler,
	}
	executor, eh := createMockBridge(args)
	sm := createStateMachine(t, executor, ethToElrond.GettingPendingBatchFromEthereum)
	numSteps := 20
	for i := 0; i < numSteps; i++ {
		err := sm.ExecuteOneStep()
		require.Nil(t, err)
	}

	assert.Equal(t, 4, executor.GetFunctionCounter(resetRetriesCountOnElrond))
	assert.Equal(t, 4, executor.GetFunctionCounter(getLastExecutedEthBatchIDFromElrond))
	assert.Equal(t, 4, executor.GetFunctionCounter(getAndStoreBatchFromEthereum))
	assert.Equal(t, 4, executor.GetFunctionCounter(verifyLastDepositNonceExecutedOnEthereumBatch))

	assert.Equal(t, 4, executor.GetFunctionCounter(wasTransferProposedOnElrond))
	assert.Equal(t, 4, executor.GetFunctionCounter(proposeTransferOnElrond))

	assert.Equal(t, 4, executor.GetFunctionCounter(getAndStoreActionID))
	assert.Equal(t, 4, executor.GetFunctionCounter(wasProposedTransferSigned))
	assert.Equal(t, 4, executor.GetFunctionCounter(signProposedTransfer))

	assert.Equal(t, 4, executor.GetFunctionCounter(processMaxRetriesOnElrond))
	assert.Equal(t, 4, executor.GetFunctionCounter(isQuorumReached))

	assert.Equal(t, 4, executor.GetFunctionCounter(wasActionIDPerformed))
	assert.Equal(t, 1, executor.GetFunctionCounter(performActionID))

	assert.Nil(t, eh.lastError)
}

func TestOneStepErrors_ShouldReturnToPendingBatch(t *testing.T) {
	stepsThatCanError := []core.StepIdentifier{
		getAndStoreActionID,
		getAndStoreBatchFromEthereum,
		getLastExecutedEthBatchIDFromElrond,
		verifyLastDepositNonceExecutedOnEthereumBatch,
		wasTransferProposedOnElrond,
		proposeTransferOnElrond,
		wasProposedTransferSigned,
		signProposedTransfer,
		isQuorumReached,
		wasActionIDPerformed,
		performActionID,
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
		wasActionIDPerformedHandler: func() bool {
			numCalled++
			if numCalled > 1 {
				return true
			}

			return false
		},
		maxRetriesReachedHandler:         falseHandler,
		wasProposedTransferSignedHandler: falseHandler,
		wasTransferProposedHandler:       falseHandler,
	}

	executor, eh := createMockBridge(args)
	sm := createStateMachine(t, executor, ethToElrond.GettingPendingBatchFromEthereum)

	maxNumSteps := 10
	for i := 0; i < maxNumSteps; i++ {
		err := sm.ExecuteOneStep()
		assert.Nil(t, err)

		if eh.lastError != nil {
			if sm.CurrentStep.Identifier() == ethToElrond.GettingPendingBatchFromEthereum {
				return
			}

			require.Fail(t, fmt.Sprintf("should have jumped to initial step, got next step %s, stepThatErrors %s",
				sm.CurrentStep.Identifier(), stepThatErrors))
		}
	}

	require.Fail(t, fmt.Sprintf("max number of steps reached but not jumped to initial step, stepThatErrors %s", stepThatErrors))
}
