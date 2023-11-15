package bridge

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-sdk-go/core"
)

// NonceTransactionsHandlerStub -
type NonceTransactionsHandlerStub struct {
	GetNonceCalled        func(ctx context.Context, address core.AddressHandler) (uint64, error)
	SendTransactionCalled func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error)
	CloseCalled           func() error
}

// GetNonce -
func (stub *NonceTransactionsHandlerStub) GetNonce(ctx context.Context, address core.AddressHandler) (uint64, error) {
	if stub.GetNonceCalled != nil {
		return stub.GetNonceCalled(ctx, address)
	}

	return 0, nil
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
