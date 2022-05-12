package ethereum

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/ethElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients/ethereum/contract"
	bridgeCore "github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/core/converters"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	bridgeTests "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridge"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

var expectedAmounts = []*big.Int{big.NewInt(20), big.NewInt(40)}
var expectedTokens = []common.Address{common.BytesToAddress([]byte("ERC20token1")), common.BytesToAddress([]byte("ERC20token2"))}
var expectedRecipients = []common.Address{common.BytesToAddress([]byte("to1")), common.BytesToAddress([]byte("to2"))}
var expectedNonces = []*big.Int{big.NewInt(10), big.NewInt(30)}

func createMockEthereumClientArgs() ArgsEthereumClient {
	sk, _ := crypto.HexToECDSA("9bb971db41e3815a669a71c3f1bcb24e0b81f21e04bf11faa7a34b9b40e7cfb1")

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
		PrivateKey:            sk,
		TokensMapper: &bridgeTests.TokensMapperStub{
			ConvertTokenCalled: func(ctx context.Context, sourceBytes []byte) ([]byte, error) {
				return append([]byte("ERC20"), sourceBytes...), nil
			},
		},
		SignatureHolder:         &testsCommon.SignaturesHolderStub{},
		SafeContractAddress:     testsCommon.CreateRandomEthereumAddress(),
		GasHandler:              &testsCommon.GasHandlerStub{},
		TransferGasLimitBase:    50,
		TransferGasLimitForEach: 20,
		AllowDelta:              5,
	}
}

func createMockTransferBatch() *clients.TransferBatch {
	return &clients.TransferBatch{
		ID: 332,
		Deposits: []*clients.DepositTransfer{
			{
				Nonce:               10,
				ToBytes:             []byte("to1"),
				DisplayableTo:       "to1",
				FromBytes:           []byte("from1"),
				DisplayableFrom:     "from1",
				TokenBytes:          []byte("token1"),
				DisplayableToken:    "token1",
				Amount:              big.NewInt(20),
				ConvertedTokenBytes: []byte("ERC20token1"),
			},
			{
				Nonce:               30,
				ToBytes:             []byte("to2"),
				DisplayableTo:       "to2",
				FromBytes:           []byte("from2"),
				DisplayableFrom:     "from2",
				TokenBytes:          []byte("token2"),
				DisplayableToken:    "token2",
				Amount:              big.NewInt(40),
				ConvertedTokenBytes: []byte("ERC20token2"),
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
	t.Run("nil private key", func(t *testing.T) {
		args := createMockEthereumClientArgs()
		args.PrivateKey = nil
		c, err := NewEthereumClient(args)

		assert.Equal(t, clients.ErrNilPrivateKey, err)
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
	t.Run("invalid AllowDelta should error", func(t *testing.T) {
		t.Parallel()

		args := createMockEthereumClientArgs()
		args.AllowDelta = 0

		c, err := NewEthereumClient(args)

		assert.True(t, check.IfNil(c))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for args.AllowedDelta"))
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
			GetBatchCalled: func(ctx context.Context, batchNonce *big.Int) (contract.Batch, error) {
				return contract.Batch{}, expectedErr
			},
		}
		batch, err := c.GetBatch(context.Background(), 1)
		assert.Nil(t, batch)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("error while getting deposits", func(t *testing.T) {
		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			GetBatchCalled: func(ctx context.Context, batchNonce *big.Int) (contract.Batch, error) {
				return contract.Batch{
					Nonce:         batchNonce,
					DepositsCount: 2,
				}, nil
			},
			GetBatchDepositsCalled: func(ctx context.Context, batchNonce *big.Int) ([]contract.Deposit, error) {
				return nil, expectedErr
			},
		}
		batch, err := c.GetBatch(context.Background(), 1)
		assert.Nil(t, batch)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("deposits mismatch - with 0", func(t *testing.T) {
		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			GetBatchCalled: func(ctx context.Context, batchNonce *big.Int) (contract.Batch, error) {
				return contract.Batch{
					Nonce:         batchNonce,
					DepositsCount: 2,
				}, nil
			},
			GetBatchDepositsCalled: func(ctx context.Context, batchNonce *big.Int) ([]contract.Deposit, error) {
				return make([]contract.Deposit, 0), nil
			},
		}
		batch, err := c.GetBatch(context.Background(), 1)
		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, errDepositsAndBatchDepositsCountDiffer))
		assert.True(t, strings.Contains(err.Error(), "batch.DepositsCount: 2, fetched deposits len: 0"))
	})
	t.Run("deposits mismatch - with non zero value", func(t *testing.T) {
		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			GetBatchCalled: func(ctx context.Context, batchNonce *big.Int) (contract.Batch, error) {
				return contract.Batch{
					Nonce:         batchNonce,
					DepositsCount: 2,
				}, nil
			},
			GetBatchDepositsCalled: func(ctx context.Context, batchNonce *big.Int) ([]contract.Deposit, error) {
				return []contract.Deposit{
					{
						Nonce: big.NewInt(22),
					},
				}, nil
			},
		}
		batch, err := c.GetBatch(context.Background(), 1)
		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, errDepositsAndBatchDepositsCountDiffer))
		assert.True(t, strings.Contains(err.Error(), "batch.DepositsCount: 2, fetched deposits len: 1"))
	})
	t.Run("returns batch should work", func(t *testing.T) {
		from1 := testsCommon.CreateRandomEthereumAddress()
		token1 := testsCommon.CreateRandomEthereumAddress()
		recipient1 := testsCommon.CreateRandomElrondAddress()

		from2 := testsCommon.CreateRandomEthereumAddress()
		token2 := testsCommon.CreateRandomEthereumAddress()
		recipient2 := testsCommon.CreateRandomElrondAddress()

		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			GetBatchCalled: func(ctx context.Context, batchNonce *big.Int) (contract.Batch, error) {
				return contract.Batch{
					Nonce:                  big.NewInt(112243),
					BlockNumber:            0,
					LastUpdatedBlockNumber: 0,
					DepositsCount:          2,
				}, nil
			},
			GetBatchDepositsCalled: func(ctx context.Context, batchNonce *big.Int) ([]contract.Deposit, error) {
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
				}, nil
			},
		}

		expectedBatch := &clients.TransferBatch{
			ID: 112243,
			Deposits: []*clients.DepositTransfer{
				{
					Nonce:               10,
					ToBytes:             recipient1.AddressBytes(),
					DisplayableTo:       recipient1.AddressAsBech32String(),
					FromBytes:           from1[:],
					DisplayableFrom:     hex.EncodeToString(from1[:]),
					TokenBytes:          token1[:],
					DisplayableToken:    hex.EncodeToString(token1[:]),
					Amount:              big.NewInt(20),
					ConvertedTokenBytes: append([]byte("ERC20"), token1[:]...),
				},
				{
					Nonce:               30,
					ToBytes:             recipient2.AddressBytes(),
					DisplayableTo:       recipient2.AddressAsBech32String(),
					FromBytes:           from2[:],
					DisplayableFrom:     hex.EncodeToString(from2[:]),
					TokenBytes:          token2[:],
					DisplayableToken:    hex.EncodeToString(token2[:]),
					Amount:              big.NewInt(40),
					ConvertedTokenBytes: append([]byte("ERC20"), token2[:]...),
				},
			},
			Statuses: make([]byte, 2),
		}

		batch, err := c.GetBatch(context.Background(), 1)
		assert.Equal(t, expectedBatch, batch)
		assert.Nil(t, err)
	})

}

func TestClient_GenerateMessageHash(t *testing.T) {
	t.Parallel()

	args := createMockEthereumClientArgs()
	batch := createMockTransferBatch()

	t.Run("nil batch should error", func(t *testing.T) {
		c, _ := NewEthereumClient(args)
		h, err := c.GenerateMessageHash(nil)

		assert.Equal(t, common.Hash{}, h)
		assert.True(t, errors.Is(err, clients.ErrNilBatch))
	})
	t.Run("should work", func(t *testing.T) {
		c, _ := NewEthereumClient(args)
		argLists, _ := c.extractList(batch)
		assert.Equal(t, expectedAmounts, argLists.amounts)
		assert.Equal(t, expectedTokens, argLists.tokens)
		assert.Equal(t, expectedRecipients, argLists.recipients)
		assert.Equal(t, expectedNonces, argLists.nonces)

		h, err := c.GenerateMessageHash(batch)
		assert.Nil(t, err)
		assert.Equal(t, "c68190e0a3b8d7c6bd966272a11d618ceddc4b38662b0a1610621f4d30ec07ca", hex.EncodeToString(h.Bytes()))
	})
}

func TestClient_BroadcastSignatureForMessageHash(t *testing.T) {
	t.Parallel()

	expectedSig := "b556014dd984183e4662dc3204e522a5a92093fd6f64bb2da9c1b66b8d5ad12d774e05728b83c76bf09bb91af93ede4118f59aa949c7d02c86051dd0fa140c9900"
	broadcastCalled := false

	hash := common.HexToHash("c99286352d865e33f1747761cbd440a7906b9bd8a5261cb6909e5ba18dd19b08")
	args := createMockEthereumClientArgs()
	args.Broadcaster = &testsCommon.BroadcasterStub{
		BroadcastSignatureCalled: func(signature []byte, messageHash []byte) {
			assert.Equal(t, hash.Bytes(), messageHash)
			assert.Equal(t, expectedSig, hex.EncodeToString(signature))
			broadcastCalled = true
		},
	}

	c, _ := NewEthereumClient(args)
	c.BroadcastSignatureForMessageHash(hash)

	assert.True(t, broadcastCalled)
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
	batch := createMockTransferBatch()
	signatures := make([][]byte, 10)
	for i := range signatures {
		signatures[i] = []byte(fmt.Sprintf("sig %d", i))
	}

	t.Run("nil batch", func(t *testing.T) {
		c, _ := NewEthereumClient(args)
		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, nil, 10)
		assert.Equal(t, "", hash)
		assert.True(t, errors.Is(err, clients.ErrNilBatch))
	})
	t.Run("get block number fails", func(t *testing.T) {
		expectedErr := errors.New("expected error get block number")
		c, _ := NewEthereumClient(args)
		c.clientWrapper = &bridgeTests.EthereumClientWrapperStub{
			BlockNumberCalled: func(ctx context.Context) (uint64, error) {
				return 0, expectedErr
			},
		}
		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, batch, 10)
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
		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, batch, 10)
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
		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, batch, 10)
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
		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, batch, 10)
		assert.Equal(t, "", hash)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("not enough quorum", func(t *testing.T) {
		c, _ := NewEthereumClient(args)
		c.signatureHolder = &testsCommon.SignaturesHolderStub{
			SignaturesCalled: func(messageHash []byte) [][]byte {
				return signatures[:9]
			},
		}
		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, batch, 10)
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
		newBatch.Deposits = append(newBatch.Deposits, &clients.DepositTransfer{
			Nonce:               40,
			ToBytes:             []byte("to3"),
			DisplayableTo:       "to3",
			FromBytes:           []byte("from3"),
			DisplayableFrom:     "from3",
			TokenBytes:          []byte("token1"),
			DisplayableToken:    "token1",
			Amount:              big.NewInt(80),
			ConvertedTokenBytes: []byte("ERC20token1"),
		})

		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, newBatch, 9)
		assert.Equal(t, "", hash)
		assert.True(t, errors.Is(err, errInsufficientBalance))
	})
	t.Run("not enough erc20 balance", func(t *testing.T) {
		c, _ := NewEthereumClient(args)
		c.signatureHolder = &testsCommon.SignaturesHolderStub{
			SignaturesCalled: func(messageHash []byte) [][]byte {
				return signatures[:9]
			},
		}
		c.erc20ContractsHandler = &bridgeTests.ERC20ContractsHolderStub{
			BalanceOfCalled: func(ctx context.Context, erc20Address common.Address, address common.Address) (*big.Int, error) {
				assert.Equal(t, c.safeContractAddress, address)
				tokenErc20 := common.BytesToAddress([]byte("ERC20token1"))
				if erc20Address.String() == tokenErc20.String() {
					return big.NewInt(99), nil
				}

				return big.NewInt(1000000), nil
			},
		}

		newBatch := batch.Clone()
		newBatch.Deposits = append(newBatch.Deposits, &clients.DepositTransfer{
			Nonce:               40,
			ToBytes:             []byte("to3"),
			DisplayableTo:       "to3",
			FromBytes:           []byte("from3"),
			DisplayableFrom:     "from3",
			TokenBytes:          []byte("token1"),
			DisplayableToken:    "token1",
			Amount:              big.NewInt(80),
			ConvertedTokenBytes: []byte("ERC20token1"),
		})

		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, newBatch, 9)
		assert.Equal(t, "", hash)
		assert.True(t, errors.Is(err, errInsufficientErc20Balance))
	})
	t.Run("erc20 balance of errors", func(t *testing.T) {
		expectedErr := errors.New("expected error erc20 balance of")
		c, _ := NewEthereumClient(args)
		c.signatureHolder = &testsCommon.SignaturesHolderStub{
			SignaturesCalled: func(messageHash []byte) [][]byte {
				return signatures[:9]
			},
		}
		c.erc20ContractsHandler = &bridgeTests.ERC20ContractsHolderStub{
			BalanceOfCalled: func(ctx context.Context, erc20Address common.Address, address common.Address) (*big.Int, error) {
				return nil, expectedErr
			},
		}

		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, batch, 9)
		assert.Equal(t, "", hash)
		assert.True(t, errors.Is(err, expectedErr))
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
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, sigs [][]byte) (*types.Transaction, error) {
				return nil, expectedErr
			},
		}

		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, batch, 9)
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
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, sigs [][]byte) (*types.Transaction, error) {
				assert.Equal(t, expectedTokens, tokens)
				assert.Equal(t, expectedRecipients, recipients)
				assert.Equal(t, expectedAmounts, amounts)
				assert.Equal(t, expectedNonces, nonces)
				assert.Equal(t, big.NewInt(332), batchNonce)
				assert.Equal(t, signatures[:9], sigs)
				wasCalled = true

				txData := &types.LegacyTx{
					Nonce: 0,
				}
				return types.NewTx(txData), nil
			},
		}

		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, batch, 9)
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
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, sigs [][]byte) (*types.Transaction, error) {
				assert.Equal(t, expectedTokens, tokens)
				assert.Equal(t, expectedRecipients, recipients)
				assert.Equal(t, expectedAmounts, amounts)
				assert.Equal(t, expectedNonces, nonces)
				assert.Equal(t, big.NewInt(332), batchNonce)
				assert.Equal(t, signatures[:5], sigs)
				wasCalled = true

				txData := &types.LegacyTx{
					Nonce: 0,
				}
				return types.NewTx(txData), nil
			},
		}

		hash, err := c.ExecuteTransfer(context.Background(), common.Hash{}, batch, 5)
		assert.Equal(t, "0xc5b2c658f5fa236c598a6e7fbf7f21413dc42e2a41dd982eb772b30707cba2eb", hash)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func TestClient_GetTransactionsStatuses(t *testing.T) {
	t.Parallel()

	expectedStatuses := []byte{1, 2, 3}
	expectedBatchID := big.NewInt(2232)
	args := createMockEthereumClientArgs()
	args.ClientWrapper = &bridgeTests.EthereumClientWrapperStub{
		GetStatusesAfterExecutionCalled: func(ctx context.Context, batchID *big.Int) ([]byte, error) {
			assert.Equal(t, expectedBatchID, batchID)
			return expectedStatuses, nil
		},
	}

	c, _ := NewEthereumClient(args)

	statuses, err := c.GetTransactionsStatuses(context.Background(), expectedBatchID.Uint64())
	assert.Nil(t, err)
	assert.Equal(t, expectedStatuses, statuses)
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
			checkStatusHandler(t, statusHandler, ethElrond.Available, "")
		}
	})
	t.Run("same current nonce should error after a while", func(t *testing.T) {
		resetClient(c)
		_ = c.CheckClientAvailability(context.Background())

		incrementor = 0

		// place a random message as to test it is reset
		statusHandler.SetStringMetric(bridgeCore.MetricElrondClientStatus, ethElrond.ClientStatus(3).String())
		statusHandler.SetStringMetric(bridgeCore.MetricLastElrondClientError, "random")

		// this will just increment the retry counter
		for i := 0; i < int(args.AllowDelta); i++ {
			err := c.CheckClientAvailability(context.Background())
			assert.Nil(t, err)
			checkStatusHandler(t, statusHandler, ethElrond.Available, "")
		}

		for i := 0; i < 10; i++ {
			message := fmt.Sprintf("block %d fetched for %d times in a row", currentNonce, args.AllowDelta+uint64(i+1))
			err := c.CheckClientAvailability(context.Background())
			assert.Nil(t, err)
			checkStatusHandler(t, statusHandler, ethElrond.Unavailable, message)
		}
	})
	t.Run("same current nonce should error after a while and then recovers", func(t *testing.T) {
		resetClient(c)
		_ = c.CheckClientAvailability(context.Background())

		incrementor = 0

		// this will just increment the retry counter
		for i := 0; i < int(args.AllowDelta); i++ {
			err := c.CheckClientAvailability(context.Background())
			assert.Nil(t, err)
			checkStatusHandler(t, statusHandler, ethElrond.Available, "")
		}

		for i := 0; i < 10; i++ {
			message := fmt.Sprintf("block %d fetched for %d times in a row", currentNonce, args.AllowDelta+uint64(i+1))
			err := c.CheckClientAvailability(context.Background())
			assert.Nil(t, err)
			checkStatusHandler(t, statusHandler, ethElrond.Unavailable, message)
		}

		incrementor = 1
		err := c.CheckClientAvailability(context.Background())
		assert.Nil(t, err)
		checkStatusHandler(t, statusHandler, ethElrond.Available, "")
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
		checkStatusHandler(t, statusHandler, ethElrond.Unavailable, expectedErr.Error())
		assert.Equal(t, expectedErr, err)
	})
}

func resetClient(c *client) {
	c.mut.Lock()
	c.retriesAvailabilityCheck = 0
	c.mut.Unlock()
	c.clientWrapper.SetStringMetric(bridgeCore.MetricElrondClientStatus, "")
	c.clientWrapper.SetStringMetric(bridgeCore.MetricLastElrondClientError, "")
}

func checkStatusHandler(t *testing.T, statusHandler *testsCommon.StatusHandlerMock, status ethElrond.ClientStatus, message string) {
	assert.Equal(t, status.String(), statusHandler.GetStringMetric(bridgeCore.MetricElrondClientStatus))
	assert.Equal(t, message, statusHandler.GetStringMetric(bridgeCore.MetricLastElrondClientError))
}
