package factory

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/status"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	p2pMocks "github.com/multiversx/mx-bridge-eth-go/testsCommon/p2p"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-sdk-go/blockchain"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockBridgeCommonArgs() ArgsBridgeCommon {
	stateMachineConfig := config.ConfigStateMachine{
		StepDurationInMillis:       1000,
		IntervalForLeaderInSeconds: 60,
	}

	cfg := config.Config{
		MultiversX: config.MultiversXConfig{
			PrivateKeyFile:                  "testdata/grace.pem",
			IntervalToResendTxsInSeconds:    60,
			NetworkAddress:                  "http://127.0.0.1:8079",
			MultisigContractAddress:         "erd1qqqqqqqqqqqqqpgqgftcwj09u0nhmskrw7xxqcqh8qmzwyexd8ss7ftcxx",
			SafeContractAddress:             "erd1qqqqqqqqqqqqqpgqgftcwj09u0nhmskrw7xxqcqh8qmzwyexd8ss7ftcxx",
			GasMap:                          testsCommon.CreateTestMultiversXGasMap(),
			MaxRetriesOnQuorumReached:       1,
			MaxRetriesOnWasTransferProposed: 1,
			ClientAvailabilityAllowDelta:    10,
			Proxy: config.ProxyConfig{
				CacherExpirationSeconds: 600,
				RestAPIEntityType:       "observer",
				MaxNoncesDelta:          10,
				FinalityCheck:           true,
			},
		},
		Relayer: config.ConfigRelayer{
			RoleProvider: config.RoleProviderConfig{
				PollingIntervalInMillis: 1000,
			},
		},
		StateMachine: map[string]config.ConfigStateMachine{
			"toMultiversX":   stateMachineConfig,
			"fromMultiversX": stateMachineConfig,
		},
	}
	configs := config.Configs{
		GeneralConfig:   cfg,
		ApiRoutesConfig: config.ApiRoutesConfig{},
		FlagsConfig: config.ContextFlagsConfig{
			RestApiInterface: core.WebServerOffString,
		},
	}

	argsProxy := blockchain.ArgsProxy{
		ProxyURL:            cfg.MultiversX.NetworkAddress,
		CacheExpirationTime: time.Minute,
		EntityType:          sdkCore.ObserverNode,
	}
	proxy, _ := blockchain.NewProxy(argsProxy)
	return ArgsBridgeCommon{
		Configs:                       configs,
		Messenger:                     &p2pMocks.MessengerStub{},
		StatusStorer:                  testsCommon.NewStorerMock(),
		Proxy:                         proxy,
		MultiversXClientStatusHandler: &testsCommon.StatusHandlerStub{},
		TimeForBootstrap:              minTimeForBootstrap,
		TimeBeforeRepeatJoin:          minTimeBeforeRepeatJoin,
		MetricsHolder:                 status.NewMetricsHolder(),
		AppStatusHandler:              &statusHandler.AppStatusHandlerStub{},
	}
}

func TestNewBaseComponents(t *testing.T) {
	t.Parallel()

	t.Run("nil Proxy", func(t *testing.T) {
		t.Parallel()
		args := createMockBridgeCommonArgs()
		args.Proxy = nil

		components, err := NewBaseComponents(args)
		assert.True(t, errors.Is(err, errNilProxy))
		assert.Nil(t, components)
	})
	t.Run("nil Messenger", func(t *testing.T) {
		t.Parallel()
		args := createMockBridgeCommonArgs()
		args.Messenger = nil

		components, err := NewBaseComponents(args)
		assert.Equal(t, errNilMessenger, err)
		assert.Nil(t, components)
	})
	t.Run("nil StatusStorer", func(t *testing.T) {
		t.Parallel()
		args := createMockBridgeCommonArgs()
		args.StatusStorer = nil

		components, err := NewBaseComponents(args)
		assert.Equal(t, errNilStatusStorer, err)
		assert.Nil(t, components)
	})
	t.Run("err on createMultiversXKeysAndAddresses, empty pk file", func(t *testing.T) {
		t.Parallel()
		args := createMockBridgeCommonArgs()
		args.Configs.GeneralConfig.MultiversX.PrivateKeyFile = ""

		components, err := NewBaseComponents(args)
		assert.NotNil(t, err)
		assert.Nil(t, components)
	})
	t.Run("err on createMultiversXKeysAndAddresses, empty multisig address", func(t *testing.T) {
		t.Parallel()
		args := createMockBridgeCommonArgs()
		args.Configs.GeneralConfig.MultiversX.MultisigContractAddress = ""

		components, err := NewBaseComponents(args)
		assert.NotNil(t, err)
		assert.Nil(t, components)
	})
	t.Run("invalid time for bootstrap", func(t *testing.T) {
		t.Parallel()
		args := createMockBridgeCommonArgs()
		args.TimeForBootstrap = minTimeForBootstrap - 1

		components, err := NewBaseComponents(args)
		assert.True(t, errors.Is(err, errInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for TimeForBootstrap"))
		assert.Nil(t, components)
	})
	t.Run("invalid time before retry", func(t *testing.T) {
		t.Parallel()
		args := createMockBridgeCommonArgs()
		args.TimeBeforeRepeatJoin = minTimeBeforeRepeatJoin - 1

		components, err := NewBaseComponents(args)
		assert.True(t, errors.Is(err, errInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for TimeBeforeRepeatJoin"))
		assert.Nil(t, components)
	})
	t.Run("nil MetricsHolder", func(t *testing.T) {
		t.Parallel()
		args := createMockBridgeCommonArgs()
		args.MetricsHolder = nil

		components, err := NewBaseComponents(args)
		assert.Equal(t, errNilMetricsHolder, err)
		assert.Nil(t, components)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		args := createMockBridgeCommonArgs()

		components, err := NewBaseComponents(args)
		require.Nil(t, err)
		require.NotNil(t, components)
		require.Equal(t, 1, len(components.closableHandlers))
	})
}

func TestBaseBridgeComponents_MvxRelayerAddresses(t *testing.T) {
	t.Parallel()

	args := createMockBridgeCommonArgs()
	components, _ := NewBaseComponents(args)

	bech32Address, _ := components.MultiversXRelayerAddress().AddressAsBech32String()
	assert.Equal(t, "erd1r69gk66fmedhhcg24g2c5kn2f2a5k4kvpr6jfw67dn2lyydd8cfswy6ede", bech32Address)
}
