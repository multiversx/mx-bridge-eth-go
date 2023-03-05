package factory

import (
	"github.com/multiversx/mx-bridge-eth-go/clients"
	batchValidatorManagement "github.com/multiversx/mx-bridge-eth-go/clients/batchValidator"
	"github.com/multiversx/mx-bridge-eth-go/clients/batchValidator/disabled"
)

// CreateBatchValidator generates an implementation of BatchValidator
func CreateBatchValidator(args batchValidatorManagement.ArgsBatchValidator, enabled bool) (clients.BatchValidator, error) {
	if enabled {
		return batchValidatorManagement.NewBatchValidator(args)
	}
	return disabled.NewDisabledBatchValidator(), nil
}
