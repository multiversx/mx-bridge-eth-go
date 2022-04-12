package disabled

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
)

type disabledBatchValidator struct{}

// NewDisabledBatchValidator will return a disabled batch validator instance
func NewDisabledBatchValidator() *disabledBatchValidator {
	return &disabledBatchValidator{}
}

// ValidateBatch returns true,nil and will result in skipping batch validation
func (dbv *disabledBatchValidator) ValidateBatch(_ context.Context, _ *clients.TransferBatch) (bool, error) {
	return true, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dbv *disabledBatchValidator) IsInterfaceNil() bool {
	return dbv == nil
}
