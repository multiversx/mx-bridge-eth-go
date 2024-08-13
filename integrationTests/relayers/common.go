package relayers

import (
	"context"
	"fmt"
	"math/big"
	"path"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/clients/chain"
	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	bridgeTests "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	chainConfig "github.com/multiversx/mx-chain-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("integrationTests/relayers")

func createMockErc20ContractsHolder(tokens []common.Address, safeContractEthAddress common.Address, availableBalances []*big.Int) *bridgeTests.ERC20ContractsHolderStub {
	return &bridgeTests.ERC20ContractsHolderStub{
		BalanceOfCalled: func(ctx context.Context, erc20Address common.Address, address common.Address) (*big.Int, error) {
			for i, tk := range tokens {
				if tk != erc20Address {
					continue
				}

				if address == safeContractEthAddress {
					return availableBalances[i], nil
				}

				return big.NewInt(0), nil
			}

			return nil, fmt.Errorf("unregistered token %s", erc20Address.Hex())
		},
	}
}

func availableTokensMapToSlices(erc20Map map[common.Address]*big.Int) ([]common.Address, []*big.Int) {
	tokens := make([]common.Address, 0, len(erc20Map))
	availableBalances := make([]*big.Int, 0, len(erc20Map))

	for addr, val := range erc20Map {
		tokens = append(tokens, addr)
		availableBalances = append(availableBalances, val)
	}

	return tokens, availableBalances
}

func closeRelayers(relayers []bridgeComponents) {
	for _, r := range relayers {
		_ = r.Close()
	}
}

// CreateBridgeComponentsConfig -
func CreateBridgeComponentsConfig(index int, workingDir string) config.Config {
	stateMachineConfig := config.ConfigStateMachine{
		StepDurationInMillis:       1000,
		IntervalForLeaderInSeconds: 60,
	}

	return config.Config{
		Eth: config.EthereumConfig{
			Chain:                        chain.Ethereum,
			NetworkAddress:               "mock",
			MultisigContractAddress:      "3009d97FfeD62E57d444e552A9eDF9Ee6Bc8644c",
			PrivateKeyFile:               fmt.Sprintf("testdata/ethereum%d.sk", index),
			IntervalToResendTxsInSeconds: 10,
			GasLimitBase:                 200000,
			GasLimitForEach:              30000,
			GasStation: config.GasStationConfig{
				Enabled: false,
			},
			MaxRetriesOnQuorumReached:          1,
			IntervalToWaitForTransferInSeconds: 1,
			ClientAvailabilityAllowDelta:       5,
			EventsBlockRangeFrom:               -5,
			EventsBlockRangeTo:                 50,
		},
		MultiversX: config.MultiversXConfig{
			NetworkAddress:                  "mock",
			MultisigContractAddress:         "erd1qqqqqqqqqqqqqpgqzyuaqg3dl7rqlkudrsnm5ek0j3a97qevd8sszj0glf",
			SafeContractAddress:             "erd1qqqqqqqqqqqqqpgqtvnswnzxxz8susupesys0hvg7q2z5nawrcjq06qdus",
			PrivateKeyFile:                  path.Join(workingDir, fmt.Sprintf("multiversx%d.pem", index)),
			IntervalToResendTxsInSeconds:    10,
			GasMap:                          testsCommon.CreateTestMultiversXGasMap(),
			MaxRetriesOnQuorumReached:       1,
			MaxRetriesOnWasTransferProposed: 3,
			ClientAvailabilityAllowDelta:    5,
			Proxy: config.ProxyConfig{
				CacherExpirationSeconds: 600,
				RestAPIEntityType:       "observer",
				MaxNoncesDelta:          10,
				FinalityCheck:           true,
			},
		},
		P2P: config.ConfigP2P{},
		StateMachine: map[string]config.ConfigStateMachine{
			"EthereumToMultiversX": stateMachineConfig,
			"MultiversXToEthereum": stateMachineConfig,
		},
		Relayer: config.ConfigRelayer{
			Marshalizer: chainConfig.MarshalizerConfig{
				Type:           "json",
				SizeCheckDelta: 10,
			},
			RoleProvider: config.RoleProviderConfig{
				PollingIntervalInMillis: 1000,
			},
		},
	}
}
