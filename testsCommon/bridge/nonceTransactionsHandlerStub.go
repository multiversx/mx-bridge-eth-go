package bridge

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-sdk-go/core"
)

// NonceTransactionsHandlerStub -
type NonceTransactionsHandlerStub struct {
	ApplyNonceAndGasPriceCalled func(ctx context.Context, address core.AddressHandler, tx *transaction.FrontendTransaction) error
	SendTransactionCalled       func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error)
	CloseCalled                 func() error
}

// ApplyNonceAndGasPrice -
func (stub *NonceTransactionsHandlerStub) ApplyNonceAndGasPrice(ctx context.Context, address core.AddressHandler, tx *transaction.FrontendTransaction) error {
	if stub.ApplyNonceAndGasPriceCalled != nil {
		return stub.ApplyNonceAndGasPriceCalled(ctx, address, tx)
	}

	return nil
}

// SendTransaction -
func (stub *NonceTransactionsHandlerStub) SendTransaction(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
	if stub.SendTransactionCalled != nil {
		return stub.SendTransactionCalled(ctx, tx)
	}

	return "", nil
}

// Close -
func (stub *NonceTransactionsHandlerStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}

	return nil
}
