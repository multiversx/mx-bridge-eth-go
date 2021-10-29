package factory

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/gasManagement"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/gasManagement/disabled"
)

func CreateGasStation(args gasManagement.ArgsGasStation, enabled bool) (bridge.GasHandler, error) {
	if enabled {
		return gasManagement.NewGasStation(args)
	}
	return &disabled.DisabledGasStation{}, nil
}
