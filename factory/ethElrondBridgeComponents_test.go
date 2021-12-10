package factory

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/config"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/stateMachine"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
	p2pMocks "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/p2p"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockEthElrondBridgeArgs() ArgsEthereumToElrondBridge {
	stateMachineConfig := config.ConfigStateMachine{
		StepDurationInMillis: 1000,
	}

	cfg := config.Config{
		Eth: config.EthereumConfig{
			NetworkAddress:               "http://127.0.0.1:8545",
			SafeContractAddress:          "5DdDe022a65F8063eE9adaC54F359CBF46166068",
			PrivateKeyFile:               "testdata/grace.sk",
			IntervalToResendTxsInSeconds: 0,
			GasLimit:                     500000,
			GasStation: config.GasStationConfig{
				Enabled:                  true,
				URL:                      "",
				PollingIntervalInSeconds: 1,
				RequestTimeInSeconds:     1,
				MaximumAllowedGasPrice:   1000,
				GasPriceSelector:         "fast",
			},
		},
		Elrond: config.ElrondConfig{
			IntervalToResendTxsInSeconds: 60,
			PrivateKeyFile:               "testdata/grace.pem",
			NetworkAddress:               "http://127.0.0.1:8079",
			MultisigContractAddress:      "erd1qqqqqqqqqqqqqpgqgftcwj09u0nhmskrw7xxqcqh8qmzwyexd8ss7ftcxx",
			GasMap:                       testsCommon.CreateTestElrondGasMap(),
		},
		Relayer: config.ConfigRelayer{
			RoleProvider: config.RoleProviderConfig{
				PollingIntervalInMillis: 1000,
			},
		},
		StateMachine: map[string]config.ConfigStateMachine{
			"EthToElrond": stateMachineConfig,
			"ElrondToEth": stateMachineConfig,
		},
	}
	configs := config.Configs{
		GeneralConfig:   cfg,
		ApiRoutesConfig: config.ApiRoutesConfig{},
		FlagsConfig: config.ContextFlagsConfig{
			RestApiInterface: core.WebServerOffString,
		},
	}

	return ArgsEthereumToElrondBridge{
		Configs:              configs,
		Messenger:            &p2pMocks.MessengerStub{},
		StatusStorer:         testsCommon.NewStorerMock(),
		Proxy:                blockchain.NewElrondProxy(cfg.Elrond.NetworkAddress, nil),
		Erc20ContractsHolder: &bridgeV2.ERC20ContractsHolderStub{},
		ClientWrapper:        &bridgeV2.EthereumClientWrapperStub{},
	}
}

func TestNewEthElrondBridgeComponents(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		args := createMockEthElrondBridgeArgs()

		components, err := NewEthElrondBridgeComponents(args)
		require.Nil(t, err)
		require.NotNil(t, components)
		require.Equal(t, 4, len(components.closableHandlers))
	})
	t.Run("nil Proxy", func(t *testing.T) {
		t.Parallel()
		args := createMockEthElrondBridgeArgs()
		args.Proxy = nil

		components, err := NewEthElrondBridgeComponents(args)
		assert.Equal(t, errNilProxy, err)
		assert.Nil(t, components)
	})
	t.Run("nil Messenger", func(t *testing.T) {
		t.Parallel()
		args := createMockEthElrondBridgeArgs()
		args.Messenger = nil

		components, err := NewEthElrondBridgeComponents(args)
		assert.Equal(t, errNilMessenger, err)
		assert.Nil(t, components)
	})
	t.Run("nil ClientWrapper", func(t *testing.T) {
		t.Parallel()
		args := createMockEthElrondBridgeArgs()
		args.ClientWrapper = nil

		components, err := NewEthElrondBridgeComponents(args)
		assert.Equal(t, errNilEthClient, err)
		assert.Nil(t, components)
	})
	t.Run("nil StatusStorer", func(t *testing.T) {
		t.Parallel()
		args := createMockEthElrondBridgeArgs()
		args.StatusStorer = nil

		components, err := NewEthElrondBridgeComponents(args)
		assert.Equal(t, errNilStatusStorer, err)
		assert.Nil(t, components)
	})
	t.Run("nil Erc20ContractsHolder", func(t *testing.T) {
		t.Parallel()
		args := createMockEthElrondBridgeArgs()
		args.Erc20ContractsHolder = nil

		components, err := NewEthElrondBridgeComponents(args)
		assert.Equal(t, errNilErc20ContractsHolder, err)
		assert.Nil(t, components)
	})
	t.Run("err on createElrondKeysAndAddresses, empty pk file", func(t *testing.T) {
		t.Parallel()
		args := createMockEthElrondBridgeArgs()
		args.Configs.GeneralConfig.Elrond.PrivateKeyFile = ""

		components, err := NewEthElrondBridgeComponents(args)
		assert.NotNil(t, err)
		assert.Nil(t, components)
	})
	t.Run("err on createElrondKeysAndAddresses, empty multisig address", func(t *testing.T) {
		t.Parallel()
		args := createMockEthElrondBridgeArgs()
		args.Configs.GeneralConfig.Elrond.MultisigContractAddress = ""

		components, err := NewEthElrondBridgeComponents(args)
		assert.NotNil(t, err)
		assert.Nil(t, components)
	})
	t.Run("err on createElrondClient", func(t *testing.T) {
		t.Parallel()
		args := createMockEthElrondBridgeArgs()
		args.Configs.GeneralConfig.Elrond.GasMap = config.ElrondGasMapConfig{}

		components, err := NewEthElrondBridgeComponents(args)
		assert.NotNil(t, err)
		assert.Nil(t, components)
	})
	t.Run("err on createElrondRoleProvider", func(t *testing.T) {
		t.Parallel()
		args := createMockEthElrondBridgeArgs()
		args.Configs.GeneralConfig.Relayer.RoleProvider.PollingIntervalInMillis = 0

		components, err := NewEthElrondBridgeComponents(args)
		assert.NotNil(t, err)
		assert.Nil(t, components)
	})
	t.Run("err on createEthereumClient, empty eth config", func(t *testing.T) {
		t.Parallel()
		args := createMockEthElrondBridgeArgs()
		args.Configs.GeneralConfig.Eth = config.EthereumConfig{}

		components, err := NewEthElrondBridgeComponents(args)
		assert.NotNil(t, err)
		assert.Nil(t, components)
	})
	t.Run("err on createEthereumClient, invalid gas price selector", func(t *testing.T) {
		t.Parallel()
		args := createMockEthElrondBridgeArgs()
		args.Configs.GeneralConfig.Eth.GasStation.GasPriceSelector = core.WebServerOffString

		components, err := NewEthElrondBridgeComponents(args)
		assert.NotNil(t, err)
		assert.Nil(t, components)
	})
	t.Run("err missing state machine config", func(t *testing.T) {
		t.Parallel()
		args := createMockEthElrondBridgeArgs()
		args.Configs.GeneralConfig.StateMachine = make(map[string]config.ConfigStateMachine)

		components, err := NewEthElrondBridgeComponents(args)
		assert.True(t, errors.Is(err, errMissingConfig))
		assert.True(t, strings.Contains(err.Error(), ethToElrondName))
		assert.Nil(t, components)
	})
}

func TestEthElrondBridgeComponents_StartAndCloseShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockEthElrondBridgeArgs()
	components, err := NewEthElrondBridgeComponents(args)
	assert.Nil(t, err)

	err = components.Start()
	assert.Nil(t, err)
	assert.Equal(t, 5, len(components.closableHandlers))

	time.Sleep(time.Second * 2) //allow go routines to start

	err = components.Close()
	assert.Nil(t, err)
}

func TestEthElrondBridgeComponents_StartWithBadInitializationShouldError(t *testing.T) {
	t.Parallel()

	args := createMockEthElrondBridgeArgs()
	components, _ := NewEthElrondBridgeComponents(args)
	components.ethToElrondMachineStates = nil

	err := components.Start()
	assert.Equal(t, stateMachine.ErrNilStepsMap, err)
}

func TestEthElrondBridgeComponents_Close(t *testing.T) {
	t.Parallel()

	t.Run("nil closable should not panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, fmt.Sprintf("should have not failed %v", r))
			}
		}()

		components := &ethElrondBridgeComponents{
			baseLogger: logger.GetOrCreate("test"),
		}
		components.addClosableComponent(nil)

		err := components.Close()
		assert.Nil(t, err)
	})
	t.Run("one component errors, should return error", func(t *testing.T) {
		t.Parallel()

		components := &ethElrondBridgeComponents{
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
