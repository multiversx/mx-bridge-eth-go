package disabled

import (
	"context"
	"math/big"
)

// DisabledGasStation implementation in case no gasStation is used
type DisabledGasStation struct{}

// GetCurrentGasPrice returns nil,nil and will cause the gas price to be determined automatically
func (dgs *DisabledGasStation) GetCurrentGasPrice() (*big.Int, error) {
	return nil, nil
}

// Execute returns nil
func (dgs *DisabledGasStation) Execute(_ context.Context) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dgs *DisabledGasStation) IsInterfaceNil() bool {
	return dgs == nil
}
