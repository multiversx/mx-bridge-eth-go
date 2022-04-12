package testsCommon

import (
	"context"
	"errors"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
)

type BatchValidatorStub struct {
	ValidateBatchCalled func(ctx context.Context, batch *clients.TransferBatch) (bool, error)
}

// ValidateBatch -
func (bvs *BatchValidatorStub) ValidateBatch(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
	if bvs.ValidateBatchCalled != nil {
		return bvs.ValidateBatchCalled(ctx, batch)
	}
	return false, errors.New("method not implemented")
}

// IsInterfaceNil -
func (bvs *BatchValidatorStub) IsInterfaceNil() bool {
	return bvs == nil
}
