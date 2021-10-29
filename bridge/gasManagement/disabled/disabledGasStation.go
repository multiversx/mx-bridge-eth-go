package disabled

import "math/big"

// DisabledGasStation implementation in case no gasStation is used
type DisabledGasStation struct{}

// GetCurrentGasPrice return nil; gas price will be automatically determined
func (dgs *DisabledGasStation) GetCurrentGasPrice() (*big.Int, error) {
	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dgs *DisabledGasStation) IsInterfaceNil() bool {
	return dgs == nil
}
