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
	_ = bridge.Bridge(&client{})
	_ = bridge.QuorumProvider(&client{})
)

const TestPrivateKey = "60f3849d7c8d93dfce1947d17c34be3e4ea974e74e15ce877f0df34d7192efab"
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
						Recipient:    buff,
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
			bcs := &mockInteractors.BridgeContractStub{
				GetNextPendingBatchCalled: func(opts *bind.CallOpts) (contract.Batch, error) {
					return tt.receivedBatch, nil
				},
			}
			c := client{
				bridgeContract: bcs,
				gasLimit:       GasLimit,
				log:            logger.GetOrCreate("testEthClient"),
			}
			c.addressConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, c.log)

			got := c.GetPending(context.TODO())

			assert.Equal(t, tt.expectedBatch, got)
		})
	}
}

func TestSign(t *testing.T) {
	buildStubs := func() (*broadcasterStub, client) {
		broadcaster := &broadcasterStub{}
		c := client{
			bridgeContract: &mockInteractors.BridgeContractStub{},
			privateKey:     privateKey(t),
			broadcaster:    broadcaster,
			mapper:         &mapperStub{},
			gasLimit:       GasLimit,
			log:            logger.GetOrCreate("testEthClient"),
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
		_, _ = c.Sign(context.TODO(), bridge.NewActionId(setStatusAction), batch)

		expectedSignature, _ := hexutil.Decode("0x524957e3081d49d98c98881abd5cf6f737722a4aa0e7915771a567e3cb45cfc625cd9fcf9ec53c86182e517c1e61dbc076722905d11b73e1ed42665ec051342701")

		assert.Equal(t, expectedSignature, broadcaster.lastBroadcastSignature)
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
		_, _ = c.Sign(context.TODO(), bridge.NewActionId(setStatusAction), batch)

		expectedSignature, _ := hexutil.Decode("0xd9b1ae38d7e24837e90e7aaac2ae9ca1eb53dc7a30c41774ad7f7f5fd2371c2d0ac6e69643f6aaa25bd9b000dcf0b8be567bcde7f0a5fb5aad122273999bad2500")

		assert.Equal(t, expectedSignature, broadcaster.lastBroadcastSignature)
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
		_, _ = c.Sign(context.TODO(), bridge.NewActionId(transferAction), batch)
		expectedSignature, _ := hexutil.Decode("0xab3ce0cdc229afc9fcd0447800142da85aa116f16a26e151b9cad95b361ab73d24694ded888a06a1e9b731af8a1b549a1fc5188117e40bea11d9e74af4a6d5fa01")

		assert.Equal(t, expectedSignature, broadcaster.lastBroadcastSignature)
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

		sk, err := crypto.HexToECDSA(TestPrivateKey)
		require.Nil(t, err)

		pk := sk.Public()
		pkECDSA, ok := pk.(*ecdsa.PublicKey)
		require.True(t, ok)

		ethAddress := crypto.PubkeyToAddress(*pkECDSA)
		broadcaster, c := buildStubs()
		_, _ = c.Sign(context.TODO(), bridge.NewActionId(setStatusAction), batch)
		testPublicKey(t, broadcaster, ethAddress)

		_, _ = c.Sign(context.TODO(), bridge.NewActionId(transferAction), batch)
		testPublicKey(t, broadcaster, ethAddress)
	})
}

func testPublicKey(t *testing.T, broadcaster *broadcasterStub, expectedAddress common.Address) {
	sig := broadcaster.lastBroadcastSignature
	msg := broadcaster.lastBroadcastMsgHash

	pkBytesRecovered, err := crypto.Ecrecover(msg, sig)
	require.Nil(t, err)

	pkRecovered, err := crypto.UnmarshalPubkey(pkBytesRecovered)
	require.Nil(t, err)

	addressRecovered := crypto.PubkeyToAddress(*pkRecovered)

	require.Equal(t, expectedAddress, addressRecovered)
}

func TestSignersCount(t *testing.T) {
	broadcaster := &broadcasterStub{lastBroadcastSignature: []byte("signature")}
	c := client{
		bridgeContract: &mockInteractors.BridgeContractStub{},
		broadcaster:    broadcaster,
		gasLimit:       GasLimit,
		log:            logger.GetOrCreate("testEthClient"),
	}

	batch := &bridge.Batch{
		Id: bridge.NewBatchId(0),
	}
	got := c.SignersCount(context.TODO(), batch, bridge.NewActionId(0))

	assert.Equal(t, uint(1), got)
}

func TestWasExecuted(t *testing.T) {
	t.Run("when action is set status", func(t *testing.T) {
		bcs := &mockInteractors.BridgeContractStub{
			WasBatchFinishedCalled: func(opts *bind.CallOpts, batchNonce *big.Int) (bool, error) {
				return true, nil
			},
		}
		c := client{
			bridgeContract: bcs,
			broadcaster:    &broadcasterStub{},
			gasLimit:       GasLimit,
			log:            logger.GetOrCreate("testEthClient"),
		}

		got := c.WasExecuted(context.TODO(), bridge.NewActionId(setStatusAction), bridge.NewBatchId(42))

		assert.Equal(t, true, got)
	})
	t.Run("when action is transfer", func(t *testing.T) {
		bcs := &mockInteractors.BridgeContractStub{
			WasBatchExecutedCalled: func(opts *bind.CallOpts, batchNonce *big.Int) (bool, error) {
				return true, nil
			},
		}
		c := client{
			bridgeContract: bcs,
			broadcaster:    &broadcasterStub{},
			gasLimit:       GasLimit,
			log:            logger.GetOrCreate("testEthClient"),
		}

		got := c.WasExecuted(context.TODO(), bridge.NewActionId(transferAction), bridge.NewBatchId(42))

		assert.Equal(t, true, got)
	})
}

func TestExecute(t *testing.T) {
	t.Run("when action is set status", func(t *testing.T) {
		expected := "0x029bc1fcae8ad9f887af3f37a9ebb223f1e535b009fc7ad7b053ba9b5ff666ae"
		bcs := &mockInteractors.BridgeContractStub{
			FinishCurrentPendingBatchCalled: func(opts *bind.TransactOpts, batchNonce *big.Int, newDepositStatuses []uint8, signatures [][]byte) (*types.Transaction, error) {
				return types.NewTx(&types.AccessListTx{}), nil
			},
		}
		c := client{
			bridgeContract:   bcs,
			privateKey:       privateKey(t),
			publicKey:        publicKey(t),
			broadcaster:      &broadcasterStub{},
			blockchainClient: &mockInteractors.BlockchainClientStub{},
			log:              logger.GetOrCreate("testEthClient"),
			gasLimit:         GasLimit,
			gasHandler:       &testsCommon.GasHandlerStub{},
		}
		batch := &bridge.Batch{Id: bridge.NewBatchId(42)}

		got, _ := c.Execute(context.TODO(), bridge.NewActionId(setStatusAction), batch)

		assert.Equal(t, expected, got)
	})
	t.Run("when action is transfer", func(t *testing.T) {
		expected := "0x029bc1fcae8ad9f887af3f37a9ebb223f1e535b009fc7ad7b053ba9b5ff666ae"
		gasPrice := 1000
		nonce := 1234
		blockNonce := 4321
		executeTransferCalled := false
		bcs := &mockInteractors.BridgeContractStub{
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
				executeTransferCalled = true

				assert.Equal(t, big.NewInt(int64(gasPrice)), opts.GasPrice)
				assert.Equal(t, opts.Nonce, big.NewInt(int64(nonce)))
				return types.NewTx(&types.AccessListTx{}), nil
			},
		}
		c := client{
			bridgeContract: bcs,
			privateKey:     privateKey(t),
			publicKey:      publicKey(t),
			broadcaster:    &broadcasterStub{},
			mapper:         &mapperStub{},
			blockchainClient: &mockInteractors.BlockchainClientStub{
				BlockNumberCalled: func(ctx context.Context) (uint64, error) {
					return uint64(blockNonce), nil
				},
				NonceAtCalled: func(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
					require.Equal(t, big.NewInt(int64(blockNonce)), blockNumber)
					return uint64(nonce), nil
				},
			},
			gasHandler: &testsCommon.GasHandlerStub{
				GetCurrentGasPriceCalled: func() (*big.Int, error) {
					return big.NewInt(int64(gasPrice)), nil
				},
			},
			gasLimit: GasLimit,
			log:      logger.GetOrCreate("testEthClient"),
		}
		batch := &bridge.Batch{Id: bridge.NewBatchId(42)}

		got, err := c.Execute(context.TODO(), bridge.NewActionId(transferAction), batch)
		require.Nil(t, err)
		assert.Equal(t, expected, got)
		assert.True(t, executeTransferCalled)
	})
	t.Run("gas price handler errors", func(t *testing.T) {
		bcs := &mockInteractors.BridgeContractStub{
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
				require.Fail(t, "should have not been called")

				return nil, nil
			},
		}
		gasPriceError := fmt.Errorf("gas price error")
		c := client{
			bridgeContract:   bcs,
			privateKey:       privateKey(t),
			publicKey:        publicKey(t),
			broadcaster:      &broadcasterStub{},
			mapper:           &mapperStub{},
			blockchainClient: &mockInteractors.BlockchainClientStub{},
			gasHandler: &testsCommon.GasHandlerStub{
				GetCurrentGasPriceCalled: func() (*big.Int, error) {
					return big.NewInt(0), gasPriceError
				},
			},
			gasLimit: GasLimit,
			log:      logger.GetOrCreate("testEthClient"),
		}
		batch := &bridge.Batch{Id: bridge.NewBatchId(42)}

		got, err := c.Execute(context.TODO(), bridge.NewActionId(transferAction), batch)
		assert.Equal(t, "", got)
		assert.Equal(t, gasPriceError, err)
	})
	t.Run("blockchain client errors on blockNumber", func(t *testing.T) {
		bcs := &mockInteractors.BridgeContractStub{
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
				require.Fail(t, "should have not been called")

				return nil, nil
			},
		}
		blockNumError := fmt.Errorf("block number error")
		c := client{
			bridgeContract: bcs,
			privateKey:     privateKey(t),
			publicKey:      publicKey(t),
			broadcaster:    &broadcasterStub{},
			mapper:         &mapperStub{},
			blockchainClient: &mockInteractors.BlockchainClientStub{
				BlockNumberCalled: func(ctx context.Context) (uint64, error) {
					return 0, blockNumError
				},
			},
			gasHandler: &testsCommon.GasHandlerStub{},
			gasLimit:   GasLimit,
			log:        logger.GetOrCreate("testEthClient"),
		}
		batch := &bridge.Batch{Id: bridge.NewBatchId(42)}

		got, err := c.Execute(context.TODO(), bridge.NewActionId(transferAction), batch)
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
			bcs := &mockInteractors.BridgeContractStub{
				QuorumCalled: func(opts *bind.CallOpts) (*big.Int, error) {
					return test.actual, nil
				},
			}
			c := client{
				bridgeContract: bcs,
				privateKey:     privateKey(t),
				broadcaster:    &broadcasterStub{},
				mapper:         &mapperStub{},
				gasLimit:       GasLimit,
				log:            logger.GetOrCreate("testEthClient"),
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
	bcs := &mockInteractors.BridgeContractStub{
		GetStatusesAfterExecutionCalled: func(opts *bind.CallOpts, batchNonceElrondETH *big.Int) ([]uint8, error) {
			methodCalled = true
			return statuses, nil
		},
	}
	c := client{
		bridgeContract: bcs,
		privateKey:     privateKey(t),
		broadcaster:    &broadcasterStub{},
		mapper:         &mapperStub{},
		gasLimit:       GasLimit,
		log:            logger.GetOrCreate("testEthClient"),
	}

	returned, err := c.GetTransactionsStatuses(context.TODO(), bridge.NewBatchId(12))
	assert.Nil(t, err)
	assert.Equal(t, statuses, returned)
	assert.True(t, methodCalled)
}

func privateKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()

	sk, err := crypto.HexToECDSA(TestPrivateKey)
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

type broadcasterStub struct {
	lastBroadcastSignature []byte
	lastBroadcastMsgHash   []byte
}


func (b *broadcasterStub) SendSignature(signature []byte, msgHash []byte) {
	b.lastBroadcastSignature = signature
	b.lastBroadcastMsgHash = msgHash
}

// Signatures -
func (b *broadcasterStub) Signatures(_ []byte) [][]byte {
	return [][]byte{b.lastBroadcastSignature}
}

// IsInterfaceNil returns true if there is no value under the interface
func (b *broadcasterStub) IsInterfaceNil() bool {
	return b == nil
}

type mapperStub struct{}

// GetTokenId -
func (m *mapperStub) GetTokenId(string) string {
	return "tokenId"
}

// GetErc20Address -
func (m *mapperStub) GetErc20Address(string) string {
	return "0x30C7c97471FB5C5238c946E549c608D27f37AAb8"
}

// IsInterfaceNil returns true if there is no value under the interface
func (m *mapperStub) IsInterfaceNil() bool {
	return m == nil
}
