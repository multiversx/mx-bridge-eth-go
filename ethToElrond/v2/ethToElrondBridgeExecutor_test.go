package v2

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"
)

func createMockEthToElrondExecutorArgs() ArgsEthToElrondBridgeExecutor {
	return ArgsEthToElrondBridgeExecutor{
		Log:              logger.GetOrCreate("test"),
		TopologyProvider: &bridgeV2.TopologyProviderStub{},
		ElrondClient:     &bridgeV2.ElrondClientStub{},
		EthereumClient:   &bridgeV2.EthereumClientStub{},
	}
}

func TestNewEthToElrondBridgeExecutor(t *testing.T) {
	t.Parallel()

	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		args.Log = nil
		executor, err := NewEthToElrondBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, errNilLogger, err)
	})
	t.Run("nil elrond client should error", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		args.ElrondClient = nil
		executor, err := NewEthToElrondBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, errNilElrondClient, err)
	})
	t.Run("nil ethereum client should error", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		args.EthereumClient = nil
		executor, err := NewEthToElrondBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, errNilEthereumClient, err)
	})
	t.Run("nil topology provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		args.TopologyProvider = nil
		executor, err := NewEthToElrondBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, errNilTopologyProvider, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		executor, err := NewEthToElrondBridgeExecutor(args)

		assert.False(t, check.IfNil(executor))
		assert.Nil(t, err)
	})
}

func TestEthToElrondBridgeExecutor_GetLogger(t *testing.T) {
	t.Parallel()

	args := createMockEthToElrondExecutorArgs()
	executor, _ := NewEthToElrondBridgeExecutor(args)

	assert.True(t, args.Log == executor.GetLogger()) //pointer testing
}

func TestEthToElrondBridgeExecutor_MyTurnAsLeader(t *testing.T) {
	t.Parallel()

	args := createMockEthToElrondExecutorArgs()
	wasCalled := false
	args.TopologyProvider = &bridgeV2.TopologyProviderStub{
		MyTurnAsLeaderCalled: func() bool {
			wasCalled = true
			return true
		},
	}

	executor, _ := NewEthToElrondBridgeExecutor(args)
	assert.True(t, executor.MyTurnAsLeader())
	assert.True(t, wasCalled)
}

func TestEthToElrondBridgeExecutor_GetAndStoreActionIDFromElrond(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		executor, _ := NewEthToElrondBridgeExecutor(args)

		actionID, err := executor.GetAndStoreActionIDFromElrond(context.Background())
		assert.Zero(t, actionID)
		assert.Equal(t, errNilBatch, err)
	})
	t.Run("elrond client errors", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		expectedErr := errors.New("expected error")
		providedBatch := &clients.TransferBatch{}

		args.ElrondClient = &bridgeV2.ElrondClientStub{
			GetActionIDForProposeTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (uint64, error) {
				assert.True(t, providedBatch == batch)
				return 0, expectedErr
			},
		}
		executor, _ := NewEthToElrondBridgeExecutor(args)
		executor.batch = providedBatch

		actionID, err := executor.GetAndStoreActionIDFromElrond(context.Background())
		assert.Zero(t, actionID)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		providedBatch := &clients.TransferBatch{}
		providedActionID := uint64(48939)

		args.ElrondClient = &bridgeV2.ElrondClientStub{
			GetActionIDForProposeTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (uint64, error) {
				assert.True(t, providedBatch == batch)
				return providedActionID, nil
			},
		}
		executor, _ := NewEthToElrondBridgeExecutor(args)
		executor.batch = providedBatch

		assert.NotEqual(t, providedActionID, executor.actionID)

		actionID, err := executor.GetAndStoreActionIDFromElrond(context.Background())
		assert.Equal(t, providedActionID, actionID)
		assert.Nil(t, err)
		assert.Equal(t, providedActionID, executor.GetStoredActionID())
		assert.Equal(t, providedActionID, executor.actionID)
	})
}

func TestEthToElrondBridgeExecutor_GetAndStoreBatchFromEthereum(t *testing.T) {
	t.Parallel()

	t.Run("ethereum client errors", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		providedNonce := uint64(8346)
		expectedErr := errors.New("expected error")
		args.EthereumClient = &bridgeV2.EthereumClientStub{
			GetBatchCalled: func(ctx context.Context, nonce uint64) (*clients.TransferBatch, error) {
				assert.Equal(t, providedNonce, nonce)
				return nil, expectedErr
			},
		}
		executor, _ := NewEthToElrondBridgeExecutor(args)
		err := executor.GetAndStoreBatchFromEthereum(context.Background(), providedNonce)

		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		providedNonce := uint64(8346)
		expectedBatch := &clients.TransferBatch{}
		args.EthereumClient = &bridgeV2.EthereumClientStub{
			GetBatchCalled: func(ctx context.Context, nonce uint64) (*clients.TransferBatch, error) {
				assert.Equal(t, providedNonce, nonce)
				return expectedBatch, nil
			},
		}
		executor, _ := NewEthToElrondBridgeExecutor(args)
		err := executor.GetAndStoreBatchFromEthereum(context.Background(), providedNonce)

		assert.Nil(t, err)
		assert.True(t, expectedBatch == executor.GetStoredBatch()) // pointer testing
		assert.True(t, expectedBatch == executor.batch)
	})
}

func TestEthToElrondBridgeExecutor_GetLastExecutedEthBatchIDFromElrond(t *testing.T) {
	t.Parallel()

	args := createMockEthToElrondExecutorArgs()
	providedBatchID := uint64(36727)
	args.ElrondClient = &bridgeV2.ElrondClientStub{
		GetLastExecutedEthBatchIDCalled: func(ctx context.Context) (uint64, error) {
			return providedBatchID, nil
		},
	}
	executor, _ := NewEthToElrondBridgeExecutor(args)

	batchID, err := executor.GetLastExecutedEthBatchIDFromElrond(context.Background())
	assert.Equal(t, providedBatchID, batchID)
	assert.Nil(t, err)
}

func TestEthToElrondBridgeExecutor_VerifyLastDepositNonceExecutedOnEthereumBatch(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		executor, _ := NewEthToElrondBridgeExecutor(args)

		err := executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
		assert.Equal(t, errNilBatch, err)
	})
	t.Run("get last executed tx id errors", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		expectedErr := errors.New("expected error")
		args.ElrondClient = &bridgeV2.ElrondClientStub{
			GetLastExecutedEthTxIDCalled: func(ctx context.Context) (uint64, error) {
				return 0, expectedErr
			},
		}
		executor, _ := NewEthToElrondBridgeExecutor(args)
		executor.batch = &clients.TransferBatch{}

		err := executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
		assert.Equal(t, expectedErr, err)
	})

	args := createMockEthToElrondExecutorArgs()
	txId := uint64(6657)
	args.ElrondClient = &bridgeV2.ElrondClientStub{
		GetLastExecutedEthTxIDCalled: func(ctx context.Context) (uint64, error) {
			return txId, nil
		},
	}

	t.Run("first deposit nonce equals last tx nonce should error", func(t *testing.T) {
		t.Parallel()

		executor, _ := NewEthToElrondBridgeExecutor(args)
		executor.batch = &clients.TransferBatch{
			Deposits: []*clients.DepositTransfer{
				{
					Nonce: txId,
				},
			},
		}

		err := executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
		assert.True(t, errors.Is(err, errInvalidDepositNonce))
		assert.True(t, strings.Contains(err.Error(), "6657"))
	})
	t.Run("first deposit nonce is smaller than the last tx nonce should error", func(t *testing.T) {
		t.Parallel()

		executor, _ := NewEthToElrondBridgeExecutor(args)
		executor.batch = &clients.TransferBatch{
			Deposits: []*clients.DepositTransfer{
				{
					Nonce: txId - 1,
				},
			},
		}

		err := executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
		assert.True(t, errors.Is(err, errInvalidDepositNonce))
		assert.True(t, strings.Contains(err.Error(), "6656"))
	})
	t.Run("gap found error", func(t *testing.T) {
		t.Parallel()

		executor, _ := NewEthToElrondBridgeExecutor(args)
		executor.batch = &clients.TransferBatch{
			Deposits: []*clients.DepositTransfer{
				{
					Nonce: txId + 1,
				},
				{
					Nonce: txId + 3,
				},
			},
		}

		err := executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
		assert.True(t, errors.Is(err, errInvalidDepositNonce))
		assert.True(t, strings.Contains(err.Error(), "6660"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		executor, _ := NewEthToElrondBridgeExecutor(args)
		executor.batch = &clients.TransferBatch{
			Deposits: []*clients.DepositTransfer{
				{
					Nonce: txId + 1,
				},
			},
		}

		err := executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
		assert.Nil(t, err)

		executor.batch = &clients.TransferBatch{
			Deposits: []*clients.DepositTransfer{
				{
					Nonce: txId + 1,
				},
				{
					Nonce: txId + 2,
				},
			},
		}

		err = executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
		assert.Nil(t, err)
	})
}

func TestEthToElrondBridgeExecutor_WasTransferProposedOnElrond(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		executor, _ := NewEthToElrondBridgeExecutor(args)

		wasTransfered, err := executor.WasTransferProposedOnElrond(context.Background())
		assert.False(t, wasTransfered)
		assert.Equal(t, errNilBatch, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		providedBatch := &clients.TransferBatch{}
		wasCalled := false
		args.ElrondClient = &bridgeV2.ElrondClientStub{
			WasProposedTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
				assert.True(t, providedBatch == batch)
				wasCalled = true
				return true, nil
			},
		}

		executor, _ := NewEthToElrondBridgeExecutor(args)
		executor.batch = providedBatch

		wasProposed, err := executor.WasTransferProposedOnElrond(context.Background())
		assert.True(t, wasProposed)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestEthToElrondBridgeExecutor_ProposeTransferOnElrond(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		executor, _ := NewEthToElrondBridgeExecutor(args)

		err := executor.ProposeTransferOnElrond(context.Background())
		assert.Equal(t, errNilBatch, err)
	})
	t.Run("propose transfer fails", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		providedBatch := &clients.TransferBatch{}
		expectedErr := errors.New("expected error")
		args.ElrondClient = &bridgeV2.ElrondClientStub{
			ProposeTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (string, error) {
				assert.True(t, providedBatch == batch)

				return "", expectedErr
			},
		}
		executor, _ := NewEthToElrondBridgeExecutor(args)
		executor.batch = providedBatch

		err := executor.ProposeTransferOnElrond(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		providedBatch := &clients.TransferBatch{}
		wasCalled := false
		args.ElrondClient = &bridgeV2.ElrondClientStub{
			ProposeTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (string, error) {
				assert.True(t, providedBatch == batch)
				wasCalled = true

				return "", nil
			},
		}
		executor, _ := NewEthToElrondBridgeExecutor(args)
		executor.batch = providedBatch

		err := executor.ProposeTransferOnElrond(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestEthToElrondBridgeExecutor_WasProposedTransferSignedOnElrond(t *testing.T) {
	t.Parallel()

	args := createMockEthToElrondExecutorArgs()
	providedActionID := uint64(378276)
	wasCalled := false
	args.ElrondClient = &bridgeV2.ElrondClientStub{
		WasExecutedCalled: func(ctx context.Context, actionID uint64) (bool, error) {
			assert.Equal(t, providedActionID, actionID)
			wasCalled = true
			return true, nil
		},
	}
	executor, _ := NewEthToElrondBridgeExecutor(args)
	executor.actionID = providedActionID

	wasSigned, err := executor.WasProposedTransferSignedOnElrond(context.Background())
	assert.True(t, wasSigned)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestEthToElrondBridgeExecutor_SignProposedTransferOnElrond(t *testing.T) {
	t.Parallel()

	t.Run("elrond client errors", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		expectedErr := errors.New("expected error")
		providedActionID := uint64(378276)
		args.ElrondClient = &bridgeV2.ElrondClientStub{
			SignCalled: func(ctx context.Context, actionID uint64) (string, error) {
				assert.Equal(t, providedActionID, actionID)
				return "", expectedErr
			},
		}

		executor, _ := NewEthToElrondBridgeExecutor(args)
		executor.actionID = providedActionID

		err := executor.SignProposedTransferOnElrond(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		providedActionID := uint64(378276)
		wasCalled := false
		args.ElrondClient = &bridgeV2.ElrondClientStub{
			SignCalled: func(ctx context.Context, actionID uint64) (string, error) {
				assert.Equal(t, providedActionID, actionID)
				wasCalled = true
				return "", nil
			},
		}

		executor, _ := NewEthToElrondBridgeExecutor(args)
		executor.actionID = providedActionID

		err := executor.SignProposedTransferOnElrond(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestEthToElrondBridgeExecutor_IsQuorumReachedOnElrond(t *testing.T) {
	t.Parallel()

	args := createMockEthToElrondExecutorArgs()
	providedActionID := uint64(378276)
	wasCalled := false
	args.ElrondClient = &bridgeV2.ElrondClientStub{
		QuorumReachedCalled: func(ctx context.Context, actionID uint64) (bool, error) {
			assert.Equal(t, providedActionID, actionID)
			wasCalled = true
			return true, nil
		},
	}
	executor, _ := NewEthToElrondBridgeExecutor(args)
	executor.actionID = providedActionID

	isQuorumReached, err := executor.IsQuorumReachedOnElrond(context.Background())
	assert.True(t, isQuorumReached)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestEthToElrondBridgeExecutor_WasActionIDPerformedOnElrond(t *testing.T) {
	t.Parallel()

	args := createMockEthToElrondExecutorArgs()
	providedActionID := uint64(378276)
	wasCalled := false
	args.ElrondClient = &bridgeV2.ElrondClientStub{
		WasExecutedCalled: func(ctx context.Context, actionID uint64) (bool, error) {
			assert.Equal(t, providedActionID, actionID)
			wasCalled = true
			return true, nil
		},
	}
	executor, _ := NewEthToElrondBridgeExecutor(args)
	executor.actionID = providedActionID

	wasPerformed, err := executor.WasActionIDPerformedOnElrond(context.Background())
	assert.True(t, wasPerformed)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestEthToElrondBridgeExecutor_PerformActionIDOnElrond(t *testing.T) {
	t.Parallel()

	t.Run("nil batch", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		executor, _ := NewEthToElrondBridgeExecutor(args)

		err := executor.PerformActionIDOnElrond(context.Background())
		assert.Equal(t, errNilBatch, err)
	})
	t.Run("elrond client errors", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		expectedErr := errors.New("expected error")
		providedActionID := uint64(7383)
		providedBatch := &clients.TransferBatch{}
		args.ElrondClient = &bridgeV2.ElrondClientStub{
			PerformActionCalled: func(ctx context.Context, actionID uint64, batch *clients.TransferBatch) (string, error) {
				assert.Equal(t, providedActionID, actionID)
				assert.True(t, providedBatch == batch)
				return "", expectedErr
			},
		}
		executor, _ := NewEthToElrondBridgeExecutor(args)
		executor.batch = providedBatch
		executor.actionID = providedActionID

		err := executor.PerformActionIDOnElrond(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockEthToElrondExecutorArgs()
		wasCalled := false
		providedActionID := uint64(7383)
		providedBatch := &clients.TransferBatch{}
		args.ElrondClient = &bridgeV2.ElrondClientStub{
			PerformActionCalled: func(ctx context.Context, actionID uint64, batch *clients.TransferBatch) (string, error) {
				assert.Equal(t, providedActionID, actionID)
				assert.True(t, providedBatch == batch)
				wasCalled = true
				return "", nil
			},
		}
		executor, _ := NewEthToElrondBridgeExecutor(args)
		executor.batch = providedBatch
		executor.actionID = providedActionID

		err := executor.PerformActionIDOnElrond(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestEthToElrondBridgeExecutor_RetriesCount(t *testing.T) {
	t.Parallel()

	expectedMaxRetries := uint64(3)
	args := createMockEthToElrondExecutorArgs()
	wasCalledOnElrondClient := false
	args.ElrondClient = &bridgeV2.ElrondClientStub{
		GetMaxNumberOfRetriesAllowedCalled: func() uint64 {
			wasCalledOnElrondClient = true
			return expectedMaxRetries
		},
	}
	executor, _ := NewEthToElrondBridgeExecutor(args)
	for i := uint64(0); i < expectedMaxRetries; i++ {
		assert.False(t, executor.ProcessMaxRetriesOnElrond())
	}

	// Test elrond
	assert.Equal(t, expectedMaxRetries, executor.retriesOnElrond)
	assert.True(t, executor.ProcessMaxRetriesOnElrond())
	executor.ResetRetriesCountOnElrond()
	assert.Equal(t, uint64(0), executor.retriesOnElrond)
	assert.True(t, wasCalledOnElrondClient)
}
