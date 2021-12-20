package roleProviders

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createEthereumMockArgs() ArgsEthereumRoleProvider {
	return ArgsEthereumRoleProvider{
		Log:                     logger.GetOrCreate("test"),
		EthereumChainInteractor: &bridgeV2.EthereumClientWrapperStub{},
	}
}

func TestNewEthereumRoleProvider(t *testing.T) {
	t.Parallel()

	t.Run("nil ethereum chain interactor should error", func(t *testing.T) {
		t.Parallel()

		args := createEthereumMockArgs()
		args.EthereumChainInteractor = nil

		erp, err := NewEthereumRoleProvider(args)
		assert.True(t, check.IfNil(erp))
		assert.Equal(t, ErrNilEthereumChainInteractor, err)
	})
	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		args := createEthereumMockArgs()
		args.Log = nil

		erp, err := NewEthereumRoleProvider(args)
		assert.True(t, check.IfNil(erp))
		assert.Equal(t, ErrNilLogger, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createEthereumMockArgs()

		erp, err := NewEthereumRoleProvider(args)
		assert.False(t, check.IfNil(erp))
		assert.Nil(t, err)
	})
}

func TestEthereumRoleProvider_ExecuteErrorsInInteractor(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	args := createEthereumMockArgs()
	args.EthereumChainInteractor = &bridgeV2.EthereumClientWrapperStub{
		GetRelayersCalled: func(ctx context.Context) ([]common.Address, error) {
			return nil, expectedErr
		},
	}

	erp, _ := NewEthereumRoleProvider(args)
	err := erp.Execute(context.TODO())
	assert.Equal(t, expectedErr, err)
}

func TestEthereumProvider_ExecuteShouldWork(t *testing.T) {
	t.Parallel()

	whitelistedAddresses := []common.Address{
		common.HexToAddress("0x132A150926691F08a693721503a38affeD18d524"),
		common.HexToAddress("0xb6e20FF4Ae7d29be233D874633F2F0Dcb326E5c0"),
	}

	t.Run("nil whitelisted", testEthereumExecuteShouldWork(nil))
	t.Run("empty whitelisted", testEthereumExecuteShouldWork(make([]common.Address, 0)))
	t.Run("with whitelisted", testEthereumExecuteShouldWork(whitelistedAddresses))
}

func testEthereumExecuteShouldWork(whitelistedAddresses []common.Address) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		args := createEthereumMockArgs()
		args.EthereumChainInteractor = &bridgeV2.EthereumClientWrapperStub{
			GetRelayersCalled: func(ctx context.Context) ([]common.Address, error) {
				return whitelistedAddresses, nil
			},
		}

		erp, _ := NewEthereumRoleProvider(args)
		err := erp.Execute(context.TODO())
		assert.Nil(t, err)

		for _, addr := range whitelistedAddresses {
			assert.True(t, erp.isWhitelisted(addr))
		}

		randomAddress := common.HexToAddress("0x093c0B280ba430A9Cc9C3649FF34FCBf6347bC50")
		assert.False(t, erp.isWhitelisted(randomAddress))
		erp.mut.RLock()
		assert.Equal(t, len(whitelistedAddresses), len(erp.whitelistedAddresses))
		erp.mut.RUnlock()
	}
}

func TestEthereumRoleProvider_VerifyEthSignature(t *testing.T) {
	t.Parallel()

	whitelistedAddresses := []common.Address{
		common.HexToAddress("0x132A150926691F08a693721503a38affeD18d524"),
		common.HexToAddress("0xb6e20FF4Ae7d29be233D874633F2F0Dcb326E5c0"),
	}
	hexMsg := "c124f221e1992619dfe3254e46a97bd7d787ed4e699f48aca715d54e7f52ff5d"
	hexSig := "b0ddb854c7c6a5c78cdbf9e7e6c204711c162220298dc1bfab58be77b8627c155ae4dac5d06197283407993b359752f8906487fc0e3a031173fd07c010e5cddc00"
	t.Run("verify should work", testEthereumVerifySigShouldWork(whitelistedAddresses, hexSig, hexMsg, nil))

	whitelistedAddresses = []common.Address{
		common.HexToAddress("0x132A150926691F08a693721503a38affeD18d524"),
	}
	hexMsg = "c124f221e1992619dfe3254e46a97bd7d787ed4e699f48aca715d54e7f52ff5d"
	hexSig = "b0ddb854c7c6a5c78cdbf9e7e6c204711c162220298dc1bfab58be77b8627c155ae4dac5d06197283407993b359752f8906487fc0e3a031173fd07c010e5cddc00"
	t.Run("address not whitelisted", testEthereumVerifySigShouldWork(whitelistedAddresses, hexSig, hexMsg, ErrAddressIsNotWhitelisted))
}

func testEthereumVerifySigShouldWork(whitelistedAddresses []common.Address, hexSig string, hexMsg string, expectedErr error) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		sig, err := hex.DecodeString(hexSig)
		require.Nil(t, err)

		msg, err := hex.DecodeString(hexMsg)
		require.Nil(t, err)

		args := createEthereumMockArgs()
		args.EthereumChainInteractor = &bridgeV2.EthereumClientWrapperStub{
			GetRelayersCalled: func(ctx context.Context) ([]common.Address, error) {
				return whitelistedAddresses, nil
			},
		}

		erp, _ := NewEthereumRoleProvider(args)
		err = erp.Execute(context.TODO())
		assert.Nil(t, err)

		err = erp.VerifyEthSignature(sig, msg)
		if expectedErr == nil {
			require.Nil(t, err)
		} else {
			require.True(t, errors.Is(err, expectedErr))
		}
	}
}
