package module

import (
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/interactors"
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
			TTLForFailedRefundIdInSeconds:   3600,
		},
		RefundExecutor: config.RefundExecutorConfig{
			GasToExecute:                  30000000,
			PollingIntervalInMillis:       10000,
			TTLForFailedRefundIdInSeconds: 86400,
		},
		Filter: config.PendingOperationsFilterConfig{
			DeniedEthAddresses:  nil,
			AllowedEthAddresses: []string{"*"},
			DeniedMvxAddresses:  nil,
			AllowedMvxAddresses: []string{"*"},
			DeniedTokens:        nil,
			AllowedTokens:       []string{"*"},
		},
		TransactionChecks: config.TransactionChecksConfig{
			TimeInSecondsBetweenChecks: 6,
			ExecutionTimeoutInSeconds:  120,
		},
	}
}

func TestNewScCallsModule(t *testing.T) {
	t.Parallel()

	t.Run("nil proxy should error", func(t *testing.T) {
		t.Parallel()
		cfg := createTestConfigs()
		argsScCallsModule := ArgsScCallsModule{
			Cfg:   cfg,
			Proxy: nil,
			Log:   &testsCommon.LoggerStub{},
		}

		module, err := NewScCallsModule(argsScCallsModule)
		assert.NotNil(t, err)
		assert.Equal(t, err, errNilProxy)
		assert.Nil(t, module)
	})
	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()
		cfg := createTestConfigs()
		argsScCallsModule := ArgsScCallsModule{
			Cfg:   cfg,
			Proxy: &interactors.ProxyStub{},
			Log:   nil,
		}

		module, err := NewScCallsModule(argsScCallsModule)
		assert.NotNil(t, err)
		assert.Equal(t, err, errNilLogger)
		assert.Nil(t, module)
	})
	t.Run("invalid filter config should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg.Filter.DeniedTokens = []string{"*"}
		argsScCallsModule := ArgsScCallsModule{
			Cfg:   cfg,
			Proxy: &interactors.ProxyStub{},
			Log:   &testsCommon.LoggerStub{},
		}

		module, err := NewScCallsModule(argsScCallsModule)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "unsupported marker * on item at index 0 in list DeniedTokens")
		assert.Nil(t, module)
	})
	t.Run("invalid resend interval should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg.General.IntervalToResendTxsInSeconds = 0
		argsScCallsModule := ArgsScCallsModule{
			Cfg:   cfg,
			Proxy: &interactors.ProxyStub{},
			Log:   &testsCommon.LoggerStub{},
		}

		module, err := NewScCallsModule(argsScCallsModule)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid value for intervalToResend in NewNonceTransactionHandlerV2")
		assert.Nil(t, module)
	})
	t.Run("invalid private key file should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg.General.PrivateKeyFile = ""
		argsScCallsModule := ArgsScCallsModule{
			Cfg:   cfg,
			Proxy: &interactors.ProxyStub{},
			Log:   &testsCommon.LoggerStub{},
		}

		module, err := NewScCallsModule(argsScCallsModule)
		assert.NotNil(t, err)
		assert.Nil(t, module)
	})
	t.Run("invalid polling interval for SC calls should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg.ScCallsExecutor.PollingIntervalInMillis = 0
		argsScCallsModule := ArgsScCallsModule{
			Cfg:   cfg,
			Proxy: &interactors.ProxyStub{},
			Log:   &testsCommon.LoggerStub{},
		}

		module, err := NewScCallsModule(argsScCallsModule)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid value for PollingInterval")
		assert.Nil(t, module)
	})
	t.Run("invalid max gas to execute for SC calls should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg.ScCallsExecutor.MaxGasLimitToUse = 1
		argsScCallsModule := ArgsScCallsModule{
			Cfg:   cfg,
			Proxy: &interactors.ProxyStub{},
			Log:   &testsCommon.LoggerStub{},
		}

		module, err := NewScCallsModule(argsScCallsModule)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "provided gas limit is less than absolute minimum required for MaxGasLimitToUse")
		assert.Nil(t, module)
	})
	t.Run("invalid polling interval for refunds should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg.RefundExecutor.PollingIntervalInMillis = 0
		argsScCallsModule := ArgsScCallsModule{
			Cfg:   cfg,
			Proxy: &interactors.ProxyStub{},
			Log:   &testsCommon.LoggerStub{},
		}

		module, err := NewScCallsModule(argsScCallsModule)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid value for PollingInterval")
		assert.Nil(t, module)
	})
	t.Run("invalid gas to execute for refunds should error", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		cfg.RefundExecutor.GasToExecute = 0
		argsScCallsModule := ArgsScCallsModule{
			Cfg:   cfg,
			Proxy: &interactors.ProxyStub{},
			Log:   &testsCommon.LoggerStub{},
		}

		module, err := NewScCallsModule(argsScCallsModule)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "provided gas limit is less than absolute minimum required for GasToExecute")
		assert.Nil(t, module)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cfg := createTestConfigs()
		argsScCallsModule := ArgsScCallsModule{
			Cfg:   cfg,
			Proxy: &interactors.ProxyStub{},
			Log:   &testsCommon.LoggerStub{},
		}

		module, err := NewScCallsModule(argsScCallsModule)
		assert.Nil(t, err)
		assert.NotNil(t, module)

		assert.Zero(t, module.GetNumSentTransaction())

		err = module.Close()
		assert.Nil(t, err)
	})
}
