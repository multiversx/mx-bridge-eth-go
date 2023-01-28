package mappers

import (
	"context"
	"errors"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/clients"
	bridgeTests "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewMapper(t *testing.T) {
	t.Parallel()
	{
		t.Run("Erc20ToMultiversX: nil dataGetter", func(t *testing.T) {
			mapper, err := NewErc20ToMultiversXMapper(nil)
			assert.Equal(t, clients.ErrNilDataGetter, err)
			assert.True(t, check.IfNil(mapper))
		})
		t.Run("Erc20ToMultiversX: should work", func(t *testing.T) {
			mapper, err := NewErc20ToMultiversXMapper(&bridgeTests.DataGetterStub{})
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))
		})
	}
	{
		t.Run("MultiversXToErc20: nil dataGetter", func(t *testing.T) {
			mapper, err := NewMultiversXToErc20Mapper(nil)
			assert.Equal(t, clients.ErrNilDataGetter, err)
			assert.True(t, check.IfNil(mapper))
		})
		t.Run("MultiversXToErc20: should work", func(t *testing.T) {
			mapper, err := NewMultiversXToErc20Mapper(&bridgeTests.DataGetterStub{})
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))
		})
	}
}

func TestConvertToken(t *testing.T) {
	t.Parallel()

	{
		t.Run("MultiversXToErc20: dataGetter returns error", func(t *testing.T) {
			expectedError := errors.New("expected error")
			dg := &bridgeTests.DataGetterStub{
				GetERC20AddressForTokenIdCalled: func(ctx context.Context, tokenId []byte) ([][]byte, error) {
					return nil, expectedError
				}}
			mapper, err := NewMultiversXToErc20Mapper(dg)
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))

			_, err = mapper.ConvertToken(context.Background(), []byte("erdAddress"))
			assert.Equal(t, expectedError, err)
		})
		t.Run("MultiversXToErc20: should work", func(t *testing.T) {
			expectedErc20Address := []byte("erc20Address")
			dg := &bridgeTests.DataGetterStub{
				GetERC20AddressForTokenIdCalled: func(ctx context.Context, tokenId []byte) ([][]byte, error) {
					return [][]byte{expectedErc20Address}, nil
				}}
			mapper, err := NewMultiversXToErc20Mapper(dg)
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))
			erc20AddressReturned, err := mapper.ConvertToken(context.Background(), []byte("erdAddress"))
			assert.Nil(t, err)
			assert.Equal(t, expectedErc20Address, erc20AddressReturned)
		})
	}
	{
		t.Run("Erc20ToMultiversX: dataGetter returns error", func(t *testing.T) {
			expectedError := errors.New("expected error")
			dg := &bridgeTests.DataGetterStub{
				GetTokenIdForErc20AddressCalled: func(ctx context.Context, erc20Address []byte) ([][]byte, error) {
					return nil, expectedError
				}}
			mapper, err := NewErc20ToMultiversXMapper(dg)
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))

			_, err = mapper.ConvertToken(context.Background(), []byte("erc20Address"))
			assert.Equal(t, expectedError, err)
		})
		t.Run("Erc20ToMultiversX: should work", func(t *testing.T) {
			expectedErdAddress := []byte("erdAddress")
			dg := &bridgeTests.DataGetterStub{
				GetTokenIdForErc20AddressCalled: func(ctx context.Context, erc20Address []byte) ([][]byte, error) {
					return [][]byte{expectedErdAddress}, nil
				}}
			mapper, err := NewErc20ToMultiversXMapper(dg)
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))
			erdAddressReturned, err := mapper.ConvertToken(context.Background(), []byte("erc20Address"))
			assert.Nil(t, err)
			assert.Equal(t, expectedErdAddress, erdAddressReturned)
		})
	}
}
