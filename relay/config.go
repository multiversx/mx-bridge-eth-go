package relay

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-go/config"
)

type Config struct {
	Eth     bridge.Config
	Elrond  bridge.Config
	P2P     ConfigP2P
	Relayer ConfigRelayer
}

type ConfigP2P struct {
	Port            string
	Seed            string
	InitialPeerList []string
	ProtocolID      string
}

type ConfigRelayer struct {
	Marshalizer config.MarshalizerConfig
	Antiflood   config.WebServerAntifloodConfig
}
