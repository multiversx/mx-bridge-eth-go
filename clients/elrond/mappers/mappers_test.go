package mappers

import (
	"context"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewMapper(t *testing.T) {
	t.Parallel()
	{
		t.Run("ElrondToErc20: nil dataGetter", func(t *testing.T) {
			mapper, err := NewErc20ToElrondMapper(nil)
			assert.Equal(t, errNilDataGetter, err)
			assert.True(t, check.IfNil(mapper))
		})
		t.Run("ElrondToErc20: should work", func(t *testing.T) {
			mapper, err := NewErc20ToElrondMapper(&bridgeV2.DataGetterStub{})
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))
		})
	}
	{
		t.Run("Erc20ToElrond: nil dataGetter", func(t *testing.T) {
			mapper, err := NewElrondToErc20Mapper(nil)
			assert.Equal(t, errNilDataGetter, err)
			assert.True(t, check.IfNil(mapper))
		})
		t.Run("Erc20ToElrond: should work", func(t *testing.T) {
			mapper, err := NewElrondToErc20Mapper(&bridgeV2.DataGetterStub{})
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))
		})
	}
}

func TestConvertToken(t *testing.T) {
	t.Parallel()

	{
		t.Run("ElrondToErc20: dataGetter returns error", func(t *testing.T) {
			expectedError := errors.New("expected error")
			dg := &bridgeV2.DataGetterStub{
				GetERC20AddressForTokenIdCalled: func(ctx context.Context, tokenId []byte) ([][]byte, error) {
					return nil, expectedError
				}}
			mapper, err := NewElrondToErc20Mapper(dg)
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))

			_, err = mapper.ConvertToken(context.Background(), []byte("erdAddress"))
			assert.Equal(t, expectedError, err)
		})
		t.Run("ElrondToErc20: should work", func(t *testing.T) {
			expectedErc20Address := []byte("erc20Address")
			dg := &bridgeV2.DataGetterStub{
				GetERC20AddressForTokenIdCalled: func(ctx context.Context, tokenId []byte) ([][]byte, error) {
					return [][]byte{expectedErc20Address}, nil
				}}
			mapper, err := NewElrondToErc20Mapper(dg)
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))
			erc20AddressReturned, err := mapper.ConvertToken(context.Background(), []byte("erdAddress"))
			assert.Nil(t, err)
			assert.Equal(t, expectedErc20Address, erc20AddressReturned)
		})
	}
	{
		t.Run("Erc20ToElrond: dataGetter returns error", func(t *testing.T) {
			expectedError := errors.New("expected error")
			dg := &bridgeV2.DataGetterStub{
				GetTokenIdForErc20AddressCalled: func(ctx context.Context, erc20Address []byte) ([][]byte, error) {
					return nil, expectedError
				}}
			mapper, err := NewErc20ToElrondMapper(dg)
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))

			_, err = mapper.ConvertToken(context.Background(), []byte("erc20Address"))
			assert.Equal(t, expectedError, err)
		})
		t.Run("Erc20ToElrond: should work", func(t *testing.T) {
			expectedErdAddress := []byte("erdAddress")
			dg := &bridgeV2.DataGetterStub{
				GetTokenIdForErc20AddressCalled: func(ctx context.Context, erc20Address []byte) ([][]byte, error) {
					return [][]byte{expectedErdAddress}, nil
				}}
			mapper, err := NewErc20ToElrondMapper(dg)
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))
			erdAddressReturned, err := mapper.ConvertToken(context.Background(), []byte("erc20Address"))
			assert.Nil(t, err)
			assert.Equal(t, expectedErdAddress, erdAddressReturned)
		})
	}
}
