package mock

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-eth-bridge/config"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	elrondConfig "github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// CreateMockRelayArgs will create a mocked ArgsRelayer instance
func CreateMockRelayArgs(
	name string,
	index int,
	messenger p2p.Messenger,
	elrondChainMock *ElrondChainMock,
	ethereumChainMock *EthereumChainMock,
) relay.ArgsRelayer {

	return relay.ArgsRelayer{
		Configs: config.Configs{
			GeneralConfig:   createMockRelayConfig(index),
			ApiRoutesConfig: config.ApiRoutesConfig{},
			FlagsConfig: config.ContextFlagsConfig{
				RestApiInterface: core.WebServerOffString,
			},
		},
		Name:                   name,
		Proxy:                  elrondChainMock,
		EthClient:              ethereumChainMock,
		EthInstance:            ethereumChainMock,
		Messenger:              messenger,
		EthClientStatusHandler: testsCommon.NewStatusHandlerMock("mock"),
		StatusStorer:           testsCommon.NewStorerMock(),
	}
}

func createMockRelayConfig(index int) config.Config {
	stateMachineConfig := config.ConfigStateMachine{
		StepDurationInMillis: 1000,
		Steps: []config.StepConfig{
			{Name: "getting the pending transactions", DurationInMillis: 1000},
			{Name: "proposing transfer", DurationInMillis: 1000},
			{Name: "waiting signatures for propose transfer", DurationInMillis: 1000},
			{Name: "executing transfer", DurationInMillis: 1000},
			{Name: "proposing set status", DurationInMillis: 1000},
			{Name: "waiting signatures for propose set status", DurationInMillis: 1000},
			{Name: "executing set status", DurationInMillis: 1000},
		},
	}

	return config.Config{
		Eth: config.EthereumConfig{
			NetworkAddress:               "mock",
			MultisigContractAddress:      "3009d97FfeD62E57d444e552A9eDF9Ee6Bc8644c",
			PrivateKeyFile:               fmt.Sprintf("testdata/ethereum%d.sk", index),
			IntervalToResendTxsInSeconds: 10,
			GasLimit:                     500000,
			GasStation: config.GasStationConfig{
				Enabled: false,
			},
		},
		Elrond: config.ElrondConfig{
			NetworkAddress:               "mock",
			MultisigContractAddress:      "erd1qqqqqqqqqqqqqpgqzyuaqg3dl7rqlkudrsnm5ek0j3a97qevd8sszj0glf",
			PrivateKeyFile:               fmt.Sprintf("testdata/elrond%d.pem", index),
			IntervalToResendTxsInSeconds: 10,
			GasMap:                       testsCommon.CreateTestElrondGasMap(),
		},
		P2P: config.ConfigP2P{},
		StateMachine: map[string]config.ConfigStateMachine{
			"EthToElrond": stateMachineConfig,
			"ElrondToEth": stateMachineConfig,
		},
		Relayer: config.ConfigRelayer{
			Marshalizer: elrondConfig.MarshalizerConfig{
				Type:           "json",
				SizeCheckDelta: 10,
			},
			RoleProvider: config.RoleProviderConfig{
				PollingIntervalInMillis: 1000,
			},
		},
	}
}
