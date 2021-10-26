package mock

// GasHandlerStub -
type GasHandlerStub struct {
	GetCurrentGasPriceCalled func() (int, error)
}

// GetCurrentGasPrice -
func (ghs *GasHandlerStub) GetCurrentGasPrice() (int, error) {
	if ghs.GetCurrentGasPriceCalled != nil {
		return ghs.GetCurrentGasPriceCalled()
	}

	return 0, nil
}

// IsInterfaceNil -
func (ghs *GasHandlerStub) IsInterfaceNil() bool {
	return ghs == nil
}
