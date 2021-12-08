package roleProviders

import (
	"context"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/stretchr/testify/assert"
)

func createElrondMockArgs() ArgsElrondRoleProvider {
	return ArgsElrondRoleProvider{
		Log:        logger.GetOrCreate("test"),
		DataGetter: &bridgeV2.DataGetterStub{},
	}
}

func TestNewElrondRoleProvider(t *testing.T) {
	t.Parallel()

	t.Run("nil data getter should error", func(t *testing.T) {
		args := createElrondMockArgs()
		args.DataGetter = nil

		erp, err := NewElrondRoleProvider(args)
		assert.True(t, check.IfNil(erp))
		assert.Equal(t, ErrNilDataGetter, err)
	})
	t.Run("nil logger should error", func(t *testing.T) {
		args := createElrondMockArgs()
		args.Log = nil

		erp, err := NewElrondRoleProvider(args)
		assert.True(t, check.IfNil(erp))
		assert.Equal(t, ErrNilLogger, err)
	})
	t.Run("should work", func(t *testing.T) {
		args := createElrondMockArgs()

		erp, err := NewElrondRoleProvider(args)
		assert.False(t, check.IfNil(erp))
		assert.Nil(t, err)
	})
}

func TestElrondRoleProvider_ExecuteErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	args := createElrondMockArgs()
	args.DataGetter = &bridgeV2.DataGetterStub{
		GetAllStakedRelayersCalled: func(ctx context.Context) ([][]byte, error) {
			return nil, expectedErr
		},
	}

	erp, _ := NewElrondRoleProvider(args)
	err := erp.Execute(context.TODO())
	assert.Equal(t, expectedErr, err)
}

func TestElrondProvider_ExecuteShouldWork(t *testing.T) {
	t.Parallel()

	whitelistedAddresses := [][]byte{
		[]byte("address 1"),
		[]byte("address 2"),
	}

	t.Run("nil whitelisted", testElrondExecuteShouldWork(nil))
	t.Run("empty whitelisted", testElrondExecuteShouldWork(make([][]byte, 0)))
	t.Run("with whitelisted", testElrondExecuteShouldWork(whitelistedAddresses))
}

func testElrondExecuteShouldWork(whitelistedAddresses [][]byte) func(t *testing.T) {
	return func(t *testing.T) {
		args := createElrondMockArgs()
		args.DataGetter = &bridgeV2.DataGetterStub{
			GetAllStakedRelayersCalled: func(ctx context.Context) ([][]byte, error) {
				return whitelistedAddresses, nil
			},
		}

		erp, _ := NewElrondRoleProvider(args)
		err := erp.Execute(context.TODO())
		assert.Nil(t, err)

		for _, addr := range whitelistedAddresses {
			addressHandler := data.NewAddressFromBytes(addr)
			assert.True(t, erp.IsWhitelisted(addressHandler))
		}

		randomAddress := data.NewAddressFromBytes([]byte("random address"))
		assert.False(t, erp.IsWhitelisted(randomAddress))
		assert.False(t, erp.IsWhitelisted(nil))
		erp.mut.RLock()
		assert.Equal(t, len(whitelistedAddresses), len(erp.whitelistedAddresses))
		erp.mut.RUnlock()
	}
}
