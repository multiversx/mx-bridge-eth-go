package relay

import "github.com/ElrondNetwork/elrond-eth-bridge/bridge"

type Config struct {
	Eth    bridge.Config
	Elrond bridge.Config
}
