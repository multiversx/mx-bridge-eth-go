package sui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/mystenbcs"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/block-vision/sui-go-sdk/transaction"
	"github.com/multiversx/mx-bridge-eth-go/clients"
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
	relayer := signer.NewSigner(bytes.Repeat([]byte{0x1}, 32))

	return ArgsSuiClient{
		Proxy:                        &interactors.SuiProxyStub{},
		TxHandler:                    &bridgeTests.SuiTxHandlerStub{},
		Log:                          logger.GetOrCreate("test"),
		Signer:                       relayer,
		PackageId:                    "0x674a8fc0a6b48c8efea86ad7ed962107c5c132a78e7cc79c9c5b9391ba8b6d83",
		SafeObjectId:                 "0x32c8ebf5853163472964ce226b194af05aa5f2e4cc47678924ebe538a1c88416",
		SafeInitialSharedVersion:     425322,
		BridgeObjectId:               "0xe15513cc93d6efbfbdc7844df141b312bb677ee564a5838b7b22a891f9f05c65",
		BridgeInitialSharedVersion:   982471,
		TreasuryObjectId:             "0x32c8ebf5853163472964ce226b194af05aa5f2e4cc47678924ebe538a1c88416",
		TreasuryInitialSharedVersion: 425322,
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

func generateHashForTokenGroup(t *testing.T, batchID uint64, group *TokenTransferGroup) []byte {
	var msg bytes.Buffer
	enc := mystenbcs.NewEncoder(&msg)

	err := enc.Encode(batchID)
	assert.NoError(t, err)

	for i := 0; i < len(group.Tokens); i++ {
		err = enc.Encode(group.Tokens[i][2:])
		assert.NoError(t, err)

		err = enc.Encode(group.Recipients[i])
		assert.NoError(t, err)

		err = enc.Encode(group.Amounts[i])
		assert.NoError(t, err)

		err = enc.Encode(group.Nonces[i])
		assert.NoError(t, err)
	}

	expectedHash := blake2b.Sum256(msg.Bytes())
	return expectedHash[:]
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
	t.Run("nil transaction handler", func(t *testing.T) {
		args := createMockSuiClientArgs()
		args.TxHandler = nil
		c, err := NewSuiClient(args)

		assert.Equal(t, errNilTxHandler, err)
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

					res, _ := createResultsRawFromValues(Batch{
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

					res, _ := createResultsRawFromValues(Batch{
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
				res, _ := createResultsRawFromValues(make([]Deposit, 4), true)
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

					res, _ := createResultsRawFromValues(Batch{
						Nonce:                  batchNonce,
						DepositsCount:          1,
						LastUpdatedTimestampMs: 90,
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
				deposits := []Deposit{
					{
						Nonce:          42,
						TokenTypeBytes: []byte("1::coin::Coin"),
						Amount:         1000,
						Sender:         testsCommon.CreateRandomSuiAddressBytes(),
						Recipient:      bytes.Repeat([]byte{0x2}, 32),
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

					res, _ := createResultsRawFromValues(Batch{
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
				deposits := []Deposit{
					{
						Nonce:          42,
						TokenTypeBytes: []byte(token1),
						Amount:         20,
						Sender:         from1,
						Recipient:      recipient1.AddressBytes(),
					},
					{
						Nonce:          43,
						TokenTypeBytes: []byte(token2),
						Amount:         40,
						Sender:         from2,
						Recipient:      recipient2.AddressBytes(),
					},
				}
				res, _ := createResultsRawFromValues(deposits, true)
				x := models.SuiTransactionBlockResponse{
					Effects: models.SuiEffects{
						Status: models.ExecutionStatus{
							Status: "success",
						},
					},
					Results: res,
				}
				return x, nil
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
					DisplayableFrom:       suiAddressFromBytes(from1[:]),
					SourceTokenBytes:      []byte(token1),
					DisplayableToken:      "0x" + token1,
					Amount:                big.NewInt(20),
					DestinationTokenBytes: append([]byte("SUI"), AppendLengthToData([]byte("0x"+token1))...),
				},
				{
					Nonce:                 43,
					ToBytes:               recipient2.AddressBytes(),
					DisplayableTo:         bech32Recipient2Address,
					FromBytes:             from2[:],
					DisplayableFrom:       suiAddressFromBytes(from2[:]),
					SourceTokenBytes:      []byte(token2),
					DisplayableToken:      "0x" + token2,
					Amount:                big.NewInt(40),
					DestinationTokenBytes: append([]byte("SUI"), AppendLengthToData([]byte("0x"+token2))...),
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

					res, _ := createResultsRawFromValues(Batch{
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
				deposits := []Deposit{
					{
						Nonce:          42,
						TokenTypeBytes: []byte(token1),
						Amount:         20,
						Sender:         from1,
						Recipient:      recipient1.AddressBytes(),
					},
					{
						Nonce:          43,
						TokenTypeBytes: []byte(token2),
						Amount:         40,
						Sender:         from2,
						Recipient:      recipient2.AddressBytes(),
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
					DisplayableFrom:       suiAddressFromBytes(from1[:]),
					SourceTokenBytes:      []byte(token1),
					DisplayableToken:      "0x" + token1,
					Amount:                big.NewInt(20),
					DestinationTokenBytes: append([]byte("SUI"), AppendLengthToData([]byte("0x"+token1))...),
				},
				{
					Nonce:                 43,
					ToBytes:               recipient2.AddressBytes(),
					DisplayableTo:         bech32Recipient2Address,
					FromBytes:             from2[:],
					DisplayableFrom:       suiAddressFromBytes(from2[:]),
					SourceTokenBytes:      []byte(token2),
					DisplayableToken:      "0x" + token2,
					Amount:                big.NewInt(40),
					DestinationTokenBytes: append([]byte("SUI"), AppendLengthToData([]byte("0x"+token2))...),
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

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedSig := "ACPnde90uco6f/ISc8tgB5luaNU+bONOvDeLBh7vMhtpJT2ebziR+1U5YetJA1vKzrFbok24CcGAbt/xOJQvEAGKiOPddAnxlf1S2y08ul1yymcJvx2UEhvzdIgBtA9vXA=="
		msgToSign := bytes.Repeat([]byte{'a'}, 32)
		broadcastCalled := false
		args := createMockSuiClientArgs()
		c, _ := NewSuiClient(args)

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
		args.Proxy = &interactors.SuiProxyStub{
			SuiDevInspectTransactionBlockCalled: func(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
				return models.SuiTransactionBlockResponse{}, expectedErr
			},
		}
		c, _ := NewSuiClient(args)

		err := c.CheckRequiredBalance(context.Background(), coinType, balance)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("not enough coins", func(t *testing.T) {
		args.Proxy = &interactors.SuiProxyStub{
			SuiDevInspectTransactionBlockCalled: func(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
				res, _ := createResultsRawFromValues(balance.Uint64())
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

		err := c.CheckRequiredBalance(context.Background(), coinType, big.NewInt(0).Add(balance, big.NewInt(1)))
		assert.True(t, errors.Is(err, errInsufficientCoinBalance))
	})
	t.Run("should work", func(t *testing.T) {
		args.Proxy = &interactors.SuiProxyStub{
			SuiDevInspectTransactionBlockCalled: func(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
				res, _ := createResultsRawFromValues(balance.Uint64())
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

	t.Run("should return correct hash for batch with same token deposits", func(t *testing.T) {
		args := createMockSuiClientArgs()
		c, _ := NewSuiClient(args)

		batch := &batchProcessor.ArgListsBatch{
			PeerTokens:    [][]byte{[]byte("token1"), []byte("token1")},
			Recipients:    [][]byte{[]byte("recipient1"), []byte("recipient2")},
			MvxTokenBytes: [][]byte{[]byte("mvxToken1"), []byte("mvxToken1")},
			Amounts:       []*big.Int{big.NewInt(100), big.NewInt(200)},
			Nonces:        []*big.Int{big.NewInt(1), big.NewInt(2)},
			Direction:     batchProcessor.FromMultiversX,
		}
		batchID := uint64(123)

		suiAddress1 := suiAddressFromBytes(batch.Recipients[0])
		suiAddressBytes1, _ := transaction.ConvertSuiAddressStringToBytes(models.SuiAddress(suiAddress1))

		suiAddress2 := suiAddressFromBytes(batch.Recipients[1])
		suiAddressBytes2, _ := transaction.ConvertSuiAddressStringToBytes(models.SuiAddress(suiAddress2))

		expectedHash := generateHashForTokenGroup(t, batchID, &TokenTransferGroup{
			Tokens:     [][]byte{batch.PeerTokens[0], batch.PeerTokens[1]},
			Recipients: []models.SuiAddressBytes{*suiAddressBytes1, *suiAddressBytes2},
			Amounts:    []uint64{batch.Amounts[0].Uint64(), batch.Amounts[1].Uint64()},
			Nonces:     []uint64{batch.Nonces[0].Uint64(), batch.Nonces[1].Uint64()},
		})

		hash, err := c.GenerateMessageHash(batch, batchID)
		assert.NoError(t, err)
		assert.Equal(t, expectedHash, hash)
	})
	t.Run("should return same hash for batches with same deposits in different order", func(t *testing.T) {
		args := createMockSuiClientArgs()
		c, _ := NewSuiClient(args)

		batchID := uint64(456)
		batch1 := &batchProcessor.ArgListsBatch{
			PeerTokens:    [][]byte{[]byte("token1"), []byte("token2")},
			Recipients:    [][]byte{[]byte("recipient1"), []byte("recipient2")},
			MvxTokenBytes: [][]byte{[]byte("mvxToken1"), []byte("mvxToken2")},
			Amounts:       []*big.Int{big.NewInt(900), big.NewInt(560)},
			Nonces:        []*big.Int{big.NewInt(640), big.NewInt(310)},
			Direction:     batchProcessor.FromMultiversX,
		}
		batch2 := &batchProcessor.ArgListsBatch{
			PeerTokens:    [][]byte{[]byte("token2"), []byte("token1")},
			Recipients:    [][]byte{[]byte("recipient2"), []byte("recipient1")},
			MvxTokenBytes: [][]byte{[]byte("mvxToken2"), []byte("mvxToken1")},
			Amounts:       []*big.Int{big.NewInt(560), big.NewInt(900)},
			Nonces:        []*big.Int{big.NewInt(310), big.NewInt(640)},
			Direction:     batchProcessor.FromMultiversX,
		}

		hash1, err := c.GenerateMessageHash(batch1, batchID)
		assert.NoError(t, err)
		assert.NotNil(t, hash1)

		hash2, err := c.GenerateMessageHash(batch2, batchID)
		assert.NoError(t, err)
		assert.NotNil(t, hash2)

		assert.True(t, bytes.Equal(hash1, hash2))

		suiAddress1 := suiAddressFromBytes(batch1.Recipients[0])
		suiAddressBytes1, _ := transaction.ConvertSuiAddressStringToBytes(models.SuiAddress(suiAddress1))

		suiAddress2 := suiAddressFromBytes(batch1.Recipients[1])
		suiAddressBytes2, _ := transaction.ConvertSuiAddressStringToBytes(models.SuiAddress(suiAddress2))

		// we use batch1 data to generate expected hashes because is ordered correctly
		expectedHashForToken1 := generateHashForTokenGroup(t, batchID, &TokenTransferGroup{
			Tokens:     [][]byte{batch1.PeerTokens[0]},
			Recipients: []models.SuiAddressBytes{*suiAddressBytes1},
			Amounts:    []uint64{batch1.Amounts[0].Uint64()},
			Nonces:     []uint64{batch1.Nonces[0].Uint64()},
		})

		expectedHashForToken2 := generateHashForTokenGroup(t, batchID, &TokenTransferGroup{
			Tokens:     [][]byte{batch1.PeerTokens[1]},
			Recipients: []models.SuiAddressBytes{*suiAddressBytes2},
			Amounts:    []uint64{batch1.Amounts[1].Uint64()},
			Nonces:     []uint64{batch1.Nonces[1].Uint64()},
		})

		assert.True(t, bytes.Equal(expectedHashForToken1, hash1[:32]))
		assert.True(t, bytes.Equal(expectedHashForToken2, hash1[32:]))
	})
	t.Run("should be deterministic over multiple invocations", func(t *testing.T) {
		args := createMockSuiClientArgs()
		c, _ := NewSuiClient(args)

		batch := &batchProcessor.ArgListsBatch{
			PeerTokens:    [][]byte{[]byte("token1"), []byte("token2")},
			Recipients:    [][]byte{[]byte("recipient1"), []byte("recipient2")},
			MvxTokenBytes: [][]byte{[]byte("mvxToken1"), []byte("mvxToken2")},
			Amounts:       []*big.Int{big.NewInt(900), big.NewInt(560)},
			Nonces:        []*big.Int{big.NewInt(640), big.NewInt(310)},
			Direction:     batchProcessor.FromMultiversX,
		}
		batchID := uint64(456)

		const iterations = 100
		var hashes [][]byte

		for i := 0; i < iterations; i++ {
			hash, err := c.GenerateMessageHash(batch, batchID)
			assert.NoError(t, err, "iteration %d failed", i)
			assert.NotNil(t, hash)
			assert.NotEmpty(t, hash)

			hashes = append(hashes, hash)
		}

		referenceHash := hashes[0]
		for i := 1; i < len(hashes); i++ {
			assert.True(t, bytes.Equal(referenceHash, hashes[i]),
				"Hash mismatch at iteration %d.\nExpected: %x\nGot: %x",
				i, referenceHash, hashes[i])
		}
	})
}

func TestClient_ExecuteTransfer(t *testing.T) {
	t.Parallel()

	args := createMockSuiClientArgs()
	batch := &core.TransferBatch{
		ID: 332,
		Deposits: []*core.DepositTransfer{
			{
				Nonce:                 10,
				ToBytes:               []byte("to1"),
				DisplayableTo:         "to1",
				FromBytes:             []byte("from1"),
				DisplayableFrom:       "from1",
				SourceTokenBytes:      []byte("source token1"),
				DisplayableToken:      "token1",
				Amount:                big.NewInt(20),
				DestinationTokenBytes: []byte("0x123::suitoken::SUItoken1"),
			},
			{
				Nonce:                 30,
				ToBytes:               []byte("to2"),
				DisplayableTo:         "to2",
				FromBytes:             []byte("from2"),
				DisplayableFrom:       "from2",
				SourceTokenBytes:      []byte("source token2"),
				DisplayableToken:      "token2",
				Amount:                big.NewInt(40),
				DestinationTokenBytes: []byte("0x123::suitoken::SUItoken2"),
			},
		},
		Statuses: make([]byte, 2),
	}
	argLists := batchProcessor.ExtractListFromMvx(batch)
	signatures := make([][]byte, 10)
	for i := range signatures {
		signatures[i] = []byte(fmt.Sprintf("sig %d", i))
	}

	t.Run("nil batch", func(t *testing.T) {
		c, _ := NewSuiClient(args)
		hash, err := c.ExecuteTransfer(context.Background(), []byte{}, nil, 0, 10)
		assert.Equal(t, "", hash)
		assert.True(t, errors.Is(err, clients.ErrNilBatch))
	})
	t.Run("check if the contract is paused fails", func(t *testing.T) {
		expectedErr := errors.New("expected error is paused")
		c, _ := NewSuiClient(args)
		c.suiClientDataGetter.proxy = createFailMockProxy(expectedErr)
		hash, err := c.ExecuteTransfer(context.Background(), []byte{}, argLists, batch.ID, 10)
		assert.Equal(t, "", hash)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("contract is paused should error", func(t *testing.T) {
		c, _ := NewSuiClient(args)
		values, err := createResultsRawFromValues(true)
		assert.NoError(t, err)
		c.suiClientDataGetter.proxy = createMockProxy(values)
		hash, err := c.ExecuteTransfer(context.Background(), []byte{}, argLists, batch.ID, 10)
		assert.Equal(t, "", hash)
		assert.True(t, errors.Is(err, clients.ErrMultisigContractPaused))
	})
	t.Run("not enough quorum", func(t *testing.T) {
		c, _ := NewSuiClient(args)

		values, err := createResultsRawFromValues(false)
		assert.NoError(t, err)

		c.suiClientDataGetter.proxy = createMockProxy(values)
		c.signatureHolder = &testsCommon.SignaturesHolderStub{
			SignaturesCalled: func(messageHash []byte) [][]byte {
				return signatures[:9]
			},
		}
		hash, err := c.ExecuteTransfer(context.Background(), []byte{}, argLists, batch.ID, 10)
		assert.Equal(t, "", hash)
		assert.True(t, errors.Is(err, clients.ErrQuorumNotReached))
		assert.True(t, strings.Contains(err.Error(), "num signatures: 9, quorum: 10"))
	})
	t.Run("error getting gas coin", func(t *testing.T) {
		expectedErr := errors.New("expected error")
		c, _ := NewSuiClient(args)

		values, err := createResultsRawFromValues(false)
		assert.NoError(t, err)

		c.signatureHolder = &testsCommon.SignaturesHolderStub{
			SignaturesCalled: func(messageHash []byte) [][]byte {
				return signatures
			},
		}
		c.suiClientDataGetter.proxy = &interactors.SuiProxyStub{
			SuiDevInspectTransactionBlockCalled: func(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
				return models.SuiTransactionBlockResponse{
					Effects: models.SuiEffects{
						Status: models.ExecutionStatus{
							Status: "success",
						},
					},
					Results: values,
				}, nil
			},
			SuiXGetCoinsCalled: func(ctx context.Context, req models.SuiXGetCoinsRequest) (models.PaginatedCoinsResponse, error) {
				return models.PaginatedCoinsResponse{}, expectedErr
			},
		}
		hash, err := c.ExecuteTransfer(context.Background(), []byte{}, argLists, batch.ID, 10)
		assert.Equal(t, "", hash)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("relayer has no coins for gas", func(t *testing.T) {
		c, _ := NewSuiClient(args)

		values, err := createResultsRawFromValues(false)
		assert.NoError(t, err)

		c.signatureHolder = &testsCommon.SignaturesHolderStub{
			SignaturesCalled: func(messageHash []byte) [][]byte {
				return signatures
			},
		}
		c.suiClientDataGetter.proxy = &interactors.SuiProxyStub{
			SuiDevInspectTransactionBlockCalled: func(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
				return models.SuiTransactionBlockResponse{
					Effects: models.SuiEffects{
						Status: models.ExecutionStatus{
							Status: "success",
						},
					},
					Results: values,
				}, nil
			},
			SuiXGetCoinsCalled: func(ctx context.Context, req models.SuiXGetCoinsRequest) (models.PaginatedCoinsResponse, error) {
				return models.PaginatedCoinsResponse{
					Data: []models.CoinData{},
				}, nil
			},
		}
		hash, err := c.ExecuteTransfer(context.Background(), []byte{}, argLists, batch.ID, 10)
		assert.Equal(t, "", hash)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), c.relayerAddress)
	})
	t.Run("execute transfer errors", func(t *testing.T) {
		expectedErr := errors.New("expected error execute transfer")

		values, err := createResultsRawFromValues(false)
		assert.NoError(t, err)

		args.Proxy = &sui.Client{}
		c, err := NewSuiClient(args)
		assert.NoError(t, err)

		c.signatureHolder = &testsCommon.SignaturesHolderStub{
			SignaturesCalled: func(messageHash []byte) [][]byte {
				return signatures[:9]
			},
		}
		c.suiClientDataGetter.proxy = &interactors.SuiProxyStub{
			SuiDevInspectTransactionBlockCalled: func(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
				return models.SuiTransactionBlockResponse{
					Effects: models.SuiEffects{
						Status: models.ExecutionStatus{
							Status: "success",
						},
					},
					Results: values,
				}, nil
			},
			SuiXGetCoinsCalled: func(ctx context.Context, req models.SuiXGetCoinsRequest) (models.PaginatedCoinsResponse, error) {
				return models.PaginatedCoinsResponse{
					Data: []models.CoinData{
						{
							CoinObjectId:        "0x0e82989b575d6ebb4e7f6062f5463bdd1cce9db77b9f34ac923b90031765fddf",
							Version:             "1",
							Digest:              "4Nd1mHZtwVaFgqsSVBz3tH6KvXZcG1oMqezLh8u6BbhE",
							Balance:             "100000",
							CoinType:            "0x2::sui::SUI",
							PreviousTransaction: "9rS8PZyT1QK5qLgN",
						},
					},
					NextCursor:  "",
					HasNextPage: false,
				}, nil
			},
		}
		c.txHandler = &bridgeTests.SuiTxHandlerStub{
			SendTransactionReturnHashCalled: func(ctx context.Context, gasCoin *transaction.SuiObjectRef, calls []core.SuiPTBOperation) (string, error) {
				return "", expectedErr
			},
		}

		hash, err := c.ExecuteTransfer(context.Background(), []byte{}, argLists, batch.ID, 9)
		assert.Equal(t, "", hash)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work - same number of signatures as quorum", func(t *testing.T) {
		args.Proxy = &sui.Client{}
		c, err := NewSuiClient(args)
		assert.NoError(t, err)

		values, err := createResultsRawFromValues(false)
		assert.NoError(t, err)
		wasCalled := false

		c.signatureHolder = &testsCommon.SignaturesHolderStub{
			SignaturesCalled: func(messageHash []byte) [][]byte {
				return signatures[:9]
			},
		}

		c.suiClientDataGetter.proxy = &interactors.SuiProxyStub{
			SuiDevInspectTransactionBlockCalled: func(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
				return models.SuiTransactionBlockResponse{
					Effects: models.SuiEffects{
						Status: models.ExecutionStatus{
							Status: "success",
						},
					},
					Results: values,
				}, nil
			},
			SuiXGetCoinsCalled: func(ctx context.Context, req models.SuiXGetCoinsRequest) (models.PaginatedCoinsResponse, error) {
				return models.PaginatedCoinsResponse{
					Data: []models.CoinData{
						{
							CoinObjectId:        "0x0e82989b575d6ebb4e7f6062f5463bdd1cce9db77b9f34ac923b90031765fddf",
							Version:             "1",
							Digest:              "4Nd1mHZtwVaFgqsSVBz3tH6KvXZcG1oMqezLh8u6BbhE",
							Balance:             "100000",
							CoinType:            "0x2::sui::SUI",
							PreviousTransaction: "9rS8PZyT1QK5qLgN",
						},
					},
					NextCursor:  "",
					HasNextPage: false,
				}, nil
			},
		}
		c.txHandler = &bridgeTests.SuiTxHandlerStub{
			SendTransactionReturnHashCalled: func(ctx context.Context, gasCoin *transaction.SuiObjectRef, calls []core.SuiPTBOperation) (string, error) {
				wasCalled = true
				return "0xc5b2c658f5fa236c598a6e7fbf7f21413dc42e2a41dd982eb772b30707cba2eb", nil
			},
		}

		hash, err := c.ExecuteTransfer(context.Background(), []byte{}, argLists, batch.ID, 9)
		assert.Equal(t, "0xc5b2c658f5fa236c598a6e7fbf7f21413dc42e2a41dd982eb772b30707cba2eb", hash)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
	t.Run("should work - more signatures should trim", func(t *testing.T) {
		args.Proxy = &sui.Client{}
		c, err := NewSuiClient(args)
		assert.NoError(t, err)

		values, err := createResultsRawFromValues(false)
		assert.NoError(t, err)
		wasCalled := false

		c.signatureHolder = &testsCommon.SignaturesHolderStub{
			SignaturesCalled: func(messageHash []byte) [][]byte {
				return signatures[:9]
			},
		}

		c.suiClientDataGetter.proxy = &interactors.SuiProxyStub{
			SuiDevInspectTransactionBlockCalled: func(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
				return models.SuiTransactionBlockResponse{
					Effects: models.SuiEffects{
						Status: models.ExecutionStatus{
							Status: "success",
						},
					},
					Results: values,
				}, nil
			},
			SuiXGetCoinsCalled: func(ctx context.Context, req models.SuiXGetCoinsRequest) (models.PaginatedCoinsResponse, error) {
				return models.PaginatedCoinsResponse{
					Data: []models.CoinData{
						{
							CoinObjectId:        "0x0e82989b575d6ebb4e7f6062f5463bdd1cce9db77b9f34ac923b90031765fddf",
							Version:             "1",
							Digest:              "4Nd1mHZtwVaFgqsSVBz3tH6KvXZcG1oMqezLh8u6BbhE",
							Balance:             "100000",
							CoinType:            "0x2::sui::SUI",
							PreviousTransaction: "9rS8PZyT1QK5qLgN",
						},
					},
					NextCursor:  "",
					HasNextPage: false,
				}, nil
			},
		}
		c.txHandler = &bridgeTests.SuiTxHandlerStub{
			SendTransactionReturnHashCalled: func(ctx context.Context, gasCoin *transaction.SuiObjectRef, calls []core.SuiPTBOperation) (string, error) {
				wasCalled = true
				return "0xc5b2c658f5fa236c598a6e7fbf7f21413dc42e2a41dd982eb772b30707cba2eb", nil
			},
		}

		hash, err := c.ExecuteTransfer(context.Background(), []byte{}, argLists, batch.ID, 5)
		assert.Equal(t, "0xc5b2c658f5fa236c598a6e7fbf7f21413dc42e2a41dd982eb772b30707cba2eb", hash)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
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
		c.suiClientDataGetter.proxy = &interactors.SuiProxyStub{
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
