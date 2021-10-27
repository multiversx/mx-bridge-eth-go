package gasManagement

import "github.com/ElrondNetwork/elrond-eth-bridge/core"

// GetLatestResponse -
func (gs *gasStation) GetLatestResponse() *gasStationResponse {
	gs.mut.RLock()
	defer gs.mut.RUnlock()

	return gs.latestResponse
}

// SetSelector -
func (gs *gasStation) SetSelector(gasPriceSelector core.EthGasPriceSelector) {
	gs.mut.Lock()
	defer gs.mut.Unlock()

	gs.gasPriceSelector = gasPriceSelector
}
