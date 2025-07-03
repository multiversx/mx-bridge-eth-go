package sui

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/converters"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	bridgeTests "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
)

func createMockSuiClientArgs() ArgsSuiClient {
	addressConverter, err := converters.NewAddressConverter()
	if err != nil {
		panic(err)
	}

	return ArgsSuiClient{
		ClientWrapper:    &bridgeTests.SuiClientWrapperStub{},
		Log:              logger.GetOrCreate("test"),
		AddressConverter: addressConverter,
		Broadcaster:      &testsCommon.BroadcasterStub{},
		TokensMapper: &bridgeTests.TokensMapperStub{
			ConvertTokenCalled: func(ctx context.Context, sourceBytes []byte) ([]byte, error) {
				return append([]byte("SUI"), sourceBytes...), nil
			},
		},
		CryptoHandler:                &bridgeTests.SuiCryptoHandlerStub{},
		SignatureHolder:              &testsCommon.SignaturesHolderStub{},
		SafeContractId:               testsCommon.CreateRandomSuiAddress(),
		TransferGasLimitBase:         50,
		TransferGasLimitForEach:      20,
		ClientAvailabilityAllowDelta: 5,
		EventsBlockRangeFrom:         -100,
		EventsBlockRangeTo:           400,
	}
}

func TestNewSuiClient(t *testing.T) {
	t.Parallel()

	t.Run("nil client wrapper", func(t *testing.T) {
		args := createMockSuiClientArgs()
		args.ClientWrapper = nil
		c, err := NewSuiClient(args)

		assert.Equal(t, clients.ErrNilClientWrapper, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("nil logger", func(t *testing.T) {
		args := createMockSuiClientArgs()
		args.Log = nil
		c, err := NewSuiClient(args)

		assert.Equal(t, clients.ErrNilLogger, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("nil address converter", func(t *testing.T) {
		args := createMockSuiClientArgs()
		args.AddressConverter = nil
		c, err := NewSuiClient(args)

		assert.Equal(t, clients.ErrNilAddressConverter, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("nil broadcaster", func(t *testing.T) {
		args := createMockSuiClientArgs()
		args.Broadcaster = nil
		c, err := NewSuiClient(args)

		assert.Equal(t, clients.ErrNilBroadcaster, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("nil crypto handler", func(t *testing.T) {
		args := createMockSuiClientArgs()
		args.CryptoHandler = nil
		c, err := NewSuiClient(args)

		assert.Equal(t, clients.ErrNilCryptoHandler, err)
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
	t.Run("0 transfer gas limit base", func(t *testing.T) {
		args := createMockSuiClientArgs()
		args.TransferGasLimitBase = 0
		c, err := NewSuiClient(args)

		assert.Equal(t, clients.ErrInvalidGasLimit, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("0 transfer gas limit for each", func(t *testing.T) {
		args := createMockSuiClientArgs()
		args.TransferGasLimitForEach = 0
		c, err := NewSuiClient(args)

		assert.Equal(t, clients.ErrInvalidGasLimit, err)
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
	t.Run("invalid events block range from should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSuiClientArgs()
		args.EventsBlockRangeFrom = 100
		args.EventsBlockRangeTo = 50

		c, err := NewSuiClient(args)

		assert.True(t, check.IfNil(c))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "args.EventsBlockRangeFrom"))
		assert.True(t, strings.Contains(err.Error(), "args.EventsBlockRangeTo"))
	})
	t.Run("nil crypto handler", func(t *testing.T) {
		args := createMockSuiClientArgs()
		args.CryptoHandler = nil
		c, err := NewSuiClient(args)

		assert.Equal(t, clients.ErrNilCryptoHandler, err)
		assert.True(t, check.IfNil(c))
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

	args := createMockSuiClientArgs()
	c, _ := NewSuiClient(args)
	expectedErr := errors.New("expected error")

	t.Run("error while getting batch", func(t *testing.T) {
		c.clientWrapper = &bridgeTests.SuiClientWrapperStub{
			GetBatchCalled: func(ctx context.Context, batchNonce *big.Int) (core.Batch, bool, error) {
				return core.Batch{}, false, expectedErr
			},
		}
		batch, isFinal, err := c.GetBatch(context.Background(), 1)
		assert.Nil(t, batch)
		assert.Equal(t, expectedErr, err)
		assert.False(t, isFinal)
	})
	t.Run("error while getting deposits", func(t *testing.T) {
		c.clientWrapper = &bridgeTests.SuiClientWrapperStub{
			GetBatchCalled: func(ctx context.Context, batchNonce *big.Int) (core.Batch, bool, error) {
				return core.Batch{
					Nonce:         batchNonce,
					DepositsCount: 3,
				}, true, nil
			},
			GetBatchDepositsCalled: func(ctx context.Context, batchNonce *big.Int) ([]core.Deposit, bool, error) {
				return nil, false, expectedErr
			},
		}
		batch, isFinal, err := c.GetBatch(context.Background(), 1)
		assert.Nil(t, batch)
		assert.Equal(t, expectedErr, err)
		assert.False(t, isFinal)
	})
	t.Run("deposits mismatch - with 0", func(t *testing.T) {
		c.clientWrapper = &bridgeTests.SuiClientWrapperStub{
			GetBatchCalled: func(ctx context.Context, batchNonce *big.Int) (core.Batch, bool, error) {
				return core.Batch{
					Nonce:         batchNonce,
					DepositsCount: 3,
				}, true, nil
			},
			GetBatchDepositsCalled: func(ctx context.Context, batchNonce *big.Int) ([]core.Deposit, bool, error) {
				return make([]core.Deposit, 0), true, nil
			},
		}
		batch, isFinal, err := c.GetBatch(context.Background(), 1)
		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, clients.ErrDepositsAndBatchDepositsCountDiffer))
		assert.True(t, strings.Contains(err.Error(), "batch.DepositsCount: 3, fetched deposits len: 0"))
		assert.False(t, isFinal)
	})
	t.Run("deposits mismatch - with non zero value", func(t *testing.T) {
		c.clientWrapper = &bridgeTests.SuiClientWrapperStub{
			GetBatchCalled: func(ctx context.Context, batchNonce *big.Int) (core.Batch, bool, error) {
				return core.Batch{
					Nonce:         batchNonce,
					DepositsCount: 3,
				}, true, nil
			},
			GetBatchDepositsCalled: func(ctx context.Context, batchNonce *big.Int) ([]core.Deposit, bool, error) {
				return make([]core.Deposit, 4), true, nil
			},
		}
		batch, isFinal, err := c.GetBatch(context.Background(), 1)
		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, clients.ErrDepositsAndBatchDepositsCountDiffer))
		assert.True(t, strings.Contains(err.Error(), "batch.DepositsCount: 3, fetched deposits len: 4"))
		assert.False(t, isFinal)
	})
	t.Run("returns batch should work", func(t *testing.T) {
		from1 := testsCommon.CreateRandomSuiAddressBytes()
		token1 := testsCommon.CreateRandomCoinId()
		recipient1 := testsCommon.CreateRandomMultiversXAddress()

		from2 := testsCommon.CreateRandomSuiAddressBytes()
		token2 := testsCommon.CreateRandomCoinId()
		recipient2 := testsCommon.CreateRandomMultiversXAddress()

		c.clientWrapper = &bridgeTests.SuiClientWrapperStub{
			GetBatchCalled: func(ctx context.Context, batchNonce *big.Int) (core.Batch, bool, error) {
				return core.Batch{
					Nonce:                  big.NewInt(537),
					BlockNumber:            0,
					LastUpdatedBlockNumber: 0,
					DepositsCount:          2,
				}, true, nil
			},
			GetBatchDepositsCalled: func(ctx context.Context, batchNonce *big.Int) ([]core.Deposit, bool, error) {
				return []core.Deposit{
					{
						Nonce:        big.NewInt(10),
						TokenAddress: token1,
						Amount:       big.NewInt(20),
						Depositor:    from1,
						Recipient:    recipient1.AddressSlice(),
					},
					{
						Nonce:        big.NewInt(30),
						TokenAddress: token2,
						Amount:       big.NewInt(40),
						Depositor:    from2,
						Recipient:    recipient2.AddressSlice(),
					},
				}, true, nil
			},
		}

		bech32Recipient1Address, _ := recipient1.AddressAsBech32String()
		bech32Recipient2Address, _ := recipient2.AddressAsBech32String()
		expectedBatch := &core.TransferBatch{
			ID: 537,
			Deposits: []*core.DepositTransfer{
				{
					Nonce:                 10,
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
					Nonce:                 30,
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

		batch, isFinal, err := c.GetBatch(context.Background(), 1)
		assert.Equal(t, expectedBatch, batch)
		assert.Nil(t, err)
		assert.True(t, isFinal)
	})
	t.Run("returns non final batch should work", func(t *testing.T) {
		from1 := testsCommon.CreateRandomSuiAddressBytes()
		token1 := testsCommon.CreateRandomCoinId()
		recipient1 := testsCommon.CreateRandomMultiversXAddress()

		from2 := testsCommon.CreateRandomSuiAddressBytes()
		token2 := testsCommon.CreateRandomCoinId()
		recipient2 := testsCommon.CreateRandomMultiversXAddress()

		c.clientWrapper = &bridgeTests.SuiClientWrapperStub{
			GetBatchCalled: func(ctx context.Context, batchNonce *big.Int) (core.Batch, bool, error) {
				return core.Batch{
					Nonce:                  big.NewInt(98765),
					BlockNumber:            0,
					LastUpdatedBlockNumber: 0,
					DepositsCount:          2,
				}, false, nil
			},
			GetBatchDepositsCalled: func(ctx context.Context, batchNonce *big.Int) ([]core.Deposit, bool, error) {
				return []core.Deposit{
					{
						Nonce:        big.NewInt(10),
						TokenAddress: token1,
						Amount:       big.NewInt(20),
						Depositor:    from1,
						Recipient:    recipient1.AddressSlice(),
					},
					{
						Nonce:        big.NewInt(30),
						TokenAddress: token2,
						Amount:       big.NewInt(40),
						Depositor:    from2,
						Recipient:    recipient2.AddressSlice(),
					},
				}, false, nil
			},
		}

		bech32Recipient1Address, _ := recipient1.AddressAsBech32String()
		bech32Recipient2Address, _ := recipient2.AddressAsBech32String()
		expectedBatch := &core.TransferBatch{
			ID: 98765,
			Deposits: []*core.DepositTransfer{
				{
					Nonce:                 10,
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
					Nonce:                 30,
					ToBytes:               recipient2.AddressBytes(),
					DisplayableTo:         bech32Recipient2Address,
					FromBytes:             from2[:],
					DisplayableFrom:       hex.EncodeToString(from2[:]),
					SourceTokenBytes:      []byte(token2),
					DisplayableToken:      token2[:],
					Amount:                big.NewInt(40),
					DestinationTokenBytes: append([]byte("SUI"), token2[:]...),
				},
			},
			Statuses: make([]byte, 2),
		}

		batch, isFinal, err := c.GetBatch(context.Background(), 1)
		assert.Equal(t, expectedBatch, batch)
		assert.Nil(t, err)
		assert.False(t, isFinal)
	})
}

func TestClient_BroadcastSignatureForMessageHash(t *testing.T) {
	t.Parallel()

	t.Run("sign failed should not broadcast", func(t *testing.T) {
		t.Parallel()

		expectedError := errors.New("expected error")
		msgToSign := []byte("message to sign")
		args := createMockSuiClientArgs()
		args.Broadcaster = &testsCommon.BroadcasterStub{
			BroadcastSignatureCalled: func(signature []byte, messageHash []byte) {
				assert.Fail(t, "should have not called broadcast")
			},
		}
		args.CryptoHandler = &bridgeTests.SuiCryptoHandlerStub{
			SignCalled: func(msg []byte) ([]byte, error) {
				assert.Equal(t, msgToSign, msg)
				return nil, expectedError
			},
		}

		c, _ := NewSuiClient(args)
		c.BroadcastSignatureForMessageHash(msgToSign)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedSig := "expected sig"
		broadcastCalled := false

		msgToSign := []byte("message to sign")
		args := createMockSuiClientArgs()
		args.Broadcaster = &testsCommon.BroadcasterStub{
			BroadcastSignatureCalled: func(signature []byte, message []byte) {
				assert.Equal(t, msgToSign, message)
				assert.Equal(t, expectedSig, string(signature))
				broadcastCalled = true
			},
		}
		args.CryptoHandler = &bridgeTests.SuiCryptoHandlerStub{
			SignCalled: func(msg []byte) ([]byte, error) {
				assert.Equal(t, msgToSign, msg)
				return []byte(expectedSig), nil
			},
		}

		c, _ := NewSuiClient(args)
		c.BroadcastSignatureForMessageHash(msgToSign)

		assert.True(t, broadcastCalled)
	})
}

func TestClient_WasExecuted(t *testing.T) {
	t.Parallel()

	wasCalled := false
	args := createMockSuiClientArgs()
	args.ClientWrapper = &bridgeTests.SuiClientWrapperStub{
		WasBatchExecutedCalled: func(ctx context.Context, batchNonce *big.Int) (bool, error) {
			wasCalled = true
			return true, nil
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
	coinType := "0x0::coin"
	balance := big.NewInt(1000000)

	t.Run("not enough coins", func(t *testing.T) {
		c, _ := NewSuiClient(args)
		c.clientWrapper = &bridgeTests.SuiClientWrapperStub{
			GetBalanceCalled: func(ctx context.Context, account string, coinType string) (models.CoinBalanceResponse, error) {
				assert.Equal(t, c.safeContractId, account)

				return models.CoinBalanceResponse{
					TotalBalance: balance.String(),
				}, nil
			},
		}
		err := c.CheckRequiredBalance(context.Background(), coinType, big.NewInt(0).Add(balance, big.NewInt(1)))
		assert.True(t, errors.Is(err, errInsufficientCoinBalance))
	})
	t.Run("erc20 balance of errors", func(t *testing.T) {
		expectedErr := errors.New("expected error erc20 balance of")
		c, _ := NewSuiClient(args)
		c.clientWrapper = &bridgeTests.SuiClientWrapperStub{
			GetBalanceCalled: func(ctx context.Context, account string, coinType string) (models.CoinBalanceResponse, error) {
				return models.CoinBalanceResponse{}, expectedErr
			},
		}
		err := c.CheckRequiredBalance(context.Background(), coinType, balance)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("should work", func(t *testing.T) {
		c, _ := NewSuiClient(args)
		c.clientWrapper = &bridgeTests.SuiClientWrapperStub{
			GetBalanceCalled: func(ctx context.Context, account string, coinType string) (models.CoinBalanceResponse, error) {
				return models.CoinBalanceResponse{
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

	t.Run("error while getting total balances", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockSuiClientArgs()
		args.ClientWrapper = &bridgeTests.SuiClientWrapperStub{
			TotalBalancesCalled: func(ctx context.Context, token string) (*big.Int, error) {
				return nil, expectedErr
			},
		}
		c, _ := NewSuiClient(args)

		balances, err := c.TotalBalances(context.Background(), "wrong coin type")
		assert.Nil(t, balances)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedBalance := big.NewInt(100)
		args := createMockSuiClientArgs()
		args.ClientWrapper = &bridgeTests.SuiClientWrapperStub{
			TotalBalancesCalled: func(ctx context.Context, token string) (*big.Int, error) {
				return providedBalance, nil
			},
		}
		c, _ := NewSuiClient(args)

		balances, err := c.TotalBalances(context.Background(), "0x0::coin::Coin")
		assert.Nil(t, err)
		assert.Equal(t, providedBalance, balances)
	})
}

func TestClient_MintBalances(t *testing.T) {
	t.Parallel()

	t.Run("error while getting mint balances", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockSuiClientArgs()
		args.ClientWrapper = &bridgeTests.SuiClientWrapperStub{
			MintBalancesCalled: func(ctx context.Context, token string) (*big.Int, error) {
				return nil, expectedErr
			},
		}
		c, _ := NewSuiClient(args)

		balances, err := c.MintBalances(context.Background(), "wrong coin type")
		assert.Nil(t, balances)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedBalance := big.NewInt(100)
		args := createMockSuiClientArgs()
		args.ClientWrapper = &bridgeTests.SuiClientWrapperStub{
			MintBalancesCalled: func(ctx context.Context, token string) (*big.Int, error) {
				return providedBalance, nil
			},
		}
		c, _ := NewSuiClient(args)

		balances, err := c.MintBalances(context.Background(), "0x0::coin::Coin")
		assert.Nil(t, err)
		assert.Equal(t, providedBalance, balances)
	})
}

func TestClient_BurnBalances(t *testing.T) {
	t.Parallel()

	t.Run("error while getting burn balances", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockSuiClientArgs()
		args.ClientWrapper = &bridgeTests.SuiClientWrapperStub{
			BurnBalancesCalled: func(ctx context.Context, token string) (*big.Int, error) {
				return nil, expectedErr
			},
		}
		c, _ := NewSuiClient(args)

		balances, err := c.BurnBalances(context.Background(), "wrong coin type")
		assert.Nil(t, balances)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedBalance := big.NewInt(100)
		args := createMockSuiClientArgs()
		args.ClientWrapper = &bridgeTests.SuiClientWrapperStub{
			BurnBalancesCalled: func(ctx context.Context, token string) (*big.Int, error) {
				return providedBalance, nil
			},
		}
		c, _ := NewSuiClient(args)

		balances, err := c.BurnBalances(context.Background(), "0x0::coin::Coin")
		assert.Nil(t, err)
		assert.Equal(t, providedBalance, balances)
	})
}

func TestClient_MintBurnTokens(t *testing.T) {
	t.Parallel()

	t.Run("error while getting mint burn tokens", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockSuiClientArgs()
		args.ClientWrapper = &bridgeTests.SuiClientWrapperStub{
			MintBurnTokensCalled: func(ctx context.Context, token string) (bool, error) {
				return false, expectedErr
			},
		}
		c, _ := NewSuiClient(args)

		isMintBurn, err := c.MintBurnTokens(context.Background(), "wrong coin type")
		assert.False(t, isMintBurn)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockSuiClientArgs()
		args.ClientWrapper = &bridgeTests.SuiClientWrapperStub{
			MintBurnTokensCalled: func(ctx context.Context, token string) (bool, error) {
				return true, nil
			},
		}
		c, _ := NewSuiClient(args)

		isMintBurn, err := c.MintBurnTokens(context.Background(), "0x0::coin::Coin")
		assert.Nil(t, err)
		assert.True(t, isMintBurn)
	})
}

func TestClient_NativeTokens(t *testing.T) {
	t.Parallel()

	t.Run("error while getting native tokens", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockSuiClientArgs()
		args.ClientWrapper = &bridgeTests.SuiClientWrapperStub{
			NativeTokensCalled: func(ctx context.Context, token string) (bool, error) {
				return false, expectedErr
			},
		}
		c, _ := NewSuiClient(args)

		isNative, err := c.NativeTokens(context.Background(), "wrong coin type")
		assert.False(t, isNative)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockSuiClientArgs()
		args.ClientWrapper = &bridgeTests.SuiClientWrapperStub{
			NativeTokensCalled: func(ctx context.Context, token string) (bool, error) {
				return true, nil
			},
		}
		c, _ := NewSuiClient(args)

		isNative, err := c.NativeTokens(context.Background(), "0x0::coin::Coin")
		assert.Nil(t, err)
		assert.True(t, isNative)
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
		args.ClientWrapper = &bridgeTests.SuiClientWrapperStub{
			GetStatusesAfterExecutionCalled: func(ctx context.Context, batchID *big.Int) ([]byte, bool, error) {
				assert.Equal(t, expectedBatchID, batchID)
				return nil, false, expectedErr
			},
		}

		c, _ := NewSuiClient(args)
		statuses, err := c.GetTransactionsStatuses(context.Background(), expectedBatchID.Uint64())
		assert.Nil(t, statuses)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("statuses are not final, should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSuiClientArgs()
		args.ClientWrapper = &bridgeTests.SuiClientWrapperStub{
			GetStatusesAfterExecutionCalled: func(ctx context.Context, batchID *big.Int) ([]byte, bool, error) {
				assert.Equal(t, expectedBatchID, batchID)
				return []byte("dummy"), false, nil
			},
		}

		c, _ := NewSuiClient(args)
		statuses, err := c.GetTransactionsStatuses(context.Background(), expectedBatchID.Uint64())
		assert.Nil(t, statuses)
		assert.Equal(t, clients.ErrStatusIsNotFinal, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockSuiClientArgs()
		args.ClientWrapper = &bridgeTests.SuiClientWrapperStub{
			GetStatusesAfterExecutionCalled: func(ctx context.Context, batchID *big.Int) ([]byte, bool, error) {
				assert.Equal(t, expectedBatchID, batchID)
				return expectedStatuses, true, nil
			},
		}

		c, _ := NewSuiClient(args)
		statuses, err := c.GetTransactionsStatuses(context.Background(), expectedBatchID.Uint64())
		assert.Equal(t, expectedStatuses, statuses)
		assert.Nil(t, err)
	})
}

func TestClient_GetQuorumSize(t *testing.T) {
	t.Parallel()

	args := createMockSuiClientArgs()
	providedValue := big.NewInt(6453)
	args.ClientWrapper = &bridgeTests.SuiClientWrapperStub{
		QuorumCalled: func(ctx context.Context) (*big.Int, error) {
			return providedValue, nil
		},
	}
	c, _ := NewSuiClient(args)

	quorum, err := c.GetQuorumSize(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, providedValue, quorum)
}

func TestClient_IsQuorumReached(t *testing.T) {
	t.Parallel()
	msg := []byte("message")

	t.Run("quorum errors", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockSuiClientArgs()
		args.ClientWrapper = &bridgeTests.SuiClientWrapperStub{
			QuorumCalled: func(ctx context.Context) (*big.Int, error) {
				return nil, expectedErr
			},
		}
		c, _ := NewSuiClient(args)

		isReached, err := c.IsQuorumReached(context.Background(), msg)
		assert.False(t, isReached)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("quorum returns less than minimum allowed", func(t *testing.T) {
		t.Parallel()

		args := createMockSuiClientArgs()
		args.ClientWrapper = &bridgeTests.SuiClientWrapperStub{
			QuorumCalled: func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(0), nil
			},
		}
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
		args.ClientWrapper = &bridgeTests.SuiClientWrapperStub{
			QuorumCalled: func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(3), nil
			},
		}
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
	args.ClientWrapper = &bridgeTests.SuiClientWrapperStub{
		StatusHandler: statusHandler,
		GetLatestCheckpointCalled: func(ctx context.Context) (uint64, error) {
			currentCheckpoint += incrementor
			return currentCheckpoint, nil
		},
	}
	expectedErr := errors.New("expected error")

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
		c.clientWrapper = &bridgeTests.SuiClientWrapperStub{
			StatusHandler: statusHandler,
			GetLatestCheckpointCalled: func(ctx context.Context) (uint64, error) {
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
	c.clientWrapper.SetStringMetric(core.MetricMultiversXClientStatus, "")
	c.clientWrapper.SetStringMetric(core.MetricLastMultiversXClientError, "")
}

func checkStatusHandler(t *testing.T, statusHandler *testsCommon.StatusHandlerMock, status core.ClientStatus, message string) {
	assert.Equal(t, status.String(), statusHandler.GetStringMetric(core.MetricMultiversXClientStatus))
	assert.Equal(t, message, statusHandler.GetStringMetric(core.MetricLastMultiversXClientError))
}
