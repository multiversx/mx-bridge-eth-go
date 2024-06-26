package disabled

import "math/big"

const defaultDisabledGasPrice = 1000

// DisabledGasStation implementation in case no gasStation is used
type DisabledGasStation struct{}

// GetCurrentGasPrice returns a default value
func (dgs *DisabledGasStation) GetCurrentGasPrice() (*big.Int, error) {
	return big.NewInt(defaultDisabledGasPrice), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dgs *DisabledGasStation) IsInterfaceNil() bool {
	return dgs == nil
}
