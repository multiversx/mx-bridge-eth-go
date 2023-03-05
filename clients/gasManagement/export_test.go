package gasManagement

import "github.com/multiversx/mx-bridge-eth-go/core"

// GetLatestGasPrice -
func (gs *gasStation) GetLatestGasPrice() int {
	gs.mut.RLock()
	defer gs.mut.RUnlock()

	return gs.latestGasPrice
}

// SetSelector -
func (gs *gasStation) SetSelector(gasPriceSelector core.EthGasPriceSelector) {
	gs.mut.Lock()
	defer gs.mut.Unlock()

	gs.gasPriceSelector = gasPriceSelector
}
