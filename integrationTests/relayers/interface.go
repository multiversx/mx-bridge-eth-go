package relayers

import (
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ethereum/go-ethereum/common"
)

type bridgeComponents interface {
	ElrondRelayerAddress() erdgoCore.AddressHandler
	EthereumRelayerAddress() common.Address
	Start() error
	Close() error
}
