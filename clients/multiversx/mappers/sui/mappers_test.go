package sui

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
		t.Run("SuiToMultiversX: nil dataGetter", func(t *testing.T) {
			mapper, err := NewSuiToMultiversXMapper(nil)
			assert.Equal(t, clients.ErrNilDataGetter, err)
			assert.True(t, check.IfNil(mapper))
		})
		t.Run("SuiToMultiversX: should work", func(t *testing.T) {
			mapper, err := NewSuiToMultiversXMapper(&bridgeTests.DataGetterStub{})
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))
		})
	}
	{
		t.Run("MultiversXToSui: nil dataGetter", func(t *testing.T) {
			mapper, err := NewMultiversXToSuiMapper(nil)
			assert.Equal(t, clients.ErrNilDataGetter, err)
			assert.True(t, check.IfNil(mapper))
		})
		t.Run("MultiversXToSui: should work", func(t *testing.T) {
			mapper, err := NewMultiversXToSuiMapper(&bridgeTests.DataGetterStub{})
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))
		})
	}
}

func TestConvertToken(t *testing.T) {
	t.Parallel()

	{
		t.Run("MultiversXToSui: dataGetter returns error", func(t *testing.T) {
			expectedError := errors.New("expected error")
			dg := &bridgeTests.DataGetterStub{
				GetSuiCoinForTokenIdCalled: func(ctx context.Context, tokenId []byte) ([][]byte, error) {
					return nil, expectedError
				}}
			mapper, err := NewMultiversXToSuiMapper(dg)
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))

			_, err = mapper.ConvertToken(context.Background(), []byte("erdAddress"))
			assert.Equal(t, expectedError, err)
		})
		t.Run("MultiversXToSui: should work", func(t *testing.T) {
			expectedSuiAddress := []byte("erc20Address")
			dg := &bridgeTests.DataGetterStub{
				GetSuiCoinForTokenIdCalled: func(ctx context.Context, tokenId []byte) ([][]byte, error) {
					return [][]byte{expectedSuiAddress}, nil
				}}
			mapper, err := NewMultiversXToSuiMapper(dg)
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))
			erc20AddressReturned, err := mapper.ConvertToken(context.Background(), []byte("erdAddress"))
			assert.Nil(t, err)
			assert.Equal(t, expectedSuiAddress, erc20AddressReturned)
		})
	}
	{
		t.Run("SuiToMultiversX: dataGetter returns error", func(t *testing.T) {
			expectedError := errors.New("expected error")
			dg := &bridgeTests.DataGetterStub{
				GetTokenIdForSuiCoinCalled: func(ctx context.Context, erc20Address []byte) ([][]byte, error) {
					return nil, expectedError
				}}
			mapper, err := NewSuiToMultiversXMapper(dg)
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))

			_, err = mapper.ConvertToken(context.Background(), []byte("erc20Address"))
			assert.Equal(t, expectedError, err)
		})
		t.Run("SuiToMultiversX: should work", func(t *testing.T) {
			expectedErdAddress := []byte("erdAddress")
			dg := &bridgeTests.DataGetterStub{
				GetTokenIdForSuiCoinCalled: func(ctx context.Context, erc20Address []byte) ([][]byte, error) {
					return [][]byte{expectedErdAddress}, nil
				}}
			mapper, err := NewSuiToMultiversXMapper(dg)
			assert.Nil(t, err)
			assert.False(t, check.IfNil(mapper))
			erdAddressReturned, err := mapper.ConvertToken(context.Background(), []byte("erc20Address"))
			assert.Nil(t, err)
			assert.Equal(t, expectedErdAddress, erdAddressReturned)
		})
	}
}
