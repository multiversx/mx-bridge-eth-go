package relayers

import (
	sdkCore "github.com/multiversx/mx-sdk-go/core"
)

type bridgeComponents interface {
	MultiversXRelayerAddress() sdkCore.AddressHandler
	PeerChainRelayerAddress() string
	Start() error
	Close() error
}
