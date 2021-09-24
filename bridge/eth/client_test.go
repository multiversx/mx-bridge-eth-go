package eth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/testHelpers"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

// verify Client implements interface
var (
	_ = bridge.Bridge(&Client{})
)

const TestPrivateKey = "60f3849d7c8d93dfce1947d17c34be3e4ea974e74e15ce877f0df34d7192efab"
const GasLimit = uint64(400000)

func TestGetPending(t *testing.T) {
	testHelpers.SetTestLogLevel()

	useCases := []struct {
		name          string
		receivedBatch Batch
		expectedBatch *bridge.Batch
	}{
		{
			name: "it will map a non empty batch",
			receivedBatch: Batch{
				Nonce: big.NewInt(1),
				Deposits: []Deposit{
					{
						TokenAddress: common.HexToAddress("0x093c0B280ba430A9Cc9C3649FF34FCBf6347bC50"),
						Amount:       big.NewInt(42),
						Depositor:    common.HexToAddress("0x132A150926691F08a693721503a38affeD18d524"),
						Recipient:    []byte("erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8"),
						Status:       0,
					},
				},
			},
			expectedBatch: &bridge.Batch{
				Id: big.NewInt(1),
				Transactions: []*bridge.DepositTransaction{
					{To: "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8",
						From:         "0x132A150926691F08a693721503a38affeD18d524",
						TokenAddress: "0x093c0B280ba430A9Cc9C3649FF34FCBf6347bC50",
						Amount:       big.NewInt(42),
					},
				},
			},
		},
		{
			name: "it will return nil for an empty batch",
			receivedBatch: Batch{
				Nonce: big.NewInt(0),
			},
			expectedBatch: nil,
		},
	}

	for _, tt := range useCases {
		t.Run(tt.name, func(t *testing.T) {
			client := Client{
				bridgeContract: &bridgeContractStub{batch: tt.receivedBatch},
				gasLimit: GasLimit,
				log:            logger.GetOrCreate("testEthClient"),
			}

			got := client.GetPending(context.TODO())

			assert.Equal(t, tt.expectedBatch, got)
		})
	}
}

func TestProposeSetStatus(t *testing.T) {
	cases := []struct {
		status       uint8
		signatureHex string
	}{
		{bridge.Executed, "0x524957e3081d49d98c98881abd5cf6f737722a4aa0e7915771a567e3cb45cfc625cd9fcf9ec53c86182e517c1e61dbc076722905d11b73e1ed42665ec051342701"},
		{bridge.Rejected, "0xd9b1ae38d7e24837e90e7aaac2ae9ca1eb53dc7a30c41774ad7f7f5fd2371c2d0ac6e69643f6aaa25bd9b000dcf0b8be567bcde7f0a5fb5aad122273999bad2500"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("status %v should return the proper signature", c.status), func(t *testing.T) {
			broadcaster := &broadcasterStub{}
			client := Client{
				bridgeContract: &bridgeContractStub{},
				privateKey:     privateKey(t),
				broadcaster:    broadcaster,
				gasLimit: GasLimit,
				log:            logger.GetOrCreate("testEthClient"),
			}

			batch := &bridge.Batch{
				Id: bridge.NewBatchId(42),
				Transactions: []*bridge.DepositTransaction{{
					Status: c.status,
				}},
			}
			client.ProposeSetStatus(context.TODO(), batch)
			expectedSignature, _ := hexutil.Decode(c.signatureHex)

			assert.Equal(t, expectedSignature, broadcaster.lastBroadcastSignature)
			assert.Equal(t, []uint8{c.status}, client.lastProposedStatuses)
		})
	}
}

func TestProposeTransfer(t *testing.T) {
	broadcaster := &broadcasterStub{}
	client := Client{
		bridgeContract: &bridgeContractStub{},
		privateKey:     privateKey(t),
		broadcaster:    broadcaster,
		mapper:         &mapperStub{},
		gasLimit: GasLimit,
		log:            logger.GetOrCreate("testEthClient"),
	}

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
	_, _ = client.ProposeTransfer(context.TODO(), batch)
	expectedSignature, _ := hexutil.Decode("0xab3ce0cdc229afc9fcd0447800142da85aa116f16a26e151b9cad95b361ab73d24694ded888a06a1e9b731af8a1b549a1fc5188117e40bea11d9e74af4a6d5fa01")

	assert.Equal(t, expectedSignature, broadcaster.lastBroadcastSignature)
	assert.Equal(t, batch, client.lastTransferBatch)
}

func TestSignersCount(t *testing.T) {
	broadcaster := &broadcasterStub{lastBroadcastSignature: []byte("signature")}
	client := Client{
		bridgeContract: &bridgeContractStub{},
		broadcaster:    broadcaster,
		gasLimit: GasLimit,
		log:            logger.GetOrCreate("testEthClient"),
	}

	got := client.SignersCount(context.TODO(), bridge.NewActionId(0))

	assert.Equal(t, uint(1), got)
}

func TestWasExecuted(t *testing.T) {
	t.Run("when there is not last batch", func(t *testing.T) {
		contract := &bridgeContractStub{wasBatchFinished: true}
		client := Client{
			bridgeContract: contract,
			broadcaster:    &broadcasterStub{},
			gasLimit: GasLimit,
			log:            logger.GetOrCreate("testEthClient"),
		}

		got := client.WasExecuted(context.TODO(), bridge.NewActionId(0), bridge.NewBatchId(42))

		assert.Equal(t, true, got)
	})
	t.Run("when there is a last batch", func(t *testing.T) {
		contract := &bridgeContractStub{wasExecuted: true}
		client := Client{
			bridgeContract:    contract,
			lastTransferBatch: &bridge.Batch{},
			broadcaster:       &broadcasterStub{},
			gasLimit: GasLimit,
			log:               logger.GetOrCreate("testEthClient"),
		}

		got := client.WasExecuted(context.TODO(), bridge.NewActionId(0), bridge.NewBatchId(42))

		assert.Equal(t, true, got)
	})
	t.Run("when is true and there is last batch it will clean the state", func(t *testing.T) {
		client := Client{
			bridgeContract:       &bridgeContractStub{wasExecuted: true, wasBatchFinished: true},
			log:                  logger.GetOrCreate("testEthClient"),
			lastTransferBatch:    &bridge.Batch{},
			gasLimit: GasLimit,
			lastProposedStatuses: []uint8{bridge.Executed},
		}

		_ = client.WasExecuted(context.TODO(), nil, nil)

		assert.Nil(t, client.lastTransferBatch)
		assert.Equal(t, []uint8{bridge.Executed}, client.lastProposedStatuses)
	})
	t.Run("when is true and there is a last status it will clean the state", func(t *testing.T) {
		client := Client{
			bridgeContract:       &bridgeContractStub{wasExecuted: true},
			log:                  logger.GetOrCreate("testEthClient"),
			gasLimit: GasLimit,
			lastProposedStatuses: []uint8{bridge.Executed},
		}

		_ = client.WasExecuted(context.TODO(), nil, nil)

		assert.Empty(t, client.lastProposedStatuses)
	})
	t.Run("when is false and there is a last batch it will not clean the state", func(t *testing.T) {
		client := Client{
			bridgeContract:       &bridgeContractStub{wasExecuted: false},
			log:                  logger.GetOrCreate("testEthClient"),
			lastProposedStatuses: []uint8{bridge.Executed},
			gasLimit: GasLimit,
			lastTransferBatch:    &bridge.Batch{},
		}

		_ = client.WasExecuted(context.TODO(), nil, nil)

		assert.NotNil(t, client.lastTransferBatch)
	})
}

func TestExecute(t *testing.T) {
	t.Run("when there is no last transfer", func(t *testing.T) {
		expected := "0x029bc1fcae8ad9f887af3f37a9ebb223f1e535b009fc7ad7b053ba9b5ff666ae"
		contract := &bridgeContractStub{executedTransaction: types.NewTx(&types.AccessListTx{})}
		client := Client{
			bridgeContract:   contract,
			privateKey:       privateKey(t),
			publicKey:        publicKey(t),
			broadcaster:      &broadcasterStub{},
			blockchainClient: &blockchainClientStub{},
			log:              logger.GetOrCreate("testEthClient"),
			gasLimit: GasLimit,
		}
		batch := &bridge.Batch{Id: bridge.NewBatchId(42)}

		got, _ := client.Execute(context.TODO(), bridge.NewActionId(0), batch)

		assert.Equal(t, expected, got)
	})
	t.Run("when there is last transfer", func(t *testing.T) {
		expected := "0x029bc1fcae8ad9f887af3f37a9ebb223f1e535b009fc7ad7b053ba9b5ff666ae"
		contract := &bridgeContractStub{transferTransaction: types.NewTx(&types.AccessListTx{})}
		client := Client{
			bridgeContract:   contract,
			privateKey:       privateKey(t),
			publicKey:        publicKey(t),
			broadcaster:      &broadcasterStub{},
			mapper:           &mapperStub{},
			blockchainClient: &blockchainClientStub{},
			lastTransferBatch: &bridge.Batch{
				Id: bridge.NewBatchId(42),
				Transactions: []*bridge.DepositTransaction{{
					TokenAddress: "0x574554482d323936313238",
				}},
			},
			gasLimit: GasLimit,
			log: logger.GetOrCreate("testEthClient"),
		}
		batch := &bridge.Batch{Id: bridge.NewBatchId(42)}

		got, _ := client.Execute(context.TODO(), bridge.NewActionId(0), batch)

		assert.Equal(t, expected, got)
	})
}

func privateKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()

	privateKey, err := crypto.HexToECDSA(TestPrivateKey)
	if err != nil {
		t.Fatal(err)
		return nil
	}

	return privateKey
}

func publicKey(t *testing.T) *ecdsa.PublicKey {
	t.Helper()

	publicKey := privateKey(t).Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatal("error casting public key to ECDSA")
	}

	return publicKeyECDSA
}

type bridgeContractStub struct {
	batch               Batch
	wasExecuted         bool
	wasBatchFinished    bool
	executedTransaction *types.Transaction
	transferTransaction *types.Transaction
}

func (c *bridgeContractStub) GetNextPendingBatch(*bind.CallOpts) (Batch, error) {
	return c.batch, nil
}

func (c *bridgeContractStub) FinishCurrentPendingBatch(*bind.TransactOpts, *big.Int, []uint8, [][]byte) (*types.Transaction, error) {
	return c.executedTransaction, nil
}

func (c *bridgeContractStub) ExecuteTransfer(*bind.TransactOpts, []common.Address, []common.Address, []*big.Int, *big.Int, [][]byte) (*types.Transaction, error) {
	return c.transferTransaction, nil
}

func (c *bridgeContractStub) WasBatchExecuted(*bind.CallOpts, *big.Int) (bool, error) {
	return c.wasExecuted, nil
}

func (c *bridgeContractStub) WasBatchFinished(*bind.CallOpts, *big.Int) (bool, error) {
	return c.wasBatchFinished, nil
}

type broadcasterStub struct {
	lastBroadcastSignature []byte
}

func (b *broadcasterStub) SendSignature(signature []byte) {
	b.lastBroadcastSignature = signature
}

func (b *broadcasterStub) Signatures() [][]byte {
	return [][]byte{b.lastBroadcastSignature}
}

type blockchainClientStub struct{}

func (b *blockchainClientStub) PendingNonceAt(context.Context, common.Address) (uint64, error) {
	return 0, nil
}

func (b *blockchainClientStub) SuggestGasPrice(context.Context) (*big.Int, error) {
	return nil, nil
}

func (b *blockchainClientStub) ChainID(context.Context) (*big.Int, error) {
	return big.NewInt(42), nil
}

type mapperStub struct{}

func (m *mapperStub) GetTokenId(string) string {
	return "tokenId"
}

func (m *mapperStub) GetErc20Address(string) string {
	return "0x30C7c97471FB5C5238c946E549c608D27f37AAb8"
}
