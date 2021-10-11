package contracts

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ethereum/go-ethereum/common"
)

type transferRequest struct {
	depositor    []byte
	to           core.AddressHandler
	tokenAddress []byte
	amount       *big.Int
	nonce        int
}

func (tr *transferRequest) toEthDepositInfo() eth.Deposit {
	return eth.Deposit{
		Nonce:        big.NewInt(int64(tr.nonce)),
		TokenAddress: common.BytesToAddress(tr.tokenAddress),
		Amount:       tr.amount,
		Depositor:    common.BytesToAddress(tr.depositor),
		Recipient:    []byte(tr.to.AddressAsBech32String()),
		Status:       0,
	}
}
