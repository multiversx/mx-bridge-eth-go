package relay

import (
	"context"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/elrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth"
)

type Relay struct {
	ethBridge    bridge.Bridge
	elrondBridge bridge.Bridge
}

func NewRelay(config *Config) (*Relay, error) {
	ethBridge, err := eth.NewClient(config.Eth)
	if err != nil {
		return nil, err
	}

	elrondBridge, err := elrond.NewClient(config.Elrond)
	if err != nil {
		return nil, err
	}

	return &Relay{
		ethBridge:    ethBridge,
		elrondBridge: elrondBridge,
	}, nil
}

func (r *Relay) Start(context.Context) {
}

func (r *Relay) Stop() {
}
