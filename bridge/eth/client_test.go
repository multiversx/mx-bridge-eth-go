package eth

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth/contract"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	mockInteractors "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/interactors"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// verify Client implements interface
var (
	_             = bridge.Bridge(&client{})
	_             = bridge.QuorumProvider(&client{})
	sigHolderStub = &testsCommon.SignaturesHolderStub{
		SignaturesCalled: func(messageHash []byte) [][]byte {
			return [][]byte{[]byte("signature")}
		},
	}
)

const TestPrivateKey = " 60f3849d7c8d93dfce1947d17c34be3e4ea974e74e15ce877f0df34d7192efab\n\t "
const GasLimit = uint64(400000)

func TestGetPending(t *testing.T) {
	pkConv, _ := pubkeyConverter.NewBech32PubkeyConverter(32, logger.GetOrCreate("test"))
	buff, _ := pkConv.Decode("erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8")

	useCases := []struct {
		name          string
		receivedBatch contract.Batch
		expectedBatch *bridge.Batch
	}{
		{
			name: "it will map a non empty batch",
			receivedBatch: contract.Batch{
				Nonce: big.NewInt(1),
				Deposits: []contract.Deposit{
					{
						TokenAddress: common.HexToAddress("0x093c0B280ba430A9Cc9C3649FF34FCBf6347bC50"),
						Amount:       big.NewInt(42),
						Depositor:    common.HexToAddress("0x132A150926691F08a693721503a38affeD18d524"),
						Recipient:    core.ConvertFromByteSliceToArray(buff),
						Status:       0,
					},
				},
			},
			expectedBatch: &bridge.Batch{
				Id: big.NewInt(1),
				Transactions: []*bridge.DepositTransaction{
					{
						To:            string(buff),
						DisplayableTo: "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8",
						From:          "0x132A150926691F08a693721503a38affeD18d524",
						TokenAddress:  "0x093c0B280ba430A9Cc9C3649FF34FCBf6347bC50",
						Amount:        big.NewInt(42),
					},
				},
			},
		},
		{
			name: "it will return nil for an empty batch",
			receivedBatch: contract.Batch{
				Nonce: big.NewInt(0),
			},
			expectedBatch: nil,
		},
	}

	for _, tt := range useCases {
		t.Run(tt.name, func(t *testing.T) {
			clientWrapper := &mockInteractors.EthereumChainInteractorStub{
				GetNextPendingBatchCalled: func(ctx context.Context) (contract.Batch, error) {
					return tt.receivedBatch, nil
				},
			}
			c := client{
				clientWrapper: clientWrapper,
				gasLimit:      GasLimit,
				log:           logger.GetOrCreate("testEthClient"),
			}
			c.addressConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, c.log)

			got, err := c.GetPending(context.TODO())

			assert.Equal(t, tt.expectedBatch, got)
			assert.Nil(t, err)
		})
	}
}

func TestSign(t *testing.T) {
	mapper := testsCommon.NewMapperMock()
	mapper.AddPair(testsCommon.CreateRandomEthereumAddress(), "tck")
	mapper.AddPair(common.HexToAddress("30C7c97471FB5C5238c946E549c608D27f37AAb8"), "574554482d323936313238")

	buildStubs := func() (*testsCommon.BroadcasterStub, *client) {
		broadcaster := &testsCommon.BroadcasterStub{}
		c := &client{
			clientWrapper: &mockInteractors.EthereumChainInteractorStub{},
			privateKey:    privateKey(t),
			broadcaster:   broadcaster,
			mapper:        mapper,
			gasLimit:      GasLimit,
			log:           logger.GetOrCreate("testEthClient"),
		}

		return broadcaster, c
	}
	t.Run("will sign propose status for executed tx", func(t *testing.T) {
		batch := &bridge.Batch{
			Id: bridge.NewBatchId(42),
			Transactions: []*bridge.DepositTransaction{{
				Status: bridge.Executed,
			}},
		}
		broadcaster, c := buildStubs()
		c.GetActionIdForSetStatusOnPendingTransfer(context.TODO(), batch)
		var lastBroadcastSignature []byte
		broadcaster.BroadcastSignatureCalled = func(signature []byte, messageHash []byte) {
			lastBroadcastSignature = signature
		}

		_, _ = c.Sign(context.TODO(), bridge.NewActionId(setStatusAction), batch)

		expectedSignature, _ := hexutil.Decode("0x524957e3081d49d98c98881abd5cf6f737722a4aa0e7915771a567e3cb45cfc625cd9fcf9ec53c86182e517c1e61dbc076722905d11b73e1ed42665ec051342701")

		assert.Equal(t, expectedSignature, lastBroadcastSignature)
	})
	t.Run("will sign propose status for rejected tx", func(t *testing.T) {
		batch := &bridge.Batch{
			Id: bridge.NewBatchId(42),
			Transactions: []*bridge.DepositTransaction{{
				Status: bridge.Rejected,
			}},
		}
		broadcaster, c := buildStubs()
		c.GetActionIdForSetStatusOnPendingTransfer(context.TODO(), batch)
		var lastBroadcastSignature []byte
		broadcaster.BroadcastSignatureCalled = func(signature []byte, messageHash []byte) {
			lastBroadcastSignature = signature
		}

		_, _ = c.Sign(context.TODO(), bridge.NewActionId(setStatusAction), batch)

		expectedSignature, _ := hexutil.Decode("0xd9b1ae38d7e24837e90e7aaac2ae9ca1eb53dc7a30c41774ad7f7f5fd2371c2d0ac6e69643f6aaa25bd9b000dcf0b8be567bcde7f0a5fb5aad122273999bad2500")

		assert.Equal(t, expectedSignature, lastBroadcastSignature)
	})
	t.Run("will sign tx for transfer", func(t *testing.T) {
		batch := &bridge.Batch{
			Id: bridge.NewBatchId(42),
			Transactions: []*bridge.DepositTransaction{{
				To:           "cf95254084ab772696643f0e05ac4711ed674ac1",
				From:         "04aa6d6029b4e136d04848f5b588c2951185666cc871982994f7ef1654282fa3",
				TokenAddress: "574554482d323936313238",
				Amount:       big.NewInt(1),
				DepositNonce: bridge.NewNonce(2),
			},
			},
		}
		broadcaster, c := buildStubs()
		c.GetActionIdForProposeTransfer(context.TODO(), batch)
		var lastBroadcastSignature []byte
		broadcaster.BroadcastSignatureCalled = func(signature []byte, messageHash []byte) {
			lastBroadcastSignature = signature
		}

		_, _ = c.Sign(context.TODO(), bridge.NewActionId(transferAction), batch)
		expectedSignature, _ := hexutil.Decode("0xab3ce0cdc229afc9fcd0447800142da85aa116f16a26e151b9cad95b361ab73d24694ded888a06a1e9b731af8a1b549a1fc5188117e40bea11d9e74af4a6d5fa01")

		assert.Equal(t, expectedSignature, lastBroadcastSignature)
	})
	t.Run("sign transfer will generate recoverable public key", func(t *testing.T) {
		batch := &bridge.Batch{
			Id: bridge.NewBatchId(42),
			Transactions: []*bridge.DepositTransaction{
				{
					To:           "cf95254084ab772696643f0e05ac4711ed674ac1",
					From:         "04aa6d6029b4e136d04848f5b588c2951185666cc871982994f7ef1654282fa3",
					TokenAddress: "574554482d323936313238",
					Amount:       big.NewInt(1),
					DepositNonce: bridge.NewNonce(2),
				},
			},
		}

		sk, err := crypto.HexToECDSA(core.TrimWhiteSpaceCharacters(TestPrivateKey))
		require.Nil(t, err)

		pk := sk.Public()
		pkECDSA, ok := pk.(*ecdsa.PublicKey)
		require.True(t, ok)

		ethAddress := crypto.PubkeyToAddress(*pkECDSA)
		broadcaster, c := buildStubs()
		var sig, msg []byte
		broadcaster.BroadcastSignatureCalled = func(signature []byte, messageHash []byte) {
			sig = signature
			msg = messageHash
		}
		_, _ = c.Sign(context.TODO(), bridge.NewActionId(setStatusAction), batch)
		testPublicKey(t, sig, msg, ethAddress)

		_, _ = c.Sign(context.TODO(), bridge.NewActionId(transferAction), batch)
		testPublicKey(t, sig, msg, ethAddress)
	})
}

func testPublicKey(t *testing.T, sig []byte, msg []byte, expectedAddress common.Address) {
	pkBytesRecovered, err := crypto.Ecrecover(msg, sig)
	require.Nil(t, err)

	pkRecovered, err := crypto.UnmarshalPubkey(pkBytesRecovered)
	require.Nil(t, err)

	addressRecovered := crypto.PubkeyToAddress(*pkRecovered)

	require.Equal(t, expectedAddress, addressRecovered)
}

func TestSignersCount(t *testing.T) {
	broadcaster := &testsCommon.BroadcasterStub{}
	c := client{
		clientWrapper: &mockInteractors.EthereumChainInteractorStub{},
		broadcaster:   broadcaster,
		gasLimit:      GasLimit,
		log:           logger.GetOrCreate("testEthClient"),
	}
	batch := &bridge.Batch{
		Id: bridge.NewBatchId(0),
	}

	t.Run("should return 0 when sig holder is nil", func(t *testing.T) {
		got := c.SignersCount(batch, bridge.NewActionId(0), nil)

		assert.Equal(t, uint(0), got)
	})
	t.Run("should return signature", func(t *testing.T) {
		got := c.SignersCount(batch, bridge.NewActionId(0), sigHolderStub)

		assert.Equal(t, uint(1), got)
	})
}

func TestWasExecuted(t *testing.T) {
	t.Run("when action is set status", func(t *testing.T) {
		clientWrapper := &mockInteractors.EthereumChainInteractorStub{
			WasBatchFinishedCalled: func(ctx context.Context, batchNonce *big.Int) (bool, error) {
				return true, nil
			},
		}
		c := client{
			clientWrapper: clientWrapper,
			broadcaster:   &testsCommon.BroadcasterStub{},
			gasLimit:      GasLimit,
			log:           logger.GetOrCreate("testEthClient"),
		}

		got := c.WasExecuted(context.TODO(), bridge.NewActionId(setStatusAction), bridge.NewBatchId(42))

		assert.Equal(t, true, got)
	})
	t.Run("when action is transfer", func(t *testing.T) {
		clientWrapper := &mockInteractors.EthereumChainInteractorStub{
			WasBatchExecutedCalled: func(ctx context.Context, batchNonce *big.Int) (bool, error) {
				return true, nil
			},
		}
		c := client{
			clientWrapper: clientWrapper,
			broadcaster:   &testsCommon.BroadcasterStub{},
			gasLimit:      GasLimit,
			log:           logger.GetOrCreate("testEthClient"),
		}

		got := c.WasExecuted(context.TODO(), bridge.NewActionId(transferAction), bridge.NewBatchId(42))

		assert.Equal(t, true, got)
	})
}

func TestExecute(t *testing.T) {
	batch := &bridge.Batch{
		Id: bridge.NewBatchId(42),
		Transactions: []*bridge.DepositTransaction{
			{
				To:           testsCommon.CreateRandomEthereumAddress().Hex(),
				TokenAddress: "tck",
				Amount:       big.NewInt(10000),
			},
			{
				To:           testsCommon.CreateRandomEthereumAddress().Hex(),
				TokenAddress: "tck",
				Amount:       big.NewInt(2),
			},
		},
	}

	mapper := testsCommon.NewMapperMock()
	erc20Address := testsCommon.CreateRandomEthereumAddress()
	mapper.AddPair(erc20Address, "tck")

	t.Run("when signatures holder is nil", func(t *testing.T) {
		c := client{
			clientWrapper: &mockInteractors.EthereumChainInteractorStub{},
			privateKey:    privateKey(t),
			publicKey:     publicKey(t),
			broadcaster:   &testsCommon.BroadcasterStub{},
			log:           logger.GetOrCreate("testEthClient"),
			gasLimit:      GasLimit,
			gasHandler:    &testsCommon.GasHandlerStub{},
		}

		got, err := c.Execute(context.TODO(), bridge.NewActionId(setStatusAction), batch, nil)
		assert.Equal(t, ErrNilSignaturesHolder, err)
		assert.Equal(t, "", got)
	})
	t.Run("when action is set status", func(t *testing.T) {
		expected := "0x029bc1fcae8ad9f887af3f37a9ebb223f1e535b009fc7ad7b053ba9b5ff666ae"
		clientWrapper := &mockInteractors.EthereumChainInteractorStub{
			FinishCurrentPendingBatchCalled: func(opts *bind.TransactOpts, batchNonce *big.Int, newDepositStatuses []uint8, signatures [][]byte) (*types.Transaction, error) {
				return types.NewTx(&types.AccessListTx{}), nil
			},
		}
		c := client{
			clientWrapper: clientWrapper,
			privateKey:    privateKey(t),
			publicKey:     publicKey(t),
			broadcaster:   &testsCommon.BroadcasterStub{},
			log:           logger.GetOrCreate("testEthClient"),
			gasLimit:      GasLimit,
			gasHandler:    &testsCommon.GasHandlerStub{},
		}

		got, _ := c.Execute(context.TODO(), bridge.NewActionId(setStatusAction), batch, sigHolderStub)

		assert.Equal(t, expected, got)
	})
	t.Run("not enough ERC20 balance", func(t *testing.T) {
		gasPrice := 1000
		nonce := 1234
		blockNonce := 4321
		executeTransferCalled := false

		safeContractAddress := testsCommon.CreateRandomEthereumAddress()
		erc20Contracts := map[common.Address]Erc20Contract{
			erc20Address: &mockInteractors.Erc20ContractStub{
				BalanceOfCalled: func(ctx context.Context, account common.Address) (*big.Int, error) {
					require.Equal(t, account, safeContractAddress)
					return big.NewInt(10001), nil
				},
			},
		}

		clientWrapper := &mockInteractors.EthereumChainInteractorStub{
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
				executeTransferCalled = true

				assert.Equal(t, big.NewInt(int64(gasPrice)), opts.GasPrice)
				assert.Equal(t, opts.Nonce, big.NewInt(int64(nonce)))
				return types.NewTx(&types.AccessListTx{}), nil
			},
			BlockNumberCalled: func(ctx context.Context) (uint64, error) {
				return uint64(blockNonce), nil
			},
			NonceAtCalled: func(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
				require.Equal(t, big.NewInt(int64(blockNonce)), blockNumber)
				return uint64(nonce), nil
			},
		}
		c := client{
			clientWrapper:       clientWrapper,
			privateKey:          privateKey(t),
			publicKey:           publicKey(t),
			broadcaster:         &testsCommon.BroadcasterStub{},
			mapper:              mapper,
			erc20Contracts:      erc20Contracts,
			safeContractAddress: safeContractAddress,
			gasHandler: &testsCommon.GasHandlerStub{
				GetCurrentGasPriceCalled: func() (*big.Int, error) {
					return big.NewInt(int64(gasPrice)), nil
				},
			},
			gasLimit: GasLimit,
			log:      logger.GetOrCreate("testEthClient"),
		}

		_, err := c.Execute(context.TODO(), bridge.NewActionId(transferAction), batch, sigHolderStub)
		require.True(t, errors.Is(err, ErrInsufficientErc20Balance))
		assert.False(t, executeTransferCalled)
	})
	t.Run("ERC20 contract not found", func(t *testing.T) {
		gasPrice := 1000
		nonce := 1234
		blockNonce := 4321
		executeTransferCalled := false

		clientWrapper := &mockInteractors.EthereumChainInteractorStub{
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
				executeTransferCalled = true

				assert.Equal(t, big.NewInt(int64(gasPrice)), opts.GasPrice)
				assert.Equal(t, opts.Nonce, big.NewInt(int64(nonce)))
				return types.NewTx(&types.AccessListTx{}), nil
			},
			BlockNumberCalled: func(ctx context.Context) (uint64, error) {
				return uint64(blockNonce), nil
			},
			NonceAtCalled: func(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
				require.Equal(t, big.NewInt(int64(blockNonce)), blockNumber)
				return uint64(nonce), nil
			},
		}
		c := client{
			clientWrapper:  clientWrapper,
			privateKey:     privateKey(t),
			publicKey:      publicKey(t),
			broadcaster:    &testsCommon.BroadcasterStub{},
			mapper:         mapper,
			erc20Contracts: make(map[common.Address]Erc20Contract),
			gasHandler: &testsCommon.GasHandlerStub{
				GetCurrentGasPriceCalled: func() (*big.Int, error) {
					return big.NewInt(int64(gasPrice)), nil
				},
			},
			gasLimit: GasLimit,
			log:      logger.GetOrCreate("testEthClient"),
		}

		_, err := c.Execute(context.TODO(), bridge.NewActionId(transferAction), batch, sigHolderStub)
		require.True(t, errors.Is(err, ErrMissingErc20ContractDefinition))
		assert.False(t, executeTransferCalled)
	})
	t.Run("when action is transfer", func(t *testing.T) {
		expected := "0x029bc1fcae8ad9f887af3f37a9ebb223f1e535b009fc7ad7b053ba9b5ff666ae"
		gasPrice := 1000
		nonce := 1234
		blockNonce := 4321
		executeTransferCalled := false

		safeContractAddress := testsCommon.CreateRandomEthereumAddress()
		erc20Contracts := map[common.Address]Erc20Contract{
			erc20Address: &mockInteractors.Erc20ContractStub{
				BalanceOfCalled: func(ctx context.Context, account common.Address) (*big.Int, error) {
					require.Equal(t, account, safeContractAddress)
					return big.NewInt(10002), nil
				},
			},
		}

		clientWrapper := &mockInteractors.EthereumChainInteractorStub{
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
				executeTransferCalled = true

				assert.Equal(t, big.NewInt(int64(gasPrice)), opts.GasPrice)
				assert.Equal(t, opts.Nonce, big.NewInt(int64(nonce)))
				return types.NewTx(&types.AccessListTx{}), nil
			},
			BlockNumberCalled: func(ctx context.Context) (uint64, error) {
				return uint64(blockNonce), nil
			},
			NonceAtCalled: func(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
				require.Equal(t, big.NewInt(int64(blockNonce)), blockNumber)
				return uint64(nonce), nil
			},
		}
		c := client{
			clientWrapper:       clientWrapper,
			privateKey:          privateKey(t),
			publicKey:           publicKey(t),
			broadcaster:         &testsCommon.BroadcasterStub{},
			mapper:              mapper,
			erc20Contracts:      erc20Contracts,
			safeContractAddress: safeContractAddress,
			gasHandler: &testsCommon.GasHandlerStub{
				GetCurrentGasPriceCalled: func() (*big.Int, error) {
					return big.NewInt(int64(gasPrice)), nil
				},
			},
			gasLimit: GasLimit,
			log:      logger.GetOrCreate("testEthClient"),
		}

		got, err := c.Execute(context.TODO(), bridge.NewActionId(transferAction), batch, sigHolderStub)
		require.Nil(t, err)
		assert.Equal(t, expected, got)
		assert.True(t, executeTransferCalled)
	})
	t.Run("gas price handler errors", func(t *testing.T) {
		clientWrapper := &mockInteractors.EthereumChainInteractorStub{
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
				require.Fail(t, "should have not been called")

				return nil, nil
			},
		}
		gasPriceError := fmt.Errorf("gas price error")
		c := client{
			clientWrapper: clientWrapper,
			privateKey:    privateKey(t),
			publicKey:     publicKey(t),
			broadcaster:   &testsCommon.BroadcasterStub{},
			mapper:        mapper,
			gasHandler: &testsCommon.GasHandlerStub{
				GetCurrentGasPriceCalled: func() (*big.Int, error) {
					return big.NewInt(0), gasPriceError
				},
			},
			gasLimit: GasLimit,
			log:      logger.GetOrCreate("testEthClient"),
		}

		got, err := c.Execute(context.TODO(), bridge.NewActionId(transferAction), batch, sigHolderStub)
		assert.Equal(t, "", got)
		assert.Equal(t, gasPriceError, err)
	})
	t.Run("blockchain client errors on blockNumber", func(t *testing.T) {
		blockNumError := fmt.Errorf("block number error")
		clientWrapper := &mockInteractors.EthereumChainInteractorStub{
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
				require.Fail(t, "should have not been called")

				return nil, nil
			},
			BlockNumberCalled: func(ctx context.Context) (uint64, error) {
				return 0, blockNumError
			},
		}
		c := client{
			clientWrapper: clientWrapper,
			privateKey:    privateKey(t),
			publicKey:     publicKey(t),
			broadcaster:   &testsCommon.BroadcasterStub{},
			mapper:        mapper,
			gasHandler:    &testsCommon.GasHandlerStub{},
			gasLimit:      GasLimit,
			log:           logger.GetOrCreate("testEthClient"),
		}

		got, err := c.Execute(context.TODO(), bridge.NewActionId(transferAction), batch, sigHolderStub)
		assert.Equal(t, "", got)
		assert.True(t, errors.Is(err, blockNumError))
	})
}

func TestGetQuorum(t *testing.T) {
	tests := []struct {
		actual   *big.Int
		expected uint
		error    error
	}{
		{actual: big.NewInt(42), expected: 42, error: nil},
		{actual: big.NewInt(math.MaxUint32 + 1), expected: 0, error: errors.New("quorum is not a uint")},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("When contract quorum is %v", test.actual), func(t *testing.T) {
			clientWrapper := &mockInteractors.EthereumChainInteractorStub{
				QuorumCalled: func(ctx context.Context) (*big.Int, error) {
					return test.actual, nil
				},
			}
			c := client{
				clientWrapper: clientWrapper,
				privateKey:    privateKey(t),
				broadcaster:   &testsCommon.BroadcasterStub{},
				mapper:        testsCommon.NewMapperMock(),
				gasLimit:      GasLimit,
				log:           logger.GetOrCreate("testEthClient"),
			}

			actual, err := c.GetQuorum(context.TODO())

			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.error, err)
		})
	}
}

func TestClient_GetTransactionsStatuses(t *testing.T) {
	t.Parallel()

	methodCalled := false
	statuses := []byte{1, 2}
	clientWrapper := &mockInteractors.EthereumChainInteractorStub{
		GetStatusesAfterExecutionCalled: func(ctx context.Context, batchNonceElrondETH *big.Int) ([]uint8, error) {
			methodCalled = true
			return statuses, nil
		},
	}
	c := client{
		clientWrapper: clientWrapper,
		privateKey:    privateKey(t),
		broadcaster:   &testsCommon.BroadcasterStub{},
		mapper:        testsCommon.NewMapperMock(),
		gasLimit:      GasLimit,
		log:           logger.GetOrCreate("testEthClient"),
	}

	returned, err := c.GetTransactionsStatuses(context.TODO(), bridge.NewBatchId(12))
	assert.Nil(t, err)
	assert.Equal(t, statuses, returned)
	assert.True(t, methodCalled)
}

func privateKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()

	sk, err := crypto.HexToECDSA(core.TrimWhiteSpaceCharacters(TestPrivateKey))
	if err != nil {
		t.Fatal(err)
		return nil
	}

	return sk
}

func publicKey(t *testing.T) *ecdsa.PublicKey {
	t.Helper()

	pk := privateKey(t).Public()
	publicKeyECDSA, ok := pk.(*ecdsa.PublicKey)
	if !ok {
		t.Fatal("error casting public key to ECDSA")
	}

	return publicKeyECDSA
}
