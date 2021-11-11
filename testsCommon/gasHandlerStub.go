package testsCommon

import (
	"context"
	"math/big"
)

// GasHandlerStub -
type GasHandlerStub struct {
	ExecuteCalled            func(ctx context.Context) error
	GetCurrentGasPriceCalled func() (*big.Int, error)
}

// Execute -
func (ghs *GasHandlerStub) Execute(ctx context.Context) error {
	if ghs.ExecuteCalled != nil {
		return ghs.ExecuteCalled(ctx)
	}

	return nil
}

// GetCurrentGasPrice -
func (ghs *GasHandlerStub) GetCurrentGasPrice() (*big.Int, error) {
	if ghs.GetCurrentGasPriceCalled != nil {
		return ghs.GetCurrentGasPriceCalled()
	}

	return big.NewInt(0), nil
}

// IsInterfaceNil -
func (ghs *GasHandlerStub) IsInterfaceNil() bool {
	return ghs == nil
}
