package sui

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/mystenbcs"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/clients/sui/dtos"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	bridgeTests "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/interactors"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/blake2b"
)

func createMockSuiClientArgs() ArgsSuiClient {
	relayer := signer.NewSigner(seedBytes)

	return ArgsSuiClient{
		Proxy:                      &interactors.SuiProxyStub{},
		Log:                        logger.GetOrCreate("test"),
		RelayerPrivateKey:          relayer.PriKey,
		SafePackageId:              "0x674a8fc0a6b48c8efea86ad7ed962107c5c132a78e7cc79c9c5b9391ba8b6d83",
		SafeObjectId:               "0x32c8ebf5853163472964ce226b194af05aa5f2e4cc47678924ebe538a1c88416",
		SafeInitialSharedVersion:   425322,
		BridgePackageId:            "0x19ecc8df0b93999f87b5e2e531bb80a49d1ab2f4644032cd31f2c53ed624c94a",
		BridgeObjectId:             "0xe15513cc93d6efbfbdc7844df141b312bb677ee564a5838b7b22a891f9f05c65",
		BridgeInitialSharedVersion: 982471,
		TokensMapper: &bridgeTests.TokensMapperStub{
			ConvertTokenCalled: func(ctx context.Context, sourceBytes []byte) ([]byte, error) {
				return append([]byte("SUI"), sourceBytes...), nil
			},
		},
		Broadcaster:                  &testsCommon.BroadcasterStub{},
		StatusHandler:                &testsCommon.StatusHandlerStub{},
		SignatureHolder:              &testsCommon.SignaturesHolderStub{},
		ClientAvailabilityAllowDelta: 5,
	}
}

func TestNewSuiClient(t *testing.T) {
	t.Parallel()

	t.Run("nil proxy", func(t *testing.T) {
		args := createMockSuiClientArgs()
		args.Proxy = nil
		c, err := NewSuiClient(args)

		assert.Equal(t, errNilProxy, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("nil logger", func(t *testing.T) {
		args := createMockSuiClientArgs()
		args.Log = nil
		c, err := NewSuiClient(args)

		assert.Equal(t, clients.ErrNilLogger, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("nil tokens mapper", func(t *testing.T) {
		args := createMockSuiClientArgs()
		args.TokensMapper = nil
		c, err := NewSuiClient(args)

		assert.Equal(t, clients.ErrNilTokensMapper, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("nil broadcaster", func(t *testing.T) {
		args := createMockSuiClientArgs()
		args.Broadcaster = nil
		c, err := NewSuiClient(args)

		assert.Equal(t, clients.ErrNilBroadcaster, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("nil status handler", func(t *testing.T) {
		args := createMockSuiClientArgs()
		args.StatusHandler = nil
		c, err := NewSuiClient(args)

		assert.Equal(t, clients.ErrNilStatusHandler, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("nil tokens mapper", func(t *testing.T) {
		args := createMockSuiClientArgs()
		args.TokensMapper = nil
		c, err := NewSuiClient(args)

		assert.Equal(t, clients.ErrNilTokensMapper, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("nil signature holder", func(t *testing.T) {
		args := createMockSuiClientArgs()
		args.SignatureHolder = nil
		c, err := NewSuiClient(args)

		assert.Equal(t, clients.ErrNilSignaturesHolder, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("invalid ClientAvailabilityAllowDelta should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSuiClientArgs()
		args.ClientAvailabilityAllowDelta = 0

		c, err := NewSuiClient(args)

		assert.True(t, check.IfNil(c))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for args.AllowedDelta"))
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockSuiClientArgs()
		c, err := NewSuiClient(args)

		assert.Nil(t, err)
		assert.False(t, check.IfNil(c))
	})
}

func TestClient_GetBatch(t *testing.T) {
	t.Parallel()
	expectedErr := errors.New("expected error")

	t.Run("get batch failed should error", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiClientArgs()

		args.Proxy = &interactors.SuiProxyStub{
			SuiDevInspectTransactionBlockCalled: func(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
				return models.SuiTransactionBlockResponse{}, expectedErr
			},
		}

		c, _ := NewSuiClient(args)
		batch, isFinal, err := c.GetBatch(context.Background(), batchNonce)
		assert.Nil(t, batch)
		assert.False(t, isFinal)
		assert.ErrorIs(t, err, expectedErr)
	})
	t.Run("get batch deposits failed should error", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiClientArgs()

		getBatchCalled := false
		args.Proxy = &interactors.SuiProxyStub{
			SuiDevInspectTransactionBlockCalled: func(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
				// Simulate a call to get batch
				if getBatchCalled == false {
					getBatchCalled = true

					res, _ := createResultsRawFromValues(dtos.Batch{
						Nonce:         batchNonce,
						DepositsCount: 2,
						TimestampMs:   100,
					}, true)
					return models.SuiTransactionBlockResponse{
						Effects: models.SuiEffects{
							Status: models.ExecutionStatus{
								Status: "success",
							},
						},
						Results: res,
					}, nil
				}

				// Simulate a call to get batch deposits
				return models.SuiTransactionBlockResponse{
					Effects: models.SuiEffects{
						Status: models.ExecutionStatus{
							Status: "failed",
						},
					},
				}, expectedErr
			},
		}

		c, _ := NewSuiClient(args)
		batch, isFinal, err := c.GetBatch(context.Background(), batchNonce)
		assert.Nil(t, batch)
		assert.False(t, isFinal)
		assert.ErrorIs(t, err, expectedErr)
	})
	t.Run("deposits count mismatch should error", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiClientArgs()

		getBatchCalled := false
		args.Proxy = &interactors.SuiProxyStub{
			SuiDevInspectTransactionBlockCalled: func(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
				// Simulate a call to get batch
				if getBatchCalled == false {
					getBatchCalled = true

					res, _ := createResultsRawFromValues(dtos.Batch{
						Nonce:         batchNonce,
						DepositsCount: 2,
						TimestampMs:   100,
					}, true)
					return models.SuiTransactionBlockResponse{
						Effects: models.SuiEffects{
							Status: models.ExecutionStatus{
								Status: "success",
							},
						},
						Results: res,
					}, nil
				}

				// Simulate a call to get batch deposits
				res, _ := createResultsRawFromValues(make([]dtos.Deposit, 4), true)
				return models.SuiTransactionBlockResponse{
					Effects: models.SuiEffects{
						Status: models.ExecutionStatus{
							Status: "success",
						},
					},
					Results: res,
				}, nil
			},
		}

		c, _ := NewSuiClient(args)
		batch, isFinal, err := c.GetBatch(context.Background(), batchNonce)
		assert.Nil(t, batch)
		assert.False(t, isFinal)
		assert.True(t, errors.Is(err, clients.ErrDepositsAndBatchDepositsCountDiffer))
	})
	t.Run("token conversion failed should error", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiClientArgs()

		wasGetBatchCalled := false
		args.TokensMapper = &bridgeTests.TokensMapperStub{
			ConvertTokenCalled: func(ctx context.Context, sourceTokenBytes []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		args.Proxy = &interactors.SuiProxyStub{
			SuiDevInspectTransactionBlockCalled: func(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
				// First call for get_batch
				if wasGetBatchCalled == false {
					wasGetBatchCalled = true

					res, _ := createResultsRawFromValues(dtos.Batch{
						Nonce:         batchNonce,
						DepositsCount: 1,
						TimestampMs:   100,
					}, true)
					return models.SuiTransactionBlockResponse{
						Effects: models.SuiEffects{
							Status: models.ExecutionStatus{
								Status: "success",
							},
						},
						Results: res,
					}, nil
				}

				// Second call for get_batch_deposits
				deposits := []dtos.Deposit{
					{
						Nonce:        42,
						TokenAddress: "coin::Coin::0x1",
						Amount:       1000,
						Depositor:    bytes.Repeat([]byte{0x1}, 32),
						Recipient:    bytes.Repeat([]byte{0x2}, 32),
						Status:       uint8(1),
					},
				}
				res, _ := createResultsRawFromValues(deposits, true)
				return models.SuiTransactionBlockResponse{
					Effects: models.SuiEffects{
						Status: models.ExecutionStatus{
							Status: "success",
						},
					},
					Results: res,
				}, nil
			},
		}

		c, _ := NewSuiClient(args)
		batch, isFinal, err := c.GetBatch(context.Background(), batchNonce)
		assert.Nil(t, batch)
		assert.False(t, isFinal)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("returns batch should work", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiClientArgs()

		from1 := testsCommon.CreateRandomSuiAddressBytes()
		token1 := testsCommon.CreateRandomCoinId()
		recipient1 := testsCommon.CreateRandomMultiversXAddress()

		from2 := testsCommon.CreateRandomSuiAddressBytes()
		token2 := testsCommon.CreateRandomCoinId()
		recipient2 := testsCommon.CreateRandomMultiversXAddress()

		wasGetBatchCalled := false

		args.Proxy = &interactors.SuiProxyStub{
			SuiDevInspectTransactionBlockCalled: func(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
				// First call for get_batch
				if wasGetBatchCalled == false {
					wasGetBatchCalled = true

					res, _ := createResultsRawFromValues(dtos.Batch{
						Nonce:                  batchNonce,
						DepositsCount:          2,
						LastUpdatedTimestampMs: 100,
						TimestampMs:            100,
					}, true)
					return models.SuiTransactionBlockResponse{
						Effects: models.SuiEffects{
							Status: models.ExecutionStatus{
								Status: "success",
							},
						},
						Results: res,
					}, nil
				}

				// Second call for get_batch_deposits
				deposits := []dtos.Deposit{
					{
						Nonce:        42,
						TokenAddress: token1,
						Amount:       20,
						Depositor:    from1[:],
						Recipient:    recipient1.AddressBytes(),
						Status:       uint8(1),
					},
					{
						Nonce:        43,
						TokenAddress: token2,
						Amount:       40,
						Depositor:    from2[:],
						Recipient:    recipient2.AddressBytes(),
						Status:       uint8(2),
					},
				}
				res, _ := createResultsRawFromValues(deposits, true)
				return models.SuiTransactionBlockResponse{
					Effects: models.SuiEffects{
						Status: models.ExecutionStatus{
							Status: "success",
						},
					},
					Results: res,
				}, nil
			},
		}

		bech32Recipient1Address, _ := recipient1.AddressAsBech32String()
		bech32Recipient2Address, _ := recipient2.AddressAsBech32String()
		expectedBatch := &core.TransferBatch{
			ID:          batchNonce,
			BlockNumber: 100,
			Deposits: []*core.DepositTransfer{
				{
					Nonce:                 42,
					ToBytes:               recipient1.AddressBytes(),
					DisplayableTo:         bech32Recipient1Address,
					FromBytes:             from1[:],
					DisplayableFrom:       hex.EncodeToString(from1[:]),
					SourceTokenBytes:      []byte(token1),
					DisplayableToken:      token1,
					Amount:                big.NewInt(20),
					DestinationTokenBytes: append([]byte("SUI"), token1[:]...),
				},
				{
					Nonce:                 43,
					ToBytes:               recipient2.AddressBytes(),
					DisplayableTo:         bech32Recipient2Address,
					FromBytes:             from2[:],
					DisplayableFrom:       hex.EncodeToString(from2[:]),
					SourceTokenBytes:      []byte(token2),
					DisplayableToken:      token2,
					Amount:                big.NewInt(40),
					DestinationTokenBytes: append([]byte("SUI"), token2[:]...),
				},
			},
			Statuses: make([]byte, 2),
		}

		c, _ := NewSuiClient(args)
		batch, isFinal, err := c.GetBatch(context.Background(), batchNonce)
		assert.Equal(t, expectedBatch, batch)
		assert.True(t, isFinal)
		assert.Nil(t, err)
	})
	t.Run("returns non final batch should work", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiClientArgs()

		from1 := testsCommon.CreateRandomSuiAddressBytes()
		token1 := testsCommon.CreateRandomCoinId()
		recipient1 := testsCommon.CreateRandomMultiversXAddress()

		from2 := testsCommon.CreateRandomSuiAddressBytes()
		token2 := testsCommon.CreateRandomCoinId()
		recipient2 := testsCommon.CreateRandomMultiversXAddress()

		wasGetBatchCalled := false

		args.Proxy = &interactors.SuiProxyStub{
			SuiDevInspectTransactionBlockCalled: func(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
				// First call for get_batch
				if wasGetBatchCalled == false {
					wasGetBatchCalled = true

					res, _ := createResultsRawFromValues(dtos.Batch{
						Nonce:                  batchNonce,
						DepositsCount:          2,
						LastUpdatedTimestampMs: 100,
						TimestampMs:            100,
					}, false)
					return models.SuiTransactionBlockResponse{
						Effects: models.SuiEffects{
							Status: models.ExecutionStatus{
								Status: "success",
							},
						},
						Results: res,
					}, nil
				}

				// Second call for get_batch_deposits
				deposits := []dtos.Deposit{
					{
						Nonce:        42,
						TokenAddress: token1,
						Amount:       20,
						Depositor:    from1[:],
						Recipient:    recipient1.AddressBytes(),
						Status:       uint8(1),
					},
					{
						Nonce:        43,
						TokenAddress: token2,
						Amount:       40,
						Depositor:    from2[:],
						Recipient:    recipient2.AddressBytes(),
						Status:       uint8(2),
					},
				}
				res, _ := createResultsRawFromValues(deposits, true)
				return models.SuiTransactionBlockResponse{
					Effects: models.SuiEffects{
						Status: models.ExecutionStatus{
							Status: "success",
						},
					},
					Results: res,
				}, nil
			},
		}

		bech32Recipient1Address, _ := recipient1.AddressAsBech32String()
		bech32Recipient2Address, _ := recipient2.AddressAsBech32String()
		expectedBatch := &core.TransferBatch{
			ID:          batchNonce,
			BlockNumber: 100,
			Deposits: []*core.DepositTransfer{
				{
					Nonce:                 42,
					ToBytes:               recipient1.AddressBytes(),
					DisplayableTo:         bech32Recipient1Address,
					FromBytes:             from1[:],
					DisplayableFrom:       hex.EncodeToString(from1[:]),
					SourceTokenBytes:      []byte(token1),
					DisplayableToken:      token1,
					Amount:                big.NewInt(20),
					DestinationTokenBytes: append([]byte("SUI"), token1[:]...),
				},
				{
					Nonce:                 43,
					ToBytes:               recipient2.AddressBytes(),
					DisplayableTo:         bech32Recipient2Address,
					FromBytes:             from2[:],
					DisplayableFrom:       hex.EncodeToString(from2[:]),
					SourceTokenBytes:      []byte(token2),
					DisplayableToken:      token2,
					Amount:                big.NewInt(40),
					DestinationTokenBytes: append([]byte("SUI"), token2[:]...),
				},
			},
			Statuses: make([]byte, 2),
		}

		c, _ := NewSuiClient(args)
		batch, isFinal, err := c.GetBatch(context.Background(), batchNonce)
		assert.Equal(t, expectedBatch, batch)
		assert.False(t, isFinal)
		assert.Nil(t, err)
	})
}

func TestClient_BroadcastSignatureForMessageHash(t *testing.T) {
	t.Parallel()

	t.Run("sign failed should not broadcast", func(t *testing.T) {
		t.Parallel()

		expectedError := errors.New("expected error")
		msgToSign := []byte("message to sign")
		args := createMockSuiClientArgs()
		c, _ := NewSuiClient(args)

		c.txHandler = &bridgeTests.SuiTxHandlerStub{
			SignCalled: func(msg []byte) ([]byte, error) {
				assert.Equal(t, msgToSign, msg)
				return nil, expectedError
			},
		}
		c.broadcaster = &testsCommon.BroadcasterStub{
			BroadcastSignatureCalled: func(signature []byte, messageHash []byte) {
				assert.Fail(t, "should have not called broadcast")
			},
		}

		c.BroadcastSignatureForMessageHash(msgToSign)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedSig := "expected sig"
		broadcastCalled := false

		msgToSign := []byte("message to sign")
		args := createMockSuiClientArgs()
		c, _ := NewSuiClient(args)

		c.txHandler = &bridgeTests.SuiTxHandlerStub{
			SignCalled: func(msg []byte) ([]byte, error) {
				assert.Equal(t, msgToSign, msg)
				return []byte(expectedSig), nil
			},
		}
		c.broadcaster = &testsCommon.BroadcasterStub{
			BroadcastSignatureCalled: func(signature []byte, message []byte) {
				assert.Equal(t, msgToSign, message)
				assert.Equal(t, expectedSig, string(signature))
				broadcastCalled = true
			},
		}

		c.BroadcastSignatureForMessageHash(msgToSign)

		assert.True(t, broadcastCalled)
	})
}

func TestClient_WasExecuted(t *testing.T) {
	t.Parallel()

	wasCalled := false
	args := createMockSuiClientArgs()
	args.Proxy = &interactors.SuiProxyStub{
		SuiDevInspectTransactionBlockCalled: func(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
			wasCalled = true
			values, _ := createResultsRawFromValues(true)

			return models.SuiTransactionBlockResponse{
				Effects: models.SuiEffects{
					Status: models.ExecutionStatus{
						Status: "success",
					},
				},
				Results: values,
			}, nil
		},
	}
	c, _ := NewSuiClient(args)
	wasExecuted, err := c.WasExecuted(context.Background(), 1)

	assert.True(t, wasExecuted)
	assert.True(t, wasCalled)
	assert.Nil(t, err)
}

func TestClient_CheckRequiredBalance(t *testing.T) {
	t.Parallel()
	args := createMockSuiClientArgs()
	coinType := []byte(testsCommon.CreateRandomCoinId())
	balance := big.NewInt(1000000)

	t.Run("get balance fails should error", func(t *testing.T) {
		expectedErr := errors.New("expected error GetBalance")
		c, _ := NewSuiClient(args)
		c.proxy = &interactors.SuiProxyStub{
			SuiXGetBalanceCalled: func(ctx context.Context, req models.SuiXGetBalanceRequest) (models.CoinBalanceResponse, error) {
				return models.CoinBalanceResponse{}, expectedErr
			},
		}

		err := c.CheckRequiredBalance(context.Background(), coinType, balance)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("not enough coins", func(t *testing.T) {
		c, _ := NewSuiClient(args)
		c.proxy = &interactors.SuiProxyStub{
			SuiXGetBalanceCalled: func(ctx context.Context, req models.SuiXGetBalanceRequest) (models.CoinBalanceResponse, error) {
				return models.CoinBalanceResponse{
					CoinType:     string(coinType),
					TotalBalance: balance.String(),
				}, nil
			},
		}

		err := c.CheckRequiredBalance(context.Background(), coinType, big.NewInt(0).Add(balance, big.NewInt(1)))
		assert.True(t, errors.Is(err, errInsufficientCoinBalance))
	})
	t.Run("should work", func(t *testing.T) {
		c, _ := NewSuiClient(args)
		c.proxy = &interactors.SuiProxyStub{
			SuiXGetBalanceCalled: func(ctx context.Context, req models.SuiXGetBalanceRequest) (models.CoinBalanceResponse, error) {
				return models.CoinBalanceResponse{
					CoinType:     string(coinType),
					TotalBalance: balance.String(),
				}, nil
			},
		}

		err := c.CheckRequiredBalance(context.Background(), coinType, balance)
		assert.Nil(t, err)
	})
}

func TestClient_TotalBalances(t *testing.T) {
	t.Parallel()
	coinType := testsCommon.CreateRandomCoinId()
	coinTypeBytes := []byte(coinType)

	t.Run("error while getting total balances", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockSuiClientArgs()
		args.Proxy = createFailMockProxy(expectedErr)
		c, _ := NewSuiClient(args)

		balances, err := c.TotalBalances(context.Background(), coinTypeBytes)
		assert.Nil(t, balances)
		fmt.Println(err)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedBalance := uint64(100)
		values, _ := createResultsRawFromValues(providedBalance)
		args := createMockSuiClientArgs()
		args.Proxy = createMockProxy(values)
		c, _ := NewSuiClient(args)

		balances, err := c.TotalBalances(context.Background(), coinTypeBytes)
		assert.Nil(t, err)
		assert.Equal(t, big.NewInt(0).SetUint64(providedBalance), balances)
	})
}

func TestClient_GetTransactionsStatuses(t *testing.T) {
	t.Parallel()

	expectedStatuses := []byte{1, 2, 3}
	expectedBatchID := big.NewInt(2232)
	expectedErr := errors.New("expected error")

	t.Run("operation error, should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSuiClientArgs()
		args.Proxy = createFailMockProxy(expectedErr)

		c, _ := NewSuiClient(args)
		statuses, err := c.GetTransactionsStatuses(context.Background(), expectedBatchID.Uint64())
		assert.Nil(t, statuses)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("statuses are not final, should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSuiClientArgs()
		values, _ := createResultsRawFromValues([]byte("dummy"), false)
		args.Proxy = createMockProxy(values)

		c, _ := NewSuiClient(args)
		statuses, err := c.GetTransactionsStatuses(context.Background(), expectedBatchID.Uint64())
		assert.Nil(t, statuses)
		assert.Equal(t, clients.ErrStatusIsNotFinal, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockSuiClientArgs()
		values, _ := createResultsRawFromValues(expectedStatuses, true)
		args.Proxy = createMockProxy(values)

		c, _ := NewSuiClient(args)
		statuses, err := c.GetTransactionsStatuses(context.Background(), expectedBatchID.Uint64())
		assert.Equal(t, expectedStatuses, statuses)
		assert.Nil(t, err)
	})
}

func TestClient_GenerateMessageHash(t *testing.T) {
	t.Parallel()

	t.Run("should return error when batch is nil", func(t *testing.T) {
		args := createMockSuiClientArgs()
		c, _ := NewSuiClient(args)
		hash, err := c.GenerateMessageHash(nil, 42)

		assert.Nil(t, hash)
		assert.True(t, errors.Is(err, clients.ErrNilBatch))
	})

	t.Run("should return correct hash for valid input", func(t *testing.T) {
		args := createMockSuiClientArgs()
		c, _ := NewSuiClient(args)

		batch := &batchProcessor.ArgListsBatch{
			PeerTokens:    [][]byte{[]byte("token1"), []byte("token2")},
			Recipients:    [][]byte{[]byte("recipient1"), []byte("recipient2")},
			MvxTokenBytes: [][]byte{[]byte("mvxToken1"), []byte("mvxToken2")},
			Amounts:       []*big.Int{big.NewInt(100), big.NewInt(200)},
			Nonces:        []*big.Int{big.NewInt(1), big.NewInt(2)},
			Direction:     batchProcessor.FromMultiversX,
		}
		batchID := uint64(123)

		uint64Amounts := make([]uint64, 0, len(batch.Amounts))
		for _, amount := range batch.Amounts {
			uint64Amounts = append(uint64Amounts, amount.Uint64())
		}

		uint64Nonces := make([]uint64, 0, len(batch.Nonces))
		for _, nonce := range batch.Nonces {
			uint64Nonces = append(uint64Nonces, nonce.Uint64())
		}

		expectedData := batchProcessor.SuiTransferData{
			Recipients: batch.Recipients,
			SuiTokens:  batch.PeerTokens,
			Amounts:    uint64Amounts,
			Nonces:     uint64Nonces,
			BatchId:    batchID,
		}

		expectedBytes, err := mystenbcs.Marshal(expectedData)
		assert.NoError(t, err)

		expectedHash := blake2b.Sum256(expectedBytes)

		hash, err := c.GenerateMessageHash(batch, batchID)
		assert.NoError(t, err)
		assert.Equal(t, expectedHash[:], hash)
	})
}

func TestClient_GetQuorumSize(t *testing.T) {
	t.Parallel()

	args := createMockSuiClientArgs()
	providedValue := uint64(6453)
	values, _ := createResultsRawFromValues(providedValue, true)
	args.Proxy = createMockProxy(values)

	c, _ := NewSuiClient(args)
	quorum, err := c.GetQuorumSize(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(0).SetUint64(providedValue), quorum)
}

func TestClient_IsQuorumReached(t *testing.T) {
	t.Parallel()
	msg := []byte("message")

	t.Run("quorum errors", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockSuiClientArgs()
		args.Proxy = createFailMockProxy(expectedErr)
		c, _ := NewSuiClient(args)

		isReached, err := c.IsQuorumReached(context.Background(), msg)
		assert.False(t, isReached)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("quorum returns less than minimum allowed", func(t *testing.T) {
		t.Parallel()

		args := createMockSuiClientArgs()
		values, _ := createResultsRawFromValues(uint64(0))
		args.Proxy = createMockProxy(values)
		c, _ := NewSuiClient(args)

		isReached, err := c.IsQuorumReached(context.Background(), msg)
		assert.False(t, isReached)
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "in IsQuorumReached, minQuorum"))
	})
	t.Run("quorum values comparison", func(t *testing.T) {
		t.Parallel()

		signatures := make([][]byte, 0)
		args := createMockSuiClientArgs()
		values, _ := createResultsRawFromValues(uint64(3))
		args.Proxy = createMockProxy(values)

		args.SignatureHolder = &testsCommon.SignaturesHolderStub{
			SignaturesCalled: func(messageHash []byte) [][]byte {
				return signatures
			},
		}
		c, _ := NewSuiClient(args)

		isReached, err := c.IsQuorumReached(context.Background(), msg)
		assert.False(t, isReached)
		assert.Nil(t, err)

		signatures = append(signatures, []byte("sig"))
		signatures = append(signatures, []byte("sig"))
		isReached, err = c.IsQuorumReached(context.Background(), msg)
		assert.False(t, isReached)
		assert.Nil(t, err)

		signatures = append(signatures, []byte("sig"))
		isReached, err = c.IsQuorumReached(context.Background(), msg)
		assert.True(t, isReached)
		assert.Nil(t, err)

		signatures = append(signatures, []byte("sig"))
		isReached, err = c.IsQuorumReached(context.Background(), msg)
		assert.True(t, isReached)
		assert.Nil(t, err)
	})
}

func TestClient_CheckClientAvailability(t *testing.T) {
	t.Parallel()

	currentCheckpoint := uint64(0)
	incrementor := uint64(1)
	args := createMockSuiClientArgs()
	statusHandler := testsCommon.NewStatusHandlerMock("test")
	expectedErr := errors.New("expected error")
	args.StatusHandler = statusHandler
	args.Proxy = &interactors.SuiProxyStub{
		SuiGetLatestCheckpointSequenceNumberCalled: func(ctx context.Context) (uint64, error) {
			currentCheckpoint += incrementor
			return currentCheckpoint, nil
		},
	}

	c, _ := NewSuiClient(args)

	t.Run("different current checkpoint should update - 10 times", func(t *testing.T) {
		resetClient(c)
		for i := 0; i < 10; i++ {
			err := c.CheckClientAvailability(context.Background())
			assert.Nil(t, err)
			checkStatusHandler(t, statusHandler, core.Available, "")
		}
		assert.True(t, statusHandler.GetIntMetric(core.MetricLastBlockNonce) > 0)
	})
	t.Run("same current checkpoint should error after a while", func(t *testing.T) {
		resetClient(c)
		_ = c.CheckClientAvailability(context.Background())

		incrementor = 0

		// place a random message as to test it is reset
		statusHandler.SetStringMetric(core.MetricMultiversXClientStatus, core.ClientStatus(3).String())
		statusHandler.SetStringMetric(core.MetricLastMultiversXClientError, "random")

		// this will just increment the retry counter
		for i := 0; i < int(args.ClientAvailabilityAllowDelta); i++ {
			err := c.CheckClientAvailability(context.Background())
			assert.Nil(t, err)
			checkStatusHandler(t, statusHandler, core.Available, "")
		}

		for i := 0; i < 10; i++ {
			message := fmt.Sprintf("block %d fetched for %d times in a row", currentCheckpoint, args.ClientAvailabilityAllowDelta+uint64(i+1))
			err := c.CheckClientAvailability(context.Background())
			assert.Nil(t, err)
			checkStatusHandler(t, statusHandler, core.Unavailable, message)
		}
	})
	t.Run("same current checkpoint should error after a while and then recovers", func(t *testing.T) {
		resetClient(c)
		_ = c.CheckClientAvailability(context.Background())

		incrementor = 0

		// this will just increment the retry counter
		for i := 0; i < int(args.ClientAvailabilityAllowDelta); i++ {
			err := c.CheckClientAvailability(context.Background())
			assert.Nil(t, err)
			checkStatusHandler(t, statusHandler, core.Available, "")
		}

		for i := 0; i < 10; i++ {
			message := fmt.Sprintf("block %d fetched for %d times in a row", currentCheckpoint, args.ClientAvailabilityAllowDelta+uint64(i+1))
			err := c.CheckClientAvailability(context.Background())
			assert.Nil(t, err)
			checkStatusHandler(t, statusHandler, core.Unavailable, message)
		}

		incrementor = 1
		err := c.CheckClientAvailability(context.Background())
		assert.Nil(t, err)
		checkStatusHandler(t, statusHandler, core.Available, "")
	})
	t.Run("get current checkpoint errors", func(t *testing.T) {
		resetClient(c)
		c.proxy = &interactors.SuiProxyStub{
			SuiGetLatestCheckpointSequenceNumberCalled: func(ctx context.Context) (uint64, error) {
				return 0, expectedErr
			},
		}

		err := c.CheckClientAvailability(context.Background())
		checkStatusHandler(t, statusHandler, core.Unavailable, expectedErr.Error())
		assert.Equal(t, expectedErr, err)
	})
}

func resetClient(c *client) {
	c.mut.Lock()
	c.retriesAvailabilityCheck = 0
	c.mut.Unlock()
	c.statusHandler.SetStringMetric(core.MetricMultiversXClientStatus, "")
	c.statusHandler.SetStringMetric(core.MetricLastMultiversXClientError, "")
}

func checkStatusHandler(t *testing.T, statusHandler *testsCommon.StatusHandlerMock, status core.ClientStatus, message string) {
	assert.Equal(t, status.String(), statusHandler.GetStringMetric(core.MetricMultiversXClientStatus))
	assert.Equal(t, message, statusHandler.GetStringMetric(core.MetricLastMultiversXClientError))
}
