package relayers

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"path"
	"syscall"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/clients/multiversx"
	"github.com/multiversx/mx-bridge-eth-go/config"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/factory"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/mock"
	"github.com/multiversx/mx-bridge-eth-go/status"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/require"
)

const (
	ownerPem                     = "testdata/wallets/owner.pem"
	safeContract                 = "testdata/contracts/esdt-safe.wasm"
	multisigContract             = "testdata/contracts/multisig.wasm"
	multiTransferContract        = "testdata/contracts/multi-transfer-esdt.wasm"
	bridgeProxyContract          = "testdata/contracts/bridge-proxy.wasm"
	aggregatorContract           = "testdata/contracts/aggregator.wasm"
	wrapperContract              = "testdata/contracts/bridged-tokens-wrapper.wasm"
	nodeConfig                   = "testdata/config/nodeConfig"
	proxyConfig                  = "testdata/config/proxyConfig"
	minRelayerStake              = "10000000000000000000" // 10egld
	minRelayerStakeHex           = "8AC7230489E80000"     // 10egld
	slashAmount                  = "00"
	quorum                       = "03"
	relayerPemPathFormat         = "testdata/multiversx%d.pem"
	roundDurationInMs            = 3000
	roundsPerEpoch               = 20
	numOfShards                  = 3
	serverPort                   = 8085
	proxyCacherExpirationSeconds = 600
	proxyMaxNoncesDelta          = 7
	zeroValue                    = "0"
	emptyAddress                 = "erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu"
)

type proxyWithChainSimulator interface {
	Proxy() multiversx.Proxy
	GetNetworkAddress() string
	DeploySC(ctx context.Context, path string, ownerPK string, ownerSK []byte, extraParams []string) (string, error)
	ScCall(ctx context.Context, senderPK string, senderSK []byte, contract string, value string, function string, parameters []string) (string, error)
	SendTx(ctx context.Context, senderPK string, senderSK []byte, receiver string, value string, dataField []byte) (string, error)
	FundWallets(wallets []string)
	Close()
}

type keysHolder struct {
	pk string
	sk []byte
}

func TestRelayersShouldExecuteTransfersFromEthToMultiversXWithChainSimulator(t *testing.T) {
	if testing.Short() {
		t.Skip("this is a long test")
	}

	safeContractEthAddress := testsCommon.CreateRandomEthereumAddress()
	token1Erc20 := testsCommon.CreateRandomEthereumAddress()
	token2Erc20 := testsCommon.CreateRandomEthereumAddress()
	value1 := big.NewInt(111111111)
	value2 := big.NewInt(222222222)
	tokens := []common.Address{token1Erc20, token2Erc20}
	availableBalances := []*big.Int{value1, value2}

	erc20ContractsHolder := createMockErc20ContractsHolder(tokens, safeContractEthAddress, availableBalances)

	numRelayers := 3
	relayersKeys := make([]keysHolder, 0, numRelayers)
	for i := 0; i < numRelayers; i++ {
		relayerSK, relayerPK, err := core.LoadSkPkFromPemFile(fmt.Sprintf(relayerPemPathFormat, i), 0)
		require.Nil(t, err)

		relayersKeys = append(relayersKeys, keysHolder{
			pk: relayerPK,
			sk: relayerSK,
		})
	}
	ethereumChainMock := mock.NewEthereumChainMock()

	multiversXProxyWithChainSimulator := startProxyWithChainSimulator(t)
	defer multiversXProxyWithChainSimulator.Close()

	// deploy all contracts and execute all txs needed
	safeAddress, multisigAddress := executeContractsTxs(t, multiversXProxyWithChainSimulator, relayersKeys)

	// start relayers
	relayers := startRelayers(t, numRelayers, multiversXProxyWithChainSimulator, ethereumChainMock, safeContractEthAddress, erc20ContractsHolder, safeAddress, multisigAddress)
	defer closeRelayers(relayers)

	// wait for signal interrupt or time out
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-interrupt:
			log.Error("signal interrupted")
			return
		case <-time.After(time.Minute * 15):
			log.Error("time out")
			return
		}
	}
}

func startProxyWithChainSimulator(t *testing.T) proxyWithChainSimulator {
	// create a new working directory
	tmpDir := path.Join(t.TempDir(), "test")
	err := os.MkdirAll(tmpDir, os.ModePerm)
	require.NoError(t, err)

	// start the chain simulator
	args := integrationTests.ArgProxyWithChainSimulator{
		BypassTxsSignature:           true,
		WorkingDir:                   tmpDir,
		RoundDurationInMs:            roundDurationInMs,
		RoundsPerEpoch:               roundsPerEpoch,
		NodeConfigs:                  nodeConfig,
		ProxyConfigs:                 proxyConfig,
		NumOfShards:                  numOfShards,
		BlockTimeInMs:                roundDurationInMs,
		ServerPort:                   serverPort,
		ProxyCacherExpirationSeconds: proxyCacherExpirationSeconds,
		ProxyMaxNoncesDelta:          proxyMaxNoncesDelta,
	}
	multiversXProxyWithChainSimulator, err := integrationTests.CreateProxyWithChainSimulator(args)
	require.NoError(t, err)

	return multiversXProxyWithChainSimulator
}

func startRelayers(
	t *testing.T,
	numRelayers int,
	multiversXProxyWithChainSimulator proxyWithChainSimulator,
	ethereumChainMock *mock.EthereumChainMock,
	safeContractEthAddress common.Address,
	erc20ContractsHolder *bridge.ERC20ContractsHolderStub,
	safeAddress string,
	multisigAddress string,
) []bridgeComponents {
	relayers := make([]bridgeComponents, 0, numRelayers)

	messengers := integrationTests.CreateLinkedMessengers(numRelayers)

	for i := 0; i < numRelayers; i++ {
		generalConfigs := createBridgeComponentsConfig(i)
		argsBridgeComponents := factory.ArgsEthereumToMultiversXBridge{
			Configs: config.Configs{
				GeneralConfig:   generalConfigs,
				ApiRoutesConfig: config.ApiRoutesConfig{},
				FlagsConfig: config.ContextFlagsConfig{
					RestApiInterface: bridgeCore.WebServerOffString,
				},
			},
			Proxy:                         multiversXProxyWithChainSimulator.Proxy(),
			ClientWrapper:                 ethereumChainMock,
			Messenger:                     messengers[i],
			StatusStorer:                  testsCommon.NewStorerMock(),
			TimeForBootstrap:              time.Second * 5,
			TimeBeforeRepeatJoin:          time.Second * 30,
			MetricsHolder:                 status.NewMetricsHolder(),
			AppStatusHandler:              &statusHandler.AppStatusHandlerStub{},
			MultiversXClientStatusHandler: &testsCommon.StatusHandlerStub{},
		}
		argsBridgeComponents.Configs.GeneralConfig.Eth.SafeContractAddress = safeContractEthAddress.Hex()
		argsBridgeComponents.Erc20ContractsHolder = erc20ContractsHolder
		argsBridgeComponents.Configs.GeneralConfig.MultiversX.NetworkAddress = multiversXProxyWithChainSimulator.GetNetworkAddress()
		argsBridgeComponents.Configs.GeneralConfig.MultiversX.SafeContractAddress = safeAddress
		argsBridgeComponents.Configs.GeneralConfig.MultiversX.MultisigContractAddress = multisigAddress
		relayer, err := factory.NewEthMultiversXBridgeComponents(argsBridgeComponents)
		require.Nil(t, err)

		go func() {
			err = relayer.Start()
			integrationTests.Log.LogIfError(err)
			require.Nil(t, err)
		}()

		relayers = append(relayers, relayer)
	}

	return relayers
}

func executeContractsTxs(
	t *testing.T,
	multiversXProxyWithChainSimulator proxyWithChainSimulator,
	relayersKeys []keysHolder,
) (string, string) {
	ownerSK, ownerPK, err := core.LoadSkPkFromPemFile(ownerPem, 0)
	require.NoError(t, err)

	// fund the involved wallets(owner + relayers)
	multiversXProxyWithChainSimulator.FundWallets([]string{
		ownerPK,
		relayersKeys[0].pk,
		relayersKeys[1].pk,
		relayersKeys[2].pk,
	})

	// wait for epoch 1 before sc deploys
	time.Sleep(time.Duration(roundDurationInMs*(roundsPerEpoch+2)) * time.Millisecond)

	// deploy aggregator
	aggregatorAddress, err := multiversXProxyWithChainSimulator.DeploySC(
		context.Background(),
		aggregatorContract,
		ownerPK,
		ownerSK,
		[]string{wasm.VMTypeHex, "0500", "01", "00", getHexAddress(t, ownerPK)},
	)
	require.NoError(t, err)
	require.NotEqual(t, emptyAddress, aggregatorAddress)

	log.Info("aggregator contract deployed", "address", aggregatorAddress)

	// deploy wrapper
	wrapperAddress, err := multiversXProxyWithChainSimulator.DeploySC(
		context.Background(),
		wrapperContract,
		ownerPK,
		ownerSK,
		[]string{wasm.VMTypeHex, "0500"},
	)
	require.NoError(t, err)
	require.NotEqual(t, emptyAddress, wrapperAddress)

	log.Info("wrapper contract deployed", "address", wrapperAddress)

	// deploy safe
	safeAddress, err := multiversXProxyWithChainSimulator.DeploySC(
		context.Background(),
		safeContract,
		ownerPK,
		ownerSK,
		[]string{wasm.VMTypeHex, "0500", getHexAddress(t, aggregatorAddress), "01"},
	)
	require.NoError(t, err)
	require.NotEqual(t, emptyAddress, safeAddress)

	log.Info("safe contract deployed", "address", safeAddress)

	// deploy multi-transfer
	multiTransferAddress, err := multiversXProxyWithChainSimulator.DeploySC(
		context.Background(),
		multiTransferContract,
		ownerPK,
		ownerSK,
		[]string{wasm.VMTypeHex, "0500"},
	)
	require.NoError(t, err)
	require.NotEqual(t, emptyAddress, multiTransferAddress)

	log.Info("multi-transfer contract deployed", "address", multiTransferAddress)

	// deploy multisig
	multisigAddress, err := multiversXProxyWithChainSimulator.DeploySC(
		context.Background(),
		multisigContract,
		ownerPK,
		ownerSK,
		[]string{wasm.VMTypeHex, "0500", getHexAddress(t, safeAddress), getHexAddress(t, multiTransferAddress), minRelayerStakeHex, slashAmount, quorum, getHexAddress(t, relayersKeys[0].pk), getHexAddress(t, relayersKeys[1].pk), getHexAddress(t, relayersKeys[2].pk)},
	)
	require.NoError(t, err)
	require.NotEqual(t, emptyAddress, multisigAddress)

	log.Info("multisig contract deployed", "address", multisigAddress)

	// deploy bridge proxy
	bridgeProxyAddress, err := multiversXProxyWithChainSimulator.DeploySC(
		context.Background(),
		bridgeProxyContract,
		ownerPK,
		ownerSK,
		[]string{wasm.VMTypeHex, "0500", getHexAddress(t, multiTransferAddress)},
	)
	require.NoError(t, err)
	require.NotEqual(t, emptyAddress, bridgeProxyAddress)

	log.Info("bridge proxy contract deployed", "address", bridgeProxyAddress)

	// setBridgeProxyContractAddress
	hash, err := multiversXProxyWithChainSimulator.ScCall(
		context.Background(),
		ownerPK,
		ownerSK,
		multiTransferAddress,
		zeroValue,
		"setBridgeProxyContractAddress",
		[]string{getHexAddress(t, bridgeProxyAddress)},
	)
	require.NoError(t, err)

	log.Info("setBridgeProxyContractAddress tx sent", "hash", hash)

	// setWrappingContractAddress
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		context.Background(),
		ownerPK,
		ownerSK,
		multiTransferAddress,
		zeroValue,
		"setWrappingContractAddress",
		[]string{getHexAddress(t, wrapperAddress)},
	)
	require.NoError(t, err)

	log.Info("setWrappingContractAddress tx sent", "hash", hash)

	// ChangeOwnerAddress for safe
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		context.Background(),
		ownerPK,
		ownerSK,
		safeAddress,
		zeroValue,
		"ChangeOwnerAddress",
		[]string{getHexAddress(t, multisigAddress)},
	)
	require.NoError(t, err)

	log.Info("ChangeOwnerAddress for safe tx sent", "hash", hash)

	// ChangeOwnerAddress for multi-transfer
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		context.Background(),
		ownerPK,
		ownerSK,
		multiTransferAddress,
		zeroValue,
		"ChangeOwnerAddress",
		[]string{getHexAddress(t, multisigAddress)},
	)
	require.NoError(t, err)

	log.Info("ChangeOwnerAddress for multi-transfer tx sent", "hash", hash)

	// ChangeOwnerAddress for bridge proxy
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		context.Background(),
		ownerPK,
		ownerSK,
		bridgeProxyAddress,
		zeroValue,
		"ChangeOwnerAddress",
		[]string{getHexAddress(t, multisigAddress)},
	)
	require.NoError(t, err)

	log.Info("ChangeOwnerAddress for bridge proxy tx sent", "hash", hash)

	// setMultiTransferOnEsdtSafe
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		context.Background(),
		ownerPK,
		ownerSK,
		multisigAddress,
		zeroValue,
		"setMultiTransferOnEsdtSafe",
		[]string{},
	)
	require.NoError(t, err)

	log.Info("setMultiTransferOnEsdtSafe tx sent", "hash", hash)

	// setEsdtSafeOnMultiTransfer
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		context.Background(),
		ownerPK,
		ownerSK,
		multisigAddress,
		zeroValue,
		"setEsdtSafeOnMultiTransfer",
		[]string{},
	)
	require.NoError(t, err)

	log.Info("setEsdtSafeOnMultiTransfer tx sent", "hash", hash)

	// stake relayers
	stakeRelayers(t, multiversXProxyWithChainSimulator, multisigAddress, relayersKeys)

	// unpause multisig
	hash = unpauseContract(t, multiversXProxyWithChainSimulator, ownerPK, ownerSK, multisigAddress, []byte("unpause"))
	log.Info("unpaused multisig", "hash", hash)

	// unpause safe
	hash = unpauseContract(t, multiversXProxyWithChainSimulator, ownerPK, ownerSK, multisigAddress, []byte("unpauseEsdtSafe"))
	log.Info("unpaused safe", "hash", hash)

	// unpause aggregator
	hash = unpauseContract(t, multiversXProxyWithChainSimulator, ownerPK, ownerSK, aggregatorAddress, []byte("unpause"))
	log.Info("unpaused aggregator", "hash", hash)

	// unpause wrapper
	hash = unpauseContract(t, multiversXProxyWithChainSimulator, ownerPK, ownerSK, wrapperAddress, []byte("unpause"))
	log.Info("unpaused wrapper", "hash", hash)

	return safeAddress, multisigAddress
}

func stakeRelayers(t *testing.T, multiversXProxyWithChainSimulator proxyWithChainSimulator, contract string, relayersKeys []keysHolder) {
	for _, relayerKeys := range relayersKeys {
		hash, err := multiversXProxyWithChainSimulator.SendTx(context.Background(), relayerKeys.pk, relayerKeys.sk, contract, minRelayerStake, []byte("stake"))
		require.NoError(t, err)

		log.Info(fmt.Sprintf("relayer %s staked with hash %s", relayerKeys.pk, hash))
	}
}

func unpauseContract(t *testing.T, multiversXProxyWithChainSimulator proxyWithChainSimulator, ownerPK string, ownerSK []byte, contract string, dataField []byte) string {
	hash, err := multiversXProxyWithChainSimulator.SendTx(context.Background(), ownerPK, ownerSK, contract, zeroValue, dataField)
	require.NoError(t, err)

	return hash
}

func getHexAddress(t *testing.T, bech32Address string) string {
	address, err := data.NewAddressFromBech32String(bech32Address)
	require.NoError(t, err)

	return hex.EncodeToString(address.AddressBytes())
}
