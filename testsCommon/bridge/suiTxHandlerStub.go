package bridge

import (
	"context"
	"github.com/block-vision/sui-go-sdk/models"
)

// SuiTxHandlerStub -
type SuiTxHandlerStub struct {
	SendTransactionReturnHashCalled func(ctx context.Context, moveCallRequest models.MoveCallRequest) (string, error)
	SignCalled                      func(message []byte) ([]byte, error)
}

// SendTransactionReturnHash -
func (t *SuiTxHandlerStub) SendTransactionReturnHash(ctx context.Context, moveCallRequest models.MoveCallRequest) (string, error) {
	if t.SendTransactionReturnHashCalled != nil {
		return t.SendTransactionReturnHashCalled(ctx, moveCallRequest)
	}

	return "", nil
}

// Sign -
func (t *SuiTxHandlerStub) Sign(message []byte) ([]byte, error) {
	if t.SignCalled != nil {
		return t.SignCalled(message)
	}

	return nil, nil
}
