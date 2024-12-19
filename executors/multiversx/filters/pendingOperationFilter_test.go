package filters

import (
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
)

const ethTestAddress1 = "0x880ec53af800b5cd051531672ef4fc4de233bd5d"
const ethTestAddress2 = "0x880ebbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
const mvxTestAddress1 = "erd1qqqqqqqqqqqqqpgqk839entmk46ykukvhpn90g6knskju3dtanaq20f66e"
const mvxTestAddress2 = "erd1qqqqqqqqqqqqqpgqptqsx2llrwh4phaf42lwwxez2hzeulxwanaqg7kgky"

var testLog = logger.GetOrCreate("filters")
var ethTestAddress1Bytes, _ = hex.DecodeString(ethTestAddress1[2:])

func createTestConfig() config.PendingOperationsFilterConfig {
	return config.PendingOperationsFilterConfig{
		DeniedEthAddresses:  nil,
		AllowedEthAddresses: []string{"*"},

		DeniedMvxAddresses:  nil,
		AllowedMvxAddresses: []string{"*"},

		DeniedTokens:  nil,
		AllowedTokens: []string{"*"},
	}
}

func TestNewPendingOperationFilter(t *testing.T) {
	t.Parallel()

	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		filter, err := NewPendingOperationFilter(createTestConfig(), nil)
		assert.Nil(t, filter)
		assert.ErrorIs(t, err, errNilLogger)
	})
	t.Run("empty config should error", func(t *testing.T) {
		t.Parallel()

		filter, err := NewPendingOperationFilter(config.PendingOperationsFilterConfig{}, testLog)
		assert.Nil(t, filter)
		assert.ErrorIs(t, err, errNoItemsAllowed)
	})
	t.Run("denied eth list contains wildcard should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.DeniedEthAddresses = []string{"	*  "}

		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.Nil(t, filter)
		assert.ErrorIs(t, err, errUnsupportedMarker)
		assert.Contains(t, err.Error(), "on item at index 0 in list DeniedEthAddresses")
	})
	t.Run("denied mvx list contains wildcard should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.DeniedMvxAddresses = []string{"	*  "}

		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.Nil(t, filter)
		assert.ErrorIs(t, err, errUnsupportedMarker)
		assert.Contains(t, err.Error(), "on item at index 0 in list DeniedMvxAddresses")
	})
	t.Run("denied tokens list contains wildcard should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.DeniedTokens = []string{"	*  "}

		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.Nil(t, filter)
		assert.ErrorIs(t, err, errUnsupportedMarker)
		assert.Contains(t, err.Error(), "on item at index 0 in list DeniedTokens")
	})
	t.Run("allowed eth list contains empty string should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.AllowedEthAddresses = append(cfg.AllowedEthAddresses, "	 ")

		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.Nil(t, filter)
		assert.ErrorIs(t, err, errUnsupportedMarker)
		assert.Contains(t, err.Error(), "on item at index 1 in list AllowedEthAddresses")
	})
	t.Run("allowed mvx list contains empty string should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.AllowedMvxAddresses = append(cfg.AllowedMvxAddresses, "	 ")

		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.Nil(t, filter)
		assert.ErrorIs(t, err, errUnsupportedMarker)
		assert.Contains(t, err.Error(), "on item at index 1 in list AllowedMvxAddresses")
	})
	t.Run("allowed tokens list contains empty string should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.AllowedTokens = append(cfg.AllowedTokens, "	 ")

		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.Nil(t, filter)
		assert.ErrorIs(t, err, errUnsupportedMarker)
		assert.Contains(t, err.Error(), "on item at index 1 in list AllowedTokens")
	})
	t.Run("invalid address in AllowedEthAddresses should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.AllowedEthAddresses = append(cfg.AllowedEthAddresses, "invalid address")

		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.Nil(t, filter)
		assert.ErrorIs(t, err, errMissingEthPrefix)
		assert.Contains(t, err.Error(), "on item at index 1 in list AllowedEthAddresses")
	})
	t.Run("invalid address in DeniedEthAddresses should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.DeniedEthAddresses = append(cfg.DeniedEthAddresses, "invalid address")

		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.Nil(t, filter)
		assert.ErrorIs(t, err, errMissingEthPrefix)
		assert.Contains(t, err.Error(), "on item at index 0 in list DeniedEthAddresses")
	})
	t.Run("invalid address in AllowedMvxAddresses should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.AllowedMvxAddresses = append(cfg.AllowedMvxAddresses, "invalid address")

		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.Nil(t, filter)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "on item at index 1 in list AllowedMvxAddresses")
	})
	t.Run("invalid address in DeniedMvxAddresses should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.DeniedMvxAddresses = append(cfg.DeniedMvxAddresses, "invalid address")

		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.Nil(t, filter)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "on item at index 0 in list DeniedMvxAddresses")
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfig()
		cfg.AllowedEthAddresses = append(cfg.AllowedMvxAddresses, ethTestAddress1)
		cfg.DeniedEthAddresses = append(cfg.DeniedEthAddresses, ethTestAddress1)
		cfg.AllowedMvxAddresses = append(cfg.AllowedMvxAddresses, mvxTestAddress1)
		cfg.DeniedMvxAddresses = append(cfg.DeniedMvxAddresses, mvxTestAddress1)
		filter, err := NewPendingOperationFilter(cfg, testLog)
		assert.NotNil(t, filter)
		assert.Nil(t, err)
	})
}

func TestPendingOperationFilter_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var instance *pendingOperationFilter
	assert.True(t, instance.IsInterfaceNil())

	instance = &pendingOperationFilter{}
	assert.False(t, instance.IsInterfaceNil())
}

func TestPendingOperationFilter_ShouldExecute(t *testing.T) {
	t.Parallel()

	t.Run("nil callData.To should return false", func(t *testing.T) {
		t.Parallel()

		callData := core.ProxySCCompleteCallData{
			To: nil,
		}

		cfg := createTestConfig()
		filter, _ := NewPendingOperationFilter(cfg, testLog)

		assert.False(t, filter.ShouldExecute(callData))
	})
	t.Run("callData.To is not a valid Mvx address should return false", func(t *testing.T) {
		t.Parallel()

		callData := core.ProxySCCompleteCallData{
			To: data.NewAddressFromBytes([]byte{0x1, 0x2}),
		}

		cfg := createTestConfig()
		filter, _ := NewPendingOperationFilter(cfg, testLog)

		assert.False(t, filter.ShouldExecute(callData))
	})
	t.Run("eth address", func(t *testing.T) {
		t.Parallel()

		callData := core.ProxySCCompleteCallData{
			From: common.BytesToAddress(ethTestAddress1Bytes),
		}
		callData.To, _ = data.NewAddressFromBech32String(mvxTestAddress1)
		t.Run("is denied should return false", func(t *testing.T) {
			t.Parallel()

			cfg := createTestConfig()
			cfg.DeniedEthAddresses = []string{ethTestAddress1}
			cfg.AllowedEthAddresses = []string{ethTestAddress1}

			filter, _ := NewPendingOperationFilter(cfg, testLog)
			assert.False(t, filter.ShouldExecute(callData))

			cfg.AllowedEthAddresses = []string{"*"}
			filter, _ = NewPendingOperationFilter(cfg, testLog)
			assert.False(t, filter.ShouldExecute(callData))
		})
		t.Run("is not denied but allowed should return true", func(t *testing.T) {
			t.Parallel()

			cfg := createTestConfig()
			cfg.AllowedEthAddresses = []string{ethTestAddress1}

			filter, _ := NewPendingOperationFilter(cfg, testLog)
			assert.True(t, filter.ShouldExecute(callData))

			cfg.AllowedEthAddresses = []string{"*"}
			filter, _ = NewPendingOperationFilter(cfg, testLog)
			assert.True(t, filter.ShouldExecute(callData))
		})
		t.Run("is not denied but not allowed should return false", func(t *testing.T) {
			t.Parallel()

			cfg := createTestConfig()
			cfg.AllowedEthAddresses = []string{ethTestAddress2}
			cfg.AllowedTokens = nil
			cfg.AllowedMvxAddresses = nil

			filter, _ := NewPendingOperationFilter(cfg, testLog)
			assert.False(t, filter.ShouldExecute(callData))
		})
	})
	t.Run("mvx address", func(t *testing.T) {
		t.Parallel()

		callData := core.ProxySCCompleteCallData{
			From: common.BytesToAddress(ethTestAddress1Bytes),
		}
		callData.To, _ = data.NewAddressFromBech32String(mvxTestAddress1)
		t.Run("is denied should return false", func(t *testing.T) {
			t.Parallel()

			cfg := createTestConfig()
			cfg.DeniedMvxAddresses = []string{mvxTestAddress1}
			cfg.AllowedMvxAddresses = []string{mvxTestAddress1}

			filter, _ := NewPendingOperationFilter(cfg, testLog)
			assert.False(t, filter.ShouldExecute(callData))

			cfg.AllowedMvxAddresses = []string{"*"}
			filter, _ = NewPendingOperationFilter(cfg, testLog)
			assert.False(t, filter.ShouldExecute(callData))
		})
		t.Run("is not denied but allowed should return true", func(t *testing.T) {
			t.Parallel()

			cfg := createTestConfig()
			cfg.AllowedMvxAddresses = []string{mvxTestAddress1}

			filter, _ := NewPendingOperationFilter(cfg, testLog)
			assert.True(t, filter.ShouldExecute(callData))

			cfg.AllowedMvxAddresses = []string{"*"}
			filter, _ = NewPendingOperationFilter(cfg, testLog)
			assert.True(t, filter.ShouldExecute(callData))
		})
		t.Run("is not denied but not allowed should return false", func(t *testing.T) {
			t.Parallel()

			cfg := createTestConfig()
			cfg.AllowedMvxAddresses = []string{mvxTestAddress2}
			cfg.AllowedTokens = nil
			cfg.AllowedEthAddresses = nil

			filter, _ := NewPendingOperationFilter(cfg, testLog)
			assert.False(t, filter.ShouldExecute(callData))
		})
	})
	t.Run("tokens", func(t *testing.T) {
		t.Parallel()

		token1 := "tkn1"
		token2 := "tkn2"
		callData := core.ProxySCCompleteCallData{
			From:  common.BytesToAddress(ethTestAddress1Bytes),
			Token: token1,
		}
		callData.To, _ = data.NewAddressFromBech32String(mvxTestAddress1)

		t.Run("is denied should return false", func(t *testing.T) {
			t.Parallel()

			cfg := createTestConfig()
			cfg.DeniedTokens = []string{token1}
			cfg.AllowedTokens = []string{token1}

			filter, _ := NewPendingOperationFilter(cfg, testLog)
			assert.False(t, filter.ShouldExecute(callData))

			cfg.AllowedTokens = []string{"*"}
			filter, _ = NewPendingOperationFilter(cfg, testLog)
			assert.False(t, filter.ShouldExecute(callData))
		})
		t.Run("is not denied but allowed should return true", func(t *testing.T) {
			t.Parallel()

			cfg := createTestConfig()
			cfg.AllowedTokens = []string{token1}

			filter, _ := NewPendingOperationFilter(cfg, testLog)
			assert.True(t, filter.ShouldExecute(callData))

			cfg.AllowedTokens = []string{"*"}
			filter, _ = NewPendingOperationFilter(cfg, testLog)
			assert.True(t, filter.ShouldExecute(callData))
		})
		t.Run("is not denied but not allowed should return false", func(t *testing.T) {
			t.Parallel()

			cfg := createTestConfig()
			cfg.AllowedTokens = []string{token2}
			cfg.AllowedMvxAddresses = nil
			cfg.AllowedEthAddresses = nil

			filter, _ := NewPendingOperationFilter(cfg, testLog)
			assert.False(t, filter.ShouldExecute(callData))
		})
	})
}
