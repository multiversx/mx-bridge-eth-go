package executors

import (
	"context"

	"github.com/multiversx/mx-bridge-eth-go/parsers"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type scCallExecutor struct {
	proxy Proxy
	log   logger.Logger
}

// Execute will execute one step: get all pending operations, call the filter and send execution transactions
func (executor *scCallExecutor) Execute(ctx context.Context) error {
	listOfPendingOperations, err := executor.getPendingOperations(ctx)
	if err != nil {
		return err
	}

	listOfPendingOperations = executor.filterOperations(listOfPendingOperations)

	return executor.executeOperations(listOfPendingOperations)
}

func (executor *scCallExecutor) getPendingOperations(ctx context.Context) ([]parsers.ProxySCCompleteCallData, error) {

}

// IsInterfaceNil returns true if there is no value under the interface
func (executor *scCallExecutor) IsInterfaceNil() bool {
	return executor == nil
}
