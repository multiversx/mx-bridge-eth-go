package bridge

import (
	"context"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
)

// TxHandlerStub -
type TxHandlerStub struct {
	SendTransactionReturnHashCalled func(ctx context.Context, builder builders.TxDataBuilder, gasLimit uint64) (string, error)
	CloseCalled                     func() error
}

// SendTransactionReturnHash -
func (stub *TxHandlerStub) SendTransactionReturnHash(ctx context.Context, builder builders.TxDataBuilder, gasLimit uint64) (string, error) {
	if stub.SendTransactionReturnHashCalled != nil {
		return stub.SendTransactionReturnHashCalled(ctx, builder, gasLimit)
	}

	return "", nil
}

// Close -
func (stub *TxHandlerStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}

	return nil
}
