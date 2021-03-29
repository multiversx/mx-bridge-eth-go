package bridge

import (
	"context"
	"fmt"
	"github.com/ElrondNetwork/elrond-eth-bridge/safe"
	"github.com/ElrondNetwork/elrond-eth-bridge/safe/elrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/safe/eth"
	"math/big"
)

type Bridge struct {
	ethSafe    safe.Safe
	elrondSafe safe.Safe

	ethChannel    safe.SafeTxChan
	elrondChannel safe.SafeTxChan
}

func NewBridge(ethNetworkAddress, ethSafeAddress string) (*Bridge, error) {
	ethSafe, err := eth.NewClient(ethNetworkAddress, ethSafeAddress)
	if err != nil {
		return nil, err
	}

	elrondSafe, err := elrond.NewClient()
	if err != nil {
		return nil, err
	}

	return &Bridge{
		ethSafe:       ethSafe,
		elrondSafe:    elrondSafe,
		ethChannel:    make(safe.SafeTxChan),
		elrondChannel: make(safe.SafeTxChan),
	}, nil
}

func (b *Bridge) Start(ctx context.Context) {
	var lastProcessedEthBlock big.Int
	var lastProcessedElrondBlock big.Int

	go b.ethSafe.GetTransactions(ctx, &lastProcessedEthBlock, b.ethChannel)
	go b.elrondSafe.GetTransactions(ctx, &lastProcessedElrondBlock, b.elrondChannel)
}

func (b *Bridge) Monitor() {
	select {
	case tx := <-b.ethChannel:
		b.bridgeToElrond(tx)
	case tx := <-b.elrondChannel:
		b.bridgeToEth(tx)
	}
}

func (b *Bridge) Stop() {
	close(b.ethChannel)
	close(b.elrondChannel)
}

func (b *Bridge) bridgeToElrond(tx *safe.DepositTransaction) {
	// TODO: log
	fmt.Printf("Briging %v to elrond\n", tx)
	b.elrondSafe.Bridge(tx)
}

func (b *Bridge) bridgeToEth(tx *safe.DepositTransaction) {
	// TODO: log
	fmt.Printf("Briging %v to eth\n", tx)
	b.ethSafe.Bridge(tx)
}
