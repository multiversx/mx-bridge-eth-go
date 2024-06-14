package relayers

import (
	"github.com/ethereum/go-ethereum/common"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
)

type bridgeComponents interface {
	MultiversXRelayerAddress() sdkCore.AddressHandler
	EthereumRelayerAddress() common.Address
	Start() error
	Close() error
}
