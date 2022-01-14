package ethElrond

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	bridgeTests "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridge"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var expectedErr = errors.New("expected error")
var providedBatch = &clients.TransferBatch{}

func createMockExecutorArgs() ArgsBridgeExecutor {
	return ArgsBridgeExecutor{
		Log:                      logger.GetOrCreate("test"),
		ElrondClient:             &bridgeTests.ElrondClientStub{},
		EthereumClient:           &bridgeTests.EthereumClientStub{},
		TopologyProvider:         &bridgeTests.TopologyProviderStub{},
		StatusHandler:            testsCommon.NewStatusHandlerMock("test"),
		TimeForTransferExecution: time.Second,
		SignaturesHolder:         &testsCommon.SignaturesHolderStub{},
	}
}

func TestNewBridgeExecutor(t *testing.T) {
	t.Parallel()

	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.Log = nil
		executor, err := NewTestBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrNilLogger, err)
	})
	t.Run("nil elrond client should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.ElrondClient = nil
		executor, err := NewTestBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrNilElrondClient, err)
	})
	t.Run("nil ethereum client should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.EthereumClient = nil
		executor, err := NewTestBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrNilEthereumClient, err)
	})
	t.Run("nil topology provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TopologyProvider = nil
		executor, err := NewTestBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrNilTopologyProvider, err)
	})
	t.Run("nil status handler", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.StatusHandler = nil
		executor, err := NewTestBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrNilStatusHandler, err)
	})
	t.Run("invalid time", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForTransferExecution = 0
		executor, err := NewTestBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrInvalidDuration, err)
	})
	t.Run("nil signatures holder", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.SignaturesHolder = nil
		executor, err := NewTestBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrNilSignaturesHolder, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, err := NewTestBridgeExecutor(args)

		assert.False(t, check.IfNil(executor))
		assert.Nil(t, err)
	})
}

func TestEthToElrondBridgeExecutor_PrintInfo(t *testing.T) {
	t.Parallel()

	logLevels := []logger.LogLevel{logger.LogTrace, logger.LogDebug, logger.LogInfo, logger.LogWarning, logger.LogError, logger.LogNone}
	for _, logLevel := range logLevels {
		shouldOutputToStatusHandler := logLevel == logger.LogError || logLevel == logger.LogWarning
		testPrintInfo(t, logLevel, shouldOutputToStatusHandler)
	}
}

func testPrintInfo(t *testing.T, logLevel logger.LogLevel, shouldOutputToStatusHandler bool) {
	providedLogLevel := logLevel
	providedMessage := "message"
	providedArgs := []interface{}{"string", 1, []byte("aaa")}
	wasCalled := false

	args := createMockExecutorArgs()
	statusHandler := testsCommon.NewStatusHandlerMock("test")
	args.StatusHandler = statusHandler
	args.Log = &testsCommon.LoggerStub{
		LogCalled: func(logLevel logger.LogLevel, message string, args ...interface{}) {
			wasCalled = true
			assert.Equal(t, providedLogLevel, logLevel)
			assert.Equal(t, providedMessage, message)
			assert.Equal(t, providedArgs, args)
		},
	}
	executor, _ := NewTestBridgeExecutor(args)
	executor.PrintInfo(providedLogLevel, providedMessage, providedArgs...)

	assert.True(t, wasCalled)

	if shouldOutputToStatusHandler {
		assert.True(t, len(statusHandler.GetStringMetric(core.MetricLastError)) > 0)
	}
}

func TestEthToElrondBridgeExecutor_MyTurnAsLeader(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	wasCalled := false
	args.TopologyProvider = &bridgeTests.TopologyProviderStub{
		MyTurnAsLeaderCalled: func() bool {
			wasCalled = true
			return true
		},
	}

	executor, _ := NewTestBridgeExecutor(args)
	assert.True(t, executor.MyTurnAsLeader())
	assert.True(t, wasCalled)
}

func TestEthToElrondBridgeExecutor_GetAndStoreActionIDForProposeTransferOnElrond(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewTestBridgeExecutor(args)

		actionID, err := executor.GetAndStoreActionIDForProposeTransferOnElrond(context.Background())
		assert.Zero(t, actionID)
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("elrond client errors", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.ElrondClient = &bridgeTests.ElrondClientStub{
			GetActionIDForProposeTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (uint64, error) {
				assert.True(t, providedBatch == batch)
				return 0, expectedErr
			},
		}
		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatch

		actionID, err := executor.GetAndStoreActionIDForProposeTransferOnElrond(context.Background())
		assert.Zero(t, actionID)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedActionID := uint64(48939)

		args.ElrondClient = &bridgeTests.ElrondClientStub{
			GetActionIDForProposeTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (uint64, error) {
				assert.True(t, providedBatch == batch)
				return providedActionID, nil
			},
		}
		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatch

		assert.NotEqual(t, providedActionID, executor.actionID)

		actionID, err := executor.GetAndStoreActionIDForProposeTransferOnElrond(context.Background())
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

		args := createMockExecutorArgs()
		providedNonce := uint64(8346)
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetBatchCalled: func(ctx context.Context, nonce uint64) (*clients.TransferBatch, error) {
				assert.Equal(t, providedNonce, nonce)
				return nil, expectedErr
			},
		}
		executor, _ := NewTestBridgeExecutor(args)
		err := executor.GetAndStoreBatchFromEthereum(context.Background(), providedNonce)

		assert.Equal(t, expectedErr, err)
	})
	t.Run("batch nonce mismatch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedNonce := uint64(8346)
		expectedBatch := &clients.TransferBatch{
			ID: 0,
		}
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetBatchCalled: func(ctx context.Context, nonce uint64) (*clients.TransferBatch, error) {
				assert.Equal(t, providedNonce, nonce)
				return expectedBatch, nil
			},
		}
		executor, _ := NewTestBridgeExecutor(args)
		err := executor.GetAndStoreBatchFromEthereum(context.Background(), providedNonce)

		assert.True(t, errors.Is(err, ErrBatchNotFound))
		assert.True(t, strings.Contains(err.Error(), fmt.Sprintf("%d", providedNonce)))
		assert.Nil(t, executor.GetStoredBatch())
		assert.Nil(t, executor.batch)
	})
	t.Run("no deposits should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedNonce := uint64(8346)
		expectedBatch := &clients.TransferBatch{
			ID: providedNonce,
		}
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetBatchCalled: func(ctx context.Context, nonce uint64) (*clients.TransferBatch, error) {
				assert.Equal(t, providedNonce, nonce)
				return expectedBatch, nil
			},
		}
		executor, _ := NewTestBridgeExecutor(args)
		err := executor.GetAndStoreBatchFromEthereum(context.Background(), providedNonce)

		assert.True(t, errors.Is(err, ErrBatchNotFound))
		assert.True(t, strings.Contains(err.Error(), fmt.Sprintf("%d", providedNonce)))
		assert.Nil(t, executor.GetStoredBatch())
		assert.Nil(t, executor.batch)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedNonce := uint64(8346)
		expectedBatch := &clients.TransferBatch{
			ID: providedNonce,
			Deposits: []*clients.DepositTransfer{
				{},
			},
		}
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetBatchCalled: func(ctx context.Context, nonce uint64) (*clients.TransferBatch, error) {
				assert.Equal(t, providedNonce, nonce)
				return expectedBatch, nil
			},
		}
		executor, _ := NewTestBridgeExecutor(args)
		err := executor.GetAndStoreBatchFromEthereum(context.Background(), providedNonce)

		assert.Nil(t, err)
		assert.True(t, expectedBatch == executor.GetStoredBatch()) // pointer testing
		assert.True(t, expectedBatch == executor.batch)
	})
}

func TestEthToElrondBridgeExecutor_GetLastExecutedEthBatchIDFromElrond(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	providedBatchID := uint64(36727)
	args.ElrondClient = &bridgeTests.ElrondClientStub{
		GetLastExecutedEthBatchIDCalled: func(ctx context.Context) (uint64, error) {
			return providedBatchID, nil
		},
	}
	executor, _ := NewTestBridgeExecutor(args)

	batchID, err := executor.GetLastExecutedEthBatchIDFromElrond(context.Background())
	assert.Equal(t, providedBatchID, batchID)
	assert.Nil(t, err)
}

func TestEthToElrondBridgeExecutor_VerifyLastDepositNonceExecutedOnEthereumBatch(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewTestBridgeExecutor(args)

		err := executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("get last executed tx id errors", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.ElrondClient = &bridgeTests.ElrondClientStub{
			GetLastExecutedEthTxIDCalled: func(ctx context.Context) (uint64, error) {
				return 0, expectedErr
			},
		}
		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = &clients.TransferBatch{}

		err := executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
		assert.Equal(t, expectedErr, err)
	})

	args := createMockExecutorArgs()
	txId := uint64(6657)
	args.ElrondClient = &bridgeTests.ElrondClientStub{
		GetLastExecutedEthTxIDCalled: func(ctx context.Context) (uint64, error) {
			return txId, nil
		},
	}

	t.Run("first deposit nonce equals last tx nonce should error", func(t *testing.T) {
		t.Parallel()

		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = &clients.TransferBatch{
			Deposits: []*clients.DepositTransfer{
				{
					Nonce: txId,
				},
			},
		}

		err := executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
		assert.True(t, errors.Is(err, ErrInvalidDepositNonce))
		assert.True(t, strings.Contains(err.Error(), "6657"))
	})
	t.Run("first deposit nonce is smaller than the last tx nonce should error", func(t *testing.T) {
		t.Parallel()

		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = &clients.TransferBatch{
			Deposits: []*clients.DepositTransfer{
				{
					Nonce: txId - 1,
				},
			},
		}

		err := executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
		assert.True(t, errors.Is(err, ErrInvalidDepositNonce))
		assert.True(t, strings.Contains(err.Error(), "6656"))
	})
	t.Run("gap found error", func(t *testing.T) {
		t.Parallel()

		executor, _ := NewTestBridgeExecutor(args)
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
		assert.True(t, errors.Is(err, ErrInvalidDepositNonce))
		assert.True(t, strings.Contains(err.Error(), "6660"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		executor, _ := NewTestBridgeExecutor(args)
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

		args := createMockExecutorArgs()
		executor, _ := NewTestBridgeExecutor(args)

		wasTransfered, err := executor.WasTransferProposedOnElrond(context.Background())
		assert.False(t, wasTransfered)
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		wasCalled := false
		args.ElrondClient = &bridgeTests.ElrondClientStub{
			WasProposedTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
				assert.True(t, providedBatch == batch)
				wasCalled = true
				return true, nil
			},
		}

		executor, _ := NewTestBridgeExecutor(args)
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

		args := createMockExecutorArgs()
		executor, _ := NewTestBridgeExecutor(args)

		err := executor.ProposeTransferOnElrond(context.Background())
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("propose transfer fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.ElrondClient = &bridgeTests.ElrondClientStub{
			ProposeTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (string, error) {
				assert.True(t, providedBatch == batch)

				return "", expectedErr
			},
		}
		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatch

		err := executor.ProposeTransferOnElrond(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		wasCalled := false
		args.ElrondClient = &bridgeTests.ElrondClientStub{
			ProposeTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (string, error) {
				assert.True(t, providedBatch == batch)
				wasCalled = true

				return "", nil
			},
		}
		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatch

		err := executor.ProposeTransferOnElrond(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestEthToElrondBridgeExecutor_WasActionSignedOnElrond(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	providedActionID := uint64(378276)
	wasCalled := false
	args.ElrondClient = &bridgeTests.ElrondClientStub{
		WasSignedCalled: func(ctx context.Context, actionID uint64) (bool, error) {
			assert.Equal(t, providedActionID, actionID)
			wasCalled = true
			return true, nil
		},
	}
	executor, _ := NewTestBridgeExecutor(args)
	executor.actionID = providedActionID

	wasSigned, err := executor.WasActionSignedOnElrond(context.Background())
	assert.True(t, wasSigned)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestEthToElrondBridgeExecutor_SignActionOnElrond(t *testing.T) {
	t.Parallel()

	t.Run("elrond client errors", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedActionID := uint64(378276)
		args.ElrondClient = &bridgeTests.ElrondClientStub{
			SignCalled: func(ctx context.Context, actionID uint64) (string, error) {
				assert.Equal(t, providedActionID, actionID)
				return "", expectedErr
			},
		}

		executor, _ := NewTestBridgeExecutor(args)
		executor.actionID = providedActionID

		err := executor.SignActionOnElrond(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedActionID := uint64(378276)
		wasCalled := false
		args.ElrondClient = &bridgeTests.ElrondClientStub{
			SignCalled: func(ctx context.Context, actionID uint64) (string, error) {
				assert.Equal(t, providedActionID, actionID)
				wasCalled = true
				return "", nil
			},
		}

		executor, _ := NewTestBridgeExecutor(args)
		executor.actionID = providedActionID

		err := executor.SignActionOnElrond(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestEthToElrondBridgeExecutor_IsQuorumReachedOnElrond(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	providedActionID := uint64(378276)
	wasCalled := false
	args.ElrondClient = &bridgeTests.ElrondClientStub{
		QuorumReachedCalled: func(ctx context.Context, actionID uint64) (bool, error) {
			assert.Equal(t, providedActionID, actionID)
			wasCalled = true
			return true, nil
		},
	}
	executor, _ := NewTestBridgeExecutor(args)
	executor.actionID = providedActionID

	isQuorumReached, err := executor.ProcessQuorumReachedOnElrond(context.Background())
	assert.True(t, isQuorumReached)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestEthToElrondBridgeExecutor_WasActionPerformedOnElrond(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	providedActionID := uint64(378276)
	wasCalled := false
	args.ElrondClient = &bridgeTests.ElrondClientStub{
		WasExecutedCalled: func(ctx context.Context, actionID uint64) (bool, error) {
			assert.Equal(t, providedActionID, actionID)
			wasCalled = true
			return true, nil
		},
	}
	executor, _ := NewTestBridgeExecutor(args)
	executor.actionID = providedActionID

	wasPerformed, err := executor.WasActionPerformedOnElrond(context.Background())
	assert.True(t, wasPerformed)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestEthToElrondBridgeExecutor_PerformActionOnElrond(t *testing.T) {
	t.Parallel()

	t.Run("nil batch", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewTestBridgeExecutor(args)

		err := executor.PerformActionOnElrond(context.Background())
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("elrond client errors", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedActionID := uint64(7383)
		args.ElrondClient = &bridgeTests.ElrondClientStub{
			PerformActionCalled: func(ctx context.Context, actionID uint64, batch *clients.TransferBatch) (string, error) {
				assert.Equal(t, providedActionID, actionID)
				assert.True(t, providedBatch == batch)
				return "", expectedErr
			},
		}
		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatch
		executor.actionID = providedActionID

		err := executor.PerformActionOnElrond(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		wasCalled := false
		providedActionID := uint64(7383)
		args.ElrondClient = &bridgeTests.ElrondClientStub{
			PerformActionCalled: func(ctx context.Context, actionID uint64, batch *clients.TransferBatch) (string, error) {
				assert.Equal(t, providedActionID, actionID)
				assert.True(t, providedBatch == batch)
				wasCalled = true
				return "", nil
			},
		}
		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatch
		executor.actionID = providedActionID

		err := executor.PerformActionOnElrond(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestEthToElrondBridgeExecutor_RetriesCountOnElrond(t *testing.T) {
	t.Parallel()

	expectedMaxRetries := uint64(3)
	args := createMockExecutorArgs()
	wasCalled := false
	args.ElrondClient = &bridgeTests.ElrondClientStub{
		GetMaxNumberOfRetriesOnQuorumReachedCalled: func() uint64 {
			wasCalled = true
			return expectedMaxRetries
		},
	}
	executor, _ := NewTestBridgeExecutor(args)
	for i := uint64(0); i < expectedMaxRetries; i++ {
		assert.False(t, executor.ProcessMaxRetriesOnElrond())
	}

	assert.Equal(t, expectedMaxRetries, executor.retriesOnElrond)
	assert.True(t, executor.ProcessMaxRetriesOnElrond())
	executor.ResetRetriesCountOnElrond()
	assert.Equal(t, uint64(0), executor.retriesOnElrond)
	assert.True(t, wasCalled)
}

func TestElrondToEthBridgeExecutor_GetAndStoreBatchFromElrond(t *testing.T) {
	t.Parallel()

	t.Run("GetBatchFromElrond fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.ElrondClient = &bridgeTests.ElrondClientStub{
			GetPendingCalled: func(ctx context.Context) (*clients.TransferBatch, error) {
				return nil, expectedErr
			},
		}

		executor, _ := NewTestBridgeExecutor(args)
		_, err := executor.GetBatchFromElrond(context.Background())
		assert.Equal(t, expectedErr, err)

		batch := executor.GetStoredBatch()
		assert.Nil(t, batch)
	})
	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.ElrondClient = &bridgeTests.ElrondClientStub{}

		executor, _ := NewTestBridgeExecutor(args)
		err := executor.StoreBatchFromElrond(nil)
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockExecutorArgs()
		args.ElrondClient = &bridgeTests.ElrondClientStub{
			GetPendingCalled: func(ctx context.Context) (*clients.TransferBatch, error) {
				wasCalled = true
				return providedBatch, nil
			},
		}

		executor, _ := NewTestBridgeExecutor(args)
		batch, err := executor.GetBatchFromElrond(context.Background())
		assert.True(t, wasCalled)
		assert.Equal(t, providedBatch, batch)
		assert.Nil(t, err)

		err = executor.StoreBatchFromElrond(batch)
		assert.Equal(t, providedBatch, executor.batch)
		assert.Nil(t, err)
	})
}

func TestElrondToEthBridgeExecutor_GetAndStoreActionIDForProposeSetStatusFromElrond(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewTestBridgeExecutor(args)

		actionId, err := executor.GetAndStoreActionIDForProposeSetStatusFromElrond(context.Background())
		assert.Equal(t, ErrNilBatch, err)
		assert.Equal(t, InvalidActionID, actionId)
	})
	t.Run("GetAndStoreActionIDForProposeSetStatusFromElrond fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.ElrondClient = &bridgeTests.ElrondClientStub{
			GetActionIDForSetStatusOnPendingTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (uint64, error) {
				return uint64(0), expectedErr
			},
		}

		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatch
		_, err := executor.GetAndStoreActionIDForProposeSetStatusFromElrond(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		providedActionId := uint64(1123)
		args := createMockExecutorArgs()
		args.ElrondClient = &bridgeTests.ElrondClientStub{
			GetActionIDForSetStatusOnPendingTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (uint64, error) {
				wasCalled = true
				return providedActionId, nil
			},
		}

		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatch
		actionId, err := executor.GetAndStoreActionIDForProposeSetStatusFromElrond(context.Background())
		assert.True(t, wasCalled)
		assert.Equal(t, providedActionId, actionId)
		assert.Nil(t, err)

		actionId = executor.GetStoredActionID()
		assert.Equal(t, providedActionId, actionId)
	})
}
func TestElrondToEthBridgeExecutor_WasSetStatusProposedOnElrond(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewTestBridgeExecutor(args)

		wasProposed, err := executor.WasSetStatusProposedOnElrond(context.Background())
		assert.Equal(t, ErrNilBatch, err)
		assert.False(t, wasProposed)
	})
	t.Run("WasSetStatusProposedOnElrond fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.ElrondClient = &bridgeTests.ElrondClientStub{
			WasProposedSetStatusCalled: func(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
				return false, expectedErr
			},
		}

		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatch
		_, err := executor.WasSetStatusProposedOnElrond(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockExecutorArgs()
		args.ElrondClient = &bridgeTests.ElrondClientStub{
			WasProposedSetStatusCalled: func(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
				assert.True(t, providedBatch == batch)
				wasCalled = true
				return true, nil
			},
		}

		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatch
		wasProposed, err := executor.WasSetStatusProposedOnElrond(context.Background())
		assert.True(t, wasCalled)
		assert.True(t, wasProposed)
		assert.Nil(t, err)
	})
}

func TestEthToElrondBridgeExecutor_ProposeSetStatusOnElrond(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewTestBridgeExecutor(args)

		err := executor.ProposeSetStatusOnElrond(context.Background())
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("ProposeSetStatusOnElrond fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.ElrondClient = &bridgeTests.ElrondClientStub{
			ProposeSetStatusCalled: func(ctx context.Context, batch *clients.TransferBatch) (string, error) {
				return "", expectedErr
			},
		}

		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatch
		err := executor.ProposeSetStatusOnElrond(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockExecutorArgs()
		args.ElrondClient = &bridgeTests.ElrondClientStub{
			ProposeSetStatusCalled: func(ctx context.Context, batch *clients.TransferBatch) (string, error) {
				assert.True(t, providedBatch == batch)
				wasCalled = true

				return "", nil
			},
		}

		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatch

		err := executor.ProposeSetStatusOnElrond(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestElrondToEthBridgeExecutor_MyTurnAsLeader(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	wasCalled := false
	args.TopologyProvider = &bridgeTests.TopologyProviderStub{
		MyTurnAsLeaderCalled: func() bool {
			wasCalled = true
			return true
		},
	}

	executor, _ := NewTestBridgeExecutor(args)
	assert.True(t, executor.MyTurnAsLeader())
	assert.True(t, wasCalled)
}

func TestElrondToEthBridgeExecutor_WasTransferPerformedOnEthereum(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewTestBridgeExecutor(args)

		_, err := executor.WasTransferPerformedOnEthereum(context.Background())
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("WasExecuted fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			WasExecutedCalled: func(ctx context.Context, batchID uint64) (bool, error) {
				return false, expectedErr
			},
		}

		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatch
		_, err := executor.WasTransferPerformedOnEthereum(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		providedBatchID := uint64(36727)
		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			WasExecutedCalled: func(ctx context.Context, batchID uint64) (bool, error) {
				assert.True(t, providedBatchID == batchID)
				wasCalled = true
				return true, nil
			},
		}

		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatch
		executor.batch.ID = providedBatchID

		_, err := executor.WasTransferPerformedOnEthereum(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestElrondToEthBridgeExecutor_SignTransferOnEthereum(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewTestBridgeExecutor(args)

		err := executor.SignTransferOnEthereum()
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("GenerateMessageHash fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GenerateMessageHashCalled: func(batch *clients.TransferBatch) (common.Hash, error) {
				return common.Hash{}, expectedErr
			},
		}

		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatch
		err := executor.SignTransferOnEthereum()
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalledGenerateMessageHashCalled := false
		wasCalledBroadcastSignatureForMessageHashCalled := false
		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GenerateMessageHashCalled: func(batch *clients.TransferBatch) (common.Hash, error) {
				wasCalledGenerateMessageHashCalled = true
				return common.Hash{}, nil
			},
			BroadcastSignatureForMessageHashCalled: func(msgHash common.Hash) {
				wasCalledBroadcastSignatureForMessageHashCalled = true
			},
		}

		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatch
		err := executor.SignTransferOnEthereum()
		assert.Nil(t, err)
		assert.True(t, wasCalledGenerateMessageHashCalled)
		assert.True(t, wasCalledBroadcastSignatureForMessageHashCalled)
	})
}

func TestElrondToEthBridgeExecutor_PerformTransferOnEthereum(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewTestBridgeExecutor(args)

		err := executor.PerformTransferOnEthereum(context.Background())
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("GetQuorumSize fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetQuorumSizeCalled: func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(0), expectedErr
			},
		}

		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatch
		err := executor.PerformTransferOnEthereum(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("ExecuteTransfer fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetQuorumSizeCalled: func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(0), nil
			},
			ExecuteTransferCalled: func(ctx context.Context, msgHash common.Hash, batch *clients.TransferBatch, quorum int) (string, error) {
				return "", expectedErr
			},
		}

		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatch
		err := executor.PerformTransferOnEthereum(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedHash := common.Hash{}
		providedQuorum := 12
		wasCalledGetQuorumSizeCalled := false
		wasCalledExecuteTransferCalled := false
		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetQuorumSizeCalled: func(ctx context.Context) (*big.Int, error) {
				wasCalledGetQuorumSizeCalled = true
				return big.NewInt(int64(providedQuorum)), nil
			},
			ExecuteTransferCalled: func(ctx context.Context, msgHash common.Hash, batch *clients.TransferBatch, quorum int) (string, error) {
				assert.True(t, providedHash == msgHash)
				assert.True(t, providedBatch == batch)
				assert.True(t, providedQuorum == quorum)

				wasCalledExecuteTransferCalled = true
				return "", nil
			},
		}

		executor, _ := NewTestBridgeExecutor(args)
		executor.msgHash = providedHash
		executor.batch = providedBatch
		err := executor.PerformTransferOnEthereum(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalledGetQuorumSizeCalled)
		assert.True(t, wasCalledExecuteTransferCalled)
	})
}

func TestElrondToEthBridgeExecutor_IsQuorumReachedOnEthereum(t *testing.T) {
	t.Parallel()

	t.Run("ProcessQuorumReachedOnEthereum fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			IsQuorumReachedCalled: func(ctx context.Context, msgHash common.Hash) (bool, error) {
				return false, expectedErr
			},
		}

		executor, _ := NewTestBridgeExecutor(args)

		_, err := executor.ProcessQuorumReachedOnEthereum(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		wasCalled := false
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			IsQuorumReachedCalled: func(ctx context.Context, msgHash common.Hash) (bool, error) {
				wasCalled = true
				return true, nil
			},
		}

		executor, _ := NewTestBridgeExecutor(args)

		isReached, err := executor.ProcessQuorumReachedOnEthereum(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
		assert.True(t, isReached)
	})
}

func TestElrondToEthBridgeExecutor_RetriesCountOnEthereum(t *testing.T) {
	t.Parallel()

	expectedMaxRetries := uint64(3)
	args := createMockExecutorArgs()
	wasCalled := false
	args.EthereumClient = &bridgeTests.EthereumClientStub{
		GetMaxNumberOfRetriesOnQuorumReachedCalled: func() uint64 {
			wasCalled = true
			return expectedMaxRetries
		},
	}
	executor, _ := NewTestBridgeExecutor(args)
	for i := uint64(0); i < expectedMaxRetries; i++ {
		assert.False(t, executor.ProcessMaxRetriesOnEthereum())
	}

	assert.Equal(t, expectedMaxRetries, executor.retriesOnEthereum)
	assert.True(t, executor.ProcessMaxRetriesOnEthereum())
	executor.ResetRetriesCountOnEthereum()
	assert.Equal(t, uint64(0), executor.retriesOnEthereum)
	assert.True(t, wasCalled)
}

func TestWaitForTransferConfirmation(t *testing.T) {
	t.Parallel()

	t.Run("normal expiration", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForTransferExecution = 2 * time.Second
		executor, _ := NewTestBridgeExecutor(args)

		start := time.Now()
		executor.WaitForTransferConfirmation(context.Background())
		elapsed := time.Since(start)

		assert.True(t, elapsed >= args.TimeForTransferExecution)
	})
	t.Run("context expiration", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForTransferExecution = 10 * time.Second
		executor, _ := NewTestBridgeExecutor(args)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		start := time.Now()
		executor.WaitForTransferConfirmation(ctx)
		elapsed := time.Since(start)

		assert.True(t, elapsed < args.TimeForTransferExecution)
	})

	t.Run("WasTransferPerformedOnEthereum always returns false/err", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForTransferExecution = 10 * time.Second
		executor, _ := NewTestBridgeExecutor(args)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		counter := 0
		executor.SetWasTransferPerformedOnEthereumHandle(func(ctx context.Context) (bool, error) {
			counter++
			return false, nil
		})
		executor.WaitForTransferConfirmation(ctx)

		assert.Equal(t, 10, counter)
	})

	t.Run("WasTransferPerformedOnEthereum always returns true only after 4 checks", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForTransferExecution = 10 * time.Second
		executor, _ := NewTestBridgeExecutor(args)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		start := time.Now()
		counter := 0
		executor.SetWasTransferPerformedOnEthereumHandle(func(ctx context.Context) (bool, error) {
			counter++
			if counter >= 5 {
				return true, nil
			}
			return false, nil
		})
		executor.WaitForTransferConfirmation(ctx)
		elapsed := time.Since(start)

		assert.True(t, elapsed < args.TimeForTransferExecution)
		assert.Equal(t, 5, counter)
	})
}

func TestGetBatchStatusesFromEthereum(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewTestBridgeExecutor(args)
		_, err := executor.GetBatchStatusesFromEthereum(context.Background())
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("GetTransactionsStatuses fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetTransactionsStatusesCalled: func(ctx context.Context, batchId uint64) ([]byte, error) {
				return nil, expectedErr
			},
		}

		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatch
		_, err := executor.GetBatchStatusesFromEthereum(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		providedStatuses := []byte{1, 2, 3}
		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetTransactionsStatusesCalled: func(ctx context.Context, batchId uint64) ([]byte, error) {
				wasCalled = true
				return providedStatuses, nil
			},
		}

		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatch
		statuses, err := executor.GetBatchStatusesFromEthereum(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
		assert.Equal(t, providedStatuses, statuses)
	})
}

func TestResolveNewDepositsStatuses(t *testing.T) {
	t.Parallel()

	providedBatchForResolve := &clients.TransferBatch{
		Deposits: []*clients.DepositTransfer{
			{
				DisplayableTo: "to1",
			},
			{
				DisplayableTo: "to2",
			},
		},
		Statuses: make([]byte, 2),
	}

	t.Run("less new deposits", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatchForResolve.Clone()

		executor.ResolveNewDepositsStatuses(uint64(0))
		assert.Equal(t, []byte{clients.Rejected, clients.Rejected}, executor.batch.Statuses)

		executor.batch = providedBatchForResolve.Clone()
		executor.batch.ResolveNewDeposits(1)
		assert.Equal(t, []byte{0, clients.Rejected}, executor.batch.Statuses)
	})
	t.Run("equal new deposits", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatchForResolve.Clone()

		executor.ResolveNewDepositsStatuses(uint64(2))
		assert.Equal(t, []byte{0, 0}, executor.batch.Statuses)
	})
	t.Run("more new deposits", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewTestBridgeExecutor(args)
		executor.batch = providedBatchForResolve.Clone()

		executor.ResolveNewDepositsStatuses(uint64(3))
		assert.Equal(t, []byte{0, 0, clients.Rejected}, executor.batch.Statuses)
	})
}

func TestEthToElrondBridgeExecutor_setExecutionMessageInStatusHandler(t *testing.T) {
	t.Parallel()

	expectedString := "DEBUG: message a = 1 b = ff c = str"

	wasCalled := false
	args := createMockExecutorArgs()
	args.StatusHandler = &testsCommon.StatusHandlerStub{
		SetStringMetricCalled: func(metric string, val string) {
			wasCalled = true

			assert.Equal(t, metric, core.MetricLastError)
			assert.Equal(t, expectedString, val)
		},
	}
	executor, _ := NewTestBridgeExecutor(args)
	executor.setExecutionMessageInStatusHandler(logger.LogDebug, "message", "a", 1, "b", []byte{255}, "c", "str")

	assert.True(t, wasCalled)
}

func TestSignaturesHolder_ClearStoredSignatures(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	wasCalled := false
	args.SignaturesHolder = &testsCommon.SignaturesHolderStub{
		ClearStoredSignaturesCalled: func() {
			wasCalled = true
		},
	}

	executor, _ := NewTestBridgeExecutor(args)
	executor.ClearStoredP2PSignaturesForEthereum()

	assert.True(t, wasCalled)
}
