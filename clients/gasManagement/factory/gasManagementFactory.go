package factory

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients/gasManagement"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients/gasManagement/disabled"
)

// CreateGasStation generates an implementation of GasHandler
func CreateGasStation(args gasManagement.ArgsGasStation, enabled bool) (clients.GasHandler, error) {
	if enabled {
		return gasManagement.NewGasStation(args)
	}
	return &disabled.DisabledGasStation{}, nil
}
