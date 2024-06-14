package ethmultiversx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	bridgeTests "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedErr = errors.New("expected error")
var providedBatch = &clients.TransferBatch{}
var expectedMaxRetries = uint64(3)
var zero = big.NewInt(0)

func createMockExecutorArgs() ArgsBridgeExecutor {
	return ArgsBridgeExecutor{
		Log:                          logger.GetOrCreate("test"),
		MultiversXClient:             &bridgeTests.MultiversXClientStub{},
		EthereumClient:               &bridgeTests.EthereumClientStub{},
		TopologyProvider:             &bridgeTests.TopologyProviderStub{},
		StatusHandler:                testsCommon.NewStatusHandlerMock("test"),
		TimeForWaitOnEthereum:        time.Second,
		SignaturesHolder:             &testsCommon.SignaturesHolderStub{},
		BatchValidator:               &testsCommon.BatchValidatorStub{},
		MaxQuorumRetriesOnEthereum:   minRetries,
		MaxQuorumRetriesOnMultiversX: minRetries,
		MaxRestriesOnWasProposed:     minRetries,
	}
}

func TestNewBridgeExecutor(t *testing.T) {
	t.Parallel()

	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.Log = nil
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrNilLogger, err)
	})
	t.Run("nil multiversx client should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.MultiversXClient = nil
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrNilMultiversXClient, err)
	})
	t.Run("nil ethereum client should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.EthereumClient = nil
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrNilEthereumClient, err)
	})
	t.Run("nil topology provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TopologyProvider = nil
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrNilTopologyProvider, err)
	})
	t.Run("nil status handler", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.StatusHandler = nil
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrNilStatusHandler, err)
	})
	t.Run("invalid time", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForWaitOnEthereum = 0
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrInvalidDuration, err)
	})
	t.Run("nil signatures holder", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.SignaturesHolder = nil
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrNilSignaturesHolder, err)
	})
	t.Run("nil batch validator", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.BatchValidator = nil
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, ErrNilBatchValidator, err)
	})
	t.Run("invalid MaxQuorumRetriesOnEthereum value", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.MaxQuorumRetriesOnEthereum = 0
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for args.MaxQuorumRetriesOnEthereum"))
	})
	t.Run("invalid MaxQuorumRetriesOnMultiversX value", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.MaxQuorumRetriesOnMultiversX = 0
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for args.MaxQuorumRetriesOnMultiversX"))
	})
	t.Run("invalid MaxRestriesOnWasProposed value", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.MaxRestriesOnWasProposed = 0
		executor, err := NewBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for args.MaxRestriesOnWasProposed"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, err := NewBridgeExecutor(args)

		assert.False(t, check.IfNil(executor))
		assert.Nil(t, err)
	})
}

func TestEthToMultiversXBridgeExecutor_PrintInfo(t *testing.T) {
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
	executor, _ := NewBridgeExecutor(args)
	executor.PrintInfo(providedLogLevel, providedMessage, providedArgs...)

	assert.True(t, wasCalled)

	if shouldOutputToStatusHandler {
		assert.True(t, len(statusHandler.GetStringMetric(core.MetricLastError)) > 0)
	}
}

func TestEthToMultiversXBridgeExecutor_MyTurnAsLeader(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	wasCalled := false
	args.TopologyProvider = &bridgeTests.TopologyProviderStub{
		MyTurnAsLeaderCalled: func() bool {
			wasCalled = true
			return true
		},
	}

	executor, _ := NewBridgeExecutor(args)
	assert.True(t, executor.MyTurnAsLeader())
	assert.True(t, wasCalled)
}

func TestEthToMultiversXBridgeExecutor_GetAndStoreActionIDForProposeTransferOnMultiversX(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

		actionID, err := executor.GetAndStoreActionIDForProposeTransferOnMultiversX(context.Background())
		assert.Zero(t, actionID)
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("multiversx client errors", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.MultiversXClient = &bridgeTests.MultiversXClientStub{
			GetActionIDForProposeTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (uint64, error) {
				assert.True(t, providedBatch == batch)
				return 0, expectedErr
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch

		actionID, err := executor.GetAndStoreActionIDForProposeTransferOnMultiversX(context.Background())
		assert.Zero(t, actionID)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedActionID := uint64(48939)

		args.MultiversXClient = &bridgeTests.MultiversXClientStub{
			GetActionIDForProposeTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (uint64, error) {
				assert.True(t, providedBatch == batch)
				return providedActionID, nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch

		assert.NotEqual(t, providedActionID, executor.actionID)

		actionID, err := executor.GetAndStoreActionIDForProposeTransferOnMultiversX(context.Background())
		assert.Equal(t, providedActionID, actionID)
		assert.Nil(t, err)
		assert.Equal(t, providedActionID, executor.GetStoredActionID())
		assert.Equal(t, providedActionID, executor.actionID)
	})
}

func TestEthToMultiversXBridgeExecutor_GetAndStoreBatchFromEthereum(t *testing.T) {
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
		executor, _ := NewBridgeExecutor(args)
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
		executor, _ := NewBridgeExecutor(args)
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
		executor, _ := NewBridgeExecutor(args)
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
			GetBatchSCMetadataCalled: func(ctx context.Context, nonce uint64) ([]*contract.SCExecProxyERC20SCDeposit, error) {
				return make([]*contract.SCExecProxyERC20SCDeposit, 0), nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		err := executor.GetAndStoreBatchFromEthereum(context.Background(), providedNonce)

		assert.Nil(t, err)
		assert.True(t, expectedBatch == executor.GetStoredBatch()) // pointer testing
		assert.True(t, expectedBatch == executor.batch)
	})
	t.Run("should add deposits metadata for sc calls", func(t *testing.T) {
		args := createMockExecutorArgs()
		providedNonce := uint64(8346)
		depositNonce := uint64(100)
		depositData := "testData"
		expectedBatch := &clients.TransferBatch{
			ID: providedNonce,
			Deposits: []*clients.DepositTransfer{
				{
					Nonce: depositNonce,
				},
			},
		}
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetBatchCalled: func(ctx context.Context, nonce uint64) (*clients.TransferBatch, error) {
				assert.Equal(t, providedNonce, nonce)
				return expectedBatch, nil
			},
			GetBatchSCMetadataCalled: func(ctx context.Context, nonce uint64) ([]*contract.SCExecProxyERC20SCDeposit, error) {
				return []*contract.SCExecProxyERC20SCDeposit{{
					DepositNonce: depositNonce,
					CallData:     depositData,
				}}, nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		err := executor.GetAndStoreBatchFromEthereum(context.Background(), providedNonce)

		assert.Nil(t, err)
		assert.True(t, expectedBatch == executor.GetStoredBatch()) // pointer testing
		assert.Equal(t, depositData, string(executor.batch.Deposits[0].Data))
	})
	t.Run("should add deposits metadata for sc calls even if with no data", func(t *testing.T) {
		args := createMockExecutorArgs()
		providedNonce := uint64(8346)
		depositNonce := uint64(100)
		depositData := ""
		expectedBatch := &clients.TransferBatch{
			ID: providedNonce,
			Deposits: []*clients.DepositTransfer{
				{
					Nonce: depositNonce,
				},
			},
		}
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetBatchCalled: func(ctx context.Context, nonce uint64) (*clients.TransferBatch, error) {
				assert.Equal(t, providedNonce, nonce)
				return expectedBatch, nil
			},
			GetBatchSCMetadataCalled: func(ctx context.Context, nonce uint64) ([]*contract.SCExecProxyERC20SCDeposit, error) {
				return []*contract.SCExecProxyERC20SCDeposit{{
					DepositNonce: depositNonce,
					CallData:     depositData,
				}}, nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		err := executor.GetAndStoreBatchFromEthereum(context.Background(), providedNonce)

		assert.Nil(t, err)
		assert.True(t, expectedBatch == executor.GetStoredBatch()) // pointer testing
		assert.Equal(t, "", string(executor.batch.Deposits[0].Data))
	})
}

func TestEthToMultiversXBridgeExecutor_GetLastExecutedEthBatchIDFromMultiversX(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	providedBatchID := uint64(36727)
	args.MultiversXClient = &bridgeTests.MultiversXClientStub{
		GetLastExecutedEthBatchIDCalled: func(ctx context.Context) (uint64, error) {
			return providedBatchID, nil
		},
	}
	setIntCalled := false
	args.StatusHandler = &testsCommon.StatusHandlerStub{
		SetIntMetricCalled: func(metric string, value int) {
			assert.Equal(t, core.MetricNumBatches, metric)
			assert.Equal(t, int(providedBatchID), value)
			setIntCalled = true
		},
	}
	executor, _ := NewBridgeExecutor(args)

	batchID, err := executor.GetLastExecutedEthBatchIDFromMultiversX(context.Background())
	assert.Equal(t, providedBatchID, batchID)
	assert.Nil(t, err)
	assert.True(t, setIntCalled)
}

func TestEthToMultiversXBridgeExecutor_VerifyLastDepositNonceExecutedOnEthereumBatch(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

		err := executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("get last executed tx id errors", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.MultiversXClient = &bridgeTests.MultiversXClientStub{
			GetLastExecutedEthTxIDCalled: func(ctx context.Context) (uint64, error) {
				return 0, expectedErr
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = &clients.TransferBatch{}

		err := executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
		assert.Equal(t, expectedErr, err)
	})

	args := createMockExecutorArgs()
	txId := uint64(6657)
	args.MultiversXClient = &bridgeTests.MultiversXClientStub{
		GetLastExecutedEthTxIDCalled: func(ctx context.Context) (uint64, error) {
			return txId, nil
		},
	}

	t.Run("first deposit nonce equals last tx nonce should error", func(t *testing.T) {
		t.Parallel()

		executor, _ := NewBridgeExecutor(args)
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

		executor, _ := NewBridgeExecutor(args)
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

		executor, _ := NewBridgeExecutor(args)
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

		executor, _ := NewBridgeExecutor(args)
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

func TestEthToMultiversXBridgeExecutor_WasTransferProposedOnMultiversX(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

		wasTransfered, err := executor.WasTransferProposedOnMultiversX(context.Background())
		assert.False(t, wasTransfered)
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		wasCalled := false
		args.MultiversXClient = &bridgeTests.MultiversXClientStub{
			WasProposedTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
				assert.True(t, providedBatch == batch)
				wasCalled = true
				return true, nil
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch

		wasProposed, err := executor.WasTransferProposedOnMultiversX(context.Background())
		assert.True(t, wasProposed)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestEthToMultiversXBridgeExecutor_ProposeTransferOnMultiversX(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

		err := executor.ProposeTransferOnMultiversX(context.Background())
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("propose transfer fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.MultiversXClient = &bridgeTests.MultiversXClientStub{
			ProposeTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (string, error) {
				assert.True(t, providedBatch == batch)

				return "", expectedErr
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch

		err := executor.ProposeTransferOnMultiversX(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		wasCalled := false
		args.MultiversXClient = &bridgeTests.MultiversXClientStub{
			ProposeTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (string, error) {
				assert.True(t, providedBatch == batch)
				wasCalled = true

				return "", nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch

		err := executor.ProposeTransferOnMultiversX(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestEthToMultiversXBridgeExecutor_WasActionSignedOnMultiversX(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	providedActionID := uint64(378276)
	wasCalled := false
	args.MultiversXClient = &bridgeTests.MultiversXClientStub{
		WasSignedCalled: func(ctx context.Context, actionID uint64) (bool, error) {
			assert.Equal(t, providedActionID, actionID)
			wasCalled = true
			return true, nil
		},
	}
	executor, _ := NewBridgeExecutor(args)
	executor.actionID = providedActionID

	wasSigned, err := executor.WasActionSignedOnMultiversX(context.Background())
	assert.True(t, wasSigned)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestEthToMultiversXBridgeExecutor_SignActionOnMultiversX(t *testing.T) {
	t.Parallel()

	t.Run("multiversx client errors", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedActionID := uint64(378276)
		args.MultiversXClient = &bridgeTests.MultiversXClientStub{
			SignCalled: func(ctx context.Context, actionID uint64) (string, error) {
				assert.Equal(t, providedActionID, actionID)
				return "", expectedErr
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.actionID = providedActionID

		err := executor.SignActionOnMultiversX(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedActionID := uint64(378276)
		wasCalled := false
		args.MultiversXClient = &bridgeTests.MultiversXClientStub{
			SignCalled: func(ctx context.Context, actionID uint64) (string, error) {
				assert.Equal(t, providedActionID, actionID)
				wasCalled = true
				return "", nil
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.actionID = providedActionID

		err := executor.SignActionOnMultiversX(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestEthToMultiversXBridgeExecutor_IsQuorumReachedOnMultiversX(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	providedActionID := uint64(378276)
	wasCalled := false
	args.MultiversXClient = &bridgeTests.MultiversXClientStub{
		QuorumReachedCalled: func(ctx context.Context, actionID uint64) (bool, error) {
			assert.Equal(t, providedActionID, actionID)
			wasCalled = true
			return true, nil
		},
	}
	executor, _ := NewBridgeExecutor(args)
	executor.actionID = providedActionID

	isQuorumReached, err := executor.ProcessQuorumReachedOnMultiversX(context.Background())
	assert.True(t, isQuorumReached)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestEthToMultiversXBridgeExecutor_WasActionPerformedOnMultiversX(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	providedActionID := uint64(378276)
	wasCalled := false
	args.MultiversXClient = &bridgeTests.MultiversXClientStub{
		WasExecutedCalled: func(ctx context.Context, actionID uint64) (bool, error) {
			assert.Equal(t, providedActionID, actionID)
			wasCalled = true
			return true, nil
		},
	}
	executor, _ := NewBridgeExecutor(args)
	executor.actionID = providedActionID

	wasPerformed, err := executor.WasActionPerformedOnMultiversX(context.Background())
	assert.True(t, wasPerformed)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestEthToMultiversXBridgeExecutor_PerformActionOnMultiversX(t *testing.T) {
	t.Parallel()

	t.Run("nil batch", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

		err := executor.PerformActionOnMultiversX(context.Background())
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("multiversx client errors", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		providedActionID := uint64(7383)
		args.MultiversXClient = &bridgeTests.MultiversXClientStub{
			PerformActionCalled: func(ctx context.Context, actionID uint64, batch *clients.TransferBatch) (string, error) {
				assert.Equal(t, providedActionID, actionID)
				assert.True(t, providedBatch == batch)
				return "", expectedErr
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		executor.actionID = providedActionID

		err := executor.PerformActionOnMultiversX(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		wasCalled := false
		providedActionID := uint64(7383)
		args.MultiversXClient = &bridgeTests.MultiversXClientStub{
			PerformActionCalled: func(ctx context.Context, actionID uint64, batch *clients.TransferBatch) (string, error) {
				assert.Equal(t, providedActionID, actionID)
				assert.True(t, providedBatch == batch)
				wasCalled = true
				return "", nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		executor.actionID = providedActionID

		err := executor.PerformActionOnMultiversX(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestEthToMultiversXBridgeExecutor_RetriesCountOnMultiversX(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	args.MaxQuorumRetriesOnMultiversX = expectedMaxRetries
	executor, _ := NewBridgeExecutor(args)
	for i := uint64(0); i < expectedMaxRetries; i++ {
		assert.False(t, executor.ProcessMaxQuorumRetriesOnMultiversX())
	}

	assert.Equal(t, expectedMaxRetries, executor.quorumRetriesOnMultiversX)
	assert.True(t, executor.ProcessMaxQuorumRetriesOnMultiversX())
	executor.ResetRetriesCountOnMultiversX()
	assert.Equal(t, uint64(0), executor.quorumRetriesOnMultiversX)
}

func TestEthToMultiversXBridgeExecutor_RetriesCountOnWasTransferProposedOnMultiversX(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	args.MaxRestriesOnWasProposed = expectedMaxRetries
	executor, _ := NewBridgeExecutor(args)
	for i := uint64(0); i < expectedMaxRetries; i++ {
		assert.False(t, executor.ProcessMaxRetriesOnWasTransferProposedOnMultiversX())
	}

	assert.Equal(t, expectedMaxRetries, executor.retriesOnWasProposed)
	assert.True(t, executor.ProcessMaxRetriesOnWasTransferProposedOnMultiversX())
	executor.ResetRetriesOnWasTransferProposedOnMultiversX()
	assert.Equal(t, uint64(0), executor.retriesOnWasProposed)
}

func TestMultiversXToEthBridgeExecutor_GetAndStoreBatchFromMultiversX(t *testing.T) {
	t.Parallel()

	t.Run("GetBatchFromMultiversX fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.MultiversXClient = &bridgeTests.MultiversXClientStub{
			GetPendingCalled: func(ctx context.Context) (*clients.TransferBatch, error) {
				return nil, expectedErr
			},
		}

		executor, _ := NewBridgeExecutor(args)
		_, err := executor.GetBatchFromMultiversX(context.Background())
		assert.Equal(t, expectedErr, err)

		batch := executor.GetStoredBatch()
		assert.Nil(t, batch)
	})
	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.MultiversXClient = &bridgeTests.MultiversXClientStub{}

		executor, _ := NewBridgeExecutor(args)
		err := executor.StoreBatchFromMultiversX(nil)
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockExecutorArgs()
		args.MultiversXClient = &bridgeTests.MultiversXClientStub{
			GetPendingCalled: func(ctx context.Context) (*clients.TransferBatch, error) {
				wasCalled = true
				return providedBatch, nil
			},
		}

		executor, _ := NewBridgeExecutor(args)
		batch, err := executor.GetBatchFromMultiversX(context.Background())
		assert.True(t, wasCalled)
		assert.Equal(t, providedBatch, batch)
		assert.Nil(t, err)

		err = executor.StoreBatchFromMultiversX(batch)
		assert.Equal(t, providedBatch, executor.batch)
		assert.Nil(t, err)
	})
}

func TestMultiversXToEthBridgeExecutor_GetAndStoreActionIDForProposeSetStatusFromMultiversX(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

		actionId, err := executor.GetAndStoreActionIDForProposeSetStatusFromMultiversX(context.Background())
		assert.Equal(t, ErrNilBatch, err)
		assert.Equal(t, InvalidActionID, actionId)
	})
	t.Run("GetAndStoreActionIDForProposeSetStatusFromMultiversX fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.MultiversXClient = &bridgeTests.MultiversXClientStub{
			GetActionIDForSetStatusOnPendingTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (uint64, error) {
				return uint64(0), expectedErr
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		_, err := executor.GetAndStoreActionIDForProposeSetStatusFromMultiversX(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		providedActionId := uint64(1123)
		args := createMockExecutorArgs()
		args.MultiversXClient = &bridgeTests.MultiversXClientStub{
			GetActionIDForSetStatusOnPendingTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (uint64, error) {
				wasCalled = true
				return providedActionId, nil
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		actionId, err := executor.GetAndStoreActionIDForProposeSetStatusFromMultiversX(context.Background())
		assert.True(t, wasCalled)
		assert.Equal(t, providedActionId, actionId)
		assert.Nil(t, err)

		actionId = executor.GetStoredActionID()
		assert.Equal(t, providedActionId, actionId)
	})
}
func TestMultiversXToEthBridgeExecutor_WasSetStatusProposedOnMultiversX(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

		wasProposed, err := executor.WasSetStatusProposedOnMultiversX(context.Background())
		assert.Equal(t, ErrNilBatch, err)
		assert.False(t, wasProposed)
	})
	t.Run("WasSetStatusProposedOnMultiversX fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.MultiversXClient = &bridgeTests.MultiversXClientStub{
			WasProposedSetStatusCalled: func(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
				return false, expectedErr
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		_, err := executor.WasSetStatusProposedOnMultiversX(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockExecutorArgs()
		args.MultiversXClient = &bridgeTests.MultiversXClientStub{
			WasProposedSetStatusCalled: func(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
				assert.True(t, providedBatch == batch)
				wasCalled = true
				return true, nil
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		wasProposed, err := executor.WasSetStatusProposedOnMultiversX(context.Background())
		assert.True(t, wasCalled)
		assert.True(t, wasProposed)
		assert.Nil(t, err)
	})
}

func TestEthToMultiversXBridgeExecutor_ProposeSetStatusOnMultiversX(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

		err := executor.ProposeSetStatusOnMultiversX(context.Background())
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("ProposeSetStatusOnMultiversX fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.MultiversXClient = &bridgeTests.MultiversXClientStub{
			ProposeSetStatusCalled: func(ctx context.Context, batch *clients.TransferBatch) (string, error) {
				return "", expectedErr
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		err := executor.ProposeSetStatusOnMultiversX(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		args := createMockExecutorArgs()
		args.MultiversXClient = &bridgeTests.MultiversXClientStub{
			ProposeSetStatusCalled: func(ctx context.Context, batch *clients.TransferBatch) (string, error) {
				assert.True(t, providedBatch == batch)
				wasCalled = true

				return "", nil
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch

		err := executor.ProposeSetStatusOnMultiversX(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestMultiversXToEthBridgeExecutor_MyTurnAsLeader(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	wasCalled := false
	args.TopologyProvider = &bridgeTests.TopologyProviderStub{
		MyTurnAsLeaderCalled: func() bool {
			wasCalled = true
			return true
		},
	}

	executor, _ := NewBridgeExecutor(args)
	assert.True(t, executor.MyTurnAsLeader())
	assert.True(t, wasCalled)
}

func TestMultiversXToEthBridgeExecutor_WasTransferPerformedOnEthereum(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

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

		executor, _ := NewBridgeExecutor(args)
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

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		executor.batch.ID = providedBatchID

		_, err := executor.WasTransferPerformedOnEthereum(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestMultiversXToEthBridgeExecutor_SignTransferOnEthereum(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

		err := executor.SignTransferOnEthereum()
		assert.Equal(t, ErrNilBatch, err)
	})
	t.Run("GenerateMessageHash fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GenerateMessageHashCalled: func(batch *batchProcessor.ArgListsBatch, batchID uint64) (common.Hash, error) {
				return common.Hash{}, expectedErr
			},
		}

		executor, _ := NewBridgeExecutor(args)
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
			GenerateMessageHashCalled: func(batch *batchProcessor.ArgListsBatch, batchID uint64) (common.Hash, error) {
				wasCalledGenerateMessageHashCalled = true
				return common.Hash{}, nil
			},
			BroadcastSignatureForMessageHashCalled: func(msgHash common.Hash) {
				wasCalledBroadcastSignatureForMessageHashCalled = true
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		err := executor.SignTransferOnEthereum()
		assert.Nil(t, err)
		assert.True(t, wasCalledGenerateMessageHashCalled)
		assert.True(t, wasCalledBroadcastSignatureForMessageHashCalled)
	})
}

func TestMultiversXToEthBridgeExecutor_PerformTransferOnEthereum(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)

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

		executor, _ := NewBridgeExecutor(args)
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
			ExecuteTransferCalled: func(ctx context.Context, msgHash common.Hash, batch *batchProcessor.ArgListsBatch, batchId uint64, quorum int) (string, error) {
				return "", expectedErr
			},
		}

		executor, _ := NewBridgeExecutor(args)
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
			ExecuteTransferCalled: func(ctx context.Context, msgHash common.Hash, batch *batchProcessor.ArgListsBatch, batchId uint64, quorum int) (string, error) {
				assert.True(t, providedHash == msgHash)
				assert.True(t, providedBatch.ID == batchId)
				for i := 0; i < len(providedBatch.Deposits); i++ {
					assert.Equal(t, providedBatch.Deposits[i].Amount, batch.Amounts[i])
					assert.Equal(t, providedBatch.Deposits[i].Nonce, batch.Nonces[i].Uint64())
					assert.Equal(t, providedBatch.Deposits[i].ToBytes, batch.Recipients[i].Bytes())
					assert.Equal(t, providedBatch.Deposits[i].SourceTokenBytes, batch.EthTokens[i].Bytes())
					assert.Equal(t, providedBatch.Deposits[i].DestinationTokenBytes, batch.MvxTokenBytes[i])
				}
				assert.True(t, providedQuorum == quorum)

				wasCalledExecuteTransferCalled = true
				return "", nil
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.msgHash = providedHash
		executor.batch = providedBatch
		err := executor.PerformTransferOnEthereum(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalledGetQuorumSizeCalled)
		assert.True(t, wasCalledExecuteTransferCalled)
	})
}

func TestMultiversXToEthBridgeExecutor_IsQuorumReachedOnEthereum(t *testing.T) {
	t.Parallel()

	t.Run("ProcessQuorumReachedOnEthereum fails", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			IsQuorumReachedCalled: func(ctx context.Context, msgHash common.Hash) (bool, error) {
				return false, expectedErr
			},
		}

		executor, _ := NewBridgeExecutor(args)

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

		executor, _ := NewBridgeExecutor(args)

		isReached, err := executor.ProcessQuorumReachedOnEthereum(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
		assert.True(t, isReached)
	})
}

func TestMultiversXToEthBridgeExecutor_RetriesCountOnEthereum(t *testing.T) {
	t.Parallel()

	args := createMockExecutorArgs()
	args.MaxQuorumRetriesOnEthereum = expectedMaxRetries
	executor, _ := NewBridgeExecutor(args)
	for i := uint64(0); i < expectedMaxRetries; i++ {
		assert.False(t, executor.ProcessMaxQuorumRetriesOnEthereum())
	}

	assert.Equal(t, expectedMaxRetries, executor.quorumRetriesOnEthereum)
	assert.True(t, executor.ProcessMaxQuorumRetriesOnEthereum())
	executor.ResetRetriesCountOnEthereum()
	assert.Equal(t, uint64(0), executor.quorumRetriesOnEthereum)
}

func TestWaitForTransferConfirmation(t *testing.T) {
	t.Parallel()

	t.Run("normal expiration", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForWaitOnEthereum = 2 * time.Second
		executor, _ := NewBridgeExecutor(args)

		start := time.Now()
		executor.WaitForTransferConfirmation(context.Background())
		elapsed := time.Since(start)

		assert.True(t, elapsed >= args.TimeForWaitOnEthereum)
	})
	t.Run("context expiration", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForWaitOnEthereum = 10 * time.Second
		executor, _ := NewBridgeExecutor(args)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		start := time.Now()
		executor.WaitForTransferConfirmation(ctx)
		elapsed := time.Since(start)

		assert.True(t, elapsed < args.TimeForWaitOnEthereum)
	})

	t.Run("WasTransferPerformedOnEthereum always returns false/err", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForWaitOnEthereum = 10 * time.Second
		counter := 0
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			WasExecutedCalled: func(ctx context.Context, batchID uint64) (bool, error) {
				counter++
				return false, nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = &clients.TransferBatch{}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		executor.WaitForTransferConfirmation(ctx)

		assert.Equal(t, 10, counter)
	})

	t.Run("WasTransferPerformedOnEthereum always returns true only after 4 checks", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForWaitOnEthereum = 10 * time.Second
		counter := 0
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			WasExecutedCalled: func(ctx context.Context, batchID uint64) (bool, error) {
				counter++
				if counter >= 5 {
					return true, nil
				}
				return false, nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = &clients.TransferBatch{}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		start := time.Now()

		executor.WaitForTransferConfirmation(ctx)
		elapsed := time.Since(start)

		assert.True(t, elapsed < args.TimeForWaitOnEthereum)
		assert.Equal(t, 5, counter)
	})
}

func TestGetBatchStatusesFromEthereum(t *testing.T) {
	t.Parallel()

	t.Run("nil batch should error", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)
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

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		_, err := executor.GetBatchStatusesFromEthereum(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		providedStatuses := []byte{clients.Executed, clients.Rejected}
		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetTransactionsStatusesCalled: func(ctx context.Context, batchId uint64) ([]byte, error) {
				wasCalled = true
				return providedStatuses, nil
			},
		}

		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatch
		statuses, err := executor.GetBatchStatusesFromEthereum(context.Background())
		assert.Nil(t, err)
		assert.True(t, wasCalled)
		assert.Equal(t, providedStatuses, statuses)
	})
}

func TestWaitAndReturnFinalBatchStatuses(t *testing.T) {
	t.Parallel()

	t.Run("normal expiration", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForWaitOnEthereum = 2 * time.Second
		executor, _ := NewBridgeExecutor(args)

		start := time.Now()
		statuses := executor.WaitAndReturnFinalBatchStatuses(context.Background())
		elapsed := time.Since(start)

		assert.True(t, elapsed >= args.TimeForWaitOnEthereum)
		assert.Nil(t, statuses)
	})
	t.Run("context expiration", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForWaitOnEthereum = 10 * time.Second
		executor, _ := NewBridgeExecutor(args)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		start := time.Now()
		statuses := executor.WaitAndReturnFinalBatchStatuses(ctx)
		elapsed := time.Since(start)

		assert.True(t, elapsed < args.TimeForWaitOnEthereum)
		assert.Nil(t, statuses)
	})
	t.Run("GetBatchStatusesFromEthereum always returns err", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		args.TimeForWaitOnEthereum = 10 * time.Second
		counter := 0
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetTransactionsStatusesCalled: func(ctx context.Context, batchId uint64) ([]byte, error) {
				counter++
				return nil, expectedErr
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = &clients.TransferBatch{}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		statuses := executor.WaitAndReturnFinalBatchStatuses(ctx)

		assert.Equal(t, 10, counter)
		assert.Nil(t, statuses)
	})
	t.Run("GetBatchStatusesFromEthereum always returns success+statuses only after 4 checks", func(t *testing.T) {
		t.Parallel()

		providedStatuses := []byte{clients.Executed, clients.Rejected}
		args := createMockExecutorArgs()
		args.TimeForWaitOnEthereum = 10 * time.Second
		counter := 0
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetTransactionsStatusesCalled: func(ctx context.Context, batchId uint64) ([]byte, error) {
				counter++
				if counter >= 5 {
					return providedStatuses, nil
				}
				return nil, expectedErr
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = &clients.TransferBatch{}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		start := time.Now()

		statuses := executor.WaitAndReturnFinalBatchStatuses(ctx)
		elapsed := time.Since(start)

		assert.True(t, elapsed < args.TimeForWaitOnEthereum)
		assert.Equal(t, 5, counter)
		assert.Equal(t, providedStatuses, statuses)
	})
	t.Run("GetBatchStatusesFromEthereum always returns success+statuses only after 4 checks, otherwise empty slice", func(t *testing.T) {
		t.Parallel()

		providedStatuses := []byte{clients.Executed, clients.Rejected}
		args := createMockExecutorArgs()
		args.TimeForWaitOnEthereum = 10 * time.Second
		counter := 0
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			GetTransactionsStatusesCalled: func(ctx context.Context, batchId uint64) ([]byte, error) {
				counter++
				if counter >= 5 {
					return providedStatuses, nil
				}
				return make([]byte, 0), nil
			},
		}
		executor, _ := NewBridgeExecutor(args)
		executor.batch = &clients.TransferBatch{}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		start := time.Now()

		statuses := executor.WaitAndReturnFinalBatchStatuses(ctx)
		elapsed := time.Since(start)

		assert.True(t, elapsed < args.TimeForWaitOnEthereum)
		assert.Equal(t, 5, counter)
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
		executor, _ := NewBridgeExecutor(args)
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
		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatchForResolve.Clone()

		executor.ResolveNewDepositsStatuses(uint64(2))
		assert.Equal(t, []byte{0, 0}, executor.batch.Statuses)
	})
	t.Run("more new deposits", func(t *testing.T) {
		t.Parallel()

		args := createMockExecutorArgs()
		executor, _ := NewBridgeExecutor(args)
		executor.batch = providedBatchForResolve.Clone()

		executor.ResolveNewDepositsStatuses(uint64(3))
		assert.Equal(t, []byte{0, 0, clients.Rejected}, executor.batch.Statuses)
	})
}

func TestEthToMultiversXBridgeExecutor_setExecutionMessageInStatusHandler(t *testing.T) {
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
	executor, _ := NewBridgeExecutor(args)
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

	executor, _ := NewBridgeExecutor(args)
	executor.ClearStoredP2PSignaturesForEthereum()

	assert.True(t, wasCalled)
}

func TestBridgeExecutor_CheckMultiversXClientAvailability(t *testing.T) {
	t.Parallel()

	checkAvailabilityCalled := false
	args := createMockExecutorArgs()
	args.MultiversXClient = &bridgeTests.MultiversXClientStub{
		CheckClientAvailabilityCalled: func(ctx context.Context) error {
			checkAvailabilityCalled = true
			return nil
		},
	}
	executor, _ := NewBridgeExecutor(args)
	err := executor.CheckMultiversXClientAvailability(context.Background())

	assert.Nil(t, err)
	assert.True(t, checkAvailabilityCalled)
}

func TestBridgeExecutor_CheckEthereumClientAvailability(t *testing.T) {
	t.Parallel()

	checkAvailabilityCalled := false
	args := createMockExecutorArgs()
	args.EthereumClient = &bridgeTests.EthereumClientStub{
		CheckClientAvailabilityCalled: func(ctx context.Context) error {
			checkAvailabilityCalled = true
			return nil
		},
	}
	executor, _ := NewBridgeExecutor(args)
	err := executor.CheckEthereumClientAvailability(context.Background())

	assert.Nil(t, err)
	assert.True(t, checkAvailabilityCalled)
}

func TestBridgeExecutor_ValidateBatch(t *testing.T) {
	t.Parallel()

	validateBatchCalled := false
	args := createMockExecutorArgs()
	validationBatch := &clients.TransferBatch{
		ID: 45,
	}
	args.BatchValidator = &testsCommon.BatchValidatorStub{
		ValidateBatchCalled: func(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
			assert.True(t, validationBatch == batch) // pointer testing
			validateBatchCalled = true

			return true, nil
		},
	}
	executor, _ := NewBridgeExecutor(args)
	result, err := executor.ValidateBatch(context.Background(), validationBatch)

	assert.Nil(t, err)
	assert.True(t, result)
	assert.True(t, validateBatchCalled)
}

func TestBridgeExecutor_CheckAvailableTokens(t *testing.T) {
	t.Parallel()

	direction := batchProcessor.FromMultiversX
	var ethIsNative, ethIsMintBurn, mvxIsNative, mvxIsMintBurn bool
	t.Run("from MultiversX", func(t *testing.T) {
		ethIsNative = false
		t.Run("ethIsNative = false", func(t *testing.T) {
			ethIsMintBurn = false
			t.Run("ethIsMintBurn = false", func(t *testing.T) {
				// mvxIsNative & mvxIsMintBurn does not matter
				testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isMintBurnOnEthereum"})(t)
			})

			ethIsMintBurn = true
			t.Run("ethIsMintBurn = true", func(t *testing.T) {
				mvxIsNative = false
				t.Run("mvxIsNative = false", func(t *testing.T) {
					mvxIsMintBurn = false
					t.Run("mvxIsMintBurn = false", func(t *testing.T) {
						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnMultiversX", "isMintBurnOnMultiversX"})(t)
					})

					mvxIsMintBurn = true
					t.Run("mvxIsMintBurn = true", func(t *testing.T) {
						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isNativeOnMultiversX"})(t)
					})
				})

				mvxIsNative = true
				t.Run("mvxIsNative = true", func(t *testing.T) {
					mvxIsMintBurn = false
					t.Run("mvxIsMintBurn = false", func(t *testing.T) {
						ethMintBalance := big.NewInt(0)
						ethBurnBalance := big.NewInt(0)
						mvxTotalBalances := big.NewInt(99)

						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, mvxTotalBalances, zero, zero, direction, ErrBalanceMismatch, []string{})(t)
						mvxTotalBalances = big.NewInt(100)
						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, mvxTotalBalances, zero, zero, direction, nil, []string{})(t)
					})
					mvxIsMintBurn = true
					t.Run("mvxIsMintBurn = true", func(t *testing.T) {
						ethMintBalance := big.NewInt(0)
						ethBurnBalance := big.NewInt(0)
						mvxMintBalance := big.NewInt(0)
						mvxBurnBalance := big.NewInt(99)

						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, ErrBalanceMismatch, []string{})(t)
						mvxBurnBalance = big.NewInt(100)
						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, nil, []string{})(t)
					})
				})
			})
		})

		ethIsNative = true
		t.Run("ethIsNative = true", func(t *testing.T) {
			ethIsMintBurn = false
			t.Run("ethIsMintBurn = false", func(t *testing.T) {
				mvxIsNative = false
				t.Run("mvxIsNative = false", func(t *testing.T) {
					mvxIsMintBurn = false
					t.Run("mvxIsMintBurn = false", func(t *testing.T) {
						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnMultiversX", "isMintBurnOnMultiversX"})(t)
					})
					mvxIsMintBurn = true
					t.Run("mvxIsMintBurn = true", func(t *testing.T) {
						ethTotalBalances := big.NewInt(100)
						mvxMintBalance := big.NewInt(100)
						mvxBurnBalance := big.NewInt(99)

						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, ethTotalBalances, zero, zero, zero, mvxMintBalance, mvxBurnBalance, direction, ErrBalanceMismatch, []string{})(t)
						mvxBurnBalance = big.NewInt(100)
						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, ethTotalBalances, zero, zero, zero, mvxMintBalance, mvxBurnBalance, direction, nil, []string{})(t)
					})
				})
				mvxIsNative = true
				// mvxIsMintBurn does not matter
				testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isNativeOnMultiversX"})(t)
			})

			ethIsMintBurn = true
			t.Run("ethIsMintBurn = true", func(t *testing.T) {
				mvxIsNative = false
				t.Run("mvxIsNative = false", func(t *testing.T) {
					mvxIsMintBurn = false
					t.Run("mvxIsMintBurn = false", func(t *testing.T) {
						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnMultiversX", "isMintBurnOnMultiversX"})(t)
					})

					mvxIsMintBurn = true
					t.Run("mvxIsMintBurn = true", func(t *testing.T) {
						ethMintBalance := big.NewInt(0)
						ethBurnBalance := big.NewInt(100)
						mvxMintBalance := big.NewInt(100)
						mvxBurnBalance := big.NewInt(99)

						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, ErrBalanceMismatch, []string{})(t)
						mvxBurnBalance = big.NewInt(100)
						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, nil, []string{})(t)
					})
				})
				mvxIsNative = true
				// mvxIsMintBurn does not matter
				testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isNativeOnMultiversX"})(t)
			})
		})
	})

	direction = batchProcessor.ToMultiversX
	t.Run("to MultiversX", func(t *testing.T) {
		ethIsNative = false
		t.Run("ethIsNative = false", func(t *testing.T) {
			ethIsMintBurn = false
			t.Run("ethIsMintBurn = false", func(t *testing.T) {
				// mvxIsNative && mvxIsMintBurn does not matter
				testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isMintBurnOnEthereum"})(t)
			})

			ethIsMintBurn = true
			t.Run("ethIsMintBurn = true", func(t *testing.T) {
				mvxIsNative = false
				t.Run("mvxIsNative = false", func(t *testing.T) {
					// mvxIsMintBurn does not matter
					testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isNativeOnMultiversX"})(t)
				})

				mvxIsNative = true
				t.Run("mvxIsNative = true", func(t *testing.T) {
					mvxIsMintBurn = false
					t.Run("mvxIsMintBurn = false", func(t *testing.T) {
						ethMintBalance := big.NewInt(0)
						ethBurnBalance := big.NewInt(0)
						mvxTotalBalances := big.NewInt(99)

						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, mvxTotalBalances, zero, zero, direction, ErrBalanceMismatch, []string{})(t)
						mvxTotalBalances = big.NewInt(100)
						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, mvxTotalBalances, zero, zero, direction, nil, []string{})(t)
					})
					mvxIsMintBurn = true
					t.Run("mvxIsMintBurn = true", func(t *testing.T) {
						ethMintBalance := big.NewInt(0)
						ethBurnBalance := big.NewInt(0)
						mvxBurnBalance := big.NewInt(99)
						mvxMintBalance := big.NewInt(0)

						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, ErrBalanceMismatch, []string{})(t)
						mvxBurnBalance = big.NewInt(100)
						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, nil, []string{})(t)
					})
				})
			})
		})

		ethIsNative = true
		t.Run("ethIsNative = true", func(t *testing.T) {
			ethIsMintBurn = false
			t.Run("ethIsMintBurn = false", func(t *testing.T) {
				mvxIsNative = false
				t.Run("mvxIsNative = false", func(t *testing.T) {
					mvxIsMintBurn = false
					t.Run("mvxIsMintBurn = false", func(t *testing.T) {
						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnMultiversX", "isMintBurnOnMultiversX"})(t)
					})
					mvxIsMintBurn = true
					t.Run("mvxIsMintBurn = true", func(t *testing.T) {
						ethTotalBalances := big.NewInt(99)
						mvxBurnBalance := big.NewInt(0)
						mvxMintBalance := big.NewInt(0)

						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, ethTotalBalances, zero, zero, zero, mvxMintBalance, mvxBurnBalance, direction, ErrBalanceMismatch, []string{})(t)
						ethTotalBalances = big.NewInt(100)
						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, ethTotalBalances, zero, zero, zero, mvxMintBalance, mvxBurnBalance, direction, nil, []string{})(t)
					})
				})
				mvxIsNative = true
				// mvxIsMintBurn does not matter
				testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isNativeOnMultiversX"})(t)
			})

			ethIsMintBurn = true
			t.Run("ethIsMintBurn = true", func(t *testing.T) {
				mvxIsNative = false
				t.Run("mvxIsNative = false", func(t *testing.T) {
					mvxIsMintBurn = false
					t.Run("mvxIsMintBurn = false", func(t *testing.T) {
						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnMultiversX", "isMintBurnOnMultiversX"})(t)
					})

					mvxIsMintBurn = true
					t.Run("mvxIsMintBurn = true", func(t *testing.T) {
						ethMintBalance := big.NewInt(0)
						ethBurnBalance := big.NewInt(100)
						mvxBurnBalance := big.NewInt(0)
						mvxMintBalance := big.NewInt(1)

						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, ErrBalanceMismatch, []string{})(t)
						mvxMintBalance = big.NewInt(0)
						testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, ethMintBalance, ethBurnBalance, zero, mvxMintBalance, mvxBurnBalance, direction, nil, []string{})(t)
					})
				})
				mvxIsNative = true
				// mvxIsMintBurn does not matter
				testBridge(ethIsMintBurn, ethIsNative, mvxIsMintBurn, mvxIsNative, zero, zero, zero, zero, zero, zero, direction, ErrInvalidSetup, []string{"isNativeOnEthereum", "isNativeOnMultiversX"})(t)
			})
		})
	})
}

func testBridge(
	ethIsMintBurn,
	ethIsNative,
	mvxIsMintBurn,
	mvxIsNative bool,
	ethTotalBalances,
	ethMintBalance,
	ethBurnBalance,
	mvxTotalBalances,
	mvxMintBalance,
	mvxBurnBalance *big.Int,
	direction batchProcessor.Direction,
	expectedErr error,
	expectedStringsInErr []string,
) func(t *testing.T) {
	return func(t *testing.T) {
		// TODO(jls): fix this
		//ethTokens := []common.Address{
		//	common.BytesToAddress([]byte("eth token")),
		//}
		//mvxTokens := [][]byte{
		//	[]byte("mvx token"),
		//}
		//amounts := []*big.Int{
		//	big.NewInt(100),
		//}

		args := createMockExecutorArgs()
		args.EthereumClient = &bridgeTests.EthereumClientStub{
			MintBurnTokensCalled: func(ctx context.Context, token common.Address) (bool, error) {
				return ethIsMintBurn, nil
			},
			NativeTokensCalled: func(ctx context.Context, token common.Address) (bool, error) {
				return ethIsNative, nil
			},
			TotalBalancesCalled: func(ctx context.Context, token common.Address) (*big.Int, error) {
				return ethTotalBalances, nil
			},
			MintBalancesCalled: func(ctx context.Context, token common.Address) (*big.Int, error) {
				return ethMintBalance, nil
			},
			BurnBalancesCalled: func(ctx context.Context, token common.Address) (*big.Int, error) {
				return ethBurnBalance, nil
			},
			CheckRequiredBalanceCalled: func(ctx context.Context, token common.Address, amount *big.Int) error {
				return nil
			},
		}

		args.MultiversXClient = &bridgeTests.MultiversXClientStub{
			IsMintBurnTokenCalled: func(ctx context.Context, token []byte) (bool, error) {
				return mvxIsMintBurn, nil
			},
			IsNativeTokenCalled: func(ctx context.Context, token []byte) (bool, error) {
				return mvxIsNative, nil
			},
			TotalBalancesCalled: func(ctx context.Context, token []byte) (*big.Int, error) {
				return mvxTotalBalances, nil
			},
			MintBalancesCalled: func(ctx context.Context, token []byte) (*big.Int, error) {
				return mvxMintBalance, nil
			},
			BurnBalancesCalled: func(ctx context.Context, token []byte) (*big.Int, error) {
				return mvxBurnBalance, nil
			},
			CheckRequiredBalanceCalled: func(ctx context.Context, token []byte, amount *big.Int) error {
				return nil
			},
		}

		// TODO(jls): fix this
		_ = direction
		_ = expectedErr
		_ = expectedStringsInErr
		// executor, _ := NewBridgeExecutor(args)
		// err := executor.CheckAvailableTokens(context.Background(), ethTokens, mvxTokens, amounts, direction)
		//
		// assert.True(t, errors.Is(err, expectedErr))
		// for _, expectedStringInErr := range expectedStringsInErr {
		//	assert.True(t, strings.Contains(err.Error(), expectedStringInErr))
		// }
	}
}

func TestConvertToDisplayableData_EmptyCallData(t *testing.T) {
	t.Parallel()

	callData := []byte{0x00}
	want := ""
	got, err := ConvertToDisplayableData(callData)
	require.Nil(t, err)
	require.Equal(t, want, got)
}

func TestConvertToDisplayableData_ValidCallDataWithNoArguments(t *testing.T) {
	t.Parallel()

	callData := func() []byte {
		b := []byte{0x01, 0x00, 0x00, 0x00, 0x03, 'a', 'b', 'c'}
		b = append(b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0, 0x04) // Gas limit length
		b = append(b, 0x1D, 0xCD, 0x65, 0x00)                        // Gas limit
		b = append(b, 0x00, 0x00, 0x00, 0x00)                        // numArguments
		return b
	}()

	want := "Endpoint: abc, Gas: 500000000, Arguments: "
	got, err := ConvertToDisplayableData(callData)
	require.Nil(t, err)
	require.Equal(t, want, got)
}

func TestConvertToDisplayableData_MultipleTypesArguments(t *testing.T) {
	t.Parallel()

	callData := func() []byte {
		b := []byte{0x01, 0x00, 0x00, 0x00, 0x03, 'a', 'b', 'c'}
		b = append(b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0, 0x04) // Gas limit length
		b = append(b, 0x1D, 0xCD, 0x65, 0x00)                        // Gas limit
		b = append(b, 0x00, 0x00, 0x00, 0x02)                        // numArguments
		b = append(b, 0x00, 0x00, 0x00, 0x05)                        // Argument 0 length
		b = append(b, bytes.Repeat([]byte{'A'}, 5)...)               // Argument 0 data
		b = append(b, 0x00, 0x00, 0x00, 0x32)                        // Argument 1 length
		b = append(b, bytes.Repeat([]byte{'B'}, 50)...)              // Argument 1 data
		return b
	}()

	want := "Endpoint: abc, Gas: 500000000, Arguments: AAAAA@BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"
	got, err := ConvertToDisplayableData(callData)
	require.Nil(t, err)
	if got != want {
		t.Errorf("Test with multiple types arguments failed. Got: %v, want: %v", got, want)
	}
}

func TestConvertToDisplayableData_TooShortForProtocolIndicator(t *testing.T) {
	t.Parallel()
	_, err := ConvertToDisplayableData([]byte{})
	require.NotNil(t, err)
	if err != nil {
		require.Equal(t, "callData too short for protocol indicator", err.Error())
	}
}

func TestConvertToDisplayableData_TooShortForEndpointNameLength(t *testing.T) {
	t.Parallel()
	_, err := ConvertToDisplayableData([]byte{0x01})
	require.NotNil(t, err)
	require.Equal(t, "callData too short while extracting the length for endpoint", err.Error())
}

func TestConvertToDisplayableData_UnexpectedProtocolIndicator(t *testing.T) {
	t.Parallel()
	_, err := ConvertToDisplayableData([]byte{0x02})
	require.NotNil(t, err)
	require.Equal(t, "callData unexpected protocol indicator: 2", err.Error())
}

func TestConvertToDisplayableData_TooShortForEndpointName(t *testing.T) {
	t.Parallel()
	_, err := ConvertToDisplayableData([]byte{0x01, 0x00, 0x00, 0x00, 0x05})
	require.NotNil(t, err)
	require.Equal(t, "callData too short while extracting the string data for endpoint", err.Error())
}

func TestConvertToDisplayableData_TooShortForGasLimitLength(t *testing.T) {
	t.Parallel()
	callData := func() []byte {
		b := []byte{0x01, 0x00, 0x00, 0x00, 0x03, 'a', 'b', 'c'}
		b = append(b, 0x00, 0x00, 0x04) // Bad Gas limit length
		return b
	}()
	_, err := ConvertToDisplayableData(callData)
	require.NotNil(t, err)
	require.Equal(t, "callData too short for gas limit length", err.Error())
}

func TestConvertToDisplayableData_TooShortForGasLimit(t *testing.T) {
	t.Parallel()
	callData := func() []byte {
		b := []byte{0x01, 0x00, 0x00, 0x00, 0x03, 'a', 'b', 'c'}
		b = append(b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0, 0x04) // Gas limit length
		b = append(b, 0x03, 0x00, 0x03)                              // Bad Gas limit
		return b
	}()
	_, err := ConvertToDisplayableData(callData)
	require.NotNil(t, err)
	require.Equal(t, "callData too short for gas limit", err.Error())
}

func TestConvertToDisplayableData_TooShortForNumberOfArgumentsLength(t *testing.T) {
	t.Parallel()
	callData := func() []byte {
		b := []byte{0x01, 0x00, 0x00, 0x00, 0x03, 'a', 'b', 'c'}
		b = append(b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0, 0x04) // Gas limit length
		b = append(b, 0x1D, 0xCD, 0x65, 0x00)                        // Gas limit
		b = append(b, 0x00, 0x00, 0x03)                              // Bad numArgument
		return b
	}()
	_, err := ConvertToDisplayableData(callData)
	require.NotNil(t, err)
	require.Equal(t, "callData too short for numArguments length", err.Error())
}

func TestConvertToDisplayableData_TooShortForArgumentLength(t *testing.T) {
	t.Parallel()
	callData := func() []byte {
		b := []byte{0x01, 0x00, 0x00, 0x00, 0x03, 'a', 'b', 'c'}
		b = append(b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0, 0x04) // Gas limit length
		b = append(b, 0x1D, 0xCD, 0x65, 0x00)                        // Gas limit
		b = append(b, 0x00, 0x00, 0x00, 0x01)                        // numArguments
		b = append(b, 0x00, 0x00, 0x04)                              // Bad Argument 0 length
		return b
	}()
	_, err := ConvertToDisplayableData(callData)
	require.NotNil(t, err)
	require.Equal(t, "callData too short while extracting the length for argument 0", err.Error())
}

func TestConvertToDisplayableData_TooShortForArgumentData(t *testing.T) {
	t.Parallel()
	callData := func() []byte {
		b := []byte{0x01, 0x00, 0x00, 0x00, 0x03, 'a', 'b', 'c'}
		b = append(b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0, 0x04) // Gas limit length
		b = append(b, 0x1D, 0xCD, 0x65, 0x00)                        // Gas limit
		b = append(b, 0x00, 0x00, 0x00, 0x01)                        // numArguments
		b = append(b, 0x00, 0x00, 0x00, 0x04)                        // Argument 0 length
		b = append(b, 0x00, 0x00, 0x04)                              // Bad Argument 0 data
		return b
	}()
	_, err := ConvertToDisplayableData(callData)
	require.NotNil(t, err)
	require.Equal(t, "callData too short while extracting the string data for argument 0", err.Error())
}
