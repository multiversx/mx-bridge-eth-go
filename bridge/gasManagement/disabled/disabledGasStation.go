package disabled

import "math/big"

type DisabledGasStation struct{}

func (dgs *DisabledGasStation) GetCurrentGasPrice() (*big.Int, error) {
	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dgs *DisabledGasStation) IsInterfaceNil() bool {
	return dgs == nil
}
