package bridgeV2

import (
	"context"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

// DataGetterStub -
type DataGetterStub struct {
	ExecuteQueryReturningBytesCalled  func(ctx context.Context, request *data.VmValueRequest) ([][]byte, error)
	ExecuteQueryReturningBoolCalled   func(ctx context.Context, request *data.VmValueRequest) (bool, error)
	ExecuteQueryReturningUint64Called func(ctx context.Context, request *data.VmValueRequest) (uint64, error)
	GetCurrentBatchAsDataBytesCalled  func(ctx context.Context) ([][]byte, error)
	GetTokenIdForErc20AddressCalled   func(ctx context.Context, erc20Address []byte) ([][]byte, error)
	GetERC20AddressForTokenIdCalled   func(ctx context.Context, tokenId []byte) ([][]byte, error)
}

// ExecuteQueryReturningBytes -
func (stub *DataGetterStub) ExecuteQueryReturningBytes(ctx context.Context, request *data.VmValueRequest) ([][]byte, error) {
	if stub.ExecuteQueryReturningBytesCalled != nil {
		return stub.ExecuteQueryReturningBytesCalled(ctx, request)
	}
	return [][]byte{}, nil
}

// ExecuteQueryReturningBool -
func (stub *DataGetterStub) ExecuteQueryReturningBool(ctx context.Context, request *data.VmValueRequest) (bool, error) {
	if stub.ExecuteQueryReturningBoolCalled != nil {
		return stub.ExecuteQueryReturningBoolCalled(ctx, request)
	}
	return false, nil
}

// ExecuteQueryReturningUint64 -
func (stub *DataGetterStub) ExecuteQueryReturningUint64(ctx context.Context, request *data.VmValueRequest) (uint64, error) {
	if stub.ExecuteQueryReturningUint64Called != nil {
		return stub.ExecuteQueryReturningUint64Called(ctx, request)
	}
	return 0, nil
}

// GetCurrentBatchAsDataBytes -
func (stub *DataGetterStub) GetCurrentBatchAsDataBytes(ctx context.Context) ([][]byte, error) {
	if stub.GetCurrentBatchAsDataBytesCalled != nil {
		return stub.GetCurrentBatchAsDataBytesCalled(ctx)
	}
	return [][]byte{}, nil
}

// GetTokenIdForErc20Address -
func (stub *DataGetterStub) GetTokenIdForErc20Address(ctx context.Context, erc20Address []byte) ([][]byte, error) {
	if stub.GetTokenIdForErc20AddressCalled != nil {
		return stub.GetTokenIdForErc20AddressCalled(ctx, erc20Address)
	}
	return [][]byte{}, nil
}

// GetERC20AddressForTokenId -
func (stub *DataGetterStub) GetERC20AddressForTokenId(ctx context.Context, tokenId []byte) ([][]byte, error) {
	if stub.GetERC20AddressForTokenIdCalled != nil {
		return stub.GetERC20AddressForTokenIdCalled(ctx, tokenId)
	}
	return [][]byte{}, nil
}

// IsInterfaceNil -
func (stub *DataGetterStub) IsInterfaceNil() bool {
	return stub == nil
}
