package factory

import (
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/clients/gasManagement"
	"github.com/multiversx/mx-bridge-eth-go/clients/gasManagement/disabled"
)

// CreateGasStation generates an implementation of GasHandler
func CreateGasStation(args gasManagement.ArgsGasStation, enabled bool) (clients.GasHandler, error) {
	if enabled {
		return gasManagement.NewGasStation(args)
	}
	return &disabled.DisabledGasStation{}, nil
}
