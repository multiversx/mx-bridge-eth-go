package testsCommon

import "math/big"

// GasHandlerStub -
type GasHandlerStub struct {
	GetCurrentGasPriceCalled func() (*big.Int, error)
}

// GetCurrentGasPrice -
func (ghs *GasHandlerStub) GetCurrentGasPrice() (*big.Int, error) {
	if ghs.GetCurrentGasPriceCalled != nil {
		return ghs.GetCurrentGasPriceCalled()
	}

	return big.NewInt(0), nil
}

// Close -
func (ghs *GasHandlerStub) Close() error {
	return nil
}

// IsInterfaceNil -
func (ghs *GasHandlerStub) IsInterfaceNil() bool {
	return ghs == nil
}
