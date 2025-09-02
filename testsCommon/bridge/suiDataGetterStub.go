package bridge

import (
	"context"

	"github.com/block-vision/sui-go-sdk/models"
)

// SuiDataGetterStub -
type SuiDataGetterStub struct {
	GetRelayersCalled func(ctx context.Context) ([]models.SuiAddress, error)
}

// GetRelayers -
func (stub *SuiDataGetterStub) GetRelayers(ctx context.Context) ([]models.SuiAddress, error) {
	if stub.GetRelayersCalled != nil {
		return stub.GetRelayersCalled(ctx)
	}
	return nil, nil
}

// IsInterfaceNil -
func (stub *SuiDataGetterStub) IsInterfaceNil() bool {
	return stub == nil
}
