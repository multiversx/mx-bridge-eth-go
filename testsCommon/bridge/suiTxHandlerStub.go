package bridge

import (
	"context"

	"github.com/block-vision/sui-go-sdk/transaction"
	"github.com/multiversx/mx-bridge-eth-go/core"
)

// SuiTxHandlerStub -
type SuiTxHandlerStub struct {
	SendTransactionReturnHashCalled func(ctx context.Context, gasCoin *transaction.SuiObjectRef, calls []core.SuiPTBOperation) (string, error)
}

// SendTransactionReturnHash -
func (stub *SuiTxHandlerStub) SendTransactionReturnHash(ctx context.Context, gasCoin *transaction.SuiObjectRef, calls []core.SuiPTBOperation) (string, error) {
	if stub.SendTransactionReturnHashCalled != nil {
		return stub.SendTransactionReturnHashCalled(ctx, gasCoin, calls)
	}

	return "", nil
}

// IsInterfaceNil -
func (stub *SuiTxHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
