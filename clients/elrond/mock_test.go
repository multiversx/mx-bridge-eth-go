package elrond

import (
	"context"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
)

type txHandlerStub struct {
	sendTransactionReturnHashCalled func(ctx context.Context, builder builders.TxDataBuilder, gasLimit uint64) (string, error)
	closeCalled                     func() error
}

func (stub *txHandlerStub) sendTransactionReturnHash(ctx context.Context, builder builders.TxDataBuilder, gasLimit uint64) (string, error) {
	if stub.sendTransactionReturnHashCalled != nil {
		return stub.sendTransactionReturnHashCalled(ctx, builder, gasLimit)
	}

	return "", nil
}

func (stub *txHandlerStub) close() error {
	if stub.closeCalled != nil {
		return stub.closeCalled()
	}

	return nil
}
