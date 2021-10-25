package bridgeExecutors

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/bridgeExecutors/mock"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDuration = time.Second

func createMockArgs() ArgsEthElrondBridgeExecutor {
	return ArgsEthElrondBridgeExecutor{
		ExecutorName:      "executorMock",
		Logger:            logger.GetOrCreate("test"),
		SourceBridge:      mock.NewBridgeStub(),
		DestinationBridge: mock.NewBridgeStub(),
		TopologyProvider:  &mock.TopologyProviderStub{},
		QuorumProvider:    &mock.QuorumProviderStub{},
		Timer:             &mock.TimerMock{},
		DurationsMap: map[core.StepIdentifier]time.Duration{
			ethToElrond.GettingPending: testDuration,
		},
	}
}

func TestNewbridgeExecutors(t *testing.T) {
	t.Parallel()
	t.Run("nil source bridge", func(t *testing.T) {
		args := createMockArgs()
		args.SourceBridge = nil
		executor, err := NewEthElrondBridgeExecutor(args)

		assert.Nil(t, executor)
		assert.Equal(t, fmt.Errorf("%w for the source bridge", ErrNilBridge), err)
	})
	t.Run("nil destination bridge", func(t *testing.T) {
		args := createMockArgs()
		args.DestinationBridge = nil
		executor, err := NewEthElrondBridgeExecutor(args)

		assert.Nil(t, executor)
		assert.Equal(t, fmt.Errorf("%w for the destination bridge", ErrNilBridge), err)
	})
	t.Run("nil logger", func(t *testing.T) {
		args := createMockArgs()
		args.Logger = nil
		executor, err := NewEthElrondBridgeExecutor(args)

		assert.Nil(t, executor)
		assert.Equal(t, ErrNilLogger, err)
	})
	t.Run("nil topology provider", func(t *testing.T) {
		args := createMockArgs()
		args.TopologyProvider = nil
		executor, err := NewEthElrondBridgeExecutor(args)

		assert.Nil(t, executor)
		assert.Equal(t, ErrNilTopologyProvider, err)
	})
	t.Run("nil logger", func(t *testing.T) {
		args := createMockArgs()
		args.QuorumProvider = nil
		executor, err := NewEthElrondBridgeExecutor(args)

		assert.Nil(t, executor)
		assert.Equal(t, ErrNilQuorumProvider, err)
	})
	t.Run("nil timer", func(t *testing.T) {
		args := createMockArgs()
		args.Timer = nil
		executor, err := NewEthElrondBridgeExecutor(args)

		assert.Nil(t, executor)
		assert.Equal(t, ErrNilTimer, err)
	})
	t.Run("nil duration map", func(t *testing.T) {
		args := createMockArgs()
		args.DurationsMap = nil
		executor, err := NewEthElrondBridgeExecutor(args)

		assert.Nil(t, executor)
		assert.Equal(t, ErrNilDurationsMap, err)
	})
}

func TestGetPending(t *testing.T) {
	t.Parallel()
	t.Run("no pending transaction", func(t *testing.T) {
		args := createMockArgs()
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())
		assert.False(t, executor.HasPendingBatch())

		executor.GetPendingBatch(nil)

		assert.False(t, executor.HasPendingBatch())
		assert.Nil(t, executor.pendingBatch)
	})
	t.Run("it will get the next pending transaction", func(t *testing.T) {
		expected := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		args := createMockArgs()
		sb := mock.NewBridgeStub()
		sb.GetPendingCalled = func(ctx context.Context) *bridge.Batch {
			return expected
		}
		args.SourceBridge = sb
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())
		assert.False(t, executor.HasPendingBatch())
		executor.GetPendingBatch(nil)

		assert.True(t, executor.HasPendingBatch())
		assert.Equal(t, expected, executor.pendingBatch)
	})
}

func TestLeader(t *testing.T) {
	t.Parallel()
	t.Run("relayer is leader", func(t *testing.T) {
		args := createMockArgs()
		tp := mock.NewTopologyProviderStub()
		tp.AmITheLeaderCalled = func() bool {
			return true
		}
		args.TopologyProvider = tp
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())
		assert.True(t, executor.IsLeader())
	})
	t.Run("relayer is NOT leader", func(t *testing.T) {
		args := createMockArgs()
		tp := mock.NewTopologyProviderStub()
		tp.AmITheLeaderCalled = func() bool {
			return false
		}
		args.TopologyProvider = tp
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())
		assert.False(t, executor.IsLeader())
	})
}

func TestWasProposeTransferExecutedOnDestination(t *testing.T) {
	t.Parallel()
	t.Run("proposed transfer has been proposed", func(t *testing.T) {
		expected := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		args := createMockArgs()
		sb := mock.NewBridgeStub()
		sb.GetPendingCalled = func(ctx context.Context) *bridge.Batch {
			return expected
		}
		args.SourceBridge = sb
		db := mock.NewBridgeStub()
		db.WasProposedTransferCalled = func(ctx context.Context, batch *bridge.Batch) bool {
			return true
		}
		args.DestinationBridge = db
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())

		executor.GetPendingBatch(nil)
		assert.True(t, executor.WasProposeTransferExecutedOnDestination(nil))
	})
}

func TestWasProposeSetStatusExecutedOnSource(t *testing.T) {
	t.Parallel()
	t.Run("", func(t *testing.T) {
		expected := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		args := createMockArgs()
		sb := mock.NewBridgeStub()
		sb.GetPendingCalled = func(ctx context.Context) *bridge.Batch {
			return expected
		}
		sb.WasProposedSetStatusCalled = func(ctx context.Context, batch *bridge.Batch) bool {
			return true
		}
		args.SourceBridge = sb
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())

		executor.GetPendingBatch(nil)
		assert.True(t, executor.WasProposeSetStatusExecutedOnSource(nil))
	})
}

func TestWasExecuted(t *testing.T) {
	t.Parallel()
	t.Run("OnDestination", func(t *testing.T) {
		expected := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		args := createMockArgs()
		sb := mock.NewBridgeStub()
		sb.GetPendingCalled = func(ctx context.Context) *bridge.Batch {
			return expected
		}
		args.SourceBridge = sb
		db := mock.NewBridgeStub()
		db.WasExecutedCalled = func(ctx context.Context, id bridge.ActionId, id2 bridge.BatchId) bool {
			return true
		}
		args.DestinationBridge = db
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())

		executor.GetPendingBatch(nil)
		assert.True(t, executor.WasTransferExecutedOnDestination(nil))
	})
	t.Run("OnDestination", func(t *testing.T) {
		expected := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		args := createMockArgs()
		sb := mock.NewBridgeStub()
		sb.GetPendingCalled = func(ctx context.Context) *bridge.Batch {
			return expected
		}
		sb.WasExecutedCalled = func(ctx context.Context, id bridge.ActionId, id2 bridge.BatchId) bool {
			return true
		}
		args.SourceBridge = sb
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())

		executor.GetPendingBatch(nil)
		assert.True(t, executor.WasSetStatusExecutedOnSource(nil))
	})
}

func TestIsQuorumReachedForProposeTransfer(t *testing.T) {
	t.Parallel()
	t.Run("quorum error", func(t *testing.T) {
		args := createMockArgs()
		args.QuorumProvider = &mock.QuorumProviderStub{
			GetQuorumCalled: func(ctx context.Context) (uint, error) {
				return 0, errors.New("some error")
			},
		}
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())

		assert.False(t, executor.IsQuorumReachedForProposeTransfer(nil))
	})
	t.Run("no signs", func(t *testing.T) {
		args := createMockArgs()
		args.QuorumProvider = &mock.QuorumProviderStub{
			GetQuorumCalled: func(ctx context.Context) (uint, error) {
				return 3, nil
			},
		}
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())

		assert.False(t, executor.IsQuorumReachedForProposeTransfer(nil))
	})
	t.Run("less < quorum", func(t *testing.T) {
		args := createMockArgs()
		args.QuorumProvider = &mock.QuorumProviderStub{
			GetQuorumCalled: func(ctx context.Context) (uint, error) {
				return 3, nil
			},
		}
		db := mock.NewBridgeStub()
		db.SignersCountCalled = func(ctx context.Context, id bridge.ActionId) uint {
			return 2
		}
		args.DestinationBridge = db
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())

		assert.False(t, executor.IsQuorumReachedForProposeTransfer(nil))
	})
	t.Run("signs == quorum", func(t *testing.T) {
		args := createMockArgs()
		args.QuorumProvider = &mock.QuorumProviderStub{
			GetQuorumCalled: func(ctx context.Context) (uint, error) {
				return 3, nil
			},
		}

		db := mock.NewBridgeStub()
		db.SignersCountCalled = func(ctx context.Context, id bridge.ActionId) uint {
			return 3
		}
		args.DestinationBridge = db
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())

		assert.True(t, executor.IsQuorumReachedForProposeTransfer(nil))
	})
	t.Run("signs > quorum", func(t *testing.T) {
		args := createMockArgs()
		args.QuorumProvider = &mock.QuorumProviderStub{
			GetQuorumCalled: func(ctx context.Context) (uint, error) {
				return 3, nil
			},
		}

		db := mock.NewBridgeStub()
		db.SignersCountCalled = func(ctx context.Context, id bridge.ActionId) uint {
			return 4
		}
		args.DestinationBridge = db
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())

		assert.True(t, executor.IsQuorumReachedForProposeTransfer(nil))
	})

}

func TestIsQuorumReachedForProposeSetStatus(t *testing.T) {
	t.Parallel()
	t.Run("quorum error", func(t *testing.T) {
		args := createMockArgs()
		args.QuorumProvider = &mock.QuorumProviderStub{
			GetQuorumCalled: func(ctx context.Context) (uint, error) {
				return 0, errors.New("some error")
			},
		}
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())

		assert.False(t, executor.IsQuorumReachedForProposeSetStatus(nil))
	})
	t.Run("no signs", func(t *testing.T) {
		args := createMockArgs()
		args.QuorumProvider = &mock.QuorumProviderStub{
			GetQuorumCalled: func(ctx context.Context) (uint, error) {
				return 3, nil
			},
		}
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())

		assert.False(t, executor.IsQuorumReachedForProposeSetStatus(nil))
	})
	t.Run("less < quorum", func(t *testing.T) {
		args := createMockArgs()
		args.QuorumProvider = &mock.QuorumProviderStub{
			GetQuorumCalled: func(ctx context.Context) (uint, error) {
				return 3, nil
			},
		}

		sb := mock.NewBridgeStub()
		sb.SignersCountCalled = func(ctx context.Context, id bridge.ActionId) uint {
			return 2
		}
		args.SourceBridge = sb
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())

		assert.False(t, executor.IsQuorumReachedForProposeSetStatus(nil))
	})
	t.Run("signs == quorum", func(t *testing.T) {
		args := createMockArgs()
		args.QuorumProvider = &mock.QuorumProviderStub{
			GetQuorumCalled: func(ctx context.Context) (uint, error) {
				return 3, nil
			},
		}

		sb := mock.NewBridgeStub()
		sb.SignersCountCalled = func(ctx context.Context, id bridge.ActionId) uint {
			return 3
		}
		args.SourceBridge = sb
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())

		assert.True(t, executor.IsQuorumReachedForProposeSetStatus(nil))
	})
	t.Run("signs > quorum", func(t *testing.T) {
		args := createMockArgs()
		args.QuorumProvider = &mock.QuorumProviderStub{
			GetQuorumCalled: func(ctx context.Context) (uint, error) {
				return 3, nil
			},
		}

		sb := mock.NewBridgeStub()
		sb.SignersCountCalled = func(ctx context.Context, id bridge.ActionId) uint {
			return 4
		}
		args.SourceBridge = sb
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())

		assert.True(t, executor.IsQuorumReachedForProposeSetStatus(nil))
	})
}

func TestPrintInfo(t *testing.T) {
	t.Parallel()
	printInfoTest := func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should not have panicked: %v", r))
		}
	}
	t.Run("Trace", func(t *testing.T) {
		args := createMockArgs()
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())
		defer printInfoTest()
		executor.PrintInfo(logger.LogTrace, "test")

	})
	t.Run("Debug", func(t *testing.T) {
		args := createMockArgs()
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())
		defer printInfoTest()
		executor.PrintInfo(logger.LogDebug, "test")
	})
	t.Run("Info", func(t *testing.T) {
		args := createMockArgs()
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())
		defer printInfoTest()
		executor.PrintInfo(logger.LogInfo, "test")
	})
	t.Run("Warn", func(t *testing.T) {
		args := createMockArgs()
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())
		defer printInfoTest()
		executor.PrintInfo(logger.LogWarning, "test")
	})
	t.Run("Error", func(t *testing.T) {
		args := createMockArgs()
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())
		defer printInfoTest()
		executor.PrintInfo(logger.LogError, "test")
	})
	t.Run("None", func(t *testing.T) {
		args := createMockArgs()
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())
		defer printInfoTest()
		executor.PrintInfo(logger.LogNone, "test")
	})
}

func TestProposeTransferOnDestination(t *testing.T) {
	t.Parallel()
	t.Run("ProposeTransferError", func(t *testing.T) {
		args := createMockArgs()
		expectedError := errors.New("some error")

		db := mock.NewBridgeStub()
		db.ProposeTransferError = expectedError
		args.DestinationBridge = db
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())

		err = executor.ProposeTransferOnDestination(nil)
		assert.Equal(t, expectedError, err)
	})
	t.Run("no error", func(t *testing.T) {
		args := createMockArgs()

		db := mock.NewBridgeStub()
		db.ProposeTransferCalled = func(ctx context.Context, batch *bridge.Batch) (string, error) {
			return "propose_tx_hash", nil
		}
		args.DestinationBridge = db

		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())

		err = executor.ProposeTransferOnDestination(nil)
		assert.Nil(t, err)
	})
}

func TestProposeSetStatusOnSource(t *testing.T) {
	t.Parallel()
	args := createMockArgs()
	sb := mock.NewBridgeStub()
	args.SourceBridge = sb
	executor, err := NewEthElrondBridgeExecutor(args)
	assert.Nil(t, err)
	assert.False(t, executor.IsInterfaceNil())

	executor.ProposeSetStatusOnSource(nil)
	assert.Equal(t, 1, sb.GetFunctionCounter("ProposeSetStatus"))
}

func TestCleanTopology(t *testing.T) {
	t.Parallel()
	args := createMockArgs()
	tp := mock.NewTopologyProviderStub()
	args.TopologyProvider = tp
	executor, err := NewEthElrondBridgeExecutor(args)
	assert.Nil(t, err)
	executor.CleanTopology()
	assert.Equal(t, 1, tp.GetFunctionCounter("Clean"))
}

func TestExecuteTransferOnDestination(t *testing.T) {
	t.Parallel()
	args := createMockArgs()
	db := mock.NewBridgeStub()
	args.DestinationBridge = db
	executor, err := NewEthElrondBridgeExecutor(args)
	assert.Nil(t, err)
	assert.False(t, executor.IsInterfaceNil())
	executor.ExecuteTransferOnDestination(nil)
	assert.Equal(t, 1, db.GetFunctionCounter("Execute"))
}

func TestExecuteTransferOnDestinationReturnsError(t *testing.T) {
	t.Parallel()
	args := createMockArgs()
	db := mock.NewBridgeStub()
	db.ExecuteError = errors.New("some error")
	args.DestinationBridge = db
	executor, err := NewEthElrondBridgeExecutor(args)
	assert.Nil(t, err)
	assert.False(t, executor.IsInterfaceNil())
	executor.ExecuteTransferOnDestination(nil)
	assert.Equal(t, 1, db.GetFunctionCounter("Execute"))
}

func TestExecuteSetStatusOnSource(t *testing.T) {
	t.Parallel()
	args := createMockArgs()
	sb := mock.NewBridgeStub()
	args.SourceBridge = sb
	executor, err := NewEthElrondBridgeExecutor(args)
	assert.Nil(t, err)
	assert.False(t, executor.IsInterfaceNil())
	executor.ExecuteSetStatusOnSource(nil)
	assert.Equal(t, 1, sb.GetFunctionCounter("Execute"))
}

func TestExecuteSetStatusOnSourceReturnsError(t *testing.T) {
	t.Parallel()
	args := createMockArgs()
	sb := mock.NewBridgeStub()
	sb.ExecuteError = errors.New("some error")
	args.SourceBridge = sb
	executor, err := NewEthElrondBridgeExecutor(args)
	assert.Nil(t, err)
	assert.False(t, executor.IsInterfaceNil())
	executor.ExecuteSetStatusOnSource(nil)
	assert.Equal(t, 1, sb.GetFunctionCounter("Execute"))
}

func TestSetStatusRejectedOnAllTransactions(t *testing.T) {
	t.Parallel()
	expected := &bridge.Batch{
		Id: bridge.NewBatchId(1),
		Transactions: []*bridge.DepositTransaction{
			{To: "address1", DepositNonce: bridge.NewNonce(0)},
			{To: "address2", DepositNonce: bridge.NewNonce(1)},
			{To: "address3", DepositNonce: bridge.NewNonce(2)},
		},
	}
	expectedError := errors.New("some error")
	args := createMockArgs()
	sb := mock.NewBridgeStub()
	sb.GetPendingCalled = func(ctx context.Context) *bridge.Batch {
		return expected
	}
	args.SourceBridge = sb
	executor, err := NewEthElrondBridgeExecutor(args)
	assert.Nil(t, err)
	assert.False(t, executor.IsInterfaceNil())
	assert.False(t, executor.HasPendingBatch())
	executor.GetPendingBatch(nil)

	assert.True(t, executor.HasPendingBatch())
	executor.SetStatusRejectedOnAllTransactions(expectedError)
	for _, transaction := range executor.pendingBatch.Transactions {
		assert.Equal(t, bridge.Rejected, transaction.Status)
		assert.Equal(t, expectedError, transaction.Error)
	}
}

func TestSignProposeTransferOnDestination(t *testing.T) {
	t.Parallel()
	args := createMockArgs()
	db := mock.NewBridgeStub()
	db.SignCalled = func(ctx context.Context, id bridge.ActionId) (string, error) {
		return "sign-tx-has", nil
	}
	db.GetActionIdForProposeTransferCalled = func(ctx context.Context, batch *bridge.Batch) bridge.ActionId {
		return bridge.NewActionId(1)
	}
	args.DestinationBridge = db
	executor, err := NewEthElrondBridgeExecutor(args)
	assert.Nil(t, err)
	assert.False(t, executor.IsInterfaceNil())
	assert.False(t, executor.HasPendingBatch())

	executor.GetPendingBatch(nil)

	assert.False(t, executor.HasPendingBatch())
	executor.SignProposeTransferOnDestination(nil)
	assert.Equal(t, 1, db.GetFunctionCounter("GetActionIdForProposeTransfer"))
	assert.Equal(t, 1, db.GetFunctionCounter("Sign"))
}

func TestSignProposeTransferOnDestinationReturnsError(t *testing.T) {
	t.Parallel()
	args := createMockArgs()
	db := mock.NewBridgeStub()
	db.SignError = errors.New("some error")
	db.GetActionIdForProposeTransferCalled = func(ctx context.Context, batch *bridge.Batch) bridge.ActionId {
		return bridge.NewActionId(1)
	}
	args.DestinationBridge = db

	executor, err := NewEthElrondBridgeExecutor(args)
	assert.Nil(t, err)
	assert.False(t, executor.IsInterfaceNil())
	assert.False(t, executor.HasPendingBatch())

	executor.GetPendingBatch(nil)

	assert.False(t, executor.HasPendingBatch())
	executor.SignProposeTransferOnDestination(nil)
	assert.Equal(t, 1, db.GetFunctionCounter("GetActionIdForProposeTransfer"))
	assert.Equal(t, 1, db.GetFunctionCounter("Sign"))
}

func TestSignProposeSetStatusOnSource(t *testing.T) {
	t.Parallel()
	expected := &bridge.Batch{
		Id:           bridge.NewBatchId(1),
		Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
	}
	args := createMockArgs()
	sb := mock.NewBridgeStub()
	sb.GetPendingCalled = func(ctx context.Context) *bridge.Batch {
		return expected
	}
	sb.SignCalled = func(ctx context.Context, id bridge.ActionId) (string, error) {
		return "sign-tx-has", nil
	}
	sb.GetActionIdForSetStatusOnPendingTransferCalled = func(ctx context.Context, batch *bridge.Batch) bridge.ActionId {
		return bridge.NewActionId(1)
	}
	args.SourceBridge = sb
	executor, err := NewEthElrondBridgeExecutor(args)
	assert.Nil(t, err)
	assert.False(t, executor.IsInterfaceNil())
	assert.False(t, executor.HasPendingBatch())

	executor.GetPendingBatch(nil)

	assert.True(t, executor.HasPendingBatch())
	executor.SignProposeSetStatusOnSource(nil)
	assert.Equal(t, 1, sb.GetFunctionCounter("GetActionIdForSetStatusOnPendingTransfer"))
	assert.Equal(t, 1, sb.GetFunctionCounter("Sign"))
}

func TestSignProposeSetStatusOnSourceReturnsError(t *testing.T) {
	t.Parallel()
	expected := &bridge.Batch{
		Id:           bridge.NewBatchId(1),
		Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
	}
	args := createMockArgs()
	sb := mock.NewBridgeStub()
	sb.GetPendingCalled = func(ctx context.Context) *bridge.Batch {
		return expected
	}
	sb.SignError = errors.New("some error")
	sb.GetActionIdForSetStatusOnPendingTransferCalled = func(ctx context.Context, batch *bridge.Batch) bridge.ActionId {
		return bridge.NewActionId(1)
	}
	args.SourceBridge = sb

	executor, err := NewEthElrondBridgeExecutor(args)
	assert.Nil(t, err)
	assert.False(t, executor.IsInterfaceNil())
	assert.False(t, executor.HasPendingBatch())

	executor.GetPendingBatch(nil)

	assert.True(t, executor.HasPendingBatch())
	executor.SignProposeSetStatusOnSource(nil)
	assert.Equal(t, 1, sb.GetFunctionCounter("GetActionIdForSetStatusOnPendingTransfer"))
	assert.Equal(t, 1, sb.GetFunctionCounter("Sign"))
}

func TestWaitStepToFinish(t *testing.T) {
	t.Parallel()
	t.Run("wait 0s", func(t *testing.T) {
		args := createMockArgs()
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*0)
		defer cancel()

		err = executor.WaitStepToFinish(ethToElrond.GettingPending, ctx)
		assert.Equal(t, ctx.Err(), err)
	})
	t.Run("wait < defaultWaitTime", func(t *testing.T) {
		args := createMockArgs()
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())
		ctx, cancel := context.WithTimeout(context.Background(), testDuration-time.Millisecond*500)
		defer cancel()

		err = executor.WaitStepToFinish(ethToElrond.GettingPending, ctx)
		assert.Equal(t, ctx.Err(), err)
	})
	t.Run("wait > defaultWaitTime", func(t *testing.T) {
		args := createMockArgs()
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.False(t, executor.IsInterfaceNil())
		ctx, cancel := context.WithTimeout(context.Background(), testDuration+time.Millisecond*500)
		defer cancel()

		err = executor.WaitStepToFinish(ethToElrond.GettingPending, ctx)
		assert.Nil(t, err)
	})
}

func TestUpdateTransactionsStatusesAccordingToDestination(t *testing.T) {
	t.Parallel()
	t.Run("destinationBridge.GetTransactionsStatuses returns error", func(t *testing.T) {
		args := createMockArgs()
		db := mock.NewBridgeStub()
		expectedErr := errors.New("expected error")
		db.GetTransactionsStatusesCalled = func(ctx context.Context, batchID bridge.BatchId) ([]uint8, error) {
			return nil, expectedErr
		}
		args.DestinationBridge = db
		executor, err := NewEthElrondBridgeExecutor(args)
		require.Nil(t, err)
		batch := &bridge.Batch{
			Transactions: []*bridge.DepositTransaction{
				{
					Status: 0,
				},
			},
		}
		executor.SetPendingBatch(batch)

		err = executor.UpdateTransactionsStatusesIfNeeded(nil)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("destinationBridge.GetTransactionsStatuses empty response", func(t *testing.T) {
		args := createMockArgs()
		db := mock.NewBridgeStub()
		db.GetTransactionsStatusesCalled = func(ctx context.Context, batchID bridge.BatchId) ([]uint8, error) {
			return make([]byte, 0), nil
		}
		args.DestinationBridge = db
		executor, err := NewEthElrondBridgeExecutor(args)
		require.Nil(t, err)
		batch := &bridge.Batch{
			Transactions: []*bridge.DepositTransaction{
				{},
			},
		}
		executor.SetPendingBatch(batch)

		err = executor.UpdateTransactionsStatusesIfNeeded(nil)
		assert.True(t, errors.Is(err, ErrBatchIDStatusMismatch))
	})
	t.Run("destinationBridge.GetTransactionsStatuses sets the status", func(t *testing.T) {
		args := createMockArgs()
		db := mock.NewBridgeStub()
		numTxs := 10
		statuses := make([]byte, numTxs)
		for i := 0; i < numTxs; i++ {
			statuses[i] = byte(i)
		}

		db.GetTransactionsStatusesCalled = func(ctx context.Context, batchID bridge.BatchId) ([]uint8, error) {
			return statuses, nil
		}
		args.DestinationBridge = db
		executor, err := NewEthElrondBridgeExecutor(args)
		require.Nil(t, err)

		batch := &bridge.Batch{}
		for i := 0; i < numTxs; i++ {
			batch.Transactions = append(batch.Transactions, &bridge.DepositTransaction{})
		}
		executor.SetPendingBatch(batch)

		err = executor.UpdateTransactionsStatusesIfNeeded(nil)
		assert.Nil(t, err)

		assert.Equal(t, numTxs, len(batch.Transactions)) // extra-protection that the number of txs was not modified
		for i := 0; i < numTxs; i++ {
			assert.Equal(t, byte(i), batch.Transactions[i].Status)
		}
	})
	t.Run("destinationBridge.GetTransactionsStatuses rejected transactions should not call destination bridge", func(t *testing.T) {
		args := createMockArgs()
		db := mock.NewBridgeStub()
		numTxs := 10
		statuses := make([]byte, numTxs)
		for i := 0; i < numTxs; i++ {
			statuses[i] = byte(i)
		}

		db.GetTransactionsStatusesCalled = func(ctx context.Context, batchID bridge.BatchId) ([]uint8, error) {
			require.Fail(t, "should have not called the destination bridge")
			return nil, nil
		}
		args.DestinationBridge = db
		executor, err := NewEthElrondBridgeExecutor(args)
		require.Nil(t, err)

		batch := &bridge.Batch{}
		for i := 0; i < numTxs; i++ {
			tx := &bridge.DepositTransaction{
				Status: bridge.Rejected,
			}
			batch.Transactions = append(batch.Transactions, tx)
		}
		executor.SetPendingBatch(batch)

		err = executor.UpdateTransactionsStatusesIfNeeded(nil)
		assert.Nil(t, err)

		assert.Equal(t, numTxs, len(batch.Transactions)) // extra-protection that the number of txs was not modified
		for i := 0; i < numTxs; i++ {
			assert.Equal(t, bridge.Rejected, batch.Transactions[i].Status)
		}
	})
	t.Run("destinationBridge.GetTransactionsStatuses nil pending batch should not call destination bridge", func(t *testing.T) {
		args := createMockArgs()
		db := mock.NewBridgeStub()
		db.GetTransactionsStatusesCalled = func(ctx context.Context, batchID bridge.BatchId) ([]uint8, error) {
			require.Fail(t, "should have not called the destination bridge")
			return nil, nil
		}
		args.DestinationBridge = db
		executor, err := NewEthElrondBridgeExecutor(args)
		require.Nil(t, err)

		err = executor.UpdateTransactionsStatusesIfNeeded(nil)
		assert.Nil(t, err)
	})
	t.Run("destinationBridge.GetTransactionsStatuses one tx was not rejected should call the destination bridge", func(t *testing.T) {
		args := createMockArgs()
		db := mock.NewBridgeStub()
		numTxs := 10
		statuses := make([]byte, numTxs)
		for i := 0; i < numTxs; i++ {
			statuses[i] = byte(i)
		}

		db.GetTransactionsStatusesCalled = func(ctx context.Context, batchID bridge.BatchId) ([]uint8, error) {
			return statuses, nil
		}
		args.DestinationBridge = db
		executor, err := NewEthElrondBridgeExecutor(args)
		require.Nil(t, err)

		batch := &bridge.Batch{}
		for i := 0; i < numTxs; i++ {
			tx := &bridge.DepositTransaction{
				Status: bridge.Rejected,
			}
			if i == numTxs-1 {
				tx.Status = bridge.Executed
			}
			batch.Transactions = append(batch.Transactions, tx)
		}
		executor.SetPendingBatch(batch)

		err = executor.UpdateTransactionsStatusesIfNeeded(nil)
		assert.Nil(t, err)

		assert.Equal(t, numTxs, len(batch.Transactions)) // extra-protection that the number of txs was not modified
		for i := 0; i < numTxs; i++ {
			assert.Equal(t, statuses[i], batch.Transactions[i].Status)
		}
	})
}
