package factory

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/gasManagement"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/gasManagement/disabled"
)

// CreateGasStation generates an implementation of GasHandler
func CreateGasStation(args gasManagement.ArgsGasStation, enabled bool) (bridge.GasHandler, error) {
	if enabled {
		return gasManagement.NewGasStation(args)
	}
	return &disabled.DisabledGasStation{}, nil
}
