package testsCommon

import (
	"context"

	"github.com/multiversx/mx-sdk-go/data"
)

// TransactionExecutorStub -
type TransactionExecutorStub struct {
	ExecuteTransactionCalled    func(ctx context.Context, networkConfig *data.NetworkConfig, receiver string, transactionType string, gasLimit uint64, dataBytes []byte) error
	GetNumSentTransactionCalled func() uint32
}

// ExecuteTransaction -
func (stub *TransactionExecutorStub) ExecuteTransaction(
	ctx context.Context,
	networkConfig *data.NetworkConfig,
	receiver string,
	transactionType string,
	gasLimit uint64,
	dataBytes []byte,
) error {
	if stub.ExecuteTransactionCalled != nil {
		return stub.ExecuteTransactionCalled(ctx, networkConfig, receiver, transactionType, gasLimit, dataBytes)
	}

	return nil
}

// GetNumSentTransaction -
func (stub *TransactionExecutorStub) GetNumSentTransaction() uint32 {
	if stub.GetNumSentTransactionCalled != nil {
		return stub.GetNumSentTransactionCalled()
	}

	return 0
}

// IsInterfaceNil -
func (stub *TransactionExecutorStub) IsInterfaceNil() bool {
	return stub == nil
}
