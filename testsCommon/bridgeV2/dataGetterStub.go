package bridgeV2

import (
	"context"
)

// DataGetterStub -
type DataGetterStub struct {
	GetTokenIdForErc20AddressCalled func(ctx context.Context, erc20Address []byte) ([][]byte, error)
	GetERC20AddressForTokenIdCalled func(ctx context.Context, tokenId []byte) ([][]byte, error)
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
