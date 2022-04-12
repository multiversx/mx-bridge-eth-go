package testsCommon

import (
	"errors"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
)

type BatchValidatorStub struct {
	ValidateBatchCalled func(batch *clients.TransferBatch) (bool, error)
}

// ValidateBatch -
func (bvs *BatchValidatorStub) ValidateBatch(batch *clients.TransferBatch) (bool, error) {
	if bvs.ValidateBatchCalled != nil {
		return bvs.ValidateBatchCalled(batch)
	}
	return false, errors.New("method not implemented")
}

// IsInterfaceNil -
func (bvs *BatchValidatorStub) IsInterfaceNil() bool {
	return bvs == nil
}
