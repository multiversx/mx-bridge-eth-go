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

func NewBridge(ethNetworkAddress, ethSafeAddress, elrondNetworkAddress, elrondSafeAddress, elrondPrivateKeyPath string) (*Bridge, error) {
	ethSafe, err := eth.NewClient(ethNetworkAddress, ethSafeAddress)
	if err != nil {
		return nil, err
	}

	elrondSafe, err := elrond.NewClient(elrondNetworkAddress, elrondSafeAddress, elrondPrivateKeyPath)
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
	lastProcessedEthBlock := big.NewInt(27)
	var lastProcessedElrondBlock big.Int

	go b.ethSafe.GetTransactions(ctx, lastProcessedEthBlock, b.ethChannel)
	go b.elrondSafe.GetTransactions(ctx, &lastProcessedElrondBlock, b.elrondChannel)

	b.monitor(ctx)
}

func (b *Bridge) monitor(ctx context.Context) {
	for {
		select {
		case tx := <-b.ethChannel:
			hash, err := b.bridgeToElrond(tx)

			// TODO: log
			if err != nil {
				fmt.Printf("Briging to elrond failed with %v\n", err)
			} else {
				fmt.Printf("Bridged to elrond with hash: %q\n", hash)
			}
		case tx := <-b.elrondChannel:
			_, err := b.bridgeToEth(tx)
			// TODO: log
			if err != nil {
				fmt.Printf("Briging to ethereum failed with %v\n", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (b *Bridge) Stop() {
	close(b.ethChannel)
	close(b.elrondChannel)
}

func (b *Bridge) bridgeToElrond(tx *safe.DepositTransaction) (string, error) {
	// TODO: log
	fmt.Printf("Briging %v to elrond\n", tx)
	return b.elrondSafe.Bridge(tx)
}

func (b *Bridge) bridgeToEth(tx *safe.DepositTransaction) (string, error) {
	// TODO: log
	fmt.Printf("Briging %v to eth\n", tx)
	return b.ethSafe.Bridge(tx)
}
