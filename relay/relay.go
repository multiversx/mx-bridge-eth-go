package relay

import (
	"context"
	"fmt"
	"github.com/ElrondNetwork/elrond-eth-bridge/safe"
	"github.com/ElrondNetwork/elrond-eth-bridge/safe/elrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/safe/eth"
	"math/big"
)

type Relay struct {
	ethSafe    safe.Safe
	elrondSafe safe.Safe

	ethChannel    safe.SafeTxChan
	elrondChannel safe.SafeTxChan
}

func NewRelay(ethNetworkAddress, ethSafeAddress, elrondNetworkAddress, elrondSafeAddress, elrondPrivateKeyPath string) (*Relay, error) {
	ethSafe, err := eth.NewClient(ethNetworkAddress, ethSafeAddress)
	if err != nil {
		return nil, err
	}

	elrondSafe, err := elrond.NewClient(elrondNetworkAddress, elrondSafeAddress, elrondPrivateKeyPath)
	if err != nil {
		return nil, err
	}

	return &Relay{
		ethSafe:       ethSafe,
		elrondSafe:    elrondSafe,
		ethChannel:    make(safe.SafeTxChan),
		elrondChannel: make(safe.SafeTxChan),
	}, nil
}

func (r *Relay) Start(ctx context.Context) {
	var lastProcessedEthBlock big.Int
	var lastProcessedElrondBlock big.Int

	go r.ethSafe.GetTransactions(ctx, &lastProcessedEthBlock, r.ethChannel)
	go r.elrondSafe.GetTransactions(ctx, &lastProcessedElrondBlock, r.elrondChannel)

	r.monitor(ctx)
}

func (r *Relay) monitor(ctx context.Context) {
	for {
		select {
		case tx := <-r.ethChannel:
			hash, err := r.bridgeToElrond(tx)

			// TODO: log
			if err != nil {
				fmt.Printf("Briging to elrond failed with %v\n", err)
			} else {
				fmt.Printf("Bridged to elrond with hash: %q\n", hash)
			}
		case tx := <-r.elrondChannel:
			_, err := r.bridgeToEth(tx)
			// TODO: log
			if err != nil {
				fmt.Printf("Briging to ethereum failed with %v\n", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (r *Relay) Stop() {
	close(r.ethChannel)
	close(r.elrondChannel)
}

func (r *Relay) bridgeToElrond(tx *safe.DepositTransaction) (string, error) {
	// TODO: log
	fmt.Printf("Briging %v to elrond\n", tx)
	return r.elrondSafe.Bridge(tx)
}

func (r *Relay) bridgeToEth(tx *safe.DepositTransaction) (string, error) {
	// TODO: log
	fmt.Printf("Briging %v to eth\n", tx)
	return r.ethSafe.Bridge(tx)
}
