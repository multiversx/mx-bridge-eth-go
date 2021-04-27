package eth

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

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
				DepositNonce: 1,
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
				contract: &contractStub{deposit: tt.receivedDeposit},
				log:      logger.GetOrCreate("testEthClient"),
			}

			got := client.GetPendingDepositTransaction(context.TODO())

			assert.Equal(t, tt.expectedDeposit, got)
		})
	}
}

func TestProposeSetStatusSuccessOnPendingTransfer(t *testing.T) {
	broadcaster := &broadcasterStub{}
	client := Client{
		contract:    &contractStub{},
		privateKey:  privateKey(t),
		broadcaster: broadcaster,
		log:         logger.GetOrCreate("testEthClient"),
	}

	client.ProposeSetStatusSuccessOnPendingTransfer(context.TODO())
	expectedSignature := "0x5de3f8db48b3e0b903e36b92854f97e1ce3f095e28343b59149150a81a7cdec95073ae0c52848734d2aedf511517cf18042300ec7e855aa3857a2842202a8fe400"
	expectedData := "\u0019Ethereum Signed Message:\n27CurrentPendingTransaction:3"

	assert.Equal(t, expectedSignature, broadcaster.lastBroadcastSignature)
	assert.Equal(t, expectedData, broadcaster.lastSignData)
}

func TestProposeSetStatusFailedOnPendingTransfer(t *testing.T) {
	broadcaster := &broadcasterStub{}
	client := Client{
		contract:    &contractStub{},
		privateKey:  privateKey(t),
		broadcaster: broadcaster,
		log:         logger.GetOrCreate("testEthClient"),
	}

	client.ProposeSetStatusFailedOnPendingTransfer(context.TODO())
	expectedSignature := "0x8b4a6ddb7362e158dc86e5e3bab9b325d2b7b897c4d51268c13551b501e77c852b14f0ff5be22eac8b82bc88b7f846afe52959027f1129d953c7982b165d0eaa00"
	expectedData := "\u0019Ethereum Signed Message:\n27CurrentPendingTransaction:4"

	assert.Equal(t, expectedSignature, broadcaster.lastBroadcastSignature)
	assert.Equal(t, expectedData, broadcaster.lastSignData)
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

type contractStub struct {
	deposit Deposit
}

func (c *contractStub) GetNextPendingTransaction(*bind.CallOpts) (Deposit, error) {
	return c.deposit, nil
}

func (c *contractStub) FinishCurrentPendingTransaction(_ *bind.TransactOpts, _ string, _ [][]byte) (*types.Transaction, error) {
	return nil, nil
}

type broadcasterStub struct {
	lastSignData           string
	lastBroadcastSignature string
}

func (b *broadcasterStub) SendSignature(signData, signature string) {
	b.lastSignData = signData
	b.lastBroadcastSignature = signature
}

func (b *broadcasterStub) Signatures() [][]byte {
	return [][]byte{[]byte(b.lastBroadcastSignature)}
}

func (b *broadcasterStub) SignData() string {
	return b.lastSignData
}
