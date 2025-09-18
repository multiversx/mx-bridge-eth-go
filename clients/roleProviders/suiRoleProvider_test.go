package roleproviders

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/clients/sui"
	bridgeTests "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createSuiMockArgs() ArgsSuiRoleProvider {
	return ArgsSuiRoleProvider{
		Log:        logger.GetOrCreate("test"),
		DataGetter: &bridgeTests.SuiDataGetterStub{},
	}
}

func TestNewSuiRoleProvider(t *testing.T) {
	t.Parallel()

	t.Run("nil data getter should error", func(t *testing.T) {
		t.Parallel()

		args := createSuiMockArgs()
		args.DataGetter = nil

		srp, err := NewSuiRoleProvider(args)
		assert.True(t, check.IfNil(srp))
		assert.Equal(t, clients.ErrNilDataGetter, err)
	})
	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		args := createSuiMockArgs()
		args.Log = nil

		srp, err := NewSuiRoleProvider(args)
		assert.True(t, check.IfNil(srp))
		assert.Equal(t, clients.ErrNilLogger, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createSuiMockArgs()

		srp, err := NewSuiRoleProvider(args)
		assert.False(t, check.IfNil(srp))
		assert.Nil(t, err)
	})
}

func TestSuiRoleProvider_ExecuteErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	args := createSuiMockArgs()
	args.DataGetter = &bridgeTests.SuiDataGetterStub{
		GetRelayersCalled: func(ctx context.Context) ([]models.SuiAddress, error) {
			return nil, expectedErr
		},
	}
	srp, _ := NewSuiRoleProvider(args)
	err := srp.Execute(context.TODO())
	assert.Equal(t, expectedErr, err)
}

func TestSuiRoleProvider_ExecuteShouldWork(t *testing.T) {
	t.Parallel()

	whitelistedAddresses := []models.SuiAddress{
		models.SuiAddress("0x" + strings.Repeat("1", 64)),
		models.SuiAddress("0x" + strings.Repeat("2", 64)),
		models.SuiAddress("0x" + strings.Repeat("3", 64)),
	}
	expectedSortedPublicKeys := [][]byte{
		bytes.Repeat([]byte("1"), 32),
		bytes.Repeat([]byte("2"), 32),
		bytes.Repeat([]byte("3"), 32),
	}

	t.Run("nil whitelisted", testSuiExecuteShouldWork(nil, make([][]byte, 0)))
	t.Run("empty whitelisted", testSuiExecuteShouldWork(make([]models.SuiAddress, 0), make([][]byte, 0)))
	t.Run("with whitelisted", testSuiExecuteShouldWork(whitelistedAddresses, expectedSortedPublicKeys))
}

func testSuiExecuteShouldWork(whitelistedAddresses []models.SuiAddress, expectedSortedPublicKeys [][]byte) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		args := createSuiMockArgs()
		args.DataGetter = &bridgeTests.SuiDataGetterStub{
			GetRelayersCalled: func(ctx context.Context) ([]models.SuiAddress, error) {
				return whitelistedAddresses, nil
			},
		}

		srp, _ := NewSuiRoleProvider(args)
		err := srp.Execute(context.TODO())
		assert.Nil(t, err)
	}
}

func TestSuiRoleProvider_MisconfiguredAddressesShouldError(t *testing.T) {
	t.Parallel()

	misconfiguredAddresses := []models.SuiAddress{
		models.SuiAddress("0x" + strings.Repeat("1", 64)),
		models.SuiAddress("0x" + strings.Repeat("2", 64)),
		models.SuiAddress("bad address"),
	}

	args := createSuiMockArgs()
	args.DataGetter = &bridgeTests.SuiDataGetterStub{
		GetRelayersCalled: func(ctx context.Context) ([]models.SuiAddress, error) {
			return misconfiguredAddresses, nil
		},
	}

	srp, _ := NewSuiRoleProvider(args)
	err := srp.Execute(context.TODO())
	assert.True(t, errors.Is(err, ErrInvalidSuiAddress))
	assert.True(t, strings.Contains(err.Error(), string(misconfiguredAddresses[2])))
	assert.Zero(t, len(srp.whitelistedAddresses))
}

func TestSuiRoleProvider_VerifySuiSignature(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		whitelistedAddresses []models.SuiAddress
		sig                  string
		hexMsg               string
		expectedErr          error
	}{
		{
			name: "verify should work",
			whitelistedAddresses: []models.SuiAddress{
				models.SuiAddress("0x" + strings.Repeat("1", 64)),
				models.SuiAddress("0x76da167b1a6f847c9de0a67d2df947c987726631472c8369cda3be6f7df3fb1d"),
			},
			sig:         "ACoA2nKrvISvIvdXpeWh5S0pOtxAA3Ub/qaE6QBWnRTL4SL5LmJdoi8oX10s4ScBiF0AwzsLZvujDAZWFdg2GQoPOOF+g0BWGaCx15vC91NFDWr1imC8X8iX3+hmAuIz1Q==",
			hexMsg:      "0aca4cfcc0023f00a3500cbfb04e25e5c3ea6379a2aa246d640c49bf56ebd870",
			expectedErr: nil,
		},
		{
			name: "address not whitelisted",
			whitelistedAddresses: []models.SuiAddress{
				models.SuiAddress("0x" + strings.Repeat("3", 64)),
			},
			sig:         "ACoA2nKrvISvIvdXpeWh5S0pOtxAA3Ub/qaE6QBWnRTL4SL5LmJdoi8oX10s4ScBiF0AwzsLZvujDAZWFdg2GQoPOOF+g0BWGaCx15vC91NFDWr1imC8X8iX3+hmAuIz1Q==",
			hexMsg:      "0aca4cfcc0023f00a3500cbfb04e25e5c3ea6379a2aa246d640c49bf56ebd870",
			expectedErr: ErrAddressIsNotWhitelisted,
		},
		{
			name: "invalid signature length",
			whitelistedAddresses: []models.SuiAddress{
				models.SuiAddress("0x" + strings.Repeat("1", 64)),
			},
			sig:         strings.Repeat("01", sui.EncodedSignatureLength+1),
			hexMsg:      strings.Repeat("02", sui.MessageLength),
			expectedErr: ErrInvalidSignaturesArray,
		},
		{
			name: "invalid message length",
			whitelistedAddresses: []models.SuiAddress{
				models.SuiAddress("0x" + strings.Repeat("1", 64)),
			},
			sig:         strings.Repeat("03", sui.EncodedSignatureLength),
			hexMsg:      strings.Repeat("04", sui.MessageLength-1),
			expectedErr: ErrInvalidMessagesArray,
		},
		{
			name: "signatures/messages count mismatch",
			whitelistedAddresses: []models.SuiAddress{
				models.SuiAddress("0x" + strings.Repeat("1", 64)),
			},
			sig:         strings.Repeat("05", sui.EncodedSignatureLength*2),
			hexMsg:      strings.Repeat("06", sui.MessageLength),
			expectedErr: ErrInvalidSignaturesCount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, testSuiVerifySigShouldWork(
			tt.whitelistedAddresses,
			tt.sig,
			tt.hexMsg,
			tt.expectedErr,
		))
	}
}

func testSuiVerifySigShouldWork(whitelistedAddresses []models.SuiAddress, sig string, hexMsg string, expectedErr error) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		msg, err := hex.DecodeString(hexMsg)
		require.Nil(t, err)

		args := createSuiMockArgs()
		args.DataGetter = &bridgeTests.SuiDataGetterStub{
			GetRelayersCalled: func(ctx context.Context) ([]models.SuiAddress, error) {
				return whitelistedAddresses, nil
			},
		}

		srp, _ := NewSuiRoleProvider(args)
		err = srp.Execute(context.TODO())
		assert.Nil(t, err)

		err = srp.VerifySignature([]byte(sig), msg)
		if expectedErr == nil {
			require.Nil(t, err)
		} else {
			require.True(t, errors.Is(err, expectedErr))
		}
	}
}
