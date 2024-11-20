package ethereum

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	"github.com/multiversx/mx-bridge-eth-go/core/converters"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	bridgeTests "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedMvxTransactions = []contract.MvxTransaction{
	{
		Token:        common.BytesToAddress([]byte("ERC20token1")),
		Sender:       [32]byte(bytes.Repeat([]byte("1"), 32)),
		Recipient:    common.BytesToAddress([]byte("to1")),
		Amount:       big.NewInt(20),
		DepositNonce: big.NewInt(10),
		CallData:     []byte("data1"),
	},
	{
		Token:        common.BytesToAddress([]byte("ERC20token2")),
		Sender:       [32]byte(bytes.Repeat([]byte("2"), 32)),
		Recipient:    common.BytesToAddress([]byte("to2")),
		Amount:       big.NewInt(40),
		DepositNonce: big.NewInt(30),
		CallData:     []byte("data2"),
	},
}

func createMockEthereumClientArgs() ArgsEthereumClient {
	addressConverter, err := converters.NewAddressConverter()
	if err != nil {
		panic(err)
	}

	return ArgsEthereumClient{
		ClientWrapper:         &bridgeTests.EthereumClientWrapperStub{},
		Erc20ContractsHandler: &bridgeTests.ERC20ContractsHolderStub{},
		Log:                   logger.GetOrCreate("test"),
		AddressConverter:      addressConverter,
		Broadcaster:           &testsCommon.BroadcasterStub{},
		CryptoHandler:         &bridgeTests.CryptoHandlerStub{},
		TokensMapper: &bridgeTests.TokensMapperStub{
			ConvertTokenCalled: func(ctx context.Context, sourceBytes []byte) ([]byte, error) {
				return append([]byte("ERC20"), sourceBytes...), nil
			},
		},
		SignatureHolder:              &testsCommon.SignaturesHolderStub{},
		SafeContractAddress:          testsCommon.CreateRandomEthereumAddress(),
		GasHandler:                   &testsCommon.GasHandlerStub{},
		TransferGasLimitBase:         50,
		TransferGasLimitForEach:      20,
		ClientAvailabilityAllowDelta: 5,
		EventsBlockRangeFrom:         -100,
		EventsBlockRangeTo:           400,
	}
}

func createMockTransferBatch() *bridgeCore.TransferBatch {
	return &bridgeCore.TransferBatch{
		ID: 332,
		Deposits: []*bridgeCore.DepositTransfer{
			{
				Nonce:                 10,
				ToBytes:               []byte("to1"),
				DisplayableTo:         "to1",
				FromBytes:             bytes.Repeat([]byte("1"), 32),
				DisplayableFrom:       strings.Repeat("1", 32),
				SourceTokenBytes:      []byte("source token1"),
				DisplayableToken:      "token1",
				Amount:                big.NewInt(20),
				DestinationTokenBytes: []byte("ERC20token1"),
				Data:                  []byte("data1"),
			},
			{
				Nonce:                 30,
				ToBytes:               []byte("to2"),
				DisplayableTo:         "to2",
				FromBytes:             bytes.Repeat([]byte("2"), 32),
				DisplayableFrom:       strings.Repeat("2", 32),
				SourceTokenBytes:      []byte("source token2"),
				DisplayableToken:      "token2",
				Amount:                big.NewInt(40),
				DestinationTokenBytes: []byte("ERC20token2"),
				Data:                  []byte("data2"),
			},
		},
		Statuses: make([]byte, 2),
	}
}

func TestNewEthereumClient(t *testing.T) {
	t.Parallel()

	t.Run("nil client wrapper", func(t *testing.T) {
		args := createMockEthereumClientArgs()
		args.ClientWrapper = nil
		c, err := NewEthereumClient(args)

		assert.Equal(t, errNilClientWrapper, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("nil erc20 contracts handler", func(t *testing.T) {
		args := createMockEthereumClientArgs()
		args.Erc20ContractsHandler = nil
		c, err := NewEthereumClient(args)

		assert.Equal(t, errNilERC20ContractsHandler, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("nil logger", func(t *testing.T) {
		args := createMockEthereumClientArgs()
		args.Log = nil
		c, err := NewEthereumClient(args)

		assert.Equal(t, clients.ErrNilLogger, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("nil address converter", func(t *testing.T) {
		args := createMockEthereumClientArgs()
		args.AddressConverter = nil
		c, err := NewEthereumClient(args)

		assert.Equal(t, clients.ErrNilAddressConverter, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("nil broadcaster", func(t *testing.T) {
		args := createMockEthereumClientArgs()
		args.Broadcaster = nil
		c, err := NewEthereumClient(args)

		assert.Equal(t, errNilBroadcaster, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("nil crypto handler", func(t *testing.T) {
		args := createMockEthereumClientArgs()
		args.CryptoHandler = nil
		c, err := NewEthereumClient(args)

		assert.Equal(t, clients.ErrNilCryptoHandler, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("nil tokens mapper", func(t *testing.T) {
		args := createMockEthereumClientArgs()
		args.TokensMapper = nil
		c, err := NewEthereumClient(args)

		assert.Equal(t, clients.ErrNilTokensMapper, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("nil signature holder", func(t *testing.T) {
		args := createMockEthereumClientArgs()
		args.SignatureHolder = nil
		c, err := NewEthereumClient(args)

		assert.Equal(t, errNilSignaturesHolder, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("nil gas handler", func(t *testing.T) {
		args := createMockEthereumClientArgs()
		args.GasHandler = nil
		c, err := NewEthereumClient(args)

		assert.Equal(t, errNilGasHandler, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("0 transfer gas limit base", func(t *testing.T) {
		args := createMockEthereumClientArgs()
		args.TransferGasLimitBase = 0
		c, err := NewEthereumClient(args)

		assert.Equal(t, errInvalidGasLimit, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("0 transfer gas limit for each", func(t *testing.T) {
		args := createMockEthereumClientArgs()
		args.TransferGasLimitForEach = 0
		c, err := NewEthereumClient(args)

		assert.Equal(t, errInvalidGasLimit, err)
		assert.True(t, check.IfNil(c))
	})
	t.Run("invalid ClientAvailabilityAllowDelta should error", func(t *testing.T) {
		t.Parallel()

		args := createMockEthereumClientArgs()
		args.ClientAvailabilityAllowDelta = 0

		c, err := NewEthereumClient(args)

		assert.True(t, check.IfNil(c))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for args.AllowedDelta"))
	})
	t.Run("invalid events block range from should error", func(t *testing.T) {
		t.Parallel()

		args := createMockEthereumClientArgs()
		args.EventsBlockRangeFrom = 100
		args.EventsBlockRangeTo = 50

		c, err := NewEthereumClient(args)

		assert.True(t, check.IfNil(c))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "args.EventsBlockRangeFrom"))
		assert.True(t, strings.Contains(err.Error(), "args.EventsBlockRangeTo"))
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockEthereumClientArgs()
		c, err := NewEthereumClient(args)

		assert.Nil(t, err)
		assert.False(t, check.IfNil(c))
	})
}

func TestClient_GetBatch(t *testing.T) {
	t.Parallel()

	args := createMockEthereumClientArgs()
	c, _ := NewEthereumClient(args)
	expectedErr := errors.New("expected error")

	t.Run("error while getting batch", func(t *testing.T) {
		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			GetBatchCalled: func(ctx context.Context, batchNonce *big.Int) (contract.Batch, bool, error) {
				return contract.Batch{}, false, expectedErr
			},
		}
		batch, isFinal, err := c.GetBatch(context.Background(), 1)
		assert.Nil(t, batch)
		assert.Equal(t, expectedErr, err)
		assert.False(t, isFinal)
	})
	t.Run("error while getting deposits", func(t *testing.T) {
		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			GetBatchCalled: func(ctx context.Context, batchNonce *big.Int) (contract.Batch, bool, error) {
				return contract.Batch{
					Nonce:         batchNonce,
					DepositsCount: 2,
				}, true, nil
			},
			GetBatchDepositsCalled: func(ctx context.Context, batchNonce *big.Int) ([]contract.Deposit, bool, error) {
				return nil, false, expectedErr
			},
		}
		batch, isFinal, err := c.GetBatch(context.Background(), 1)
		assert.Nil(t, batch)
		assert.Equal(t, expectedErr, err)
		assert.False(t, isFinal)
	})
	t.Run("deposits mismatch - with 0", func(t *testing.T) {
		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			GetBatchCalled: func(ctx context.Context, batchNonce *big.Int) (contract.Batch, bool, error) {
				return contract.Batch{
					Nonce:         batchNonce,
					DepositsCount: 2,
				}, true, nil
			},
			GetBatchDepositsCalled: func(ctx context.Context, batchNonce *big.Int) ([]contract.Deposit, bool, error) {
				return make([]contract.Deposit, 0), true, nil
			},
		}
		batch, isFinal, err := c.GetBatch(context.Background(), 1)
		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, errDepositsAndBatchDepositsCountDiffer))
		assert.True(t, strings.Contains(err.Error(), "batch.DepositsCount: 2, fetched deposits len: 0"))
		assert.False(t, isFinal)
	})
	t.Run("deposits mismatch - with non zero value", func(t *testing.T) {
		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			GetBatchCalled: func(ctx context.Context, batchNonce *big.Int) (contract.Batch, bool, error) {
				return contract.Batch{
					Nonce:         batchNonce,
					DepositsCount: 2,
				}, true, nil
			},
			GetBatchDepositsCalled: func(ctx context.Context, batchNonce *big.Int) ([]contract.Deposit, bool, error) {
				return []contract.Deposit{
					{
						Nonce: big.NewInt(22),
					},
				}, true, nil
			},
		}
		batch, isFinal, err := c.GetBatch(context.Background(), 1)
		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, errDepositsAndBatchDepositsCountDiffer))
		assert.True(t, strings.Contains(err.Error(), "batch.DepositsCount: 2, fetched deposits len: 1"))
		assert.False(t, isFinal)
	})
	t.Run("returns batch should work", func(t *testing.T) {
		from1 := testsCommon.CreateRandomEthereumAddress()
		token1 := testsCommon.CreateRandomEthereumAddress()
		recipient1 := testsCommon.CreateRandomMultiversXAddress()

		from2 := testsCommon.CreateRandomEthereumAddress()
		token2 := testsCommon.CreateRandomEthereumAddress()
		recipient2 := testsCommon.CreateRandomMultiversXAddress()

		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			GetBatchCalled: func(ctx context.Context, batchNonce *big.Int) (contract.Batch, bool, error) {
				return contract.Batch{
					Nonce:                  big.NewInt(112243),
					BlockNumber:            0,
					LastUpdatedBlockNumber: 0,
					DepositsCount:          2,
				}, true, nil
			},
			GetBatchDepositsCalled: func(ctx context.Context, batchNonce *big.Int) ([]contract.Deposit, bool, error) {
				return []contract.Deposit{
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
		expectedBatch := &bridgeCore.TransferBatch{
			ID: 112243,
			Deposits: []*bridgeCore.DepositTransfer{
				{
					Nonce:                 10,
					ToBytes:               recipient1.AddressBytes(),
					DisplayableTo:         bech32Recipient1Address,
					FromBytes:             from1[:],
					DisplayableFrom:       hex.EncodeToString(from1[:]),
					SourceTokenBytes:      token1[:],
					DisplayableToken:      hex.EncodeToString(token1[:]),
					Amount:                big.NewInt(20),
					DestinationTokenBytes: append([]byte("ERC20"), token1[:]...),
				},
				{
					Nonce:                 30,
					ToBytes:               recipient2.AddressBytes(),
					DisplayableTo:         bech32Recipient2Address,
					FromBytes:             from2[:],
					DisplayableFrom:       hex.EncodeToString(from2[:]),
					SourceTokenBytes:      token2[:],
					DisplayableToken:      hex.EncodeToString(token2[:]),
					Amount:                big.NewInt(40),
					DestinationTokenBytes: append([]byte("ERC20"), token2[:]...),
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
		from1 := testsCommon.CreateRandomEthereumAddress()
		token1 := testsCommon.CreateRandomEthereumAddress()
		recipient1 := testsCommon.CreateRandomMultiversXAddress()

		from2 := testsCommon.CreateRandomEthereumAddress()
		token2 := testsCommon.CreateRandomEthereumAddress()
		recipient2 := testsCommon.CreateRandomMultiversXAddress()

		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			GetBatchCalled: func(ctx context.Context, batchNonce *big.Int) (contract.Batch, bool, error) {
				return contract.Batch{
					Nonce:                  big.NewInt(112243),
					BlockNumber:            0,
					LastUpdatedBlockNumber: 0,
					DepositsCount:          2,
				}, false, nil
			},
			GetBatchDepositsCalled: func(ctx context.Context, batchNonce *big.Int) ([]contract.Deposit, bool, error) {
				return []contract.Deposit{
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
		expectedBatch := &bridgeCore.TransferBatch{
			ID: 112243,
			Deposits: []*bridgeCore.DepositTransfer{
				{
					Nonce:                 10,
					ToBytes:               recipient1.AddressBytes(),
					DisplayableTo:         bech32Recipient1Address,
					FromBytes:             from1[:],
					DisplayableFrom:       hex.EncodeToString(from1[:]),
					SourceTokenBytes:      token1[:],
					DisplayableToken:      hex.EncodeToString(token1[:]),
					Amount:                big.NewInt(20),
					DestinationTokenBytes: append([]byte("ERC20"), token1[:]...),
				},
				{
					Nonce:                 30,
					ToBytes:               recipient2.AddressBytes(),
					DisplayableTo:         bech32Recipient2Address,
					FromBytes:             from2[:],
					DisplayableFrom:       hex.EncodeToString(from2[:]),
					SourceTokenBytes:      token2[:],
					DisplayableToken:      hex.EncodeToString(token2[:]),
					Amount:                big.NewInt(40),
					DestinationTokenBytes: append([]byte("ERC20"), token2[:]...),
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

func TestClient_GenerateMessageHash(t *testing.T) {
	t.Parallel()

	args := createMockEthereumClientArgs()
	batch := createMockTransferBatch()

	t.Run("nil batch should error", func(t *testing.T) {
		c, _ := NewEthereumClient(args)
		h, err := c.GenerateMessageHash(nil, 0)

		assert.Equal(t, common.Hash{}, h)
		assert.True(t, errors.Is(err, clients.ErrNilBatch))
	})
	t.Run("should work", func(t *testing.T) {
		c, _ := NewEthereumClient(args)
		argLists, err := batchProcessor.ExtractListMvxToEth(batch)
		require.Nil(t, err)

		h, err := c.GenerateMessageHash(argLists, batch.ID)
		assert.Nil(t, err)
		assert.Equal(t, "c68190e0a3b8d7c6bd966272a11d618ceddc4b38662b0a1610621f4d30ec07ca", hex.EncodeToString(h.Bytes()))
	})
}

func TestClient_BroadcastSignatureForMessageHash(t *testing.T) {
	t.Parallel()

	t.Run("sign failed should not broadcast", func(t *testing.T) {
		t.Parallel()

		expectedError := errors.New("expected error")
		hash := common.HexToHash("hash")
		args := createMockEthereumClientArgs()
		args.Broadcaster = &testsCommon.BroadcasterStub{
			BroadcastSignatureCalled: func(signature []byte, messageHash []byte) {
				assert.Fail(t, "should have not called bradcast")
			},
		}
		args.CryptoHandler = &bridgeTests.CryptoHandlerStub{
			SignCalled: func(msgHash common.Hash) ([]byte, error) {
				assert.Equal(t, hash.Bytes(), msgHash.Bytes())
				return nil, expectedError
			},
		}

		c, _ := NewEthereumClient(args)
		c.BroadcastSignatureForMessageHash(hash)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedSig := "expected sig"
		broadcastCalled := false

		hash := common.HexToHash("hash")
		args := createMockEthereumClientArgs()
		args.Broadcaster = &testsCommon.BroadcasterStub{
			BroadcastSignatureCalled: func(signature []byte, messageHash []byte) {
				assert.Equal(t, hash.Bytes(), messageHash)
				assert.Equal(t, expectedSig, string(signature))
				broadcastCalled = true
			},
		}
		args.CryptoHandler = &bridgeTests.CryptoHandlerStub{
			SignCalled: func(msgHash common.Hash) ([]byte, error) {
				assert.Equal(t, hash.Bytes(), msgHash.Bytes())
				return []byte(expectedSig), nil
			},
		}

		c, _ := NewEthereumClient(args)
		c.BroadcastSignatureForMessageHash(hash)

		assert.True(t, broadcastCalled)
	})
}

func TestClient_WasExecuted(t *testing.T) {
	t.Parallel()

	wasCalled := false
	args := createMockEthereumClientArgs()
	args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
		WasBatchExecutedCalled: func(ctx context.Context, batchNonce *big.Int) (bool, error) {
			wasCalled = true
			return true, nil
		},
	}
	c, _ := NewEthereumClient(args)
	wasExecuted, err := c.WasExecuted(context.Background(), 1)

	assert.True(t, wasExecuted)
	assert.True(t, wasCalled)
	assert.Nil(t, err)
}

func TestClient_ExecuteTransfer(t *testing.T) {
	t.Parallel()

	args := createMockEthereumClientArgs()
	args.CryptoHandler = &bridgeTests.CryptoHandlerStub{
		CreateKeyedTransactorCalled: func(chainId *big.Int) (*bind.TransactOpts, error) {
			return &bind.TransactOpts{}, nil
		},
	}
	batch := createMockTransferBatch()
	argLists, errExtract := batchProcessor.ExtractListMvxToEth(batch)
	require.Nil(t, errExtract)

	signatures := make([][]byte, 10)
	for i := range signatures {
		signatures[i] = []byte(fmt.Sprintf("sig %d", i))
	}

	t.Run("nil batch", func(t *testing.T) {
		c, _ := NewEthereumClient(args)
		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, nil, 0, 10)
		assert.Equal(t, "", hash)
		assert.True(t, errors.Is(err, clients.ErrNilBatch))
	})
	t.Run("check if the contract is paused fails", func(t *testing.T) {
		expectedErr := errors.New("expected error is paused")
		c, _ := NewEthereumClient(args)
		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			IsPausedCalled: func(ctx context.Context) (bool, error) {
				return false, expectedErr
			},
		}
		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, argLists, batch.ID, 10)
		assert.Equal(t, "", hash)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("contract is paused should error", func(t *testing.T) {
		c, _ := NewEthereumClient(args)
		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			IsPausedCalled: func(ctx context.Context) (bool, error) {
				return true, nil
			},
		}
		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, argLists, batch.ID, 10)
		assert.Equal(t, "", hash)
		assert.True(t, errors.Is(err, clients.ErrMultisigContractPaused))
	})
	t.Run("get block number fails", func(t *testing.T) {
		expectedErr := errors.New("expected error get block number")
		c, _ := NewEthereumClient(args)
		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			BlockNumberCalled: func(ctx context.Context) (uint64, error) {
				return 0, expectedErr
			},
		}
		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, argLists, batch.ID, 10)
		assert.Equal(t, "", hash)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("get nonce fails", func(t *testing.T) {
		expectedErr := errors.New("expected error get nonce")
		c, _ := NewEthereumClient(args)
		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			NonceAtCalled: func(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
				return 0, expectedErr
			},
		}
		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, argLists, batch.ID, 10)
		assert.Equal(t, "", hash)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("get chain ID fails", func(t *testing.T) {
		expectedErr := errors.New("expected error get chain ID")
		c, _ := NewEthereumClient(args)
		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			ChainIDCalled: func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(0), expectedErr
			},
		}
		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, argLists, batch.ID, 10)
		assert.Equal(t, "", hash)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("create keyed transactor fails", func(t *testing.T) {
		expectedErr := errors.New("expected error create keyed transactor")
		c, _ := NewEthereumClient(args)
		c.cryptoHandler = &bridgeTests.CryptoHandlerStub{
			CreateKeyedTransactorCalled: func(chainId *big.Int) (*bind.TransactOpts, error) {
				return nil, expectedErr
			},
		}
		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, argLists, batch.ID, 10)
		assert.Equal(t, "", hash)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("get current gas price fails", func(t *testing.T) {
		expectedErr := errors.New("expected error get current gas price")
		c, _ := NewEthereumClient(args)
		c.gasHandler = &testsCommon.GasHandlerStub{
			GetCurrentGasPriceCalled: func() (*big.Int, error) {
				return nil, expectedErr
			},
		}
		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, argLists, batch.ID, 10)
		assert.Equal(t, "", hash)
		assert.ErrorIs(t, err, expectedErr)
	})
	t.Run("not enough quorum", func(t *testing.T) {
		c, _ := NewEthereumClient(args)
		c.signatureHolder = &testsCommon.SignaturesHolderStub{
			SignaturesCalled: func(messageHash []byte) [][]byte {
				return signatures[:9]
			},
		}
		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, argLists, batch.ID, 10)
		assert.Equal(t, "", hash)
		assert.True(t, errors.Is(err, errQuorumNotReached))
		assert.True(t, strings.Contains(err.Error(), "num signatures: 9, quorum: 10"))
	})
	t.Run("not enough balance for fees", func(t *testing.T) {
		gasPrice := big.NewInt(1000000000)
		t.Parallel()
		c, _ := NewEthereumClient(args)
		c.gasHandler = &testsCommon.GasHandlerStub{GetCurrentGasPriceCalled: func() (*big.Int, error) {
			return gasPrice, nil
		}}
		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			BalanceAtCalled: func(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
				return gasPrice, nil
			},
		}
		c.signatureHolder = &testsCommon.SignaturesHolderStub{
			SignaturesCalled: func(messageHash []byte) [][]byte {
				return signatures[:9]
			},
		}
		c.erc20ContractsHandler = &bridgeTests.ERC20ContractsHolderStub{
			BalanceOfCalled: func(ctx context.Context, erc20Address common.Address, address common.Address) (*big.Int, error) {
				return big.NewInt(10000), nil
			},
		}
		newBatch := batch.Clone()
		newBatch.Deposits = append(newBatch.Deposits, &bridgeCore.DepositTransfer{
			Nonce:                 40,
			ToBytes:               []byte("to3"),
			DisplayableTo:         "to3",
			FromBytes:             bytes.Repeat([]byte("3"), 32),
			DisplayableFrom:       strings.Repeat("3", 32),
			SourceTokenBytes:      []byte("source token1"),
			DisplayableToken:      "token1",
			Amount:                big.NewInt(80),
			DestinationTokenBytes: []byte("ERC20token1"),
		})
		newArgLists, err := batchProcessor.ExtractListMvxToEth(newBatch)
		require.Nil(t, err)

		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, newArgLists, newBatch.ID, 9)
		assert.Equal(t, "", hash)
		assert.True(t, errors.Is(err, errInsufficientBalance))
	})
	t.Run("execute transfer errors", func(t *testing.T) {
		expectedErr := errors.New("expected error execute transfer")
		c, _ := NewEthereumClient(args)
		c.signatureHolder = &testsCommon.SignaturesHolderStub{
			SignaturesCalled: func(messageHash []byte) [][]byte {
				return signatures[:9]
			},
		}
		c.erc20ContractsHandler = &bridgeTests.ERC20ContractsHolderStub{
			BalanceOfCalled: func(ctx context.Context, erc20Address common.Address, address common.Address) (*big.Int, error) {
				return big.NewInt(10000), nil
			},
		}
		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			ExecuteTransferCalled: func(opts *bind.TransactOpts, mvxTransactions []contract.MvxTransaction, batchNonce *big.Int, sigs [][]byte) (*types.Transaction, error) {
				return nil, expectedErr
			},
		}

		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, argLists, batch.ID, 9)
		assert.Equal(t, "", hash)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work - same number of signatures as quorum", func(t *testing.T) {
		c, _ := NewEthereumClient(args)
		c.signatureHolder = &testsCommon.SignaturesHolderStub{
			SignaturesCalled: func(messageHash []byte) [][]byte {
				return signatures[:9]
			},
		}
		c.erc20ContractsHandler = &bridgeTests.ERC20ContractsHolderStub{
			BalanceOfCalled: func(ctx context.Context, erc20Address common.Address, address common.Address) (*big.Int, error) {
				return big.NewInt(10000), nil
			},
		}
		wasCalled := false
		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			ExecuteTransferCalled: func(opts *bind.TransactOpts, mvxTransactions []contract.MvxTransaction, batchNonce *big.Int, sigs [][]byte) (*types.Transaction, error) {
				assert.Equal(t, expectedMvxTransactions, mvxTransactions)
				assert.Equal(t, big.NewInt(332), batchNonce)
				assert.Equal(t, signatures[:9], sigs)
				wasCalled = true

				txData := &types.LegacyTx{
					Nonce: 0,
				}
				return types.NewTx(txData), nil
			},
		}

		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, argLists, batch.ID, 9)
		assert.Equal(t, "0xc5b2c658f5fa236c598a6e7fbf7f21413dc42e2a41dd982eb772b30707cba2eb", hash)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
	t.Run("should work - more signatures should trim", func(t *testing.T) {
		c, _ := NewEthereumClient(args)
		c.signatureHolder = &testsCommon.SignaturesHolderStub{
			SignaturesCalled: func(messageHash []byte) [][]byte {
				return signatures[:9]
			},
		}
		c.erc20ContractsHandler = &bridgeTests.ERC20ContractsHolderStub{
			BalanceOfCalled: func(ctx context.Context, erc20Address common.Address, address common.Address) (*big.Int, error) {
				return big.NewInt(10000), nil
			},
		}
		wasCalled := false
		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			ExecuteTransferCalled: func(opts *bind.TransactOpts, mvxTransactions []contract.MvxTransaction, batchNonce *big.Int, sigs [][]byte) (*types.Transaction, error) {
				assert.Equal(t, expectedMvxTransactions, mvxTransactions)
				assert.Equal(t, big.NewInt(332), batchNonce)
				assert.Equal(t, signatures[:5], sigs)
				wasCalled = true

				txData := &types.LegacyTx{
					Nonce: 0,
				}
				return types.NewTx(txData), nil
			},
		}

		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, argLists, batch.ID, 5)
		assert.Equal(t, "0xc5b2c658f5fa236c598a6e7fbf7f21413dc42e2a41dd982eb772b30707cba2eb", hash)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestClient_CheckRequiredBalance(t *testing.T) {
	t.Parallel()
	args := createMockEthereumClientArgs()

	tokenErc20 := common.BytesToAddress([]byte("ERC20token1"))
	balance := big.NewInt(1000000)
	t.Run("not enough erc20 balance", func(t *testing.T) {
		c, _ := NewEthereumClient(args)
		c.erc20ContractsHandler = &bridgeTests.ERC20ContractsHolderStub{
			BalanceOfCalled: func(ctx context.Context, erc20Address common.Address, address common.Address) (*big.Int, error) {
				assert.Equal(t, c.safeContractAddress, address)

				return balance, nil
			},
		}
		err := c.CheckRequiredBalance(context.Background(), tokenErc20, big.NewInt(0).Add(balance, big.NewInt(1)))
		assert.True(t, errors.Is(err, errInsufficientErc20Balance))
	})
	t.Run("erc20 balance of errors", func(t *testing.T) {
		expectedErr := errors.New("expected error erc20 balance of")
		c, _ := NewEthereumClient(args)
		c.erc20ContractsHandler = &bridgeTests.ERC20ContractsHolderStub{
			BalanceOfCalled: func(ctx context.Context, erc20Address common.Address, address common.Address) (*big.Int, error) {
				return nil, expectedErr
			},
		}

		err := c.CheckRequiredBalance(context.Background(), tokenErc20, balance)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("should work", func(t *testing.T) {
		c, _ := NewEthereumClient(args)
		c.erc20ContractsHandler = &bridgeTests.ERC20ContractsHolderStub{
			BalanceOfCalled: func(ctx context.Context, erc20Address common.Address, address common.Address) (*big.Int, error) {
				assert.Equal(t, c.safeContractAddress, address)

				return balance, nil
			},
		}
		err := c.CheckRequiredBalance(context.Background(), tokenErc20, balance)
		assert.Nil(t, err)
	})
}

func TestClient_TotalBalances(t *testing.T) {
	t.Parallel()

	t.Run("error while getting total balances", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockEthereumClientArgs()
		args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
			TotalBalancesCalled: func(ctx context.Context, token common.Address) (*big.Int, error) {
				return nil, expectedErr
			},
		}
		c, _ := NewEthereumClient(args)

		balances, err := c.TotalBalances(context.Background(), common.Address{})
		assert.Nil(t, balances)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedBalance := big.NewInt(100)
		args := createMockEthereumClientArgs()
		args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
			TotalBalancesCalled: func(ctx context.Context, token common.Address) (*big.Int, error) {
				return providedBalance, nil
			},
		}
		c, _ := NewEthereumClient(args)

		balances, err := c.TotalBalances(context.Background(), common.Address{})
		assert.Nil(t, err)
		assert.Equal(t, providedBalance, balances)
	})
}

func TestClient_MintBalances(t *testing.T) {
	t.Parallel()

	t.Run("error while getting mint balances", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockEthereumClientArgs()
		args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
			MintBalancesCalled: func(ctx context.Context, token common.Address) (*big.Int, error) {
				return nil, expectedErr
			},
		}
		c, _ := NewEthereumClient(args)

		balances, err := c.MintBalances(context.Background(), common.Address{})
		assert.Nil(t, balances)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedBalance := big.NewInt(100)
		args := createMockEthereumClientArgs()
		args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
			MintBalancesCalled: func(ctx context.Context, token common.Address) (*big.Int, error) {
				return providedBalance, nil
			},
		}
		c, _ := NewEthereumClient(args)

		balances, err := c.MintBalances(context.Background(), common.Address{})
		assert.Nil(t, err)
		assert.Equal(t, providedBalance, balances)
	})
}

func TestClient_BurnBalances(t *testing.T) {
	t.Parallel()

	t.Run("error while getting burn balances", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockEthereumClientArgs()
		args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
			BurnBalancesCalled: func(ctx context.Context, token common.Address) (*big.Int, error) {
				return nil, expectedErr
			},
		}
		c, _ := NewEthereumClient(args)

		balances, err := c.BurnBalances(context.Background(), common.Address{})
		assert.Nil(t, balances)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedBalance := big.NewInt(100)
		args := createMockEthereumClientArgs()
		args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
			BurnBalancesCalled: func(ctx context.Context, token common.Address) (*big.Int, error) {
				return providedBalance, nil
			},
		}
		c, _ := NewEthereumClient(args)

		balances, err := c.BurnBalances(context.Background(), common.Address{})
		assert.Nil(t, err)
		assert.Equal(t, providedBalance, balances)
	})
}

func TestClient_MintBurnTokens(t *testing.T) {
	t.Parallel()

	t.Run("error while getting mint burn tokens", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockEthereumClientArgs()
		args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
			MintBurnTokensCalled: func(ctx context.Context, token common.Address) (bool, error) {
				return false, expectedErr
			},
		}
		c, _ := NewEthereumClient(args)

		isMintBurn, err := c.MintBurnTokens(context.Background(), common.Address{})
		assert.False(t, isMintBurn)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockEthereumClientArgs()
		args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
			MintBurnTokensCalled: func(ctx context.Context, token common.Address) (bool, error) {
				return true, nil
			},
		}
		c, _ := NewEthereumClient(args)

		isMintBurn, err := c.MintBurnTokens(context.Background(), common.Address{})
		assert.Nil(t, err)
		assert.True(t, isMintBurn)
	})
}

func TestClient_NativeTokens(t *testing.T) {
	t.Parallel()

	t.Run("error while getting native tokens", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockEthereumClientArgs()
		args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
			NativeTokensCalled: func(ctx context.Context, token common.Address) (bool, error) {
				return false, expectedErr
			},
		}
		c, _ := NewEthereumClient(args)

		isNative, err := c.NativeTokens(context.Background(), common.Address{})
		assert.False(t, isNative)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockEthereumClientArgs()
		args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
			NativeTokensCalled: func(ctx context.Context, token common.Address) (bool, error) {
				return true, nil
			},
		}
		c, _ := NewEthereumClient(args)

		isNative, err := c.NativeTokens(context.Background(), common.Address{})
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

		args := createMockEthereumClientArgs()
		args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
			GetStatusesAfterExecutionCalled: func(ctx context.Context, batchID *big.Int) ([]byte, bool, error) {
				assert.Equal(t, expectedBatchID, batchID)
				return nil, false, expectedErr
			},
		}

		c, _ := NewEthereumClient(args)
		statuses, err := c.GetTransactionsStatuses(context.Background(), expectedBatchID.Uint64())
		assert.Nil(t, statuses)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("statuses are not final, should error", func(t *testing.T) {
		t.Parallel()

		args := createMockEthereumClientArgs()
		args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
			GetStatusesAfterExecutionCalled: func(ctx context.Context, batchID *big.Int) ([]byte, bool, error) {
				assert.Equal(t, expectedBatchID, batchID)
				return []byte("dummy"), false, nil
			},
		}

		c, _ := NewEthereumClient(args)
		statuses, err := c.GetTransactionsStatuses(context.Background(), expectedBatchID.Uint64())
		assert.Nil(t, statuses)
		assert.Equal(t, errStatusIsNotFinal, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockEthereumClientArgs()
		args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
			GetStatusesAfterExecutionCalled: func(ctx context.Context, batchID *big.Int) ([]byte, bool, error) {
				assert.Equal(t, expectedBatchID, batchID)
				return expectedStatuses, true, nil
			},
		}

		c, _ := NewEthereumClient(args)
		statuses, err := c.GetTransactionsStatuses(context.Background(), expectedBatchID.Uint64())
		assert.Equal(t, expectedStatuses, statuses)
		assert.Nil(t, err)
	})
}

func TestClient_GetQuorumSize(t *testing.T) {
	t.Parallel()

	args := createMockEthereumClientArgs()
	providedValue := big.NewInt(6453)
	args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
		QuorumCalled: func(ctx context.Context) (*big.Int, error) {
			return providedValue, nil
		},
	}
	c, _ := NewEthereumClient(args)

	quorum, err := c.GetQuorumSize(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, providedValue, quorum)
}

func TestClient_IsQuorumReached(t *testing.T) {
	t.Parallel()

	t.Run("quorum errors", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockEthereumClientArgs()
		args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
			QuorumCalled: func(ctx context.Context) (*big.Int, error) {
				return nil, expectedErr
			},
		}
		c, _ := NewEthereumClient(args)

		isReached, err := c.IsQuorumReached(context.Background(), common.Hash{})
		assert.False(t, isReached)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("quorum returns less than minimum allowed", func(t *testing.T) {
		t.Parallel()

		args := createMockEthereumClientArgs()
		args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
			QuorumCalled: func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(0), nil
			},
		}
		c, _ := NewEthereumClient(args)

		isReached, err := c.IsQuorumReached(context.Background(), common.Hash{})
		assert.False(t, isReached)
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "in IsQuorumReached, minQuorum"))
	})
	t.Run("quorum values comparison", func(t *testing.T) {
		t.Parallel()

		signatures := make([][]byte, 0)
		args := createMockEthereumClientArgs()
		args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
			QuorumCalled: func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(3), nil
			},
		}
		args.SignatureHolder = &testsCommon.SignaturesHolderStub{
			SignaturesCalled: func(messageHash []byte) [][]byte {
				return signatures
			},
		}
		c, _ := NewEthereumClient(args)

		isReached, err := c.IsQuorumReached(context.Background(), common.Hash{})
		assert.False(t, isReached)
		assert.Nil(t, err)

		signatures = append(signatures, []byte("sig"))
		signatures = append(signatures, []byte("sig"))
		isReached, err = c.IsQuorumReached(context.Background(), common.Hash{})
		assert.False(t, isReached)
		assert.Nil(t, err)

		signatures = append(signatures, []byte("sig"))
		isReached, err = c.IsQuorumReached(context.Background(), common.Hash{})
		assert.True(t, isReached)
		assert.Nil(t, err)

		signatures = append(signatures, []byte("sig"))
		isReached, err = c.IsQuorumReached(context.Background(), common.Hash{})
		assert.True(t, isReached)
		assert.Nil(t, err)
	})
}

func TestClient_CheckClientAvailability(t *testing.T) {
	t.Parallel()

	currentNonce := uint64(0)
	incrementor := uint64(1)
	args := createMockEthereumClientArgs()
	statusHandler := testsCommon.NewStatusHandlerMock("test")
	args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
		StatusHandler: statusHandler,
		BlockNumberCalled: func(ctx context.Context) (uint64, error) {
			currentNonce += incrementor
			return currentNonce, nil
		},
	}
	expectedErr := errors.New("expected error")

	c, _ := NewEthereumClient(args)
	t.Run("different current nonce should update - 10 times", func(t *testing.T) {
		resetClient(c)
		for i := 0; i < 10; i++ {
			err := c.CheckClientAvailability(context.Background())
			assert.Nil(t, err)
			checkStatusHandler(t, statusHandler, bridgeCore.Available, "")
		}
		assert.True(t, statusHandler.GetIntMetric(bridgeCore.MetricLastBlockNonce) > 0)
	})
	t.Run("same current nonce should error after a while", func(t *testing.T) {
		resetClient(c)
		_ = c.CheckClientAvailability(context.Background())

		incrementor = 0

		// place a random message as to test it is reset
		statusHandler.SetStringMetric(bridgeCore.MetricMultiversXClientStatus, bridgeCore.ClientStatus(3).String())
		statusHandler.SetStringMetric(bridgeCore.MetricLastMultiversXClientError, "random")

		// this will just increment the retry counter
		for i := 0; i < int(args.ClientAvailabilityAllowDelta); i++ {
			err := c.CheckClientAvailability(context.Background())
			assert.Nil(t, err)
			checkStatusHandler(t, statusHandler, bridgeCore.Available, "")
		}

		for i := 0; i < 10; i++ {
			message := fmt.Sprintf("block %d fetched for %d times in a row", currentNonce, args.ClientAvailabilityAllowDelta+uint64(i+1))
			err := c.CheckClientAvailability(context.Background())
			assert.Nil(t, err)
			checkStatusHandler(t, statusHandler, bridgeCore.Unavailable, message)
		}
	})
	t.Run("same current nonce should error after a while and then recovers", func(t *testing.T) {
		resetClient(c)
		_ = c.CheckClientAvailability(context.Background())

		incrementor = 0

		// this will just increment the retry counter
		for i := 0; i < int(args.ClientAvailabilityAllowDelta); i++ {
			err := c.CheckClientAvailability(context.Background())
			assert.Nil(t, err)
			checkStatusHandler(t, statusHandler, bridgeCore.Available, "")
		}

		for i := 0; i < 10; i++ {
			message := fmt.Sprintf("block %d fetched for %d times in a row", currentNonce, args.ClientAvailabilityAllowDelta+uint64(i+1))
			err := c.CheckClientAvailability(context.Background())
			assert.Nil(t, err)
			checkStatusHandler(t, statusHandler, bridgeCore.Unavailable, message)
		}

		incrementor = 1
		err := c.CheckClientAvailability(context.Background())
		assert.Nil(t, err)
		checkStatusHandler(t, statusHandler, bridgeCore.Available, "")
	})
	t.Run("get current nonce errors", func(t *testing.T) {
		resetClient(c)
		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			StatusHandler: statusHandler,
			BlockNumberCalled: func(ctx context.Context) (uint64, error) {
				return 0, expectedErr
			},
		}

		err := c.CheckClientAvailability(context.Background())
		checkStatusHandler(t, statusHandler, bridgeCore.Unavailable, expectedErr.Error())
		assert.Equal(t, expectedErr, err)
	})
}

func TestClient_GetBatchSCMetadata(t *testing.T) {
	t.Parallel()

	t.Run("returns error on filter logs error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("filter logs err")
		args := createMockEthereumClientArgs()
		args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
			FilterLogsCalled: func(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
				return nil, expectedErr
			},
		}
		c, _ := NewEthereumClient(args)
		batch, err := c.GetBatchSCMetadata(context.Background(), 0, 0)
		assert.Nil(t, batch)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("returns expected logs", func(t *testing.T) {
		scExecAbi, _ := contract.ERC20SafeMetaData.GetAbi()
		expectedEvent := &contract.ERC20SafeERC20SCDeposit{
			BatchId:      big.NewInt(1),
			DepositNonce: big.NewInt(2),
			CallData:     []byte("call_data_to_unpack"),
		}

		eventInputs := scExecAbi.Events["ERC20SCDeposit"].Inputs.NonIndexed()
		packedArgs, err := eventInputs.Pack(expectedEvent.DepositNonce, expectedEvent.CallData)
		require.Nil(t, err)

		args := createMockEthereumClientArgs()
		args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
			FilterLogsCalled: func(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
				return []types.Log{
					{
						Data: packedArgs,
					},
				}, nil
			},
		}
		c, _ := NewEthereumClient(args)
		batch, err := c.GetBatchSCMetadata(context.Background(), expectedEvent.BatchId.Uint64(), 500)

		assert.Nil(t, err)
		assert.Equal(t, 1, len(batch))
		assert.Equal(t, expectedEvent.BatchId, batch[0].BatchId)
		assert.Equal(t, expectedEvent.DepositNonce, batch[0].DepositNonce)
		assert.Equal(t, expectedEvent.CallData, batch[0].CallData)
	})
}

func resetClient(c *client) {
	c.mut.Lock()
	c.retriesAvailabilityCheck = 0
	c.mut.Unlock()
	c.clientWrapper.SetStringMetric(bridgeCore.MetricMultiversXClientStatus, "")
	c.clientWrapper.SetStringMetric(bridgeCore.MetricLastMultiversXClientError, "")
}

func checkStatusHandler(t *testing.T, statusHandler *testsCommon.StatusHandlerMock, status bridgeCore.ClientStatus, message string) {
	assert.Equal(t, status.String(), statusHandler.GetStringMetric(bridgeCore.MetricMultiversXClientStatus))
	assert.Equal(t, message, statusHandler.GetStringMetric(bridgeCore.MetricLastMultiversXClientError))
}
