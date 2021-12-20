package bridge

import (
	"context"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

// NonceTransactionsHandlerStub -
type NonceTransactionsHandlerStub struct {
	GetNonceCalled        func(ctx context.Context, address core.AddressHandler) (uint64, error)
	SendTransactionCalled func(ctx context.Context, tx *data.Transaction) (string, error)
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
func (stub *NonceTransactionsHandlerStub) SendTransaction(ctx context.Context, tx *data.Transaction) (string, error) {
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
