package v2

import (
    "context"
    "errors"
    "math/big"
    "strings"
    "testing"

    "github.com/ElrondNetwork/elrond-eth-bridge/clients"
    "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
    "github.com/ElrondNetwork/elrond-go-core/core/check"
    logger "github.com/ElrondNetwork/elrond-go-logger"
    "github.com/ethereum/go-ethereum/common"
    "github.com/stretchr/testify/assert"
)

var expectedErr = errors.New("expected error")
var providedBatch = &clients.TransferBatch{}

func createMockBaseExecutorArgs() ArgsBaseBridgeExecutor {
    return ArgsBaseBridgeExecutor{
        Log:              logger.GetOrCreate("test"),
        ElrondClient:     &bridgeV2.ElrondClientStub{},
        EthereumClient:   &bridgeV2.EthereumClientStub{},
    }
}

func createMockEthToElrondExecutorArgs() ArgsEthToElrondBridgeExectutor {
    return ArgsEthToElrondBridgeExectutor{
        ArgsBaseBridgeExecutor:   createMockBaseExecutorArgs(),
        TopologyProviderOnElrond: &bridgeV2.TopologyProviderStub{},
    }
}

func createMockElrondToEthExecutorArgs() ArgsElrondToEthBridgeExectutor {
    return ArgsElrondToEthBridgeExectutor{
        ArgsBaseBridgeExecutor:     createMockBaseExecutorArgs(),
        TopologyProviderOnElrond:   &bridgeV2.TopologyProviderStub{},
        TopologyProviderOnEthereum: &bridgeV2.TopologyProviderStub{},
    }
}

func TestCreateEthToElrondBridgeExecutor(t *testing.T) {
    t.Parallel()

    t.Run("nil logger should error", func(t *testing.T) {
        t.Parallel()

        args := createMockEthToElrondExecutorArgs()
        args.Log = nil
        executor, err := CreateEthToElrondBridgeExecutor(args)

        assert.True(t, check.IfNil(executor))
        assert.Equal(t, ErrNilLogger, err)
    })
    t.Run("nil elrond client should error", func(t *testing.T) {
        t.Parallel()

        args := createMockEthToElrondExecutorArgs()
        args.ElrondClient = nil
        executor, err := CreateEthToElrondBridgeExecutor(args)

        assert.True(t, check.IfNil(executor))
        assert.Equal(t, ErrNilElrondClient, err)
    })
    t.Run("nil ethereum client should error", func(t *testing.T) {
        t.Parallel()

        args := createMockEthToElrondExecutorArgs()
        args.EthereumClient = nil
        executor, err := CreateEthToElrondBridgeExecutor(args)

        assert.True(t, check.IfNil(executor))
        assert.Equal(t, ErrNilEthereumClient, err)
    })
    t.Run("nil topology provider should error", func(t *testing.T) {
        t.Parallel()

        args := createMockEthToElrondExecutorArgs()
        args.TopologyProviderOnElrond = nil
        executor, err := CreateEthToElrondBridgeExecutor(args)

        assert.True(t, check.IfNil(executor))
        assert.Equal(t, ErrNilElrondTopologyProvider, err)
    })
    t.Run("should work", func(t *testing.T) {
        t.Parallel()

        args := createMockEthToElrondExecutorArgs()
        executor, err := CreateEthToElrondBridgeExecutor(args)

        assert.False(t, check.IfNil(executor))
        assert.Nil(t, err)
    })
}

func TestEthToElrondBridgeExecutor_GetLogger(t *testing.T) {
    t.Parallel()

    args := createMockEthToElrondExecutorArgs()
    executor, _ := CreateEthToElrondBridgeExecutor(args)

    assert.True(t, args.Log == executor.GetLogger()) // pointer testing
}

func TestEthToElrondBridgeExecutor_MyTurnAsLeaderOnElrond(t *testing.T) {
    t.Parallel()

    args := createMockEthToElrondExecutorArgs()
    wasCalled := false
    args.TopologyProviderOnElrond = &bridgeV2.TopologyProviderStub{
        MyTurnAsLeaderCalled: func() bool {
            wasCalled = true
            return true
        },
    }

    executor, _ := CreateEthToElrondBridgeExecutor(args)
    assert.True(t, executor.MyTurnAsLeaderOnElrond())
    assert.True(t, wasCalled)
}

func TestEthToElrondBridgeExecutor_GetAndStoreActionIDForProposeTransferOnElrond(t *testing.T) {
    t.Parallel()

    t.Run("nil batch should error", func(t *testing.T) {
        t.Parallel()

        args := createMockEthToElrondExecutorArgs()
        executor, _ := CreateEthToElrondBridgeExecutor(args)

        actionID, err := executor.GetAndStoreActionIDForProposeTransferOnElrond(context.Background())
        assert.Zero(t, actionID)
        assert.Equal(t, ErrNilBatch, err)
    })
    t.Run("elrond client errors", func(t *testing.T) {
        t.Parallel()

        args := createMockEthToElrondExecutorArgs()
        args.ElrondClient = &bridgeV2.ElrondClientStub{
            GetActionIDForProposeTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (uint64, error) {
                assert.True(t, providedBatch == batch)
                return 0, expectedErr
            },
        }
        executor, _ := CreateEthToElrondBridgeExecutor(args)
        executor.batch = providedBatch

        actionID, err := executor.GetAndStoreActionIDForProposeTransferOnElrond(context.Background())
        assert.Zero(t, actionID)
        assert.Equal(t, expectedErr, err)
    })
    t.Run("should work", func(t *testing.T) {
        t.Parallel()

        args := createMockEthToElrondExecutorArgs()
        providedActionID := uint64(48939)

        args.ElrondClient = &bridgeV2.ElrondClientStub{
            GetActionIDForProposeTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (uint64, error) {
                assert.True(t, providedBatch == batch)
                return providedActionID, nil
            },
        }
        executor, _ := CreateEthToElrondBridgeExecutor(args)
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

        args := createMockEthToElrondExecutorArgs()
        providedNonce := uint64(8346)
        args.EthereumClient = &bridgeV2.EthereumClientStub{
            GetBatchCalled: func(ctx context.Context, nonce uint64) (*clients.TransferBatch, error) {
                assert.Equal(t, providedNonce, nonce)
                return nil, expectedErr
            },
        }
        executor, _ := CreateEthToElrondBridgeExecutor(args)
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
        executor, _ := CreateEthToElrondBridgeExecutor(args)
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
    executor, _ := CreateEthToElrondBridgeExecutor(args)

    batchID, err := executor.GetLastExecutedEthBatchIDFromElrond(context.Background())
    assert.Equal(t, providedBatchID, batchID)
    assert.Nil(t, err)
}

func TestEthToElrondBridgeExecutor_VerifyLastDepositNonceExecutedOnEthereumBatch(t *testing.T) {
    t.Parallel()

    t.Run("nil batch should error", func(t *testing.T) {
        t.Parallel()

        args := createMockEthToElrondExecutorArgs()
        executor, _ := CreateEthToElrondBridgeExecutor(args)

        err := executor.VerifyLastDepositNonceExecutedOnEthereumBatch(context.Background())
        assert.Equal(t, ErrNilBatch, err)
    })
    t.Run("get last executed tx id errors", func(t *testing.T) {
        t.Parallel()

        args := createMockEthToElrondExecutorArgs()
        args.ElrondClient = &bridgeV2.ElrondClientStub{
            GetLastExecutedEthTxIDCalled: func(ctx context.Context) (uint64, error) {
                return 0, expectedErr
            },
        }
        executor, _ := CreateEthToElrondBridgeExecutor(args)
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

        executor, _ := CreateEthToElrondBridgeExecutor(args)
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

        executor, _ := CreateEthToElrondBridgeExecutor(args)
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

        executor, _ := CreateEthToElrondBridgeExecutor(args)
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

        executor, _ := CreateEthToElrondBridgeExecutor(args)
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
        executor, _ := CreateEthToElrondBridgeExecutor(args)

        wasTransfered, err := executor.WasTransferProposedOnElrond(context.Background())
        assert.False(t, wasTransfered)
        assert.Equal(t, ErrNilBatch, err)
    })
    t.Run("should work", func(t *testing.T) {
        t.Parallel()

        args := createMockEthToElrondExecutorArgs()
        wasCalled := false
        args.ElrondClient = &bridgeV2.ElrondClientStub{
            WasProposedTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
                assert.True(t, providedBatch == batch)
                wasCalled = true
                return true, nil
            },
        }

        executor, _ := CreateEthToElrondBridgeExecutor(args)
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
        executor, _ := CreateEthToElrondBridgeExecutor(args)

        err := executor.ProposeTransferOnElrond(context.Background())
        assert.Equal(t, ErrNilBatch, err)
    })
    t.Run("propose transfer fails", func(t *testing.T) {
        t.Parallel()

        args := createMockEthToElrondExecutorArgs()
        args.ElrondClient = &bridgeV2.ElrondClientStub{
            ProposeTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (string, error) {
                assert.True(t, providedBatch == batch)

                return "", expectedErr
            },
        }
        executor, _ := CreateEthToElrondBridgeExecutor(args)
        executor.batch = providedBatch

        err := executor.ProposeTransferOnElrond(context.Background())
        assert.Equal(t, expectedErr, err)
    })
    t.Run("should work", func(t *testing.T) {
        t.Parallel()

        args := createMockEthToElrondExecutorArgs()
        wasCalled := false
        args.ElrondClient = &bridgeV2.ElrondClientStub{
            ProposeTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (string, error) {
                assert.True(t, providedBatch == batch)
                wasCalled = true

                return "", nil
            },
        }
        executor, _ := CreateEthToElrondBridgeExecutor(args)
        executor.batch = providedBatch

        err := executor.ProposeTransferOnElrond(context.Background())
        assert.Nil(t, err)
        assert.True(t, wasCalled)
    })
}

func TestEthToElrondBridgeExecutor_WasActionSignedOnElrond(t *testing.T) {
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
    executor, _ := CreateEthToElrondBridgeExecutor(args)
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

        args := createMockEthToElrondExecutorArgs()
        providedActionID := uint64(378276)
        args.ElrondClient = &bridgeV2.ElrondClientStub{
            SignCalled: func(ctx context.Context, actionID uint64) (string, error) {
                assert.Equal(t, providedActionID, actionID)
                return "", expectedErr
            },
        }

        executor, _ := CreateEthToElrondBridgeExecutor(args)
        executor.actionID = providedActionID

        err := executor.SignActionOnElrond(context.Background())
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

        executor, _ := CreateEthToElrondBridgeExecutor(args)
        executor.actionID = providedActionID

        err := executor.SignActionOnElrond(context.Background())
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
    executor, _ := CreateEthToElrondBridgeExecutor(args)
    executor.actionID = providedActionID

    isQuorumReached, err := executor.IsQuorumReachedOnElrond(context.Background())
    assert.True(t, isQuorumReached)
    assert.Nil(t, err)
    assert.True(t, wasCalled)
}

func TestEthToElrondBridgeExecutor_WasActionPerformedOnElrond(t *testing.T) {
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
    executor, _ := CreateEthToElrondBridgeExecutor(args)
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

        args := createMockEthToElrondExecutorArgs()
        executor, _ := CreateEthToElrondBridgeExecutor(args)

        err := executor.PerformActionOnElrond(context.Background())
        assert.Equal(t, ErrNilBatch, err)
    })
    t.Run("elrond client errors", func(t *testing.T) {
        t.Parallel()

        args := createMockEthToElrondExecutorArgs()
        providedActionID := uint64(7383)
        args.ElrondClient = &bridgeV2.ElrondClientStub{
            PerformActionCalled: func(ctx context.Context, actionID uint64, batch *clients.TransferBatch) (string, error) {
                assert.Equal(t, providedActionID, actionID)
                assert.True(t, providedBatch == batch)
                return "", expectedErr
            },
        }
        executor, _ := CreateEthToElrondBridgeExecutor(args)
        executor.batch = providedBatch
        executor.actionID = providedActionID

        err := executor.PerformActionOnElrond(context.Background())
        assert.Equal(t, expectedErr, err)
    })
    t.Run("should work", func(t *testing.T) {
        t.Parallel()

        args := createMockEthToElrondExecutorArgs()
        wasCalled := false
        providedActionID := uint64(7383)
        args.ElrondClient = &bridgeV2.ElrondClientStub{
            PerformActionCalled: func(ctx context.Context, actionID uint64, batch *clients.TransferBatch) (string, error) {
                assert.Equal(t, providedActionID, actionID)
                assert.True(t, providedBatch == batch)
                wasCalled = true
                return "", nil
            },
        }
        executor, _ := CreateEthToElrondBridgeExecutor(args)
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
    args := createMockEthToElrondExecutorArgs()
    wasCalled := false
    args.ElrondClient = &bridgeV2.ElrondClientStub{
        GetMaxNumberOfRetriesOnQuorumReachedCalled: func() uint64 {
            wasCalled = true
            return expectedMaxRetries
        },
    }
    executor, _ := CreateEthToElrondBridgeExecutor(args)
    for i := uint64(0); i < expectedMaxRetries; i++ {
        assert.False(t, executor.ProcessMaxRetriesOnElrond())
    }

    assert.Equal(t, expectedMaxRetries, executor.retriesOnElrond)
    assert.True(t, executor.ProcessMaxRetriesOnElrond())
    executor.ResetRetriesCountOnElrond()
    assert.Equal(t, uint64(0), executor.retriesOnElrond)
    assert.True(t, wasCalled)
}

func TestCreateElrondToEthBridgeExecutor(t *testing.T) {
    t.Parallel()

    t.Run("nil logger should error", func(t *testing.T) {
        t.Parallel()

        args := createMockElrondToEthExecutorArgs()
        args.Log = nil
        executor, err := CreateElrondToEthBridgeExecutor(args)

        assert.True(t, check.IfNil(executor))
        assert.Equal(t, ErrNilLogger, err)
    })
    t.Run("nil elrond client should error", func(t *testing.T) {
        t.Parallel()

        args := createMockElrondToEthExecutorArgs()
        args.ElrondClient = nil
        executor, err := CreateElrondToEthBridgeExecutor(args)

        assert.True(t, check.IfNil(executor))
        assert.Equal(t, ErrNilElrondClient, err)
    })
    t.Run("nil ethereum client should error", func(t *testing.T) {
        t.Parallel()

        args := createMockElrondToEthExecutorArgs()
        args.EthereumClient = nil
        executor, err := CreateElrondToEthBridgeExecutor(args)

        assert.True(t, check.IfNil(executor))
        assert.Equal(t, ErrNilEthereumClient, err)
    })
    t.Run("nil elrond topology provider should error", func(t *testing.T) {
        t.Parallel()

        args := createMockElrondToEthExecutorArgs()
        args.TopologyProviderOnElrond = nil
        executor, err := CreateElrondToEthBridgeExecutor(args)

        assert.True(t, check.IfNil(executor))
        assert.Equal(t, ErrNilElrondTopologyProvider, err)
    })
    t.Run("nil ethereum topology provider should error", func(t *testing.T) {
        t.Parallel()

        args := createMockElrondToEthExecutorArgs()
        args.TopologyProviderOnEthereum = nil
        executor, err := CreateElrondToEthBridgeExecutor(args)

        assert.True(t, check.IfNil(executor))
        assert.Equal(t, ErrNilEthereumTopologyProvider, err)
    })
    t.Run("should work", func(t *testing.T) {
        t.Parallel()

        args := createMockElrondToEthExecutorArgs()
        executor, err := CreateElrondToEthBridgeExecutor(args)

        assert.False(t, check.IfNil(executor))
        assert.Nil(t, err)
    })
}

func TestElrondToEthBridgeExecutor_GetAndStoreBatchFromElrond(t *testing.T) {
    t.Parallel()

    t.Run("GetAndStoreBatchFromElrond fails", func(t *testing.T) {
        t.Parallel()

        args := createMockElrondToEthExecutorArgs()
        args.ElrondClient = &bridgeV2.ElrondClientStub{
            GetPendingCalled: func(ctx context.Context) (*clients.TransferBatch, error) {
                return nil, expectedErr
            },
        }

        executor, _ := CreateElrondToEthBridgeExecutor(args)
        err := executor.GetAndStoreBatchFromElrond(context.Background())
        assert.Equal(t, expectedErr, err)

        batch := executor.GetStoredBatch()
        assert.Nil(t, batch)
    })
    t.Run("should work", func(t *testing.T) {
        t.Parallel()

        wasCalled := false
        args := createMockElrondToEthExecutorArgs()
        args.ElrondClient = &bridgeV2.ElrondClientStub{
            GetPendingCalled: func(ctx context.Context) (*clients.TransferBatch, error) {
                wasCalled = true
                return providedBatch, nil
            },
        }

        executor, _ := CreateElrondToEthBridgeExecutor(args)
        err := executor.GetAndStoreBatchFromElrond(context.Background())
        assert.True(t, wasCalled)
        assert.Equal(t, providedBatch, executor.GetStoredBatch())
        assert.Nil(t, err)
    })
}

func TestElrondToEthBridgeExecutor_GetAndStoreActionIDForProposeSetStatusFromElrond(t *testing.T) {
    t.Parallel()

    t.Run("nil batch should error", func(t *testing.T) {
        t.Parallel()

        args := createMockElrondToEthExecutorArgs()
        executor, _ := CreateElrondToEthBridgeExecutor(args)

        actionId, err := executor.GetAndStoreActionIDForProposeSetStatusFromElrond(context.Background())
        assert.Equal(t, ErrNilBatch, err)
        assert.Equal(t, InvalidActionID, actionId)
    })
    t.Run("GetAndStoreActionIDForProposeSetStatusFromElrond fails", func(t *testing.T) {
        t.Parallel()

        args := createMockElrondToEthExecutorArgs()
        args.ElrondClient = &bridgeV2.ElrondClientStub{
            GetActionIDForSetStatusOnPendingTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (uint64, error) {
                return uint64(0), expectedErr
            },
        }

        executor, _ := CreateElrondToEthBridgeExecutor(args)
        executor.batch = providedBatch
        _, err := executor.GetAndStoreActionIDForProposeSetStatusFromElrond(context.Background())
        assert.Equal(t, expectedErr, err)
    })
    t.Run("should work", func(t *testing.T) {
        t.Parallel()

        wasCalled := false
        providedActionId := uint64(1123)
        args := createMockElrondToEthExecutorArgs()
        args.ElrondClient = &bridgeV2.ElrondClientStub{
            GetActionIDForSetStatusOnPendingTransferCalled: func(ctx context.Context, batch *clients.TransferBatch) (uint64, error) {
                wasCalled = true
                return providedActionId, nil
            },
        }

        executor, _ := CreateElrondToEthBridgeExecutor(args)
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

        args := createMockElrondToEthExecutorArgs()
        executor, _ := CreateElrondToEthBridgeExecutor(args)

        wasProposed, err := executor.WasSetStatusProposedOnElrond(context.Background())
        assert.Equal(t, ErrNilBatch, err)
        assert.False(t, wasProposed)
    })
    t.Run("WasSetStatusProposedOnElrond fails", func(t *testing.T) {
        t.Parallel()

        args := createMockElrondToEthExecutorArgs()
        args.ElrondClient = &bridgeV2.ElrondClientStub{
            WasProposedSetStatusCalled: func(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
                return false, expectedErr
            },
        }

        executor, _ := CreateElrondToEthBridgeExecutor(args)
        executor.batch = providedBatch
        _, err := executor.WasSetStatusProposedOnElrond(context.Background())
        assert.Equal(t, expectedErr, err)
    })
    t.Run("should work", func(t *testing.T) {
        t.Parallel()

        wasCalled := false
        args := createMockElrondToEthExecutorArgs()
        args.ElrondClient = &bridgeV2.ElrondClientStub{
            WasProposedSetStatusCalled: func(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
                assert.True(t, providedBatch == batch)
                wasCalled = true
                return true, nil
            },
        }

        executor, _ := CreateElrondToEthBridgeExecutor(args)
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

        args := createMockElrondToEthExecutorArgs()
        executor, _ := CreateElrondToEthBridgeExecutor(args)

        err := executor.ProposeSetStatusOnElrond(context.Background())
        assert.Equal(t, ErrNilBatch, err)
    })
    t.Run("ProposeSetStatusOnElrond fails", func(t *testing.T) {
        t.Parallel()

        args := createMockElrondToEthExecutorArgs()
        args.ElrondClient = &bridgeV2.ElrondClientStub{
            ProposeSetStatusCalled: func(ctx context.Context, batch *clients.TransferBatch) (string, error) {
                return "", expectedErr
            },
        }

        executor, _ := CreateElrondToEthBridgeExecutor(args)
        executor.batch = providedBatch
        err := executor.ProposeSetStatusOnElrond(context.Background())
        assert.Equal(t, expectedErr, err)
    })
    t.Run("should work", func(t *testing.T) {
        t.Parallel()

        wasCalled := false
        args := createMockElrondToEthExecutorArgs()
        args.ElrondClient = &bridgeV2.ElrondClientStub{
            ProposeSetStatusCalled: func(ctx context.Context, batch *clients.TransferBatch) (string, error) {
                assert.True(t, providedBatch == batch)
                wasCalled = true

                return "", nil
            },
        }

        executor, _ := CreateElrondToEthBridgeExecutor(args)
        executor.batch = providedBatch

        err := executor.ProposeSetStatusOnElrond(context.Background())
        assert.Nil(t, err)
        assert.True(t, wasCalled)
    })
}

func TestEthToElrondBridgeExecutor_MyTurnAsLeaderOnEthereum(t *testing.T) {
    t.Parallel()

    args := createMockElrondToEthExecutorArgs()
    wasCalled := false
    args.TopologyProviderOnEthereum = &bridgeV2.TopologyProviderStub{
        MyTurnAsLeaderCalled: func() bool {
            wasCalled = true
            return true
        },
    }

    executor, _ := CreateElrondToEthBridgeExecutor(args)
    assert.True(t, executor.MyTurnAsLeaderOnEthereum())
    assert.True(t, wasCalled)
}

func TestElrondToEthBridgeExecutor_WasTransferPerformedOnEthereum(t *testing.T) {
    t.Parallel()

    t.Run("nil batch should error", func(t *testing.T) {
        t.Parallel()

        args := createMockElrondToEthExecutorArgs()
        executor, _ := CreateElrondToEthBridgeExecutor(args)

        _, err := executor.WasTransferPerformedOnEthereum(context.Background())
        assert.Equal(t, ErrNilBatch, err)
    })
    t.Run("WasExecuted fails", func(t *testing.T) {
        t.Parallel()

        args := createMockElrondToEthExecutorArgs()
        args.EthereumClient = &bridgeV2.EthereumClientStub{
            WasExecutedCalled: func(ctx context.Context, batchID uint64) (bool, error) {
                return false, expectedErr
            },
        }

        executor, _ := CreateElrondToEthBridgeExecutor(args)
        executor.batch = providedBatch
        _, err := executor.WasTransferPerformedOnEthereum(context.Background())
        assert.Equal(t, expectedErr, err)
    })
    t.Run("should work", func(t *testing.T) {
        t.Parallel()

        wasCalled := false
        providedBatchID := uint64(36727)
        args := createMockElrondToEthExecutorArgs()
        args.EthereumClient = &bridgeV2.EthereumClientStub{
            WasExecutedCalled: func(ctx context.Context, batchID uint64) (bool, error) {
                assert.True(t, providedBatchID == batchID)
                wasCalled = true
                return true, nil
            },
        }

        executor, _ := CreateElrondToEthBridgeExecutor(args)
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

        args := createMockElrondToEthExecutorArgs()
        executor, _ := CreateElrondToEthBridgeExecutor(args)

        err := executor.SignTransferOnEthereum(context.Background())
        assert.Equal(t, ErrNilBatch, err)
    })
    t.Run("GenerateMessageHash fails", func(t *testing.T) {
        t.Parallel()

        args := createMockElrondToEthExecutorArgs()
        args.EthereumClient = &bridgeV2.EthereumClientStub{
            GenerateMessageHashCalled: func(batch *clients.TransferBatch) (common.Hash, error) {
                return common.Hash{}, expectedErr
            },
        }

        executor, _ := CreateElrondToEthBridgeExecutor(args)
        executor.batch = providedBatch
        err := executor.SignTransferOnEthereum(context.Background())
        assert.Equal(t, expectedErr, err)
    })
    t.Run("should work", func(t *testing.T) {
        t.Parallel()

        wasCalledGenerateMessageHashCalled := false
        wasCalledBroadcastSignatureForMessageHashCalled := false
        args := createMockElrondToEthExecutorArgs()
        args.EthereumClient = &bridgeV2.EthereumClientStub{
            GenerateMessageHashCalled: func(batch *clients.TransferBatch) (common.Hash, error) {
                wasCalledGenerateMessageHashCalled = true
                return common.Hash{}, nil
            },
            BroadcastSignatureForMessageHashCalled: func(msgHash common.Hash) {
                wasCalledBroadcastSignatureForMessageHashCalled = true
            },
        }

        executor, _ := CreateElrondToEthBridgeExecutor(args)
        executor.batch = providedBatch
        err := executor.SignTransferOnEthereum(context.Background())
        assert.Nil(t, err)
        assert.True(t, wasCalledGenerateMessageHashCalled)
        assert.True(t, wasCalledBroadcastSignatureForMessageHashCalled)
    })
}

func TestElrondToEthBridgeExecutor_PerformTransferOnEthereum(t *testing.T) {
    t.Parallel()

    t.Run("nil batch should error", func(t *testing.T) {
        t.Parallel()

        args := createMockElrondToEthExecutorArgs()
        executor, _ := CreateElrondToEthBridgeExecutor(args)

        err := executor.PerformTransferOnEthereum(context.Background())
        assert.Equal(t, ErrNilBatch, err)
    })
    t.Run("GetQuorumSize fails", func(t *testing.T) {
        t.Parallel()

        args := createMockElrondToEthExecutorArgs()
        args.EthereumClient = &bridgeV2.EthereumClientStub{
            GetQuorumSizeCalled: func() (*big.Int, error) {
                return big.NewInt(0), expectedErr
            },
        }

        executor, _ := CreateElrondToEthBridgeExecutor(args)
        executor.batch = providedBatch
        err := executor.PerformTransferOnEthereum(context.Background())
        assert.Equal(t, expectedErr, err)
    })
    t.Run("ExecuteTransfer fails", func(t *testing.T) {
        t.Parallel()

        args := createMockElrondToEthExecutorArgs()
        args.EthereumClient = &bridgeV2.EthereumClientStub{
            GetQuorumSizeCalled: func() (*big.Int, error) {
                return big.NewInt(0), nil
            },
            ExecuteTransferCalled: func(ctx context.Context, msgHash common.Hash, batch *clients.TransferBatch, quorum int) (string, error) {
                return "", expectedErr
            },
        }

        executor, _ := CreateElrondToEthBridgeExecutor(args)
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
        args := createMockElrondToEthExecutorArgs()
        args.EthereumClient = &bridgeV2.EthereumClientStub{
            GetQuorumSizeCalled: func() (*big.Int, error) {
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

        executor, _ := CreateElrondToEthBridgeExecutor(args)
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

    t.Run("IsQuorumReached fails", func(t *testing.T) {
        t.Parallel()

        args := createMockElrondToEthExecutorArgs()
        args.EthereumClient = &bridgeV2.EthereumClientStub{
            IsQuorumReachedCalled: func() (bool, error) {
                return false, expectedErr
            },
        }

        executor, _ := CreateElrondToEthBridgeExecutor(args)

        _, err := executor.IsQuorumReachedOnEthereum(context.Background())
        assert.Equal(t, expectedErr, err)
    })
    t.Run("should work", func(t *testing.T) {
        t.Parallel()

        args := createMockElrondToEthExecutorArgs()
        wasCalled := false
        args.EthereumClient = &bridgeV2.EthereumClientStub{
            IsQuorumReachedCalled: func() (bool, error) {
                wasCalled = true
                return true, nil
            },
        }

        executor, _ := CreateElrondToEthBridgeExecutor(args)

        isReached, err := executor.IsQuorumReachedOnEthereum(context.Background())
        assert.Nil(t, err)
        assert.True(t, wasCalled)
        assert.True(t, isReached)
    })
}

func TestElrondToEthBridgeExecutor_RetriesCountOnEthereum(t *testing.T) {
    t.Parallel()

    expectedMaxRetries := uint64(3)
    args := createMockElrondToEthExecutorArgs()
    wasCalled := false
    args.EthereumClient = &bridgeV2.EthereumClientStub{
        GetMaxNumberOfRetriesOnQuorumReachedCalled: func() uint64 {
            wasCalled = true
            return expectedMaxRetries
        },
    }
    executor, _ := CreateElrondToEthBridgeExecutor(args)
    for i := uint64(0); i < expectedMaxRetries; i++ {
        assert.False(t, executor.ProcessMaxRetriesOnEthereum())
    }

    assert.Equal(t, expectedMaxRetries, executor.retriesOnEthereum)
    assert.True(t, executor.ProcessMaxRetriesOnEthereum())
    executor.ResetRetriesCountOnEthereum()
    assert.Equal(t, uint64(0), executor.retriesOnEthereum)
    assert.True(t, wasCalled)
}
