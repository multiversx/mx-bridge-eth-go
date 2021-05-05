package eth

import (
	"context"
	"crypto/ecdsa"
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

func TestProposeSetStatusSuccessOnPendingTransfer(t *testing.T) {
	broadcaster := &broadcasterStub{}
	client := Client{
		bridgeContract: &bridgeContractStub{},
		privateKey:     privateKey(t),
		broadcaster:    broadcaster,
		log:            logger.GetOrCreate("testEthClient"),
	}

	client.ProposeSetStatusSuccessOnPendingTransfer(context.TODO(), bridge.NewNonce(1))
	expectedSignature, _ := hexutil.Decode("0x04f1148226b2902a5eac4631109996c2bc0af7a59b88483b3e67719ae1f1399320fc13b0a639cab0243dc5c5930f629244b5098cf1f6e1fdef102974a5ca0a8200")

	assert.Equal(t, expectedSignature, broadcaster.lastBroadcastSignature)
	assert.Equal(t, bridge.Executed, client.lastProposedStatus)
}

func TestProposeSetStatusFailedOnPendingTransfer(t *testing.T) {
	broadcaster := &broadcasterStub{}
	client := Client{
		bridgeContract: &bridgeContractStub{},
		privateKey:     privateKey(t),
		broadcaster:    broadcaster,
		log:            logger.GetOrCreate("testEthClient"),
	}

	client.ProposeSetStatusFailedOnPendingTransfer(context.TODO(), bridge.NewNonce(1))
	expectedSignature, _ := hexutil.Decode("0xf700e2f7a17879770f4a91cd044dd4c052d2cf04608fe6809ea6940b13795b76040301e51a7d8612afef89b0d15652b1a4a7351e7ba5123c8cc907b3be9eaaac01")

	assert.Equal(t, expectedSignature, broadcaster.lastBroadcastSignature)
	assert.Equal(t, bridge.Rejected, client.lastProposedStatus)
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
	contract := &bridgeContractStub{wasExecuted: true}
	client := Client{
		bridgeContract: contract,
		broadcaster:    &broadcasterStub{},
		log:            logger.GetOrCreate("testEthClient"),
	}

	got := client.WasExecuted(context.TODO(), bridge.NewActionId(0), bridge.NewNonce(42))

	assert.Equal(t, true, got)
}

func TestExecute(t *testing.T) {
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
	executedTransaction *types.Transaction
}

func (c *bridgeContractStub) GetNextPendingTransaction(*bind.CallOpts) (Deposit, error) {
	return c.deposit, nil
}

func (c *bridgeContractStub) FinishCurrentPendingTransaction(*bind.TransactOpts, *big.Int, uint8, [][]byte) (*types.Transaction, error) {
	return c.executedTransaction, nil
}

func (c *bridgeContractStub) WasTransactionExecuted(*bind.CallOpts, *big.Int) (bool, error) {
	return c.wasExecuted, nil
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
