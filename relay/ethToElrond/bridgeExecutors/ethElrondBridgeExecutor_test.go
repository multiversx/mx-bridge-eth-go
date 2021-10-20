package bridgeExecutors

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond/bridgeExecutors/mock"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"
)

func createMockArgs() ArgsEthElrondBridgeExecutor {
	return ArgsEthElrondBridgeExecutor{
		ExecutorName:      "executorMock",
		Logger:            logger.GetOrCreate("test"),
		SourceBridge:      &mock.BridgeStub{},
		DestinationBridge: &mock.BridgeStub{},
		TopologyProvider:  &mock.TopologyProviderStub{},
		QuorumProvider:    &mock.QuorumProviderStub{},
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
}

func TestGetPending(t *testing.T) {
	t.Parallel()
	t.Run("no pending transaction", func(t *testing.T) {
		args := createMockArgs()
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.Equal(t, false, executor.HasPendingBatch())
		ctx := context.Background()
		executor.GetPendingBatch(ctx)

		assert.Equal(t, false, executor.HasPendingBatch())
		assert.Nil(t, executor.pendingBatch)
	})
	t.Run("it will get the next pending transaction", func(t *testing.T) {
		expected := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		args := createMockArgs()
		args.SourceBridge = &mock.BridgeStub{
			GetPendingCalled: func(ctx context.Context) *bridge.Batch {
				return expected
			},
		}
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.Equal(t, false, executor.HasPendingBatch())
		executor.GetPendingBatch(nil)

		assert.Equal(t, true, executor.HasPendingBatch())
		assert.Equal(t, expected, executor.pendingBatch)
	})
}

func TestLider(t *testing.T) {
	t.Parallel()
	t.Run("relayer is lider", func(t *testing.T) {
		args := createMockArgs()
		args.TopologyProvider = &mock.TopologyProviderStub{
			AmITheLeaderCalled: func() bool {
				return true
			}}
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.Equal(t, true, executor.IsLeader())
	})
	t.Run("relayer is NOT lider", func(t *testing.T) {
		args := createMockArgs()
		args.TopologyProvider = &mock.TopologyProviderStub{
			AmITheLeaderCalled: func() bool {
				return false
			}}
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		assert.Equal(t, false, executor.IsLeader())
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
		args.SourceBridge = &mock.BridgeStub{
			GetPendingCalled: func(ctx context.Context) *bridge.Batch {
				return expected
			},
		}
		args.DestinationBridge = &mock.BridgeStub{
			WasProposedTransferCalled: func(ctx context.Context, batch *bridge.Batch) bool {
				return true
			},
		}
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		ctx := context.Background()
		executor.GetPendingBatch(ctx)
		assert.Equal(t, true, executor.WasProposeTransferExecutedOnDestination(ctx))
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
		args.SourceBridge = &mock.BridgeStub{
			GetPendingCalled: func(ctx context.Context) *bridge.Batch {
				return expected
			},
			WasProposedSetStatusCalled: func(ctx context.Context, batch *bridge.Batch) bool {
				return true
			},
		}
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		ctx := context.Background()
		executor.GetPendingBatch(ctx)
		assert.Equal(t, true, executor.WasProposeSetStatusExecutedOnSource(ctx))
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
		args.SourceBridge = &mock.BridgeStub{
			GetPendingCalled: func(ctx context.Context) *bridge.Batch {
				return expected
			},
		}
		args.DestinationBridge = &mock.BridgeStub{
			WasExecutedCalled: func(ctx context.Context, id bridge.ActionId, id2 bridge.BatchId) bool {
				return true
			},
		}
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		ctx := context.Background()
		executor.GetPendingBatch(ctx)
		assert.Equal(t, true, executor.WasExecutedOnDestination(ctx))
	})
	t.Run("OnDestination", func(t *testing.T) {
		expected := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		args := createMockArgs()
		args.SourceBridge = &mock.BridgeStub{
			GetPendingCalled: func(ctx context.Context) *bridge.Batch {
				return expected
			},
			WasExecutedCalled: func(ctx context.Context, id bridge.ActionId, id2 bridge.BatchId) bool {
				return true
			},
		}
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		ctx := context.Background()
		executor.GetPendingBatch(ctx)
		assert.Equal(t, true, executor.WasExecutedOnSource(ctx))
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
		ctx := context.Background()
		assert.Equal(t, false, executor.IsQuorumReachedForProposeTransfer(ctx))
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
		ctx := context.Background()
		assert.Equal(t, false, executor.IsQuorumReachedForProposeTransfer(ctx))
	})
	t.Run("less < quorum", func(t *testing.T) {
		args := createMockArgs()
		args.QuorumProvider = &mock.QuorumProviderStub{
			GetQuorumCalled: func(ctx context.Context) (uint, error) {
				return 3, nil
			},
		}

		args.DestinationBridge = &mock.BridgeStub{
			SignersCountCalled: func(ctx context.Context, id bridge.ActionId) uint {
				return 2
			},
		}
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		ctx := context.Background()
		assert.Equal(t, false, executor.IsQuorumReachedForProposeTransfer(ctx))
	})
	t.Run("signs == quorum", func(t *testing.T) {
		args := createMockArgs()
		args.QuorumProvider = &mock.QuorumProviderStub{
			GetQuorumCalled: func(ctx context.Context) (uint, error) {
				return 3, nil
			},
		}

		args.DestinationBridge = &mock.BridgeStub{
			SignersCountCalled: func(ctx context.Context, id bridge.ActionId) uint {
				return 3
			},
		}
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		ctx := context.Background()
		assert.Equal(t, true, executor.IsQuorumReachedForProposeTransfer(ctx))
	})
	t.Run("signs > quorum", func(t *testing.T) {
		args := createMockArgs()
		args.QuorumProvider = &mock.QuorumProviderStub{
			GetQuorumCalled: func(ctx context.Context) (uint, error) {
				return 3, nil
			},
		}

		args.DestinationBridge = &mock.BridgeStub{
			SignersCountCalled: func(ctx context.Context, id bridge.ActionId) uint {
				return 4
			},
		}
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		ctx := context.Background()
		assert.Equal(t, true, executor.IsQuorumReachedForProposeTransfer(ctx))
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
		ctx := context.Background()
		assert.Equal(t, false, executor.IsQuorumReachedForProposeSetStatus(ctx))
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
		ctx := context.Background()
		assert.Equal(t, false, executor.IsQuorumReachedForProposeSetStatus(ctx))
	})
	t.Run("less < quorum", func(t *testing.T) {
		args := createMockArgs()
		args.QuorumProvider = &mock.QuorumProviderStub{
			GetQuorumCalled: func(ctx context.Context) (uint, error) {
				return 3, nil
			},
		}

		args.SourceBridge = &mock.BridgeStub{
			SignersCountCalled: func(ctx context.Context, id bridge.ActionId) uint {
				return 2
			},
		}
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		ctx := context.Background()
		assert.Equal(t, false, executor.IsQuorumReachedForProposeSetStatus(ctx))
	})
	t.Run("signs == quorum", func(t *testing.T) {
		args := createMockArgs()
		args.QuorumProvider = &mock.QuorumProviderStub{
			GetQuorumCalled: func(ctx context.Context) (uint, error) {
				return 3, nil
			},
		}

		args.SourceBridge = &mock.BridgeStub{
			SignersCountCalled: func(ctx context.Context, id bridge.ActionId) uint {
				return 3
			},
		}
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		ctx := context.Background()
		assert.Equal(t, true, executor.IsQuorumReachedForProposeSetStatus(ctx))
	})
	t.Run("signs > quorum", func(t *testing.T) {
		args := createMockArgs()
		args.QuorumProvider = &mock.QuorumProviderStub{
			GetQuorumCalled: func(ctx context.Context) (uint, error) {
				return 3, nil
			},
		}

		args.SourceBridge = &mock.BridgeStub{
			SignersCountCalled: func(ctx context.Context, id bridge.ActionId) uint {
				return 4
			},
		}
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		ctx := context.Background()
		assert.Equal(t, true, executor.IsQuorumReachedForProposeSetStatus(ctx))
	})

}

func TestProposeTransferOnDestination(t *testing.T) {
	t.Parallel()
	t.Run("ProposeTransferError", func(t *testing.T) {
		args := createMockArgs()
		expected_error := errors.New("some error")
		args.DestinationBridge = &mock.BridgeStub{
			ProposeTransferError: expected_error,
		}
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		ctx := context.Background()
		err = executor.ProposeTransferOnDestination(ctx)
		assert.Equal(t, expected_error, err)
	})
	t.Run("no error", func(t *testing.T) {
		args := createMockArgs()
		args.DestinationBridge = &mock.BridgeStub{
			ProposeTransferCalled: func(ctx context.Context, batch *bridge.Batch) (string, error) {
				return "propose_tx_hash", nil
			},
		}
		executor, err := NewEthElrondBridgeExecutor(args)
		assert.Nil(t, err)
		ctx := context.Background()
		err = executor.ProposeTransferOnDestination(ctx)
		assert.Nil(t, err)
	})
}
