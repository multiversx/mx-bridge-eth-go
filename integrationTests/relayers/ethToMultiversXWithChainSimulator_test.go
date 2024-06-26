//go:build slow

// To run these slow tests, simply add the slow tag on the go test command. Also, provide a chain simulator instance on the 8085 port
// example: go test -tags slow

package relayers

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/pem"
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
	"github.com/multiversx/mx-bridge-eth-go/clients"
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
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/require"
)

var addressPubkeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, "erd")

const (
	safeContract                                 = "testdata/contracts/esdt-safe.wasm"
	multisigContract                             = "testdata/contracts/multisig.wasm"
	multiTransferContract                        = "testdata/contracts/multi-transfer-esdt.wasm"
	bridgeProxyContract                          = "testdata/contracts/bridge-proxy.wasm"
	aggregatorContract                           = "testdata/contracts/aggregator.wasm"
	wrapperContract                              = "testdata/contracts/bridged-tokens-wrapper.wasm"
	minRelayerStake                              = "10000000000000000000" // 10egld
	slashAmount                                  = "00"
	quorum                                       = "03"
	relayerPemPathFormat                         = "multiversx%d.pem"
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
	createTransactionParam                       = "createTransaction"
	unwrapToken                                  = "unwrapToken"
	setPairDecimals                              = "setPairDecimals"
	initSupplyFromChildContract                  = "initSupplyFromChildContract"
)

type chainSimulatorWrapper interface {
	Proxy() multiversx.Proxy
	GetNetworkAddress() string
	DeploySC(ctx context.Context, path string, ownerPK string, ownerSK []byte, extraParams []string) (string, error)
	ScCall(ctx context.Context, senderPK string, senderSK []byte, contract string, value string, function string, parameters []string) (string, error)
	SendTx(ctx context.Context, senderPK string, senderSK []byte, receiver string, value string, dataField []byte) (string, error)
	GetTransactionResult(ctx context.Context, hash string) (*data.TransactionOnNetwork, error)
	FundWallets(ctx context.Context, wallets []string)
	GenerateBlocksUntilEpochReached(ctx context.Context, epoch uint32)
	GenerateBlocks(ctx context.Context, numBlocks int)
	GetESDTBalance(ctx context.Context, address sdkCore.AddressHandler, token string) (string, error)
}

type keysHolder struct {
	pk string
	sk []byte
}

type testConfig struct {
	isMintBurnOnMvX bool
	isNativeOnMvx   bool
	isMintBurnOnEth bool
	isNativeOnEth   bool
}

func TestTransfersBothWaysWithChainSimulator(t *testing.T) {
	t.Run("Eth: native, MvX: mint & burn", func(t *testing.T) {
		cfg := testConfig{
			isMintBurnOnMvX: true,
			isNativeOnMvx:   false,
			isMintBurnOnEth: false,
			isNativeOnEth:   true,
		}

		testTransfersBothWaysWithChainSimulatorAndConfig(t, cfg)
	})
	t.Run("Eth: mint & burn, MvX: native", func(t *testing.T) {
		cfg := testConfig{
			isMintBurnOnMvX: false,
			isNativeOnMvx:   true,
			isMintBurnOnEth: true,
			isNativeOnEth:   false,
		}

		testTransfersBothWaysWithChainSimulatorAndConfig(t, cfg)
	})
}

func testTransfersBothWaysWithChainSimulatorAndConfig(t *testing.T, cfg testConfig) {
	safeContractEthAddress := testsCommon.CreateRandomEthereumAddress()
	token1Erc20 := testsCommon.CreateRandomEthereumAddress()
	value1 := big.NewInt(200000000)

	receiverKeys := generateMvxPrivatePublicKey(t)
	log.Info("receiver is", "address", receiverKeys.pk)

	receiverAddress, err := data.NewAddressFromBech32String(receiverKeys.pk)
	require.NoError(t, err)

	depositor := testsCommon.CreateRandomEthereumAddress()

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
	// prepare ethereum chain mock for ETH->MVX
	ethereumChainMock.UpdateMintBurnTokens(token1Erc20, cfg.isMintBurnOnEth)
	ethereumChainMock.UpdateNativeTokens(token1Erc20, cfg.isNativeOnEth)
	ethereumChainMock.AddBatch(batch)
	ethereumChainMock.AddDepositToBatch(batchNonceOnEthereum, contract.Deposit{
		Nonce:        big.NewInt(int64(txNonceOnEthereum) + 1),
		TokenAddress: token1Erc20,
		Amount:       value1,
		Depositor:    depositor,
		Recipient:    receiverAddress.AddressSlice(),
		Status:       0,
	})
	ethereumChainMock.AddBatch(batch)
	ethereumChainMock.SetQuorum(numRelayers)
	if cfg.isNativeOnEth {
		ethereumChainMock.UpdateTotalBalances(token1Erc20, value1)
	} else {
		mvxSafeContractBalance, _ := big.NewInt(0).SetString(valueToMint, 10)
		mvxSafeContractBalance.Div(mvxSafeContractBalance, big.NewInt(2))
		ethereumChainMock.UpdateMintBalances(token1Erc20, big.NewInt(0).Sub(mvxSafeContractBalance, value1))
	}

	// prepare ethereum chain mock for MVX->ETH
	expectedStatuses := []byte{clients.Executed}
	ethereumChainMock.GetStatusesAfterExecutionHandler = func() []byte {
		return expectedStatuses
	}

	tempDir := t.TempDir()
	relayersKeys := make([]keysHolder, 0, numRelayers)
	for i := 0; i < numRelayers; i++ {
		relayerKeys := generateMvxPrivatePublicKey(t)
		log.Info("generated relayer", "index", i, "address", relayerKeys.pk)
		relayersKeys = append(relayersKeys, relayerKeys)
		saveRelayerKey(t, tempDir, i, relayerKeys)
	}

	mvxChainSimulatorWrapper := startChainSimulatorWrapper(t)

	// create a test context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ownerKeys := generateMvxPrivatePublicKey(t)
	log.Info("owner is", "address", receiverKeys.pk)

	// we need to wait until epoch 1 is reached so SC deployment will work
	mvxChainSimulatorWrapper.GenerateBlocksUntilEpochReached(ctx, 1)

	// deploy all contracts and execute all txs needed
	safeAddress, multisigAddress, wrapperAddress, aggregatorAddress := executeContractsTxs(t, ctx, mvxChainSimulatorWrapper, relayersKeys, ownerKeys, receiverKeys)

	// issue and whitelist token
	newUniversalToken, newChainSpecificToken := issueAndWhitelistToken(
		t,
		ctx,
		mvxChainSimulatorWrapper,
		ownerKeys,
		wrapperAddress,
		safeAddress,
		multisigAddress,
		aggregatorAddress,
		hex.EncodeToString(token1Erc20.Bytes()),
		cfg.isNativeOnMvx,
		cfg.isMintBurnOnMvX,
	)

	// start relayers
	relayers := startRelayers(t, tempDir, numRelayers, mvxChainSimulatorWrapper, ethereumChainMock, safeContractEthAddress, erc20ContractsHolder, safeAddress, multisigAddress)
	defer closeRelayers(relayers)

	checkESDTBalance(t, ctx, mvxChainSimulatorWrapper, receiverAddress, newUniversalToken, "0", true)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	ethToMVXDone := false
	mvxToETHDone := false
	for {
		select {
		case <-interrupt:
			require.Fail(t, "signal interrupted")
			return
		case <-time.After(time.Minute * 15):
			require.Fail(t, "time out")
			return
		default:
			isTransferDoneFromETH := checkESDTBalance(t, ctx, mvxChainSimulatorWrapper, receiverAddress, newUniversalToken, value1.String(), false)
			if !ethToMVXDone && isTransferDoneFromETH {
				ethToMVXDone = true
				log.Info("ETH->MVX transfer finished, now sending back to ETH...")

				// send half of the amount back to ETH
				value := big.NewInt(0).Div(value1, big.NewInt(2))
				sendMVXToEthTransaction(t, ctx, mvxChainSimulatorWrapper, value.Bytes(), newUniversalToken, newChainSpecificToken, receiverKeys, safeAddress, wrapperAddress, depositor.Bytes())
			}

			isTransferDoneFromMVX := checkETHStatus(ethereumChainMock, depositor, token1Erc20)
			if !mvxToETHDone && isTransferDoneFromMVX {
				mvxToETHDone = true
			}

			if ethToMVXDone && mvxToETHDone {
				log.Info("MVX<->ETH transfers done")
				return
			}

			mvxChainSimulatorWrapper.GenerateBlocks(ctx, 1)
		}
	}
}

func generateMvxPrivatePublicKey(t *testing.T) keysHolder {
	keyGenerator := signing.NewKeyGenerator(ed25519.NewEd25519())
	sk, pk := keyGenerator.GeneratePair()

	skBytes, err := sk.ToByteArray()
	require.Nil(t, err)

	pkBytes, err := pk.ToByteArray()
	require.Nil(t, err)

	address, err := addressPubkeyConverter.Encode(pkBytes)
	require.Nil(t, err)

	return keysHolder{
		pk: address,
		sk: skBytes,
	}
}

func saveRelayerKey(t *testing.T, tempDir string, index int, key keysHolder) {
	blk := pem.Block{
		Type:  "PRIVATE KEY for " + key.pk,
		Bytes: []byte(hex.EncodeToString(key.sk)),
	}

	buff := bytes.NewBuffer(make([]byte, 0))
	err := pem.Encode(buff, &blk)
	require.Nil(t, err)

	err = os.WriteFile(path.Join(tempDir, fmt.Sprintf(relayerPemPathFormat, index)), buff.Bytes(), os.ModePerm)
	require.Nil(t, err)
}

func startChainSimulatorWrapper(t *testing.T) chainSimulatorWrapper {
	// create a new working directory
	tmpDir := path.Join(t.TempDir(), "test")
	err := os.MkdirAll(tmpDir, os.ModePerm)
	require.NoError(t, err)

	// start the chain simulator
	args := integrationTests.ArgChainSimulatorWrapper{
		ProxyCacherExpirationSeconds: proxyCacherExpirationSeconds,
		ProxyMaxNoncesDelta:          proxyMaxNoncesDelta,
	}
	mvxChainSimulatorWrapper, err := integrationTests.CreateChainSimulatorWrapper(args)
	require.NoError(t, err)

	return mvxChainSimulatorWrapper
}

func startRelayers(
	t *testing.T,
	workingDir string,
	numRelayers int,
	mvxChainSimulator chainSimulatorWrapper,
	ethereumChainMock *mock.EthereumChainMock,
	safeContractEthAddress common.Address,
	erc20ContractsHolder *bridge.ERC20ContractsHolderStub,
	safeAddress string,
	multisigAddress string,
) []bridgeComponents {
	relayers := make([]bridgeComponents, 0, numRelayers)

	messengers := integrationTests.CreateLinkedMessengers(numRelayers)

	for i := 0; i < numRelayers; i++ {
		generalConfigs := createBridgeComponentsConfig(i, workingDir)
		argsBridgeComponents := factory.ArgsEthereumToMultiversXBridge{
			Configs: config.Configs{
				GeneralConfig:   generalConfigs,
				ApiRoutesConfig: config.ApiRoutesConfig{},
				FlagsConfig: config.ContextFlagsConfig{
					RestApiInterface: bridgeCore.WebServerOffString,
				},
			},
			Proxy:                         mvxChainSimulator.Proxy(),
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
		argsBridgeComponents.Configs.GeneralConfig.MultiversX.NetworkAddress = mvxChainSimulator.GetNetworkAddress()
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

		ethereumChainMock.AddRelayer(relayer.EthereumRelayerAddress())
	}

	return relayers
}

func executeContractsTxs(
	t *testing.T,
	ctx context.Context,
	mvxChainSimulator chainSimulatorWrapper,
	relayersKeys []keysHolder,
	ownerKeys keysHolder,
	receiver keysHolder,
) (string, string, string, string) {
	// fund the involved wallets(owner + relayers)
	walletsToFund := make([]string, 0, len(relayersKeys)+2)
	for _, relayerKeys := range relayersKeys {
		walletsToFund = append(walletsToFund, relayerKeys.pk)
	}
	walletsToFund = append(walletsToFund, ownerKeys.pk)
	walletsToFund = append(walletsToFund, receiver.pk)
	mvxChainSimulator.FundWallets(ctx, walletsToFund)

	mvxChainSimulator.GenerateBlocks(ctx, 1)

	// deploy aggregator
	stakeValue, _ := big.NewInt(0).SetString(minRelayerStake, 10)
	aggregatorDeployParams := []string{
		hex.EncodeToString([]byte("EGLD")),
		hex.EncodeToString(stakeValue.Bytes()),
		"01",
		"01",
		"01",
		getHexAddress(t, ownerKeys.pk),
	}

	aggregatorAddress, err := mvxChainSimulator.DeploySC(
		ctx,
		aggregatorContract,
		ownerKeys.pk,
		ownerKeys.sk,
		aggregatorDeployParams,
	)
	require.NoError(t, err)
	require.NotEqual(t, emptyAddress, aggregatorAddress)

	log.Info("aggregator contract deployed", "address", aggregatorAddress)

	// deploy wrapper
	wrapperAddress, err := mvxChainSimulator.DeploySC(
		ctx,
		wrapperContract,
		ownerKeys.pk,
		ownerKeys.sk,
		[]string{},
	)
	require.NoError(t, err)
	require.NotEqual(t, emptyAddress, wrapperAddress)

	log.Info("wrapper contract deployed", "address", wrapperAddress)

	// deploy safe
	safeAddress, err := mvxChainSimulator.DeploySC(
		ctx,
		safeContract,
		ownerKeys.pk,
		ownerKeys.sk,
		[]string{getHexAddress(t, aggregatorAddress), "01"},
	)
	require.NoError(t, err)
	require.NotEqual(t, emptyAddress, safeAddress)

	log.Info("safe contract deployed", "address", safeAddress)

	// deploy multi-transfer
	multiTransferAddress, err := mvxChainSimulator.DeploySC(
		ctx,
		multiTransferContract,
		ownerKeys.pk,
		ownerKeys.sk,
		[]string{},
	)
	require.NoError(t, err)
	require.NotEqual(t, emptyAddress, multiTransferAddress)

	log.Info("multi-transfer contract deployed", "address", multiTransferAddress)

	// deploy multisig
	minRelayerStakeInt, _ := big.NewInt(0).SetString(minRelayerStake, 10)
	minRelayerStakeHex := hex.EncodeToString(minRelayerStakeInt.Bytes())
	params := []string{getHexAddress(t, safeAddress), getHexAddress(t, multiTransferAddress), minRelayerStakeHex, slashAmount, quorum}
	for _, relayerKeys := range relayersKeys {
		params = append(params, getHexAddress(t, relayerKeys.pk))
	}
	multisigAddress, err := mvxChainSimulator.DeploySC(
		ctx,
		multisigContract,
		ownerKeys.pk,
		ownerKeys.sk,
		params,
	)
	require.NoError(t, err)
	require.NotEqual(t, emptyAddress, multisigAddress)

	log.Info("multisig contract deployed", "address", multisigAddress)

	// deploy bridge proxy
	bridgeProxyAddress, err := mvxChainSimulator.DeploySC(
		ctx,
		bridgeProxyContract,
		ownerKeys.pk,
		ownerKeys.sk,
		[]string{getHexAddress(t, multiTransferAddress)},
	)
	require.NoError(t, err)
	require.NotEqual(t, emptyAddress, bridgeProxyAddress)

	log.Info("bridge proxy contract deployed", "address", bridgeProxyAddress)

	// setBridgeProxyContractAddress
	hash, err := mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		multiTransferAddress,
		zeroValue,
		setBridgeProxyContractAddress,
		[]string{getHexAddress(t, bridgeProxyAddress)},
	)
	require.NoError(t, err)
	txResult, err := mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("setBridgeProxyContractAddress tx executed", "hash", hash, "status", txResult.Status)

	// setWrappingContractAddress
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		multiTransferAddress,
		zeroValue,
		setWrappingContractAddress,
		[]string{getHexAddress(t, wrapperAddress)},
	)
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("setWrappingContractAddress tx executed", "hash", hash, "status", txResult.Status)

	// ChangeOwnerAddress for safe
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		safeAddress,
		zeroValue,
		changeOwnerAddress,
		[]string{getHexAddress(t, multisigAddress)},
	)
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("ChangeOwnerAddress for safe tx executed", "hash", hash, "status", txResult.Status)

	// ChangeOwnerAddress for multi-transfer
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		multiTransferAddress,
		zeroValue,
		changeOwnerAddress,
		[]string{getHexAddress(t, multisigAddress)},
	)
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("ChangeOwnerAddress for multi-transfer tx executed", "hash", hash, "status", txResult.Status)

	// ChangeOwnerAddress for bridge proxy
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		bridgeProxyAddress,
		zeroValue,
		changeOwnerAddress,
		[]string{getHexAddress(t, multisigAddress)},
	)
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("ChangeOwnerAddress for bridge proxy tx executed", "hash", hash, "status", txResult.Status)

	// setMultiTransferOnEsdtSafe
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		multisigAddress,
		zeroValue,
		setMultiTransferOnEsdtSafe,
		[]string{},
	)
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("setMultiTransferOnEsdtSafe tx executed", "hash", hash, "status", txResult.Status)

	// setEsdtSafeOnMultiTransfer
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		multisigAddress,
		zeroValue,
		setEsdtSafeOnMultiTransfer,
		[]string{},
	)
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("setEsdtSafeOnMultiTransfer tx executed", "hash", hash, "status", txResult.Status)

	// setPairDecimals on aggregator
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		aggregatorAddress,
		zeroValue,
		setPairDecimals,
		[]string{hex.EncodeToString([]byte(gwei)), hex.EncodeToString([]byte(chainSpecificTokenTicker)), numOfDecimalsChainSpecific})
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("setPairDecimals tx executed", "hash", hash, "status", txResult.Status)

	// stake relayers on multisig
	stakeAddressesOnContract(t, ctx, mvxChainSimulator, multisigAddress, relayersKeys)

	// stake relayers on price aggregator
	stakeAddressesOnContract(t, ctx, mvxChainSimulator, aggregatorAddress, []keysHolder{ownerKeys})

	// unpause multisig
	hash = unpauseContract(t, ctx, mvxChainSimulator, ownerKeys, multisigAddress, []byte(unpause))
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)
	log.Info("unpaused multisig executed", "hash", hash, "status", txResult.Status)

	// unpause safe
	hash = unpauseContract(t, ctx, mvxChainSimulator, ownerKeys, multisigAddress, []byte(unpauseEsdtSafe))
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)
	log.Info("unpaused safe executed", "hash", hash, "status", txResult.Status)

	// unpause aggregator
	hash = unpauseContract(t, ctx, mvxChainSimulator, ownerKeys, aggregatorAddress, []byte(unpause))
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)
	log.Info("unpaused aggregator executed", "hash", hash, "status", txResult.Status)

	// unpause wrapper
	hash = unpauseContract(t, ctx, mvxChainSimulator, ownerKeys, wrapperAddress, []byte(unpause))
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)
	log.Info("unpaused wrapper executed", "hash", hash, "status", txResult.Status)

	return safeAddress, multisigAddress, wrapperAddress, aggregatorAddress
}

func stakeAddressesOnContract(t *testing.T, ctx context.Context, mvxChainSimulator chainSimulatorWrapper, contract string, allKeys []keysHolder) {
	for _, keys := range allKeys {
		hash, err := mvxChainSimulator.SendTx(ctx, keys.pk, keys.sk, contract, minRelayerStake, []byte("stake"))
		require.NoError(t, err)
		txResult, err := mvxChainSimulator.GetTransactionResult(ctx, hash)
		require.NoError(t, err)

		log.Info(fmt.Sprintf("address %s staked on contract %s with hash %s, status %s", keys.pk, contract, hash, txResult.Status))
	}
}

func unpauseContract(t *testing.T, ctx context.Context, mvxChainSimulator chainSimulatorWrapper, ownerKeys keysHolder, contract string, dataField []byte) string {
	hash, err := mvxChainSimulator.SendTx(ctx, ownerKeys.pk, ownerKeys.sk, contract, zeroValue, dataField)
	require.NoError(t, err)

	return hash
}

func issueAndWhitelistToken(
	t *testing.T,
	ctx context.Context,
	mvxChainSimulator chainSimulatorWrapper,
	ownerKeys keysHolder,
	wrapperAddress string,
	safeAddress string,
	multisigAddress string,
	aggregatorAddress string,
	erc20Token string,
	isNativeOnMvX bool,
	isMintBurnOnMvX bool,
) (string, string) {
	// issue universal token
	hash, err := mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		esdtSystemSCAddress,
		esdtIssueCost,
		issue,
		[]string{hex.EncodeToString([]byte(universalTokenDisplayName)), hex.EncodeToString([]byte(universalTokenTicker)), "00", numOfDecimalsUniversal, hex.EncodeToString([]byte(canAddSpecialRoles)), hex.EncodeToString([]byte(trueStr))})
	require.NoError(t, err)
	txResult, err := mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)
	newUniversalToken := getTokenNameFromResult(t, *txResult)

	log.Info("issue universal token tx executed", "hash", hash, "status", txResult.Status, "token", newUniversalToken, "owner", ownerKeys.pk)

	// issue chain specific token
	valueToMintInt, _ := big.NewInt(0).SetString(valueToMint, 10)
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		esdtSystemSCAddress,
		esdtIssueCost,
		issue,
		[]string{hex.EncodeToString([]byte(chainSpecificTokenDisplayName)), hex.EncodeToString([]byte(chainSpecificTokenTicker)), hex.EncodeToString(valueToMintInt.Bytes()), numOfDecimalsChainSpecific, hex.EncodeToString([]byte(canAddSpecialRoles)), hex.EncodeToString([]byte(trueStr))})
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)
	newChainSpecificToken := getTokenNameFromResult(t, *txResult)

	log.Info("issue chain specific token tx executed", "hash", hash, "status", txResult.Status, "token", newChainSpecificToken, "owner", ownerKeys.pk)

	// set local roles bridged tokens wrapper
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		esdtSystemSCAddress,
		zeroValue,
		setSpecialRole,
		[]string{hex.EncodeToString([]byte(newUniversalToken)), getHexAddress(t, wrapperAddress), hex.EncodeToString([]byte(esdtRoleLocalMint)), hex.EncodeToString([]byte(esdtRoleLocalBurn))})
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("set local roles bridged tokens wrapper tx executed", "hash", hash, "status", txResult.Status)

	// transfer to wrapper SC half the liquidity
	valueToTransfer := big.NewInt(0).Div(valueToMintInt, big.NewInt(2))
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		wrapperAddress,
		zeroValue,
		esdtTransfer,
		[]string{hex.EncodeToString([]byte(newChainSpecificToken)), hex.EncodeToString(valueToTransfer.Bytes()), hex.EncodeToString([]byte(depositLiquidity))})
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("transfer to wrapper sc tx executed", "hash", hash, "status", txResult.Status, "ESDT value", valueToTransfer.String())

	// add wrapped token
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		wrapperAddress,
		zeroValue,
		addWrappedToken,
		[]string{hex.EncodeToString([]byte(newUniversalToken)), numOfDecimalsUniversal})
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("add wrapped token tx executed", "hash", hash, "status", txResult.Status)

	// wrapper whitelist token
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		wrapperAddress,
		zeroValue,
		whitelistToken,
		[]string{hex.EncodeToString([]byte(newChainSpecificToken)), numOfDecimalsChainSpecific, hex.EncodeToString([]byte(newUniversalToken))})
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("wrapper whitelist token tx executed", "hash", hash, "status", txResult.Status)

	// set local roles esdt safe
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		esdtSystemSCAddress,
		zeroValue,
		setSpecialRole,
		[]string{hex.EncodeToString([]byte(newChainSpecificToken)), getHexAddress(t, safeAddress), hex.EncodeToString([]byte(esdtRoleLocalMint)), hex.EncodeToString([]byte(esdtRoleLocalBurn))})
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("set local roles esdt safe tx executed", "hash", hash, "status", txResult.Status)

	// add mapping
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		multisigAddress,
		zeroValue,
		addMapping,
		[]string{erc20Token, hex.EncodeToString([]byte(newChainSpecificToken))})
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("add mapping tx executed", "hash", hash, "status", txResult.Status)

	// whitelist token
	isNative := "00"
	if isNativeOnMvX {
		isNative = "01"
	}
	isMintBurn := "00"
	if isMintBurnOnMvX {
		isMintBurn = "01"
	}
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		multisigAddress,
		zeroValue,
		esdtSafeAddTokenToWhitelist,
		[]string{hex.EncodeToString([]byte(newChainSpecificToken)), hex.EncodeToString([]byte(chainSpecificTokenTicker)), isMintBurn, isNative})
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("whitelist token tx executed", "hash", hash, "status", txResult.Status)

	// transfer to safe SC the other half of the liquidity
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		multisigAddress,
		zeroValue,
		esdtTransfer,
		[]string{
			hex.EncodeToString([]byte(newChainSpecificToken)),
			hex.EncodeToString(valueToTransfer.Bytes()),
			hex.EncodeToString([]byte(initSupplyFromChildContract)), // function initSupplyFromChildContract
			hex.EncodeToString([]byte(newChainSpecificToken)),       // provide the token, as required by the function
			hex.EncodeToString(valueToTransfer.Bytes()),             // provide also the amount, as required by the function
		})
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("transfer to multisig sc tx executed", "hash", hash, "status", txResult.Status, "ESDT value", valueToTransfer.String())

	// submit aggregator batch
	submitAggregatorBatch(t, ctx, mvxChainSimulator, aggregatorAddress, ownerKeys)

	// safe set max bridge amount for token
	maxBridgedAmountForTokenInt, _ := big.NewInt(0).SetString(maxBridgedAmountForToken, 10)
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		multisigAddress,
		zeroValue,
		esdtSafeSetMaxBridgedAmountForToken,
		[]string{hex.EncodeToString([]byte(newChainSpecificToken)), hex.EncodeToString(maxBridgedAmountForTokenInt.Bytes())})
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("safe set max bridge amount for token tx executed", "hash", hash, "status", txResult.Status)

	// multi-transfer set max bridge amount for token
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		multisigAddress,
		zeroValue,
		multiTransferEsdtSetMaxBridgedAmountForToken,
		[]string{hex.EncodeToString([]byte(newChainSpecificToken)), hex.EncodeToString(maxBridgedAmountForTokenInt.Bytes())})
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("multi-transfer set max bridge amount for token tx executed", "hash", hash, "status", txResult.Status)

	return newUniversalToken, newChainSpecificToken
}

func getTokenNameFromResult(t *testing.T, txResult data.TransactionOnNetwork) string {
	resultData := txResult.ScResults[0].Data
	splittedData := strings.Split(resultData, "@")
	if len(splittedData) < 2 {
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

func checkESDTBalance(
	t *testing.T,
	ctx context.Context,
	mvxChainSimulator chainSimulatorWrapper,
	address sdkCore.AddressHandler,
	token string,
	expectedBalance string,
	checkResult bool,
) bool {
	balance, err := mvxChainSimulator.GetESDTBalance(ctx, address, token)
	require.NoError(t, err)

	if checkResult {
		require.Equal(t, expectedBalance, balance)
	}

	return expectedBalance == balance
}

func checkETHStatus(ethereumChainMock *mock.EthereumChainMock, receiver common.Address, token common.Address) bool {
	transfer := ethereumChainMock.GetLastProposedTransfer()
	if transfer == nil {
		return false
	}

	return transfer.Recipients[0] == receiver && transfer.Tokens[0] == token
}

func sendMVXToEthTransaction(
	t *testing.T,
	ctx context.Context,
	mvxChainSimulator chainSimulatorWrapper,
	value []byte,
	universalToken string,
	chainSpecificToken string,
	senderKeys keysHolder,
	safeAddress string,
	wrapperAddress string,
	receiver []byte,
) string {
	// unwrap token
	paramsUnwrap := []string{
		hex.EncodeToString([]byte(universalToken)),
		hex.EncodeToString(value),
		hex.EncodeToString([]byte(unwrapToken)),
		hex.EncodeToString([]byte(chainSpecificToken)),
	}

	hash, err := mvxChainSimulator.ScCall(ctx, senderKeys.pk, senderKeys.sk, wrapperAddress, zeroValue, esdtTransfer, paramsUnwrap)
	require.NoError(t, err)
	txResult, err := mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("unwrap transaction sent", "hash", hash, "token", universalToken, "status", txResult.Status)

	// send tx to safe contract
	params := []string{
		hex.EncodeToString([]byte(chainSpecificToken)),
		hex.EncodeToString(value),
		hex.EncodeToString([]byte(createTransactionParam)),
		hex.EncodeToString(receiver),
	}

	hash, err = mvxChainSimulator.ScCall(ctx, senderKeys.pk, senderKeys.sk, safeAddress, zeroValue, esdtTransfer, params)
	require.NoError(t, err)

	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("MVX->ETH transaction sent", "hash", hash, "status", txResult.Status)

	return hash
}

func submitAggregatorBatch(t *testing.T, ctx context.Context, mvxChainSimulator chainSimulatorWrapper, aggregatorAddress string, ownerKeys keysHolder) {
	feeInt, _ := big.NewInt(0).SetString(fee, 10)
	timestamp := big.NewInt(time.Now().Unix())
	hash, err := mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		aggregatorAddress,
		zeroValue,
		submitBatch,
		[]string{hex.EncodeToString([]byte(gwei)), hex.EncodeToString([]byte(chainSpecificTokenTicker)), hex.EncodeToString(timestamp.Bytes()), hex.EncodeToString(feeInt.Bytes()), numOfDecimalsChainSpecific})
	require.NoError(t, err)
	txResult, err := mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("submit aggregator batch tx executed", "hash", hash, "submitter", ownerKeys.pk, "status", txResult.Status)

}
