package testsCommon

import (
	"context"
	"math/big"
)

// GasHandlerStub -
type GasHandlerStub struct {
	ExecuteCalled                 func(ctx context.Context) error
	GetCurrentGasPriceInWeiCalled func() (*big.Int, error)
}

// Execute -
func (ghs *GasHandlerStub) Execute(ctx context.Context) error {
	if ghs.ExecuteCalled != nil {
		return ghs.ExecuteCalled(ctx)
	}

	return nil
}

// GetCurrentGasPriceInWei -
func (ghs *GasHandlerStub) GetCurrentGasPriceInWei() (*big.Int, error) {
	if ghs.GetCurrentGasPriceInWeiCalled != nil {
		return ghs.GetCurrentGasPriceInWeiCalled()
	}

	return big.NewInt(0), nil
}

// IsInterfaceNil -
func (ghs *GasHandlerStub) IsInterfaceNil() bool {
	return ghs == nil
}
