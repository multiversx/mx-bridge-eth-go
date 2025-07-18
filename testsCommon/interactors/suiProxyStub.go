package interactors

import (
	"context"
	"fmt"

	"github.com/block-vision/sui-go-sdk/models"
)

// SuiProxyStub -
type SuiProxyStub struct {
	SuiGetLatestCheckpointSequenceNumberCalled func(ctx context.Context) (uint64, error)
	SuiXGetBalanceCalled                       func(ctx context.Context, req models.SuiXGetBalanceRequest) (models.CoinBalanceResponse, error)
	SuiDevInspectTransactionBlockCalled        func(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error)
	MoveCallCalled                             func(ctx context.Context, req models.MoveCallRequest) (models.TxnMetaData, error)
	SignAndExecuteTransactionBlockCalled       func(ctx context.Context, req models.SignAndExecuteTransactionBlockRequest) (models.SuiTransactionBlockResponse, error)
}

// SuiGetLatestCheckpointSequenceNumber -
func (sps *SuiProxyStub) SuiGetLatestCheckpointSequenceNumber(ctx context.Context) (uint64, error) {
	if sps.SuiGetLatestCheckpointSequenceNumberCalled != nil {
		return sps.SuiGetLatestCheckpointSequenceNumberCalled(ctx)
	}
	return 0, fmt.Errorf("not implemented")
}

// SuiXGetBalance -
func (sps *SuiProxyStub) SuiXGetBalance(ctx context.Context, req models.SuiXGetBalanceRequest) (models.CoinBalanceResponse, error) {
	if sps.SuiXGetBalanceCalled != nil {
		return sps.SuiXGetBalanceCalled(ctx, req)
	}
	return models.CoinBalanceResponse{}, fmt.Errorf("not implemented")
}

// SuiDevInspectTransactionBlock -
func (sps *SuiProxyStub) SuiDevInspectTransactionBlock(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
	if sps.SuiDevInspectTransactionBlockCalled != nil {
		return sps.SuiDevInspectTransactionBlockCalled(ctx, req)
	}
	return models.SuiTransactionBlockResponse{}, fmt.Errorf("not implemented")
}

// MoveCall -
func (sps *SuiProxyStub) MoveCall(ctx context.Context, req models.MoveCallRequest) (models.TxnMetaData, error) {
	if sps.MoveCallCalled != nil {
		return sps.MoveCallCalled(ctx, req)
	}
	return models.TxnMetaData{}, fmt.Errorf("not implemented")
}

// SignAndExecuteTransactionBlock -
func (sps *SuiProxyStub) SignAndExecuteTransactionBlock(ctx context.Context, req models.SignAndExecuteTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
	if sps.SignAndExecuteTransactionBlockCalled != nil {
		return sps.SignAndExecuteTransactionBlockCalled(ctx, req)
	}
	return models.SuiTransactionBlockResponse{}, fmt.Errorf("not implemented")
}
