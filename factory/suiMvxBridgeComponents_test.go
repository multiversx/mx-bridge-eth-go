package factory

import (
	"errors"
	"fmt"
	"github.com/block-vision/sui-go-sdk/sui"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-bridge-eth-go/clients/chain"
	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/status"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	p2pMocks "github.com/multiversx/mx-bridge-eth-go/testsCommon/p2p"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/blockchain"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockSuiMultiversXBridgeArgs() ArgsSuiToMultiversXBridge {
	stateMachineConfig := config.ConfigStateMachine{
		StepDurationInMillis:       1000,
		IntervalForLeaderInSeconds: 60,
	}

	cfg := config.Config{
		Sui: config.SuiConfig{
			Chain:                            chain.Sui,
			NetworkAddress:                   "http://127.0.0.1:8545",
			PrivateKeyFile:                   "testdata/grace.seed",
			BridgePackageId:                  "0xd85d37d10bb925c9e598169478c518f3da1090fbb8e027362e1c9c227f6fc4e0",
			BridgeObjectId:                   "0x8e3dc49b158d7cd7a72720160b7e7aa0859cda4a7ebbcb4391dd4d7190777db1",
			BridgeObjectInitialSharedVersion: 123456,
			SafePackageId:                    "0x5ea6aafe995ce6506f07335a40942024106a57f6311cb341239abf2c3ac7b82f",
			SafeObjectId:                     "0x80d7de9c4a56194087e0ba0bf59492aa8e6a5ee881606226930827085ddf2332",
			SafeObjectInitialSharedVersion:   654321,
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
			"SuiToMultiversX": stateMachineConfig,
			"MultiversXToSui": stateMachineConfig,
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

	return ArgsSuiToMultiversXBridge{
		Configs:                       configs,
		Messenger:                     &p2pMocks.MessengerStub{},
		StatusStorer:                  testsCommon.NewStorerMock(),
		Proxy:                         proxy,
		MultiversXClientStatusHandler: &testsCommon.StatusHandlerStub{},
		SuiProxy:                      sui.NewSuiClient(cfg.Sui.NetworkAddress),
		SuiClientStatusHandler:        &testsCommon.StatusHandlerStub{},
		TimeForBootstrap:              minTimeForBootstrap,
		TimeBeforeRepeatJoin:          minTimeBeforeRepeatJoin,
		MetricsHolder:                 status.NewMetricsHolder(),
		AppStatusHandler:              &statusHandler.AppStatusHandlerStub{},
	}
}

func TestNewSuiMvxBridgeComponents(t *testing.T) {
	t.Parallel()

	t.Run("nil Proxy", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiMultiversXBridgeArgs()
		args.Proxy = nil

		components, err := NewSuiMvxBridgeComponents(args)
		assert.True(t, errors.Is(err, errNilProxy))
		assert.Nil(t, components)
	})
	t.Run("nil Messenger", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiMultiversXBridgeArgs()
		args.Messenger = nil

		components, err := NewSuiMvxBridgeComponents(args)
		assert.Equal(t, errNilMessenger, err)
		assert.Nil(t, components)
	})
	t.Run("nil StatusStorer", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiMultiversXBridgeArgs()
		args.StatusStorer = nil

		components, err := NewSuiMvxBridgeComponents(args)
		assert.Equal(t, errNilStatusStorer, err)
		assert.Nil(t, components)
	})
	t.Run("err on createMultiversXKeysAndAddresses, empty pk file", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiMultiversXBridgeArgs()
		args.Configs.GeneralConfig.MultiversX.PrivateKeyFile = ""

		components, err := NewSuiMvxBridgeComponents(args)
		assert.NotNil(t, err)
		assert.Nil(t, components)
	})
	t.Run("err on createMultiversXKeysAndAddresses, empty multisig address", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiMultiversXBridgeArgs()
		args.Configs.GeneralConfig.MultiversX.MultisigContractAddress = ""

		components, err := NewSuiMvxBridgeComponents(args)
		assert.NotNil(t, err)
		assert.Nil(t, components)
	})
	t.Run("err on createMultiversXClient", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiMultiversXBridgeArgs()
		args.Configs.GeneralConfig.MultiversX.GasMap = config.MultiversXGasMapConfig{}

		components, err := NewSuiMvxBridgeComponents(args)
		assert.NotNil(t, err)
		assert.Nil(t, components)
	})
	t.Run("err on createMultiversXRoleProvider", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiMultiversXBridgeArgs()
		args.Configs.GeneralConfig.Relayer.RoleProvider.PollingIntervalInMillis = 0

		components, err := NewSuiMvxBridgeComponents(args)
		assert.NotNil(t, err)
		assert.Nil(t, components)
	})
	t.Run("err nil sui proxy", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiMultiversXBridgeArgs()
		args.Proxy = nil

		components, err := NewSuiMvxBridgeComponents(args)
		assert.True(t, errors.Is(err, errNilProxy))
		assert.Nil(t, components)
	})
	t.Run("err on createSuiClient, empty sui config", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiMultiversXBridgeArgs()
		args.Configs.GeneralConfig.Sui = config.SuiConfig{}

		components, err := NewSuiMvxBridgeComponents(args)
		assert.NotNil(t, err)
		assert.Nil(t, components)
	})
	t.Run("err on createSuiClient, invalid gas price selector", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiMultiversXBridgeArgs()
		args.Configs.GeneralConfig.Sui.GasStation.GasPriceSelector = core.WebServerOffString

		components, err := NewSuiMvxBridgeComponents(args)
		assert.NotNil(t, err)
		assert.Nil(t, components)
	})
	t.Run("err missing state machine config", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiMultiversXBridgeArgs()
		args.Configs.GeneralConfig.StateMachine = make(map[string]config.ConfigStateMachine)

		components, err := NewSuiMvxBridgeComponents(args)
		assert.True(t, errors.Is(err, errMissingConfig))
		assert.True(t, strings.Contains(err.Error(), args.Configs.GeneralConfig.Sui.Chain.PeerChainToMultiversXName()))
		assert.Nil(t, components)
	})
	t.Run("invalid time for bootstrap", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiMultiversXBridgeArgs()
		args.TimeForBootstrap = minTimeForBootstrap - 1

		components, err := NewSuiMvxBridgeComponents(args)
		assert.True(t, errors.Is(err, errInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for TimeForBootstrap"))
		assert.Nil(t, components)
	})
	t.Run("invalid time before retry", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiMultiversXBridgeArgs()
		args.TimeBeforeRepeatJoin = minTimeBeforeRepeatJoin - 1

		components, err := NewSuiMvxBridgeComponents(args)
		assert.True(t, errors.Is(err, errInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for TimeBeforeRepeatJoin"))
		assert.Nil(t, components)
	})
	t.Run("nil MetricsHolder", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiMultiversXBridgeArgs()
		args.MetricsHolder = nil

		components, err := NewSuiMvxBridgeComponents(args)
		assert.Equal(t, errNilMetricsHolder, err)
		assert.Nil(t, components)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		args := createMockSuiMultiversXBridgeArgs()

		components, err := NewSuiMvxBridgeComponents(args)
		require.Nil(t, err)
		require.NotNil(t, components)
		require.Equal(t, 7, len(components.closableHandlers))
		require.False(t, check.IfNil(components.toMultiversXStatusHandler))
		require.False(t, check.IfNil(components.fromMultiversXStatusHandler))
	})
}

func TestSuiMultiversXBridgeComponents_StartAndCloseShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockSuiMultiversXBridgeArgs()
	components, err := NewSuiMvxBridgeComponents(args)
	assert.Nil(t, err)

	err = components.Start()
	assert.Nil(t, err)
	assert.Equal(t, 7, len(components.closableHandlers))

	time.Sleep(time.Second * 2) // allow go routines to start

	err = components.Close()
	assert.Nil(t, err)
}

func TestSuiMultiversXBridgeComponents_Start(t *testing.T) {
	t.Parallel()

	t.Run("messenger errors on bootstrap", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockSuiMultiversXBridgeArgs()
		args.Messenger = &p2pMocks.MessengerStub{
			BootstrapCalled: func() error {
				return expectedErr
			},
		}
		components, _ := NewSuiMvxBridgeComponents(args)

		err := components.Start()
		assert.Equal(t, expectedErr, err)
	})
	t.Run("broadcaster errors on RegisterOnTopics", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockSuiMultiversXBridgeArgs()
		components, _ := NewSuiMvxBridgeComponents(args)
		components.broadcaster = &testsCommon.BroadcasterStub{
			RegisterOnTopicsCalled: func() error {
				return expectedErr
			},
		}

		err := components.Start()
		assert.Equal(t, expectedErr, err)
	})
}

func TestSuiMultiversXBridgeComponents_Close(t *testing.T) {
	t.Parallel()

	t.Run("nil closable should not panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, fmt.Sprintf("should have not failed %v", r))
			}
		}()

		components := &ethMvxBridgeComponents{
			baseBridgeComponents: &baseBridgeComponents{
				baseLogger: logger.GetOrCreate("test"),
			},
		}
		components.addClosableComponent(nil)

		err := components.Close()
		assert.Nil(t, err)
	})
	t.Run("one component errors, should return error", func(t *testing.T) {
		t.Parallel()

		components := &ethMvxBridgeComponents{
			baseBridgeComponents: &baseBridgeComponents{
				baseLogger: logger.GetOrCreate("test"),
			},
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

func TestSuiMultiversXBridgeComponents_startBroadcastJoinRetriesLoop(t *testing.T) {
	t.Parallel()

	t.Run("close before minTimeBeforeRepeatJoin", func(t *testing.T) {
		t.Parallel()

		numberOfCalls := uint32(0)
		args := createMockSuiMultiversXBridgeArgs()
		components, _ := NewSuiMvxBridgeComponents(args)

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
		args := createMockSuiMultiversXBridgeArgs()
		components, _ := NewSuiMvxBridgeComponents(args)
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

func TestSuiMultiversXBridgeComponents_SuiRelayerAddresses(t *testing.T) {
	t.Parallel()

	args := createMockSuiMultiversXBridgeArgs()
	components, _ := NewSuiMvxBridgeComponents(args)

	assert.Equal(t, "0x6519752d8a59e2fe533dee6657ec96703a3886b99c372410baf89e89377eaf47", components.PeerChainRelayerAddress())
}
