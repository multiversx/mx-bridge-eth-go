package disabled

import "math/big"

// DisabledGasStation implementation in case no gasStation is used
type DisabledGasStation struct{}

// GetCurrentGasPrice returns nil,nil and will cause the gas price to be determined automatically
func (dgs *DisabledGasStation) GetCurrentGasPrice() (*big.Int, error) {
	return big.NewInt(0), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dgs *DisabledGasStation) IsInterfaceNil() bool {
	return dgs == nil
}
