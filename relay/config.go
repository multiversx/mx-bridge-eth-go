package relay

import "github.com/ElrondNetwork/elrond-eth-bridge/bridge"

type Config struct {
	Eth    bridge.Config
	Elrond bridge.Config
	P2P    ConfigP2P
}

type ConfigP2P struct {
	Port            string
	Seed            string
	InitialPeerList []string
	ProtocolID      string
}
