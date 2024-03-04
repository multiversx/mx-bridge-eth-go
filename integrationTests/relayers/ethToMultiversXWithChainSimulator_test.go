package relayers

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
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
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/require"
)

const (
	ownerPem                                     = "testdata/wallets/owner.pem"
	safeContract                                 = "testdata/contracts/esdt-safe.wasm"
	multisigContract                             = "testdata/contracts/multisig.wasm"
	multiTransferContract                        = "testdata/contracts/multi-transfer-esdt.wasm"
	bridgeProxyContract                          = "testdata/contracts/bridge-proxy.wasm"
	aggregatorContract                           = "testdata/contracts/aggregator.wasm"
	wrapperContract                              = "testdata/contracts/bridged-tokens-wrapper.wasm"
	nodeConfig                                   = "testdata/config/nodeConfig"
	proxyConfig                                  = "testdata/config/proxyConfig"
	minRelayerStake                              = "10000000000000000000" // 10egld
	slashAmount                                  = "00"
	quorum                                       = "03"
	relayerPemPathFormat                         = "testdata/multiversx%d.pem"
	roundDurationInMs                            = 3000
	roundsPerEpoch                               = 20
	numOfShards                                  = 3
	serverPort                                   = 8085
	proxyCacherExpirationSeconds                 = 600
	proxyMaxNoncesDelta                          = 7
	zeroValue                                    = "0"
	unpause                                      = "unpause"
	unpauseEsdtSafe                              = "unpauseEsdtSafe"
	setEsdtSafeOnMultiTransfer                   = "setEsdtSafeOnMultiTransfer"
	setMultiTransferOnEsdtSafe                   = "setMultiTransferOnEsdtSafe"
	changeOwnerAddress                           = "ChangeOwnerAddress"
	setWrappingContractAddress                   = "setWrappingContractAddress"
	setBridgeProxyContractAddress                = "setBridgeProxyContractAddress"
	emptyAddress                                 = "erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu"
	esdtSystemSCAddress                          = "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"
	esdtIssueCost                                = "5000000000000000000" // 5egld
	universalTokenTicker                         = "USDC"
	universalTokenDisplayName                    = "WrappedUSDC"
	chainSpecificTokenTicker                     = "ETHUSDC"
	chainSpecificTokenDisplayName                = "EthereumWrappedUSDC"
	numOfDecimalsUniversal                       = "06"
	numOfDecimalsChainSpecific                   = "06"
	issue                                        = "issue"
	canAddSpecialRoles                           = "canAddSpecialRoles"
	trueStr                                      = "true"
	valueToMint                                  = "10000000000"
	setSpecialRole                               = "setSpecialRole"
	esdtRoleLocalMint                            = "ESDTRoleLocalMint"
	esdtRoleLocalBurn                            = "ESDTRoleLocalBurn"
	esdtTransfer                                 = "ESDTTransfer"
	depositLiquidity                             = "depositLiquidity"
	addWrappedToken                              = "addWrappedToken"
	whitelistToken                               = "whitelistToken"
	addMapping                                   = "addMapping"
	esdtSafeAddTokenToWhitelist                  = "esdtSafeAddTokenToWhitelist"
	submitBatch                                  = "submitBatch"
	esdtSafeSetMaxBridgedAmountForToken          = "esdtSafeSetMaxBridgedAmountForToken"
	multiTransferEsdtSetMaxBridgedAmountForToken = "multiTransferEsdtSetMaxBridgedAmountForToken"
	gwei                                         = "GWEI"
	fee                                          = "50000000"
	maxBridgedAmountForToken                     = "50000000000"
)

type proxyWithChainSimulator interface {
	Proxy() multiversx.Proxy
	GetNetworkAddress() string
	DeploySC(ctx context.Context, path string, ownerPK string, ownerSK []byte, extraParams []string) (string, error)
	ScCall(ctx context.Context, senderPK string, senderSK []byte, contract string, value string, function string, parameters []string) (string, error)
	SendTx(ctx context.Context, senderPK string, senderSK []byte, receiver string, value string, dataField []byte) (string, error)
	GetTransactionResult(hash string) (*transaction.ApiTransactionResult, error)
	FundWallets(wallets []string)
	Close()
}

type keysHolder struct {
	pk string
	sk []byte
}

func TestRelayersShouldExecuteTransfersFromEthToMultiversXWithChainSimulator(t *testing.T) {
	t.Skip("this is a long test")

	safeContractEthAddress := testsCommon.CreateRandomEthereumAddress()
	token1Erc20 := testsCommon.CreateRandomEthereumAddress()
	value1 := big.NewInt(1000)
	destination1 := testsCommon.CreateRandomMultiversXAddress()
	bech32Destination1, _ := destination1.AddressAsBech32String()
	log.Info("destination account", "address", bech32Destination1)

	depositor1 := testsCommon.CreateRandomEthereumAddress()

	tokens := []common.Address{token1Erc20}
	availableBalances := []*big.Int{value1}

	erc20ContractsHolder := createMockErc20ContractsHolder(tokens, safeContractEthAddress, availableBalances)

	batchNonceOnEthereum := uint64(1)
	txNonceOnEthereum := uint64(0)
	batch := contract.Batch{
		Nonce:                  big.NewInt(int64(batchNonceOnEthereum)),
		BlockNumber:            0,
		LastUpdatedBlockNumber: 0,
		DepositsCount:          1,
	}

	numRelayers := 3
	ethereumChainMock := mock.NewEthereumChainMock()
	token1NativeBalance := big.NewInt(0)
	ethereumChainMock.AddWhitelistedTokensMintBurn(token1Erc20, token1NativeBalance)
	ethereumChainMock.AddBatch(batch)
	ethereumChainMock.AddDepositToBatch(batchNonceOnEthereum, contract.Deposit{
		Nonce:        big.NewInt(int64(txNonceOnEthereum) + 1),
		TokenAddress: token1Erc20,
		Amount:       value1,
		Depositor:    depositor1,
		Recipient:    destination1.AddressSlice(),
		Status:       0,
	})
	ethereumChainMock.AddBatch(batch)
	ethereumChainMock.SetQuorum(numRelayers)

	relayersKeys := make([]keysHolder, 0, numRelayers)
	for i := 0; i < numRelayers; i++ {
		relayerSK, relayerPK, err := core.LoadSkPkFromPemFile(fmt.Sprintf(relayerPemPathFormat, i), 0)
		require.Nil(t, err)

		relayersKeys = append(relayersKeys, keysHolder{
			pk: relayerPK,
			sk: relayerSK,
		})
	}

	multiversXProxyWithChainSimulator := startProxyWithChainSimulator(t)
	defer multiversXProxyWithChainSimulator.Close()

	// create a test context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// read the owner keys
	ownerSK, ownerPK, err := core.LoadSkPkFromPemFile(ownerPem, 0)
	require.NoError(t, err)
	ownerKeys := keysHolder{
		pk: ownerPK,
		sk: ownerSK,
	}

	// deploy all contracts and execute all txs needed
	safeAddress, multisigAddress, wrapperAddress, aggregatorAddress := executeContractsTxs(t, ctx, multiversXProxyWithChainSimulator, relayersKeys, ownerKeys)

	// issue and whitelist token
	issueAndWhitelistToken(t, ctx, multiversXProxyWithChainSimulator, ownerKeys, wrapperAddress, safeAddress, multisigAddress, aggregatorAddress, hex.EncodeToString(token1Erc20.Bytes()))

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
		argsBridgeComponents.Configs.GeneralConfig.MultiversX.GasMap = config.MultiversXGasMapConfig{
			Sign:                   8000000,
			ProposeTransferBase:    11000000,
			ProposeTransferForEach: 5500000,
			ProposeStatusBase:      10000000,
			ProposeStatusForEach:   7000000,
			PerformActionBase:      40000000,
			PerformActionForEach:   5500000,
			ScCallPerByte:          100000,
			ScCallPerformForEach:   10000000,
		}
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
	ctx context.Context,
	multiversXProxyWithChainSimulator proxyWithChainSimulator,
	relayersKeys []keysHolder,
	ownerKeys keysHolder,
) (string, string, string, string) {
	// fund the involved wallets(owner + relayers)
	multiversXProxyWithChainSimulator.FundWallets([]string{
		ownerKeys.pk,
		relayersKeys[0].pk,
		relayersKeys[1].pk,
		relayersKeys[2].pk,
	})

	// wait for epoch 1 before sc deploys
	time.Sleep(time.Duration(roundDurationInMs*(roundsPerEpoch+2)) * time.Millisecond)

	// deploy aggregator
	aggregatorAddress, err := multiversXProxyWithChainSimulator.DeploySC(
		ctx,
		aggregatorContract,
		ownerKeys.pk,
		ownerKeys.sk,
		[]string{wasm.VMTypeHex, "0500", "01", "00", getHexAddress(t, ownerKeys.pk)},
	)
	require.NoError(t, err)
	require.NotEqual(t, emptyAddress, aggregatorAddress)

	log.Info("aggregator contract deployed", "address", aggregatorAddress)

	// deploy wrapper
	wrapperAddress, err := multiversXProxyWithChainSimulator.DeploySC(
		ctx,
		wrapperContract,
		ownerKeys.pk,
		ownerKeys.sk,
		[]string{wasm.VMTypeHex, "0500"},
	)
	require.NoError(t, err)
	require.NotEqual(t, emptyAddress, wrapperAddress)

	log.Info("wrapper contract deployed", "address", wrapperAddress)

	// deploy safe
	safeAddress, err := multiversXProxyWithChainSimulator.DeploySC(
		ctx,
		safeContract,
		ownerKeys.pk,
		ownerKeys.sk,
		[]string{wasm.VMTypeHex, "0500", getHexAddress(t, aggregatorAddress), "01"},
	)
	require.NoError(t, err)
	require.NotEqual(t, emptyAddress, safeAddress)

	log.Info("safe contract deployed", "address", safeAddress)

	// deploy multi-transfer
	multiTransferAddress, err := multiversXProxyWithChainSimulator.DeploySC(
		ctx,
		multiTransferContract,
		ownerKeys.pk,
		ownerKeys.sk,
		[]string{wasm.VMTypeHex, "0500"},
	)
	require.NoError(t, err)
	require.NotEqual(t, emptyAddress, multiTransferAddress)

	log.Info("multi-transfer contract deployed", "address", multiTransferAddress)

	// deploy multisig
	minRelayerStakeInt, _ := big.NewInt(0).SetString(minRelayerStake, 10)
	minRelayerStakeHex := hex.EncodeToString(minRelayerStakeInt.Bytes())
	multisigAddress, err := multiversXProxyWithChainSimulator.DeploySC(
		ctx,
		multisigContract,
		ownerKeys.pk,
		ownerKeys.sk,
		[]string{wasm.VMTypeHex, "0500", getHexAddress(t, safeAddress), getHexAddress(t, multiTransferAddress), minRelayerStakeHex, slashAmount, quorum, getHexAddress(t, relayersKeys[0].pk), getHexAddress(t, relayersKeys[1].pk), getHexAddress(t, relayersKeys[2].pk)},
	)
	require.NoError(t, err)
	require.NotEqual(t, emptyAddress, multisigAddress)

	log.Info("multisig contract deployed", "address", multisigAddress)

	// deploy bridge proxy
	bridgeProxyAddress, err := multiversXProxyWithChainSimulator.DeploySC(
		ctx,
		bridgeProxyContract,
		ownerKeys.pk,
		ownerKeys.sk,
		[]string{wasm.VMTypeHex, "0500", getHexAddress(t, multiTransferAddress)},
	)
	require.NoError(t, err)
	require.NotEqual(t, emptyAddress, bridgeProxyAddress)

	log.Info("bridge proxy contract deployed", "address", bridgeProxyAddress)

	// setBridgeProxyContractAddress
	hash, err := multiversXProxyWithChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		multiTransferAddress,
		zeroValue,
		setBridgeProxyContractAddress,
		[]string{getHexAddress(t, bridgeProxyAddress)},
	)
	require.NoError(t, err)
	txResult, err := multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)

	log.Info("setBridgeProxyContractAddress tx executed", "hash", hash, "status", txResult.Status.String())

	// setWrappingContractAddress
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		multiTransferAddress,
		zeroValue,
		setWrappingContractAddress,
		[]string{getHexAddress(t, wrapperAddress)},
	)
	require.NoError(t, err)
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)

	log.Info("setWrappingContractAddress tx executed", "hash", hash, "status", txResult.Status.String())

	// ChangeOwnerAddress for safe
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		safeAddress,
		zeroValue,
		changeOwnerAddress,
		[]string{getHexAddress(t, multisigAddress)},
	)
	require.NoError(t, err)
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)

	log.Info("ChangeOwnerAddress for safe tx executed", "hash", hash, "status", txResult.Status.String())

	// ChangeOwnerAddress for multi-transfer
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		multiTransferAddress,
		zeroValue,
		changeOwnerAddress,
		[]string{getHexAddress(t, multisigAddress)},
	)
	require.NoError(t, err)
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)

	log.Info("ChangeOwnerAddress for multi-transfer tx executed", "hash", hash, "status", txResult.Status.String())

	// ChangeOwnerAddress for bridge proxy
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		bridgeProxyAddress,
		zeroValue,
		changeOwnerAddress,
		[]string{getHexAddress(t, multisigAddress)},
	)
	require.NoError(t, err)
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)

	log.Info("ChangeOwnerAddress for bridge proxy tx executed", "hash", hash, "status", txResult.Status.String())

	// setMultiTransferOnEsdtSafe
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		multisigAddress,
		zeroValue,
		setMultiTransferOnEsdtSafe,
		[]string{},
	)
	require.NoError(t, err)
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)

	log.Info("setMultiTransferOnEsdtSafe tx executed", "hash", hash, "status", txResult.Status.String())

	// setEsdtSafeOnMultiTransfer
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		multisigAddress,
		zeroValue,
		setEsdtSafeOnMultiTransfer,
		[]string{},
	)
	require.NoError(t, err)
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)

	log.Info("setEsdtSafeOnMultiTransfer tx executed", "hash", hash, "status", txResult.Status.String())

	// stake relayers
	stakeRelayers(t, ctx, multiversXProxyWithChainSimulator, multisigAddress, relayersKeys)

	// unpause multisig
	hash = unpauseContract(t, ctx, multiversXProxyWithChainSimulator, ownerKeys, multisigAddress, []byte(unpause))
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)
	log.Info("unpaused multisig executed", "hash", hash, "status", txResult.Status.String())

	// unpause safe
	hash = unpauseContract(t, ctx, multiversXProxyWithChainSimulator, ownerKeys, multisigAddress, []byte(unpauseEsdtSafe))
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)
	log.Info("unpaused safe executed", "hash", hash, "status", txResult.Status.String())

	// unpause aggregator
	hash = unpauseContract(t, ctx, multiversXProxyWithChainSimulator, ownerKeys, aggregatorAddress, []byte(unpause))
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)
	log.Info("unpaused aggregator executed", "hash", hash, "status", txResult.Status.String())

	// unpause wrapper
	hash = unpauseContract(t, ctx, multiversXProxyWithChainSimulator, ownerKeys, wrapperAddress, []byte(unpause))
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)
	log.Info("unpaused wrapper executed", "hash", hash, "status", txResult.Status.String())

	return safeAddress, multisigAddress, wrapperAddress, aggregatorAddress
}

func stakeRelayers(t *testing.T, ctx context.Context, multiversXProxyWithChainSimulator proxyWithChainSimulator, contract string, relayersKeys []keysHolder) {
	for _, relayerKeys := range relayersKeys {
		hash, err := multiversXProxyWithChainSimulator.SendTx(ctx, relayerKeys.pk, relayerKeys.sk, contract, minRelayerStake, []byte("stake"))
		require.NoError(t, err)

		log.Info(fmt.Sprintf("relayer %s staked with hash %s", relayerKeys.pk, hash))
	}
}

func unpauseContract(t *testing.T, ctx context.Context, multiversXProxyWithChainSimulator proxyWithChainSimulator, ownerKeys keysHolder, contract string, dataField []byte) string {
	hash, err := multiversXProxyWithChainSimulator.SendTx(ctx, ownerKeys.pk, ownerKeys.sk, contract, zeroValue, dataField)
	require.NoError(t, err)

	return hash
}

func issueAndWhitelistToken(
	t *testing.T,
	ctx context.Context,
	multiversXProxyWithChainSimulator proxyWithChainSimulator,
	ownerKeys keysHolder,
	wrapperAddress string,
	safeAddress string,
	multisigAddress string,
	aggregatorAddress string,
	erc20Token string,
) {
	// issue universal token
	hash, err := multiversXProxyWithChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		esdtSystemSCAddress,
		esdtIssueCost,
		issue,
		[]string{hex.EncodeToString([]byte(universalTokenDisplayName)), hex.EncodeToString([]byte(universalTokenTicker)), "00", numOfDecimalsUniversal, hex.EncodeToString([]byte(canAddSpecialRoles)), hex.EncodeToString([]byte(trueStr))})
	require.NoError(t, err)
	txResult, err := multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)
	newUniversalToken := getTokenNameFromResult(t, txResult)

	log.Info("issue universal token tx executed", "hash", hash, "status", txResult.Status.String(), "token", newUniversalToken)

	// issue chain specific token
	valueToMintInt, _ := big.NewInt(0).SetString(valueToMint, 10)
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		esdtSystemSCAddress,
		esdtIssueCost,
		issue,
		[]string{hex.EncodeToString([]byte(chainSpecificTokenDisplayName)), hex.EncodeToString([]byte(chainSpecificTokenTicker)), hex.EncodeToString(valueToMintInt.Bytes()), numOfDecimalsChainSpecific, hex.EncodeToString([]byte(canAddSpecialRoles)), hex.EncodeToString([]byte(trueStr))})
	require.NoError(t, err)
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)
	newChainSpecificToken := getTokenNameFromResult(t, txResult)

	log.Info("issue chain specific token tx executed", "hash", hash, "status", txResult.Status.String(), "token", newChainSpecificToken)

	// set local roles bridged tokens wrapper
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		esdtSystemSCAddress,
		zeroValue,
		setSpecialRole,
		[]string{hex.EncodeToString([]byte(newUniversalToken)), getHexAddress(t, wrapperAddress), hex.EncodeToString([]byte(esdtRoleLocalMint)), hex.EncodeToString([]byte(esdtRoleLocalBurn))})
	require.NoError(t, err)
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)

	log.Info("set local roles bridged tokens wrapper tx executed", "hash", hash, "status", txResult.Status.String())

	// transfer to sc
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		wrapperAddress,
		zeroValue,
		esdtTransfer,
		[]string{hex.EncodeToString([]byte(newChainSpecificToken)), hex.EncodeToString(valueToMintInt.Bytes()), hex.EncodeToString([]byte(depositLiquidity))})
	require.NoError(t, err)
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)

	log.Info("transfer to sc tx executed", "hash", hash, "status", txResult.Status.String())

	// add wrapped token
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		wrapperAddress,
		zeroValue,
		addWrappedToken,
		[]string{hex.EncodeToString([]byte(newUniversalToken)), numOfDecimalsUniversal})
	require.NoError(t, err)
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)

	log.Info("add wrapped token tx executed", "hash", hash, "status", txResult.Status.String())

	// wrapper whitelist token
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		wrapperAddress,
		zeroValue,
		whitelistToken,
		[]string{hex.EncodeToString([]byte(newChainSpecificToken)), numOfDecimalsChainSpecific, hex.EncodeToString([]byte(newUniversalToken))})
	require.NoError(t, err)
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)

	log.Info("wrapper whitelist token tx executed", "hash", hash, "status", txResult.Status.String())

	// set local roles esdt safe
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		esdtSystemSCAddress,
		zeroValue,
		setSpecialRole,
		[]string{hex.EncodeToString([]byte(newChainSpecificToken)), getHexAddress(t, safeAddress), hex.EncodeToString([]byte(esdtRoleLocalMint)), hex.EncodeToString([]byte(esdtRoleLocalBurn))})
	require.NoError(t, err)
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)

	log.Info("set local roles esdt safe tx executed", "hash", hash, "status", txResult.Status.String())

	// add mapping
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		multisigAddress,
		zeroValue,
		addMapping,
		[]string{erc20Token, hex.EncodeToString([]byte(newChainSpecificToken))})
	require.NoError(t, err)
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)

	log.Info("add mapping tx executed", "hash", hash, "status", txResult.Status.String())

	// whitelist token
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		multisigAddress,
		zeroValue,
		esdtSafeAddTokenToWhitelist,
		[]string{hex.EncodeToString([]byte(newChainSpecificToken)), hex.EncodeToString([]byte(chainSpecificTokenTicker)), "01"})
	require.NoError(t, err)
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)

	log.Info("whitelist token tx executed", "hash", hash, "status", txResult.Status.String())

	// submit aggregator batch
	feeInt, _ := big.NewInt(0).SetString(fee, 10)
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		aggregatorAddress,
		zeroValue,
		submitBatch,
		[]string{hex.EncodeToString([]byte(gwei)), hex.EncodeToString([]byte(chainSpecificTokenTicker)), hex.EncodeToString(feeInt.Bytes())})
	require.NoError(t, err)
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)

	log.Info("submit aggregator batch tx executed", "hash", hash, "status", txResult.Status.String())

	// safe set max bridge amount for token
	maxBridgedAmountForTokenInt, _ := big.NewInt(0).SetString(maxBridgedAmountForToken, 10)
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		multisigAddress,
		zeroValue,
		esdtSafeSetMaxBridgedAmountForToken,
		[]string{hex.EncodeToString([]byte(newChainSpecificToken)), hex.EncodeToString(maxBridgedAmountForTokenInt.Bytes())})
	require.NoError(t, err)
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)

	log.Info("safe set max bridge amount for token tx executed", "hash", hash, "status", txResult.Status.String())

	// multi-transfer set max bridge amount for token
	hash, err = multiversXProxyWithChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		multisigAddress,
		zeroValue,
		multiTransferEsdtSetMaxBridgedAmountForToken,
		[]string{hex.EncodeToString([]byte(newChainSpecificToken)), hex.EncodeToString(maxBridgedAmountForTokenInt.Bytes())})
	require.NoError(t, err)
	txResult, err = multiversXProxyWithChainSimulator.GetTransactionResult(hash)
	require.NoError(t, err)

	log.Info("multi-transfer set max bridge amount for token tx executed", "hash", hash, "status", txResult.Status.String())
}

func getTokenNameFromResult(t *testing.T, txResult *transaction.ApiTransactionResult) string {
	resultData := txResult.SmartContractResults[0].Data
	splittedData := strings.Split(resultData, "@")
	if len(splittedData) != 2 {
		require.Fail(t, fmt.Sprintf("received invalid data received while issuing: %s", resultData))
	}

	newUniversalTokenBytes, err := hex.DecodeString(splittedData[1])
	require.NoError(t, err)

	return string(newUniversalTokenBytes)
}

func getHexAddress(t *testing.T, bech32Address string) string {
	address, err := data.NewAddressFromBech32String(bech32Address)
	require.NoError(t, err)

	return hex.EncodeToString(address.AddressBytes())
}
