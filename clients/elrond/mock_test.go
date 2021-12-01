package elrond

import (
	"context"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
)

type txHandlerStub struct {
	sendTransactionReturningHashCalled func(ctx context.Context, builder builders.TxDataBuilder, gasLimit uint64) (string, error)
	closeCalled                        func() error
}

func (stub *txHandlerStub) sendTransactionReturningHash(ctx context.Context, builder builders.TxDataBuilder, gasLimit uint64) (string, error) {
	if stub.sendTransactionReturningHashCalled != nil {
		return stub.sendTransactionReturningHashCalled(ctx, builder, gasLimit)
	}

	return "", nil
}

func (stub *txHandlerStub) close() error {
	if stub.closeCalled != nil {
		return stub.closeCalled()
	}

	return nil
}
