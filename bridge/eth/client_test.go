package eth

import (
	"context"
	"math/big"
	"testing"

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
				contract: &bridgeContract{deposit: tt.receivedDeposit},
				log:      logger.GetOrCreate("testEthClient"),
			}

			got := client.GetPendingDepositTransaction(context.TODO())

			assert.Equal(t, tt.expectedDeposit, got)
		})
	}
}

type bridgeContract struct {
	deposit Deposit
}

func (c *bridgeContract) GetNextPendingTransaction(*bind.CallOpts) (Deposit, error) {
	return c.deposit, nil
}
