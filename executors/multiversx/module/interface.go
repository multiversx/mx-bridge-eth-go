package module

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-sdk-go/core"
)

type nonceTransactionsHandler interface {
	ApplyNonceAndGasPrice(ctx context.Context, address core.AddressHandler, tx *transaction.FrontendTransaction) error
	SendTransaction(ctx context.Context, tx *transaction.FrontendTransaction) (string, error)
	Close() error
	IsInterfaceNil() bool
}

type pollingHandler interface {
	StartProcessingLoop() error
	Close() error
	IsInterfaceNil() bool
}

type executor interface {
	Execute(ctx context.Context) error
	IsInterfaceNil() bool
}
