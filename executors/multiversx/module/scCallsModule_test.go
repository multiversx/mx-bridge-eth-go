package module

import (
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/stretchr/testify/assert"
)

func createTestConfigs() config.ScCallsModuleConfig {
	return config.ScCallsModuleConfig{
		General: config.GeneralScCallsModuleConfig{
			ScProxyBech32Addresses: []string{
				"erd1qqqqqqqqqqqqqpgqgftcwj09u0nhmskrw7xxqcqh8qmzwyexd8ss7ftcxx",
			},
			NetworkAddress:               "http://127.0.0.1:8079",
			ProxyMaxNoncesDelta:          5,
			ProxyFinalityCheck:           false,
			ProxyCacherExpirationSeconds: 60,
			ProxyRestAPIEntityType:       string(sdkCore.ObserverNode),
			IntervalToResendTxsInSeconds: 1,
			PrivateKeyFile:               "testdata/grace.pem",
		},
		ScCallsExecutor: config.ScCallsExecutorConfig{
			ExtraGasToExecute:               6000000,
			MaxGasLimitToUse:                249999999,
			GasLimitForOutOfGasTransactions: 30000000,
			PollingIntervalInMillis:         10000,
		},
		Filter: config.PendingOperationsFilterConfig{
			DeniedEthAddresses:  nil,
			AllowedEthAddresses: []string{"*"},
			DeniedMvxAddresses:  nil,
			AllowedMvxAddresses: []string{"*"},
			DeniedTokens:        nil,
			AllowedTokens:       []string{"*"},
		},
	}
}

func TestNewScCallsModule(t *testing.T) {
	t.Parallel()

	t.Run("invalid filter config should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg.Filter.DeniedTokens = []string{"*"}

		module, err := NewScCallsModule(cfg, &testsCommon.LoggerStub{}, nil)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "unsupported marker * on item at index 0 in list DeniedTokens")
		assert.Nil(t, module)
	})
	t.Run("invalid proxy cacher interval expiration should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg.General.ProxyCacherExpirationSeconds = 0

		module, err := NewScCallsModule(cfg, &testsCommon.LoggerStub{}, nil)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid caching duration, provided: 0s, minimum: 1s")
		assert.Nil(t, module)
	})
	t.Run("invalid resend interval should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg.General.IntervalToResendTxsInSeconds = 0

		module, err := NewScCallsModule(cfg, &testsCommon.LoggerStub{}, nil)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid value for intervalToResend in NewNonceTransactionHandlerV2")
		assert.Nil(t, module)
	})
	t.Run("invalid private key file should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg.General.PrivateKeyFile = ""

		module, err := NewScCallsModule(cfg, &testsCommon.LoggerStub{}, nil)
		assert.NotNil(t, err)
		assert.Nil(t, module)
	})
	t.Run("invalid polling interval should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg.ScCallsExecutor.PollingIntervalInMillis = 0

		module, err := NewScCallsModule(cfg, &testsCommon.LoggerStub{}, nil)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid value for PollingInterval")
		assert.Nil(t, module)
	})
	t.Run("should work with nil close app chan", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		module, err := NewScCallsModule(cfg, &testsCommon.LoggerStub{}, nil)
		assert.Nil(t, err)
		assert.NotNil(t, module)

		assert.Zero(t, module.GetNumSentTransaction())

		err = module.Close()
		assert.Nil(t, err)
	})
	t.Run("should work with not nil close app chan", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg.TransactionChecks.CheckTransactionResults = true
		cfg.TransactionChecks.TimeInSecondsBetweenChecks = 1
		cfg.TransactionChecks.ExecutionTimeoutInSeconds = 1
		cfg.TransactionChecks.CloseAppOnError = true
		module, err := NewScCallsModule(cfg, &testsCommon.LoggerStub{}, make(chan struct{}, 1))
		assert.Nil(t, err)
		assert.NotNil(t, module)

		assert.Zero(t, module.GetNumSentTransaction())

		err = module.Close()
		assert.Nil(t, err)
	})
	t.Run("should work with not nil close app chan and 2 sc proxy addresses", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg.General.ScProxyBech32Addresses = append(cfg.General.ScProxyBech32Addresses, "erd1qqqqqqqqqqqqqpgqzyuaqg3dl7rqlkudrsnm5ek0j3a97qevd8sszj0glf")
		cfg.TransactionChecks.CheckTransactionResults = true
		cfg.TransactionChecks.TimeInSecondsBetweenChecks = 1
		cfg.TransactionChecks.ExecutionTimeoutInSeconds = 1
		cfg.TransactionChecks.CloseAppOnError = true
		module, err := NewScCallsModule(cfg, &testsCommon.LoggerStub{}, make(chan struct{}, 1))
		assert.Nil(t, err)
		assert.NotNil(t, module)

		assert.Zero(t, module.GetNumSentTransaction())

		err = module.Close()
		assert.Nil(t, err)
	})
}
