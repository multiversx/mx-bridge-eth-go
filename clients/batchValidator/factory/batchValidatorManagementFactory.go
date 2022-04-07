package factory

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	batchValidatorManagement "github.com/ElrondNetwork/elrond-eth-bridge/clients/batchValidator"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients/batchValidator/disabled"
)

// CreateBatchValidator generates an implementation of BatchValidator
func CreateBatchValidator(args batchValidatorManagement.ArgsBatchValidator, enabled bool) (clients.BatchValidator, error) {
	if enabled {
		return batchValidatorManagement.NewBatchValidator(args)
	}
	return &disabled.DisabledBatchValidator{}, nil
}
