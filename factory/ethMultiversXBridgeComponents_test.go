package factory

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-bridge-eth-go/clients/chain"
	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/status"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	bridgeTests "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	p2pMocks "github.com/multiversx/mx-bridge-eth-go/testsCommon/p2p"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/blockchain"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockEthMultiversXBridgeArgs() ArgsEthereumToMultiversXBridge {
	stateMachineConfig := config.ConfigStateMachine{
		StepDurationInMillis:       1000,
		IntervalForLeaderInSeconds: 60,
	}

	cfg := config.Config{
		Eth: config.EthereumConfig{
			Chain:                        chain.Ethereum,
			NetworkAddress:               "http://127.0.0.1:8545",
			SafeContractAddress:          "5DdDe022a65F8063eE9adaC54F359CBF46166068",
			PrivateKeyFile:               "testdata/grace.sk",
			IntervalToResendTxsInSeconds: 0,
			GasLimitBase:                 200000,
			GasLimitForEach:              30000,
			GasStation: config.GasStationConfig{
				Enabled:                    true,
				URL:                        "",
				PollingIntervalInSeconds:   1,
				RequestRetryDelayInSeconds: 1,
				MaxFetchRetries:            3,
				RequestTimeInSeconds:       1,
				MaximumAllowedGasPrice:     100,
				GasPriceSelector:           "FastGasPrice",
				GasPriceMultiplier:         1,
			},
			MaxRetriesOnQuorumReached:          1,
			IntervalToWaitForTransferInSeconds: 1,
			ClientAvailabilityAllowDelta:       10,
		},
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
			"EthereumToMultiversX": stateMachineConfig,
			"MultiversXToEthereum": stateMachineConfig,
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
	return ArgsEthereumToMultiversXBridge{
		Configs:                       configs,
		Messenger:                     &p2pMocks.MessengerStub{},
		StatusStorer:                  testsCommon.NewStorerMock(),
		Proxy:                         proxy,
		MultiversXClientStatusHandler: &testsCommon.StatusHandlerStub{},
		Erc20ContractsHolder:          &bridgeTests.ERC20ContractsHolderStub{},
		ClientWrapper:                 &bridgeTests.EthereumClientWrapperStub{},
		TimeForBootstrap:              minTimeForBootstrap,
		TimeBeforeRepeatJoin:          minTimeBeforeRepeatJoin,
		MetricsHolder:                 status.NewMetricsHolder(),
		AppStatusHandler:              &statusHandler.AppStatusHandlerStub{},
	}
}

func TestNewEthMultiversXBridgeComponents(t *testing.T) {
	t.Parallel()

	t.Run("nil Proxy", func(t *testing.T) {
		t.Parallel()
		args := createMockEthMultiversXBridgeArgs()
		args.Proxy = nil

		components, err := NewEthMultiversXBridgeComponents(args)
		assert.Equal(t, errNilProxy, err)
		assert.Nil(t, components)
	})
	t.Run("nil Messenger", func(t *testing.T) {
		t.Parallel()
		args := createMockEthMultiversXBridgeArgs()
		args.Messenger = nil

		components, err := NewEthMultiversXBridgeComponents(args)
		assert.Equal(t, errNilMessenger, err)
		assert.Nil(t, components)
	})
	t.Run("nil ClientWrapper", func(t *testing.T) {
		t.Parallel()
		args := createMockEthMultiversXBridgeArgs()
		args.ClientWrapper = nil

		components, err := NewEthMultiversXBridgeComponents(args)
		assert.Equal(t, errNilEthClient, err)
		assert.Nil(t, components)
	})
	t.Run("nil StatusStorer", func(t *testing.T) {
		t.Parallel()
		args := createMockEthMultiversXBridgeArgs()
		args.StatusStorer = nil

		components, err := NewEthMultiversXBridgeComponents(args)
		assert.Equal(t, errNilStatusStorer, err)
		assert.Nil(t, components)
	})
	t.Run("nil Erc20ContractsHolder", func(t *testing.T) {
		t.Parallel()
		args := createMockEthMultiversXBridgeArgs()
		args.Erc20ContractsHolder = nil

		components, err := NewEthMultiversXBridgeComponents(args)
		assert.Equal(t, errNilErc20ContractsHolder, err)
		assert.Nil(t, components)
	})
	t.Run("err on createMultiversXKeysAndAddresses, empty pk file", func(t *testing.T) {
		t.Parallel()
		args := createMockEthMultiversXBridgeArgs()
		args.Configs.GeneralConfig.MultiversX.PrivateKeyFile = ""

		components, err := NewEthMultiversXBridgeComponents(args)
		assert.NotNil(t, err)
		assert.Nil(t, components)
	})
	t.Run("err on createMultiversXKeysAndAddresses, empty multisig address", func(t *testing.T) {
		t.Parallel()
		args := createMockEthMultiversXBridgeArgs()
		args.Configs.GeneralConfig.MultiversX.MultisigContractAddress = ""

		components, err := NewEthMultiversXBridgeComponents(args)
		assert.NotNil(t, err)
		assert.Nil(t, components)
	})
	t.Run("err on createMultiversXClient", func(t *testing.T) {
		t.Parallel()
		args := createMockEthMultiversXBridgeArgs()
		args.Configs.GeneralConfig.MultiversX.GasMap = config.MultiversXGasMapConfig{}

		components, err := NewEthMultiversXBridgeComponents(args)
		assert.NotNil(t, err)
		assert.Nil(t, components)
	})
	t.Run("err on createMultiversXRoleProvider", func(t *testing.T) {
		t.Parallel()
		args := createMockEthMultiversXBridgeArgs()
		args.Configs.GeneralConfig.Relayer.RoleProvider.PollingIntervalInMillis = 0

		components, err := NewEthMultiversXBridgeComponents(args)
		assert.NotNil(t, err)
		assert.Nil(t, components)
	})
	t.Run("err on createEthereumClient, empty eth config", func(t *testing.T) {
		t.Parallel()
		args := createMockEthMultiversXBridgeArgs()
		args.Configs.GeneralConfig.Eth = config.EthereumConfig{}

		components, err := NewEthMultiversXBridgeComponents(args)
		assert.NotNil(t, err)
		assert.Nil(t, components)
	})
	t.Run("err on createEthereumClient, invalid gas price selector", func(t *testing.T) {
		t.Parallel()
		args := createMockEthMultiversXBridgeArgs()
		args.Configs.GeneralConfig.Eth.GasStation.GasPriceSelector = core.WebServerOffString

		components, err := NewEthMultiversXBridgeComponents(args)
		assert.NotNil(t, err)
		assert.Nil(t, components)
	})
	t.Run("err missing state machine config", func(t *testing.T) {
		t.Parallel()
		args := createMockEthMultiversXBridgeArgs()
		args.Configs.GeneralConfig.StateMachine = make(map[string]config.ConfigStateMachine)

		components, err := NewEthMultiversXBridgeComponents(args)
		assert.True(t, errors.Is(err, errMissingConfig))
		assert.True(t, strings.Contains(err.Error(), args.Configs.GeneralConfig.Eth.Chain.EvmCompatibleChainToMultiversXName()))
		assert.Nil(t, components)
	})
	t.Run("invalid time for bootstrap", func(t *testing.T) {
		t.Parallel()
		args := createMockEthMultiversXBridgeArgs()
		args.TimeForBootstrap = minTimeForBootstrap - 1

		components, err := NewEthMultiversXBridgeComponents(args)
		assert.True(t, errors.Is(err, errInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for TimeForBootstrap"))
		assert.Nil(t, components)
	})
	t.Run("invalid time before retry", func(t *testing.T) {
		t.Parallel()
		args := createMockEthMultiversXBridgeArgs()
		args.TimeBeforeRepeatJoin = minTimeBeforeRepeatJoin - 1

		components, err := NewEthMultiversXBridgeComponents(args)
		assert.True(t, errors.Is(err, errInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for TimeBeforeRepeatJoin"))
		assert.Nil(t, components)
	})
	t.Run("nil MetricsHolder", func(t *testing.T) {
		t.Parallel()
		args := createMockEthMultiversXBridgeArgs()
		args.MetricsHolder = nil

		components, err := NewEthMultiversXBridgeComponents(args)
		assert.Equal(t, errNilMetricsHolder, err)
		assert.Nil(t, components)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		args := createMockEthMultiversXBridgeArgs()

		components, err := NewEthMultiversXBridgeComponents(args)
		require.Nil(t, err)
		require.NotNil(t, components)
		require.Equal(t, 6, len(components.closableHandlers))
		require.False(t, check.IfNil(components.ethToMultiversXStatusHandler))
		require.False(t, check.IfNil(components.multiversXToEthStatusHandler))
	})
}

func TestEthMultiversXBridgeComponents_StartAndCloseShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockEthMultiversXBridgeArgs()
	components, err := NewEthMultiversXBridgeComponents(args)
	assert.Nil(t, err)

	err = components.Start()
	assert.Nil(t, err)
	assert.Equal(t, 6, len(components.closableHandlers))

	time.Sleep(time.Second * 2) // allow go routines to start

	err = components.Close()
	assert.Nil(t, err)
}

func TestEthMultiversXBridgeComponents_Start(t *testing.T) {
	t.Parallel()

	t.Run("messenger errors on bootstrap", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockEthMultiversXBridgeArgs()
		args.Messenger = &p2pMocks.MessengerStub{
			BootstrapCalled: func() error {
				return expectedErr
			},
		}
		components, _ := NewEthMultiversXBridgeComponents(args)

		err := components.Start()
		assert.Equal(t, expectedErr, err)
	})
	t.Run("broadcaster errors on RegisterOnTopics", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockEthMultiversXBridgeArgs()
		components, _ := NewEthMultiversXBridgeComponents(args)
		components.broadcaster = &testsCommon.BroadcasterStub{
			RegisterOnTopicsCalled: func() error {
				return expectedErr
			},
		}

		err := components.Start()
		assert.Equal(t, expectedErr, err)
	})
}

func TestEthMultiversXBridgeComponents_Close(t *testing.T) {
	t.Parallel()

	t.Run("nil closable should not panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, fmt.Sprintf("should have not failed %v", r))
			}
		}()

		components := &ethMultiversXBridgeComponents{
			baseLogger: logger.GetOrCreate("test"),
		}
		components.addClosableComponent(nil)

		err := components.Close()
		assert.Nil(t, err)
	})
	t.Run("one component errors, should return error", func(t *testing.T) {
		t.Parallel()

		components := &ethMultiversXBridgeComponents{
			baseLogger: logger.GetOrCreate("test"),
		}

		expectedErr := errors.New("expected error")

		numCalls := 0
		components.addClosableComponent(&testsCommon.CloserStub{
			CloseCalled: func() error {
				numCalls++
				return nil
			},
		})
		components.addClosableComponent(&testsCommon.CloserStub{
			CloseCalled: func() error {
				numCalls++
				return expectedErr
			},
		})
		components.addClosableComponent(&testsCommon.CloserStub{
			CloseCalled: func() error {
				numCalls++
				return nil
			},
		})

		err := components.Close()
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 3, numCalls)
	})
}

func TestEthMultiversXBridgeComponents_startBroadcastJoinRetriesLoop(t *testing.T) {
	t.Parallel()

	t.Run("close before minTimeBeforeRepeatJoin", func(t *testing.T) {
		t.Parallel()

		numberOfCalls := uint32(0)
		args := createMockEthMultiversXBridgeArgs()
		components, _ := NewEthMultiversXBridgeComponents(args)

		components.broadcaster = &testsCommon.BroadcasterStub{
			BroadcastJoinTopicCalled: func() {
				atomic.AddUint32(&numberOfCalls, 1)
			},
		}

		err := components.Start()
		assert.Nil(t, err)
		time.Sleep(time.Second * 3)

		err = components.Close()
		assert.Nil(t, err)
		assert.Equal(t, uint32(1), atomic.LoadUint32(&numberOfCalls)) // one call expected from Start
	})
	t.Run("broadcast should be called again", func(t *testing.T) {
		t.Parallel()

		numberOfCalls := uint32(0)
		args := createMockEthMultiversXBridgeArgs()
		components, _ := NewEthMultiversXBridgeComponents(args)
		components.timeBeforeRepeatJoin = time.Second * 3
		components.broadcaster = &testsCommon.BroadcasterStub{
			BroadcastJoinTopicCalled: func() {
				atomic.AddUint32(&numberOfCalls, 1)
			},
		}

		err := components.Start()
		assert.Nil(t, err)
		time.Sleep(time.Second * 7)

		err = components.Close()
		assert.Nil(t, err)
		assert.Equal(t, uint32(3), atomic.LoadUint32(&numberOfCalls)) // 3 calls expected: Start + 2 times from loop
	})
}

func TestEthMultiversXBridgeComponents_RelayerAddresses(t *testing.T) {
	t.Parallel()

	args := createMockEthMultiversXBridgeArgs()
	components, _ := NewEthMultiversXBridgeComponents(args)

	bech32Address, _ := components.MultiversXRelayerAddress().AddressAsBech32String()
	assert.Equal(t, "erd1r69gk66fmedhhcg24g2c5kn2f2a5k4kvpr6jfw67dn2lyydd8cfswy6ede", bech32Address)
	assert.Equal(t, "0x3FE464Ac5aa562F7948322F92020F2b668D543d8", components.EthereumRelayerAddress().String())
}
