package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/config"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
	p2pMocks "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/p2p"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	"github.com/stretchr/testify/assert"
)

func createMockEthElrondBridgeArgs() ArgsEthereumToElrondBridge {
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
		assert.Nil(t, err)
		assert.NotNil(t, components)
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
}
