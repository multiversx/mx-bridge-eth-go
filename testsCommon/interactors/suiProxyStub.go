package interactors

import (
	"context"

	"github.com/block-vision/sui-go-sdk/models"
)

// SuiProxyStub -
type SuiProxyStub struct {
	PayCalled                                  func(ctx context.Context, req models.PayRequest) (models.TxnMetaData, error)
	PublishCalled                              func(ctx context.Context, req models.PublishRequest) (models.TxnMetaData, error)
	SuiXGetCoinsCalled                         func(ctx context.Context, req models.SuiXGetCoinsRequest) (models.PaginatedCoinsResponse, error)
	SplitCoinCalled                            func(ctx context.Context, req models.SplitCoinRequest) (models.TxnMetaData, error)
	SuiGetLatestCheckpointSequenceNumberCalled func(ctx context.Context) (uint64, error)
	SuiXGetBalanceCalled                       func(ctx context.Context, req models.SuiXGetBalanceRequest) (models.CoinBalanceResponse, error)
	SuiDevInspectTransactionBlockCalled        func(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error)
	MoveCallCalled                             func(ctx context.Context, req models.MoveCallRequest) (models.TxnMetaData, error)
	SignAndExecuteTransactionBlockCalled       func(ctx context.Context, req models.SignAndExecuteTransactionBlockRequest) (models.SuiTransactionBlockResponse, error)
	SuiGetObjectBlockCalled                    func(ctx context.Context, req models.SuiGetObjectRequest) (models.SuiObjectResponse, error)
	SuiExecuteTransactionBlockCalled           func(ctx context.Context, req models.SuiExecuteTransactionBlockRequest) (models.SuiTransactionBlockResponse, error)
	SuiXGetOwnedObjectsCalled                  func(ctx context.Context, req models.SuiXGetOwnedObjectsRequest) (models.PaginatedObjectsResponse, error)
	TransferObjectCalled                       func(ctx context.Context, req models.TransferObjectRequest) (models.TxnMetaData, error)
}

// Pay -
func (sps *SuiProxyStub) Pay(ctx context.Context, req models.PayRequest) (models.TxnMetaData, error) {
	if sps.PayCalled != nil {
		return sps.PayCalled(ctx, req)
	}
	return models.TxnMetaData{}, nil
}

// Publish -
func (sps *SuiProxyStub) Publish(ctx context.Context, req models.PublishRequest) (models.TxnMetaData, error) {
	if sps.PublishCalled != nil {
		return sps.PublishCalled(ctx, req)
	}
	return models.TxnMetaData{}, nil
}

// SuiXGetCoins -
func (sps *SuiProxyStub) SuiXGetCoins(ctx context.Context, req models.SuiXGetCoinsRequest) (models.PaginatedCoinsResponse, error) {
	if sps.SuiXGetCoinsCalled != nil {
		return sps.SuiXGetCoinsCalled(ctx, req)
	}
	return models.PaginatedCoinsResponse{}, nil
}

// SplitCoin -
func (sps *SuiProxyStub) SplitCoin(ctx context.Context, req models.SplitCoinRequest) (models.TxnMetaData, error) {
	if sps.SplitCoinCalled != nil {
		return sps.SplitCoinCalled(ctx, req)
	}
	return models.TxnMetaData{}, nil
}

// SuiGetLatestCheckpointSequenceNumber -
func (sps *SuiProxyStub) SuiGetLatestCheckpointSequenceNumber(ctx context.Context) (uint64, error) {
	if sps.SuiGetLatestCheckpointSequenceNumberCalled != nil {
		return sps.SuiGetLatestCheckpointSequenceNumberCalled(ctx)
	}
	return 0, nil
}

// SuiXGetBalance -
func (sps *SuiProxyStub) SuiXGetBalance(ctx context.Context, req models.SuiXGetBalanceRequest) (models.CoinBalanceResponse, error) {
	if sps.SuiXGetBalanceCalled != nil {
		return sps.SuiXGetBalanceCalled(ctx, req)
	}
	return models.CoinBalanceResponse{}, nil
}

// SuiDevInspectTransactionBlock -
func (sps *SuiProxyStub) SuiDevInspectTransactionBlock(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
	if sps.SuiDevInspectTransactionBlockCalled != nil {
		return sps.SuiDevInspectTransactionBlockCalled(ctx, req)
	}
	return models.SuiTransactionBlockResponse{}, nil
}

// MoveCall -
func (sps *SuiProxyStub) MoveCall(ctx context.Context, req models.MoveCallRequest) (models.TxnMetaData, error) {
	if sps.MoveCallCalled != nil {
		return sps.MoveCallCalled(ctx, req)
	}
	return models.TxnMetaData{}, nil
}

// SignAndExecuteTransactionBlock -
func (sps *SuiProxyStub) SignAndExecuteTransactionBlock(ctx context.Context, req models.SignAndExecuteTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
	if sps.SignAndExecuteTransactionBlockCalled != nil {
		return sps.SignAndExecuteTransactionBlockCalled(ctx, req)
	}
	return models.SuiTransactionBlockResponse{}, nil
}

// SuiExecuteTransactionBlock -
func (sps *SuiProxyStub) SuiExecuteTransactionBlock(ctx context.Context, req models.SuiExecuteTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
	if sps.SuiExecuteTransactionBlockCalled != nil {
		return sps.SuiExecuteTransactionBlockCalled(ctx, req)
	}
	return models.SuiTransactionBlockResponse{}, nil
}

// SuiXGetOwnedObjects -
func (sps *SuiProxyStub) SuiXGetOwnedObjects(ctx context.Context, req models.SuiXGetOwnedObjectsRequest) (models.PaginatedObjectsResponse, error) {
	if sps.SuiXGetOwnedObjectsCalled != nil {
		return sps.SuiXGetOwnedObjectsCalled(ctx, req)
	}
	return models.PaginatedObjectsResponse{}, nil
}

// TransferObject -
func (sps *SuiProxyStub) TransferObject(ctx context.Context, req models.TransferObjectRequest) (models.TxnMetaData, error) {
	if sps.TransferObjectCalled != nil {
		return sps.TransferObjectCalled(ctx, req)
	}
	return models.TxnMetaData{}, nil
}
