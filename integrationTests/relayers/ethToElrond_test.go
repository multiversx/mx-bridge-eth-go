package relayers

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients/ethereum/contract"
	"github.com/ElrondNetwork/elrond-eth-bridge/config"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/factory"
	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests"
	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-eth-bridge/status"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	elrondConfig "github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelayersShouldExecuteTransferFromEthToElrond(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	safeContractEthAddress := testsCommon.CreateRandomEthereumAddress()
	token1Erc20 := testsCommon.CreateRandomEthereumAddress()
	ticker1 := "tck-000001"

	token2Erc20 := testsCommon.CreateRandomEthereumAddress()
	ticker2 := "tck-000002"

	value1 := big.NewInt(111111111)
	destination1 := testsCommon.CreateRandomElrondAddress()
	depositor1 := testsCommon.CreateRandomEthereumAddress()

	value2 := big.NewInt(222222222)
	destination2 := testsCommon.CreateRandomElrondAddress()
	depositor2 := testsCommon.CreateRandomEthereumAddress()

	tokens := []common.Address{token1Erc20, token2Erc20}
	availableBalances := []*big.Int{value1, value2}

	erc20ContractsHolder := createMockErc20ContractsHolder(tokens, safeContractEthAddress, availableBalances)

	batchNonceOnEthereum := uint64(345)
	txNonceOnEthereum := uint64(772634)
	batch := contract.Batch{
		Nonce:                  big.NewInt(int64(batchNonceOnEthereum) + 1),
		Timestamp:              big.NewInt(0),
		LastUpdatedBlockNumber: big.NewInt(0),
		Deposits: []contract.Deposit{
			{
				Nonce:        big.NewInt(int64(txNonceOnEthereum) + 1),
				TokenAddress: token1Erc20,
				Amount:       value1,
				Depositor:    depositor1,
				Recipient:    destination1.AddressSlice(),
				Status:       0,
			},
			{
				Nonce:        big.NewInt(int64(txNonceOnEthereum) + 2),
				TokenAddress: token2Erc20,
				Amount:       value2,
				Depositor:    depositor2,
				Recipient:    destination2.AddressSlice(),
				Status:       0,
			},
		},
	}

	numRelayers := 3
	ethereumChainMock := mock.NewEthereumChainMock()
	ethereumChainMock.AddBatch(batch)
	ethereumChainMock.SetQuorum(numRelayers)

	elrondChainMock := mock.NewElrondChainMock()
	elrondChainMock.AddTokensPair(token1Erc20, ticker1)
	elrondChainMock.AddTokensPair(token2Erc20, ticker2)
	elrondChainMock.SetLastExecutedEthBatchID(batchNonceOnEthereum)
	elrondChainMock.SetLastExecutedEthTxId(txNonceOnEthereum)
	elrondChainMock.GetStatusesAfterExecutionHandler = func() []byte {
		return []byte{clients.Executed, clients.Rejected}
	}
	elrondChainMock.SetQuorum(numRelayers)

	relayers := make([]bridgeComponents, 0, numRelayers)
	defer func() {
		for _, r := range relayers {
			_ = r.Close()
		}
	}()

	messengers := integrationTests.CreateLinkedMessengers(numRelayers)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()
	elrondChainMock.ProcessFinishedHandler = func() {
		log.Info("elrondChainMock.ProcessFinishedHandler called")
		asyncCancelCall(cancel, time.Second*5)
	}

	for i := 0; i < numRelayers; i++ {
		argsBridgeComponents := createMockBridgeComponentsArgs(i, messengers[i], elrondChainMock, ethereumChainMock)
		argsBridgeComponents.Configs.GeneralConfig.Eth.SafeContractAddress = safeContractEthAddress.Hex()
		argsBridgeComponents.Erc20ContractsHolder = erc20ContractsHolder
		relayer, err := factory.NewEthElrondBridgeComponents(argsBridgeComponents)
		require.Nil(t, err)

		elrondChainMock.AddRelayer(relayer.ElrondRelayerAddress())
		ethereumChainMock.AddRelayer(relayer.EthereumRelayerAddress())

		go func() {
			err = relayer.Start()
			integrationTests.Log.LogIfError(err)
			require.Nil(t, err)
		}()

		relayers = append(relayers, relayer)
	}

	<-ctx.Done()
	time.Sleep(time.Second * 5)

	assert.NotNil(t, elrondChainMock.PerformedActionID())
	transfer := elrondChainMock.ProposedTransfer()
	require.NotNil(t, transfer)
	require.Equal(t, 2, len(transfer.Transfers))
	assert.Equal(t, batchNonceOnEthereum+1, transfer.BatchId.Uint64())

	assert.Equal(t, destination1.AddressBytes(), transfer.Transfers[0].To)
	assert.Equal(t, hex.EncodeToString([]byte(ticker1)), transfer.Transfers[0].Token)
	assert.Equal(t, value1, transfer.Transfers[0].Amount)
	assert.Equal(t, depositor1, common.BytesToAddress(transfer.Transfers[0].From))
	assert.Equal(t, txNonceOnEthereum+1, transfer.Transfers[0].Nonce.Uint64())

	assert.Equal(t, destination2.AddressBytes(), transfer.Transfers[1].To)
	assert.Equal(t, hex.EncodeToString([]byte(ticker2)), transfer.Transfers[1].Token)
	assert.Equal(t, value2, transfer.Transfers[1].Amount)
	assert.Equal(t, depositor2, common.BytesToAddress(transfer.Transfers[1].From))
	assert.Equal(t, txNonceOnEthereum+2, transfer.Transfers[1].Nonce.Uint64())
}

func createMockBridgeComponentsArgs(
	index int,
	messenger p2p.Messenger,
	elrondChainMock *mock.ElrondChainMock,
	ethereumChainMock *mock.EthereumChainMock,
) factory.ArgsEthereumToElrondBridge {

	generalConfigs := createBridgeComponentsConfig(index)
	return factory.ArgsEthereumToElrondBridge{
		Configs: config.Configs{
			GeneralConfig:   generalConfigs,
			ApiRoutesConfig: config.ApiRoutesConfig{},
			FlagsConfig: config.ContextFlagsConfig{
				RestApiInterface: core.WebServerOffString,
			},
		},
		Proxy:                elrondChainMock,
		ClientWrapper:        ethereumChainMock,
		Messenger:            messenger,
		StatusStorer:         testsCommon.NewStorerMock(),
		TimeForBootstrap:     time.Second * 5,
		TimeBeforeRepeatJoin: time.Second * 30,
		MetricsHolder:        status.NewMetricsHolder(),
		AppStatusHandler:     &statusHandler.AppStatusHandlerStub{},
	}
}

func createBridgeComponentsConfig(index int) config.Config {
	stateMachineConfig := config.ConfigStateMachine{
		StepDurationInMillis:       1000,
		IntervalForLeaderInSeconds: 60,
	}

	return config.Config{
		Eth: config.EthereumConfig{
			NetworkAddress:               "mock",
			MultisigContractAddress:      "3009d97FfeD62E57d444e552A9eDF9Ee6Bc8644c",
			PrivateKeyFile:               fmt.Sprintf("testdata/ethereum%d.sk", index),
			IntervalToResendTxsInSeconds: 10,
			GasLimitBase:                 50000,
			GasLimitForEach:              30000,
			GasStation: config.GasStationConfig{
				Enabled: false,
			},
			MaxRetriesOnQuorumReached:          1,
			IntervalToWaitForTransferInSeconds: 1,
		},
		Elrond: config.ElrondConfig{
			NetworkAddress:                  "mock",
			MultisigContractAddress:         "erd1qqqqqqqqqqqqqpgqzyuaqg3dl7rqlkudrsnm5ek0j3a97qevd8sszj0glf",
			PrivateKeyFile:                  fmt.Sprintf("testdata/elrond%d.pem", index),
			IntervalToResendTxsInSeconds:    10,
			GasMap:                          testsCommon.CreateTestElrondGasMap(),
			MaxRetriesOnQuorumReached:       1,
			MaxRetriesOnWasTransferProposed: 3,
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
