package eth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ElrondNetwork/elrond-eth-bridge/testHelpers"

	"github.com/ethereum/go-ethereum/common"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
)

// verify Client implements interface
var (
	_ = bridge.Bridge(&Client{})
)

const TestPrivateKey = "60f3849d7c8d93dfce1947d17c34be3e4ea974e74e15ce877f0df34d7192efab"

func TestGetPendingDepositTransaction(t *testing.T) {
	testHelpers.SetTestLogLevel()

	useCases := []struct {
		name            string
		receivedDeposit Deposit
		expectedDeposit *bridge.DepositTransaction
	}{
		{
			name: "it will map a non empty transaction",
			receivedDeposit: Deposit{
				Nonce:        big.NewInt(1),
				TokenAddress: common.HexToAddress("0x093c0B280ba430A9Cc9C3649FF34FCBf6347bC50"),
				Amount:       big.NewInt(42),
				Depositor:    common.HexToAddress("0x132A150926691F08a693721503a38affeD18d524"),
				Recipient:    []byte("erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8"),
				Status:       0,
			},
			expectedDeposit: &bridge.DepositTransaction{
				To:           "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8",
				From:         "0x132A150926691F08a693721503a38affeD18d524",
				TokenAddress: "0x093c0B280ba430A9Cc9C3649FF34FCBf6347bC50",
				Amount:       big.NewInt(42),
				DepositNonce: big.NewInt(1),
			},
		},
		{
			name: "it will return nil for an empty transaction",
			receivedDeposit: Deposit{
				Nonce:        big.NewInt(0),
				TokenAddress: common.Address{},
				Amount:       big.NewInt(0),
				Depositor:    common.Address{},
				Recipient:    []byte(""),
				Status:       0,
			},
			expectedDeposit: nil,
		},
	}

	for _, tt := range useCases {
		t.Run(tt.name, func(t *testing.T) {
			client := Client{
				bridgeContract: &bridgeContractStub{deposit: tt.receivedDeposit},
				log:            logger.GetOrCreate("testEthClient"),
			}

			got := client.GetPendingDepositTransaction(context.TODO())

			assert.Equal(t, tt.expectedDeposit, got)
		})
	}
}

func TestProposeSetStatus(t *testing.T) {
	cases := []struct {
		status       uint8
		signatureHex string
	}{
		{bridge.Executed, "0x04f1148226b2902a5eac4631109996c2bc0af7a59b88483b3e67719ae1f1399320fc13b0a639cab0243dc5c5930f629244b5098cf1f6e1fdef102974a5ca0a8200"},
		{bridge.Rejected, "0xf700e2f7a17879770f4a91cd044dd4c052d2cf04608fe6809ea6940b13795b76040301e51a7d8612afef89b0d15652b1a4a7351e7ba5123c8cc907b3be9eaaac01"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("status %v should return the proper signature", c.status), func(t *testing.T) {
			broadcaster := &broadcasterStub{}
			client := Client{
				bridgeContract: &bridgeContractStub{},
				privateKey:     privateKey(t),
				broadcaster:    broadcaster,
				log:            logger.GetOrCreate("testEthClient"),
			}

			client.ProposeSetStatus(context.TODO(), c.status, bridge.NewNonce(1))
			expectedSignature, _ := hexutil.Decode(c.signatureHex)

			assert.Equal(t, expectedSignature, broadcaster.lastBroadcastSignature)
			assert.Equal(t, c.status, client.lastProposedStatus)
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
		log:            logger.GetOrCreate("testEthClient"),
	}

	tx := &bridge.DepositTransaction{
		To:           "cf95254084ab772696643f0e05ac4711ed674ac1",
		From:         "04aa6d6029b4e136d04848f5b588c2951185666cc871982994f7ef1654282fa3",
		TokenAddress: "574554482d323936313238",
		Amount:       big.NewInt(1),
		DepositNonce: bridge.NewNonce(2),
	}
	_, _ = client.ProposeTransfer(context.TODO(), tx)
	expectedSignature, _ := hexutil.Decode("0x94ab570d96c659ebabcb8e00657c04b9159a157f873bd23ece55bb7a958f88c859fbb99f6157d81cf2759c46774bc80a24a64f05b01cd796648ec8bb0f3ced6701")

	assert.Equal(t, expectedSignature, broadcaster.lastBroadcastSignature)
	assert.Equal(t, tx, client.lastTransferBatch)
}

func TestSignersCount(t *testing.T) {
	broadcaster := &broadcasterStub{lastBroadcastSignature: []byte("signature")}
	client := Client{
		bridgeContract: &bridgeContractStub{},
		broadcaster:    broadcaster,
		log:            logger.GetOrCreate("testEthClient"),
	}

	got := client.SignersCount(context.TODO(), bridge.NewActionId(0))

	assert.Equal(t, uint(1), got)
}

func TestWasExecuted(t *testing.T) {
	t.Run("when there is not last transaction", func(t *testing.T) {
		contract := &bridgeContractStub{wasExecuted: true}
		client := Client{
			bridgeContract: contract,
			broadcaster:    &broadcasterStub{},
			log:            logger.GetOrCreate("testEthClient"),
		}

		got := client.WasExecuted(context.TODO(), bridge.NewActionId(0), bridge.NewNonce(42))

		assert.Equal(t, true, got)
	})
	t.Run("when there is a last transaction", func(t *testing.T) {
		contract := &bridgeContractStub{wasTransferExecuted: true}
		client := Client{
			bridgeContract:    contract,
			lastTransferBatch: &bridge.DepositTransaction{},
			broadcaster:       &broadcasterStub{},
			log:               logger.GetOrCreate("testEthClient"),
		}

		got := client.WasExecuted(context.TODO(), bridge.NewActionId(0), bridge.NewNonce(42))

		assert.Equal(t, true, got)
	})
	t.Run("when is true and there is last transaction it will clean the state", func(t *testing.T) {
		client := Client{
			bridgeContract:     &bridgeContractStub{wasExecuted: true, wasTransferExecuted: true},
			log:                logger.GetOrCreate("testEthClient"),
			lastTransferBatch:  &bridge.DepositTransaction{},
			lastProposedStatus: bridge.Executed,
		}

		_ = client.WasExecuted(context.TODO(), nil, nil)

		assert.Nil(t, client.lastTransferBatch)
		assert.Equal(t, client.lastProposedStatus, bridge.Executed)
	})
	t.Run("when is true and there is a last status it will clean the state", func(t *testing.T) {
		client := Client{
			bridgeContract:     &bridgeContractStub{wasExecuted: true},
			log:                logger.GetOrCreate("testEthClient"),
			lastProposedStatus: bridge.Executed,
		}

		_ = client.WasExecuted(context.TODO(), nil, nil)

		assert.Equal(t, client.lastProposedStatus, uint8(0))
	})
	t.Run("when is false and there is a last transaction it will not clean the state", func(t *testing.T) {
		client := Client{
			bridgeContract:     &bridgeContractStub{wasExecuted: true},
			log:                logger.GetOrCreate("testEthClient"),
			lastProposedStatus: bridge.Executed,
			lastTransferBatch:  &bridge.DepositTransaction{},
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
		}

		got, _ := client.Execute(context.TODO(), bridge.NewActionId(0), bridge.NewNonce(42))

		assert.Equal(t, expected, got)
	})
	t.Run("when there is last transfer", func(t *testing.T) {
		expected := "0x029bc1fcae8ad9f887af3f37a9ebb223f1e535b009fc7ad7b053ba9b5ff666ae"
		contract := &bridgeContractStub{transferTransaction: types.NewTx(&types.AccessListTx{})}
		client := Client{
			bridgeContract:    contract,
			privateKey:        privateKey(t),
			publicKey:         publicKey(t),
			broadcaster:       &broadcasterStub{},
			mapper:            &mapperStub{},
			blockchainClient:  &blockchainClientStub{},
			lastTransferBatch: &bridge.DepositTransaction{TokenAddress: "0x574554482d323936313238"},
			log:               logger.GetOrCreate("testEthClient"),
		}

		got, _ := client.Execute(context.TODO(), bridge.NewActionId(0), bridge.NewNonce(42))

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
	deposit             Deposit
	wasExecuted         bool
	wasTransferExecuted bool
	executedTransaction *types.Transaction
	transferTransaction *types.Transaction
}

func (c *bridgeContractStub) GetNextPendingTransaction(*bind.CallOpts) (Deposit, error) {
	return c.deposit, nil
}

func (c *bridgeContractStub) FinishCurrentPendingTransaction(*bind.TransactOpts, *big.Int, uint8, [][]byte) (*types.Transaction, error) {
	return c.executedTransaction, nil
}

func (c *bridgeContractStub) ExecuteTransfer(*bind.TransactOpts, common.Address, common.Address, *big.Int, *big.Int, [][]byte) (*types.Transaction, error) {
	return c.transferTransaction, nil
}

func (c *bridgeContractStub) WasTransactionExecuted(*bind.CallOpts, *big.Int) (bool, error) {
	return c.wasExecuted, nil
}

func (c *bridgeContractStub) WasTransferExecuted(*bind.CallOpts, *big.Int) (bool, error) {
	return c.wasTransferExecuted, nil
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
