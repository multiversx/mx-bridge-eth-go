//go:build slow

package slowTests

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	ethCore "github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/wrappers"
	"github.com/multiversx/mx-bridge-eth-go/config"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	"github.com/multiversx/mx-bridge-eth-go/core/converters"
	"github.com/multiversx/mx-bridge-eth-go/executors/multiversx/module"
	"github.com/multiversx/mx-bridge-eth-go/factory"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests"
	testsRelayers "github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers"
	"github.com/multiversx/mx-bridge-eth-go/parsers"
	"github.com/multiversx/mx-bridge-eth-go/status"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	logger "github.com/multiversx/mx-chain-logger-go"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/require"
)

const (
	// Bridge consts
	numRelayers              = 3
	quorum                   = "03"
	relayerPemPathFormat     = "multiversx%d.pem"
	relayerETHKeyPathFormat  = "../testdata/ethereum%d.sk"
	scCallerFilename         = "scCaller.pem"
	minRelayerStake          = "10000000000000000000" // 10 EGLD
	fee                      = "50"
	maxBridgedAmountForToken = "500000"

	// MultiversX related consts
	proxyCacherExpirationSeconds                 = 600
	proxyMaxNoncesDelta                          = 7
	hexTrue                                      = "01"
	hexFalse                                     = "00"
	trueStr                                      = "true"
	zeroValue                                    = "0"
	slashAmount                                  = "00"
	emptyAddress                                 = "erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu"
	esdtSystemSCAddress                          = "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"
	aggregatorContract                           = "testdata/contracts/mvx/aggregator.wasm"
	wrapperContract                              = "testdata/contracts/mvx/bridged-tokens-wrapper.wasm"
	multiTransferContract                        = "testdata/contracts/mvx/multi-transfer-esdt.wasm"
	safeContract                                 = "testdata/contracts/mvx/esdt-safe.wasm"
	multisigContract                             = "testdata/contracts/mvx/multisig.wasm"
	bridgeProxyContract                          = "testdata/contracts/mvx/bridge-proxy.wasm"
	testCallerContract                           = "testdata/contracts/mvx/test-caller.wasm"
	setWrappingContractAddress                   = "setWrappingContractAddress"
	setBridgeProxyContractAddress                = "setBridgeProxyContractAddress"
	changeOwnerAddress                           = "ChangeOwnerAddress"
	unpause                                      = "unpause"
	unpauseEsdtSafe                              = "unpauseEsdtSafe"
	setEsdtSafeOnMultiTransfer                   = "setEsdtSafeOnMultiTransfer"
	setPairDecimals                              = "setPairDecimals"
	issue                                        = "issue"
	canAddSpecialRoles                           = "canAddSpecialRoles"
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
	unwrapToken                                  = "unwrapToken"
	createTransactionParam                       = "createTransaction"
	gwei                                         = "GWEI"
	chainSpecificTokenTicker                     = "ETHUSDC"
	numOfDecimalsUniversal                       = "06"
	numOfDecimalsChainSpecific                   = "06"
	universalTokenTicker                         = "USDC"
	universalTokenDisplayName                    = "WrappedUSDC"
	chainSpecificTokenDisplayName                = "EthereumWrappedUSDC"
	esdtIssueCost                                = "5000000000000000000" // 5egld
	valueToMint                                  = "10000000000"

	// Ethereum related consts
	ethSimulatedGasLimit          = 9000000
	erc20SafeABI                  = "testdata/contracts/eth/ERC20Safe.json"
	erc20SafeBytecode             = "testdata/contracts/eth/ERC20Safe.hex"
	bridgeABI                     = "testdata/contracts/eth/Bridge.json"
	bridgeBytecode                = "testdata/contracts/eth/Bridge.hex"
	scExecProxyABI                = "testdata/contracts/eth/SCExecProxy.json"
	scExecProxyBytecode           = "testdata/contracts/eth/SCExecProxy.hex"
	genericERC20ABI               = "testdata/contracts/eth/GenericERC20.json"
	genericERC20Bytecode          = "testdata/contracts/eth/GenericERC20.hex"
	ethTokenName                  = "ETHTOKEN"
	ethTokenSymbol                = "ETHT"
	ethMinAmountAllowedToTransfer = 25
	ethMaxAmountAllowedToTransfer = 500000
	ethStatusSuccess              = uint64(1)
)

var (
	addressPubkeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, "erd")
	log                       = logger.GetOrCreate("integrationTests/relayers/slowTests")
	ethOwnerSK, _             = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	ethDepositorSK, _         = crypto.HexToECDSA("9bb971db41e3815a669a71c3f1bcb24e0b81f21e04bf11faa7a34b9b40e7cfb1")
	mintAmount                = big.NewInt(20000)
	feeInt, _                 = big.NewInt(0).SetString(fee, 10)
)

type keysHolder struct {
	pk         string
	sk         []byte
	ethSK      *ecdsa.PrivateKey
	ethAddress common.Address
}

type argSimulatedSetup struct {
	t                    *testing.T
	mvxIsMintBurn        bool
	mvxIsNative          bool
	ethIsMintBurn        bool
	ethIsNative          bool
	ethSCCallMethod      string
	ethSCCallGasLimit    uint64
	ethSCCallArguments   []string
	transferBackAndForth bool
}

type simulatedSetup struct {
	*testing.T
	testContextCancel        func()
	testContext              context.Context
	simulatedETHChain        *backends.SimulatedBackend
	simulatedETHChainWrapper blockchainClient
	mvxChainSimulator        chainSimulatorWrapper
	relayers                 []bridgeComponents
	relayersKeys             []keysHolder
	scCallerKeys             keysHolder
	mvxUniversalToken        string
	mvxChainSpecificToken    string
	mvxReceiverAddress       sdkCore.AddressHandler
	mvxReceiverKeys          keysHolder
	mvxSafeAddress           string
	mvxWrapperAddress        string
	mvxMultisigAddress       string
	mvxAggregatorAddress     string
	mvxScProxyAddress        string
	mvxTestCallerAddress     sdkCore.AddressHandler
	ethSCCallMethod          string
	ethSCCallGasLimit        uint64
	ethSCCallArguments       []string
	mvxOwnerKeys             keysHolder
	ethOwnerAddress          common.Address
	ethGenericTokenAddress   common.Address
	ethGenericTokenContract  *contract.GenericERC20
	ethChainID               *big.Int
	ethSafeAddress           common.Address
	ethSafeContract          *contract.ERC20Safe
	ethSCProxyAddress        common.Address
	ethSCProxyContract       *contract.SCExecProxy
	ethBridgeAddress         common.Address
	ethBridgeContract        *contract.Bridge
	workingDir               string
	scCallerModule           io.Closer
}

func prepareSimulatedSetup(args argSimulatedSetup) *simulatedSetup {
	var err error
	testSetup := &simulatedSetup{
		T:                  args.t,
		workingDir:         args.t.TempDir(),
		ethSCCallMethod:    args.ethSCCallMethod,
		ethSCCallGasLimit:  args.ethSCCallGasLimit,
		ethSCCallArguments: args.ethSCCallArguments,
	}

	// create a test context
	testSetup.testContext, testSetup.testContextCancel = context.WithCancel(context.Background())

	testSetup.generateKeys()

	// generate the mvx receiver keys
	testSetup.mvxReceiverKeys = generateMvxPrivatePublicKey(args.t)
	testSetup.mvxReceiverAddress, err = data.NewAddressFromBech32String(testSetup.mvxReceiverKeys.pk)
	require.NoError(args.t, err)

	testSetup.ethOwnerAddress = crypto.PubkeyToAddress(ethOwnerSK.PublicKey)
	ethDepositorAddr := crypto.PubkeyToAddress(ethDepositorSK.PublicKey)

	// create ethereum simulator
	testSetup.createEthereumSimulatorAndDeployContracts(ethDepositorAddr, args.ethIsMintBurn, args.ethIsNative)

	// generate the mvx owner keys
	testSetup.mvxOwnerKeys = generateMvxPrivatePublicKey(args.t)

	erc20ContractsHolder, err := ethereum.NewErc20SafeContractsHolder(ethereum.ArgsErc20SafeContractsHolder{
		EthClient:              testSetup.simulatedETHChain,
		EthClientStatusHandler: &testsCommon.StatusHandlerStub{},
	})
	require.NoError(args.t, err)

	ethChainWrapper, err := wrappers.NewEthereumChainWrapper(wrappers.ArgsEthereumChainWrapper{
		StatusHandler:    &testsCommon.StatusHandlerStub{},
		MultiSigContract: testSetup.ethBridgeContract,
		SafeContract:     testSetup.ethSafeContract,
		BlockchainClient: testSetup.simulatedETHChainWrapper,
	})
	require.NoError(args.t, err)

	testSetup.startChainSimulatorWrapper()
	testSetup.mvxChainSimulator.GenerateBlocksUntilEpochReached(testSetup.testContext, 1)

	// deploy all contracts and execute all txs needed
	testSetup.executeContractsTxs()

	// issue and whitelist token
	testSetup.issueAndWhitelistToken(args.mvxIsMintBurn, args.mvxIsNative)

	// start relayers
	testSetup.startRelayers(ethChainWrapper, erc20ContractsHolder)

	testSetup.startScCallerModule()

	return testSetup
}

func (testSetup *simulatedSetup) close() {
	testSetup.closeRelayers()

	require.NoError(testSetup, testSetup.simulatedETHChain.Close())

	testSetup.testContextCancel()
	_ = testSetup.scCallerModule.Close()
}

func (testSetup *simulatedSetup) closeRelayers() {
	for _, r := range testSetup.relayers {
		_ = r.Close()
	}
}

func (testSetup *simulatedSetup) generateKeys() {
	relayersKeys := make([]keysHolder, 0, numRelayers)
	for i := 0; i < numRelayers; i++ {
		relayerKeys := generateMvxPrivatePublicKey(testSetup)
		log.Info("generated relayer", "index", i, "address", relayerKeys.pk)

		relayerETHSKBytes, err := os.ReadFile(fmt.Sprintf(relayerETHKeyPathFormat, i))
		require.Nil(testSetup, err)
		relayerKeys.ethSK, err = crypto.HexToECDSA(string(relayerETHSKBytes))
		require.Nil(testSetup, err)
		relayerKeys.ethAddress = crypto.PubkeyToAddress(relayerKeys.ethSK.PublicKey)

		relayersKeys = append(relayersKeys, relayerKeys)

		filename := path.Join(testSetup.workingDir, fmt.Sprintf(relayerPemPathFormat, i))
		saveMvxKey(testSetup, filename, relayerKeys)
	}

	testSetup.relayersKeys = relayersKeys
	testSetup.scCallerKeys = generateMvxPrivatePublicKey(testSetup)
	filename := path.Join(testSetup.workingDir, scCallerFilename)
	saveMvxKey(testSetup, filename, testSetup.scCallerKeys)
}

func generateMvxPrivatePublicKey(tb testing.TB) keysHolder {
	keyGenerator := signing.NewKeyGenerator(ed25519.NewEd25519())
	sk, pk := keyGenerator.GeneratePair()

	skBytes, err := sk.ToByteArray()
	require.Nil(tb, err)

	pkBytes, err := pk.ToByteArray()
	require.Nil(tb, err)

	address, err := addressPubkeyConverter.Encode(pkBytes)
	require.Nil(tb, err)

	return keysHolder{
		pk: address,
		sk: skBytes,
	}
}

func saveMvxKey(tb testing.TB, filename string, key keysHolder) {
	blk := pem.Block{
		Type:  "PRIVATE KEY for " + key.pk,
		Bytes: []byte(hex.EncodeToString(key.sk)),
	}

	buff := bytes.NewBuffer(make([]byte, 0))
	err := pem.Encode(buff, &blk)
	require.Nil(tb, err)

	err = os.WriteFile(filename, buff.Bytes(), os.ModePerm)
	require.Nil(tb, err)
}

func (testSetup *simulatedSetup) startChainSimulatorWrapper() {
	// create a new working directory
	tmpDir := path.Join(testSetup.TempDir(), "test")
	err := os.MkdirAll(tmpDir, os.ModePerm)
	require.NoError(testSetup, err)

	// start the chain simulator
	args := integrationTests.ArgChainSimulatorWrapper{
		ProxyCacherExpirationSeconds: proxyCacherExpirationSeconds,
		ProxyMaxNoncesDelta:          proxyMaxNoncesDelta,
	}
	testSetup.mvxChainSimulator, err = integrationTests.CreateChainSimulatorWrapper(args)
	require.NoError(testSetup, err)
}

func (testSetup *simulatedSetup) startRelayers(
	ethereumChain ethereum.ClientWrapper,
	erc20ContractsHolder ethereum.Erc20ContractsHolder,
) {
	relayers := make([]bridgeComponents, 0, numRelayers)

	messengers := integrationTests.CreateLinkedMessengers(numRelayers)

	for i := 0; i < numRelayers; i++ {
		generalConfigs := testsRelayers.CreateBridgeComponentsConfig(i, testSetup.workingDir)
		generalConfigs.Eth.PrivateKeyFile = fmt.Sprintf(relayerETHKeyPathFormat, i)
		argsBridgeComponents := factory.ArgsEthereumToMultiversXBridge{
			Configs: config.Configs{
				GeneralConfig:   generalConfigs,
				ApiRoutesConfig: config.ApiRoutesConfig{},
				FlagsConfig: config.ContextFlagsConfig{
					RestApiInterface: bridgeCore.WebServerOffString,
				},
			},
			Proxy:                         testSetup.mvxChainSimulator.Proxy(),
			ClientWrapper:                 ethereumChain,
			Messenger:                     messengers[i],
			StatusStorer:                  testsCommon.NewStorerMock(),
			TimeForBootstrap:              time.Second * 5,
			TimeBeforeRepeatJoin:          time.Second * 30,
			MetricsHolder:                 status.NewMetricsHolder(),
			AppStatusHandler:              &statusHandler.AppStatusHandlerStub{},
			MultiversXClientStatusHandler: &testsCommon.StatusHandlerStub{},
		}
		argsBridgeComponents.Configs.GeneralConfig.Eth.SafeContractAddress = testSetup.ethSafeAddress.Hex()
		argsBridgeComponents.Configs.GeneralConfig.Eth.SCExecProxyAddress = testSetup.ethSCProxyAddress.Hex()
		argsBridgeComponents.Erc20ContractsHolder = erc20ContractsHolder
		argsBridgeComponents.Configs.GeneralConfig.MultiversX.NetworkAddress = testSetup.mvxChainSimulator.GetNetworkAddress()
		argsBridgeComponents.Configs.GeneralConfig.MultiversX.SafeContractAddress = testSetup.mvxSafeAddress
		argsBridgeComponents.Configs.GeneralConfig.MultiversX.MultisigContractAddress = testSetup.mvxMultisigAddress
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
		require.Nil(testSetup, err)

		go func() {
			err = relayer.Start()
			integrationTests.Log.LogIfError(err)
			require.Nil(testSetup, err)
		}()

		relayers = append(relayers, relayer)
	}

	testSetup.relayers = relayers
}

func (testSetup *simulatedSetup) startScCallerModule() {
	cfg := config.ScCallsModuleConfig{
		ScProxyBech32Address:         testSetup.mvxScProxyAddress,
		ExtraGasToExecute:            20_000_000, // 20 million
		NetworkAddress:               testSetup.mvxChainSimulator.GetNetworkAddress(),
		ProxyMaxNoncesDelta:          5,
		ProxyFinalityCheck:           false,
		ProxyCacherExpirationSeconds: 60, // 1 minute
		ProxyRestAPIEntityType:       string(sdkCore.Proxy),
		IntervalToResendTxsInSeconds: 1,
		PrivateKeyFile:               path.Join(testSetup.workingDir, scCallerFilename),
		PollingIntervalInMillis:      1000, // 1 second
		FilterConfig: config.PendingOperationsFilterConfig{
			AllowedEthAddresses: []string{"*"},
			AllowedMvxAddresses: []string{"*"},
			AllowedTokens:       []string{"*"},
		},
	}

	var err error
	testSetup.scCallerModule, err = module.NewScCallsModule(cfg, log)
	require.Nil(testSetup, err)

	log.Info("started SC calls module", "monitoring SC proxy address", testSetup.mvxScProxyAddress)
}

func (testSetup *simulatedSetup) executeContractsTxs() {
	var err error

	// fund the involved wallets(owner + relayers)
	walletsToFund := make([]string, 0, len(testSetup.relayersKeys)+2)
	for _, relayerKeys := range testSetup.relayersKeys {
		walletsToFund = append(walletsToFund, relayerKeys.pk)
	}
	walletsToFund = append(walletsToFund, testSetup.mvxOwnerKeys.pk)
	walletsToFund = append(walletsToFund, testSetup.mvxReceiverKeys.pk)
	walletsToFund = append(walletsToFund, testSetup.scCallerKeys.pk)
	testSetup.mvxChainSimulator.FundWallets(testSetup.testContext, walletsToFund)

	testSetup.mvxChainSimulator.GenerateBlocks(testSetup.testContext, 1)

	// deploy aggregator
	stakeValue, _ := big.NewInt(0).SetString(minRelayerStake, 10)
	aggregatorDeployParams := []string{
		hex.EncodeToString([]byte("EGLD")),
		hex.EncodeToString(stakeValue.Bytes()),
		"01",
		"01",
		"01",
		getHexAddress(testSetup, testSetup.mvxOwnerKeys.pk),
	}

	testSetup.mvxAggregatorAddress, err = testSetup.mvxChainSimulator.DeploySC(
		testSetup.testContext,
		aggregatorContract,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		aggregatorDeployParams,
	)
	require.NoError(testSetup, err)
	require.NotEqual(testSetup, emptyAddress, testSetup.mvxAggregatorAddress)

	log.Info("aggregator contract deployed", "address", testSetup.mvxAggregatorAddress)

	// deploy wrapper
	testSetup.mvxWrapperAddress, err = testSetup.mvxChainSimulator.DeploySC(
		testSetup.testContext,
		wrapperContract,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		[]string{},
	)
	require.NoError(testSetup, err)
	require.NotEqual(testSetup, emptyAddress, testSetup.mvxWrapperAddress)

	log.Info("wrapper contract deployed", "address", testSetup.mvxWrapperAddress)

	// deploy multi-transfer
	multiTransferAddress, err := testSetup.mvxChainSimulator.DeploySC(
		testSetup.testContext,
		multiTransferContract,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		[]string{},
	)
	require.NoError(testSetup, err)
	require.NotEqual(testSetup, emptyAddress, multiTransferAddress)

	log.Info("multi-transfer contract deployed", "address", multiTransferAddress)

	// deploy safe
	testSetup.mvxSafeAddress, err = testSetup.mvxChainSimulator.DeploySC(
		testSetup.testContext,
		safeContract,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		[]string{
			getHexAddress(testSetup, testSetup.mvxAggregatorAddress),
			getHexAddress(testSetup, multiTransferAddress),
			"01",
		},
	)
	require.NoError(testSetup, err)
	require.NotEqual(testSetup, emptyAddress, testSetup.mvxSafeAddress)

	log.Info("safe contract deployed", "address", testSetup.mvxSafeAddress)

	// deploy multisig
	minRelayerStakeInt, _ := big.NewInt(0).SetString(minRelayerStake, 10)
	minRelayerStakeHex := hex.EncodeToString(minRelayerStakeInt.Bytes())
	params := []string{getHexAddress(testSetup, testSetup.mvxSafeAddress), getHexAddress(testSetup, multiTransferAddress), minRelayerStakeHex, slashAmount, quorum}
	for _, relayerKeys := range testSetup.relayersKeys {
		params = append(params, getHexAddress(testSetup, relayerKeys.pk))
	}
	testSetup.mvxMultisigAddress, err = testSetup.mvxChainSimulator.DeploySC(
		testSetup.testContext,
		multisigContract,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		params,
	)
	require.NoError(testSetup, err)
	require.NotEqual(testSetup, emptyAddress, testSetup.mvxMultisigAddress)

	log.Info("multisig contract deployed", "address", testSetup.mvxMultisigAddress)

	// deploy bridge proxy
	testSetup.mvxScProxyAddress, err = testSetup.mvxChainSimulator.DeploySC(
		testSetup.testContext,
		bridgeProxyContract,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		[]string{getHexAddress(testSetup, multiTransferAddress)},
	)
	require.NoError(testSetup, err)
	require.NotEqual(testSetup, emptyAddress, testSetup.mvxScProxyAddress)

	log.Info("bridge proxy contract deployed", "address", testSetup.mvxScProxyAddress)

	// deploy test-caller
	mvxTestCallerAddress, err := testSetup.mvxChainSimulator.DeploySC(
		testSetup.testContext,
		testCallerContract,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		[]string{},
	)
	require.NoError(testSetup, err)
	require.NotEqual(testSetup, emptyAddress, mvxTestCallerAddress)
	testSetup.mvxTestCallerAddress, err = data.NewAddressFromBech32String(mvxTestCallerAddress)
	require.NoError(testSetup, err)

	log.Info("test-caller contract deployed", "address", mvxTestCallerAddress)

	// setBridgeProxyContractAddress
	hash, err := testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		multiTransferAddress,
		zeroValue,
		setBridgeProxyContractAddress,
		[]string{getHexAddress(testSetup, testSetup.mvxScProxyAddress)},
	)
	require.NoError(testSetup, err)
	txResult, err := testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("setBridgeProxyContractAddress tx executed", "hash", hash, "status", txResult.Status)

	// setWrappingContractAddress
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		multiTransferAddress,
		zeroValue,
		setWrappingContractAddress,
		[]string{getHexAddress(testSetup, testSetup.mvxWrapperAddress)},
	)
	require.NoError(testSetup, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("setWrappingContractAddress tx executed", "hash", hash, "status", txResult.Status)

	// ChangeOwnerAddress for safe
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		testSetup.mvxSafeAddress,
		zeroValue,
		changeOwnerAddress,
		[]string{getHexAddress(testSetup, testSetup.mvxMultisigAddress)},
	)
	require.NoError(testSetup, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("ChangeOwnerAddress for safe tx executed", "hash", hash, "status", txResult.Status)

	// ChangeOwnerAddress for multi-transfer
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		multiTransferAddress,
		zeroValue,
		changeOwnerAddress,
		[]string{getHexAddress(testSetup, testSetup.mvxMultisigAddress)},
	)
	require.NoError(testSetup, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("ChangeOwnerAddress for multi-transfer tx executed", "hash", hash, "status", txResult.Status)

	// unpause sc proxy
	hash = testSetup.unpauseContract(testSetup.mvxScProxyAddress, []byte(unpause))
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)
	log.Info("unpaused sc proxy executed", "hash", hash, "status", txResult.Status)

	// ChangeOwnerAddress for bridge proxy
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		testSetup.mvxScProxyAddress,
		zeroValue,
		changeOwnerAddress,
		[]string{getHexAddress(testSetup, testSetup.mvxMultisigAddress)},
	)
	require.NoError(testSetup, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("ChangeOwnerAddress for bridge proxy tx executed", "hash", hash, "status", txResult.Status)

	// setEsdtSafeOnMultiTransfer
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		testSetup.mvxMultisigAddress,
		zeroValue,
		setEsdtSafeOnMultiTransfer,
		[]string{},
	)
	require.NoError(testSetup, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("setEsdtSafeOnMultiTransfer tx executed", "hash", hash, "status", txResult.Status)

	// setPairDecimals on aggregator
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		testSetup.mvxAggregatorAddress,
		zeroValue,
		setPairDecimals,
		[]string{hex.EncodeToString([]byte(gwei)), hex.EncodeToString([]byte(chainSpecificTokenTicker)), numOfDecimalsChainSpecific})
	require.NoError(testSetup, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("setPairDecimals tx executed", "hash", hash, "status", txResult.Status)

	// stake relayers on multisig
	testSetup.stakeAddressesOnContract(testSetup.mvxMultisigAddress, testSetup.relayersKeys)

	// stake relayers on price aggregator
	testSetup.stakeAddressesOnContract(testSetup.mvxAggregatorAddress, []keysHolder{testSetup.mvxOwnerKeys})

	// unpause multisig
	hash = testSetup.unpauseContract(testSetup.mvxMultisigAddress, []byte(unpause))
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)
	log.Info("unpaused multisig executed", "hash", hash, "status", txResult.Status)

	// unpause safe
	hash = testSetup.unpauseContract(testSetup.mvxMultisigAddress, []byte(unpauseEsdtSafe))
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)
	log.Info("unpaused safe executed", "hash", hash, "status", txResult.Status)

	// unpause aggregator
	hash = testSetup.unpauseContract(testSetup.mvxAggregatorAddress, []byte(unpause))
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)
	log.Info("unpaused aggregator executed", "hash", hash, "status", txResult.Status)

	// unpause wrapper
	hash = testSetup.unpauseContract(testSetup.mvxWrapperAddress, []byte(unpause))
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)
	log.Info("unpaused wrapper executed", "hash", hash, "status", txResult.Status)
}

func (testSetup *simulatedSetup) stakeAddressesOnContract(contract string, allKeys []keysHolder) {
	for _, keys := range allKeys {
		hash, err := testSetup.mvxChainSimulator.SendTx(testSetup.testContext, keys.pk, keys.sk, contract, minRelayerStake, []byte("stake"))
		require.NoError(testSetup, err)
		txResult, err := testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
		require.NoError(testSetup, err)

		log.Info(fmt.Sprintf("address %s staked on contract %s with hash %s, status %s", keys.pk, contract, hash, txResult.Status))
	}
}

func (testSetup *simulatedSetup) unpauseContract(contract string, dataField []byte) string {
	hash, err := testSetup.mvxChainSimulator.SendTx(testSetup.testContext, testSetup.mvxOwnerKeys.pk, testSetup.mvxOwnerKeys.sk, contract, zeroValue, dataField)
	require.NoError(testSetup, err)

	return hash
}

func (testSetup *simulatedSetup) issueAndWhitelistToken(
	isMintBurn bool,
	isNative bool,
) {
	// issue universal token
	hash, err := testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		esdtSystemSCAddress,
		esdtIssueCost,
		issue,
		[]string{hex.EncodeToString([]byte(universalTokenDisplayName)), hex.EncodeToString([]byte(universalTokenTicker)), "00", numOfDecimalsUniversal, hex.EncodeToString([]byte(canAddSpecialRoles)), hex.EncodeToString([]byte(trueStr))})
	require.NoError(testSetup, err)
	txResult, err := testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)
	testSetup.mvxUniversalToken = getTokenNameFromResult(testSetup, *txResult)

	log.Info("issue universal token tx executed", "hash", hash, "status", txResult.Status, "token", testSetup.mvxUniversalToken, "owner", testSetup.mvxOwnerKeys.pk)

	// issue chain specific token
	valueToMintInt, _ := big.NewInt(0).SetString(valueToMint, 10)
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		esdtSystemSCAddress,
		esdtIssueCost,
		issue,
		[]string{hex.EncodeToString([]byte(chainSpecificTokenDisplayName)), hex.EncodeToString([]byte(chainSpecificTokenTicker)), hex.EncodeToString(valueToMintInt.Bytes()), numOfDecimalsChainSpecific, hex.EncodeToString([]byte(canAddSpecialRoles)), hex.EncodeToString([]byte(trueStr))})
	require.NoError(testSetup, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)
	testSetup.mvxChainSpecificToken = getTokenNameFromResult(testSetup, *txResult)

	log.Info("issue chain specific token tx executed", "hash", hash, "status", txResult.Status, "token", testSetup.mvxChainSpecificToken, "owner", testSetup.mvxOwnerKeys.pk)

	// set local roles bridged tokens wrapper
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		esdtSystemSCAddress,
		zeroValue,
		setSpecialRole,
		[]string{hex.EncodeToString([]byte(testSetup.mvxUniversalToken)), getHexAddress(testSetup, testSetup.mvxWrapperAddress), hex.EncodeToString([]byte(esdtRoleLocalMint)), hex.EncodeToString([]byte(esdtRoleLocalBurn))})
	require.NoError(testSetup, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("set local roles bridged tokens wrapper tx executed", "hash", hash, "status", txResult.Status)

	// transfer to wrapper sc
	initialMintValue := valueToMintInt.Div(valueToMintInt, big.NewInt(3))
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		testSetup.mvxWrapperAddress,
		zeroValue,
		esdtTransfer,
		[]string{hex.EncodeToString([]byte(testSetup.mvxChainSpecificToken)), hex.EncodeToString(initialMintValue.Bytes()), hex.EncodeToString([]byte(depositLiquidity))})
	require.NoError(testSetup, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("transfer to wrapper sc tx executed", "hash", hash, "status", txResult.Status)

	// transfer to safe sc
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		testSetup.mvxSafeAddress,
		zeroValue,
		esdtTransfer,
		[]string{hex.EncodeToString([]byte(testSetup.mvxChainSpecificToken)), hex.EncodeToString(initialMintValue.Bytes())})
	require.NoError(testSetup, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("transfer to safe sc tx executed", "hash", hash, "status", txResult.Status)

	// add wrapped token
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		testSetup.mvxWrapperAddress,
		zeroValue,
		addWrappedToken,
		[]string{hex.EncodeToString([]byte(testSetup.mvxUniversalToken)), numOfDecimalsUniversal})
	require.NoError(testSetup, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("add wrapped token tx executed", "hash", hash, "status", txResult.Status)

	// wrapper whitelist token
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		testSetup.mvxWrapperAddress,
		zeroValue,
		whitelistToken,
		[]string{hex.EncodeToString([]byte(testSetup.mvxChainSpecificToken)), numOfDecimalsChainSpecific, hex.EncodeToString([]byte(testSetup.mvxUniversalToken))})
	require.NoError(testSetup, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("wrapper whitelist token tx executed", "hash", hash, "status", txResult.Status)

	// set local roles esdt safe
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		esdtSystemSCAddress,
		zeroValue,
		setSpecialRole,
		[]string{hex.EncodeToString([]byte(testSetup.mvxChainSpecificToken)), getHexAddress(testSetup, testSetup.mvxSafeAddress), hex.EncodeToString([]byte(esdtRoleLocalMint)), hex.EncodeToString([]byte(esdtRoleLocalBurn))})
	require.NoError(testSetup, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("set local roles esdt safe tx executed", "hash", hash, "status", txResult.Status)

	// add mapping
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		testSetup.mvxMultisigAddress,
		zeroValue,
		addMapping,
		[]string{hex.EncodeToString(testSetup.ethGenericTokenAddress.Bytes()), hex.EncodeToString([]byte(testSetup.mvxChainSpecificToken))})
	require.NoError(testSetup, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("add mapping tx executed", "hash", hash, "status", txResult.Status)

	// whitelist token
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		testSetup.mvxMultisigAddress,
		zeroValue,
		esdtSafeAddTokenToWhitelist,
		[]string{hex.EncodeToString([]byte(testSetup.mvxChainSpecificToken)), hex.EncodeToString([]byte(chainSpecificTokenTicker)), getHexBool(isMintBurn), getHexBool(isNative)})
	require.NoError(testSetup, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("whitelist token tx executed", "hash", hash, "status", txResult.Status)

	// submit aggregator batch
	testSetup.submitAggregatorBatch()

	// safe set max bridge amount for token
	maxBridgedAmountForTokenInt, _ := big.NewInt(0).SetString(maxBridgedAmountForToken, 10)
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		testSetup.mvxMultisigAddress,
		zeroValue,
		esdtSafeSetMaxBridgedAmountForToken,
		[]string{hex.EncodeToString([]byte(testSetup.mvxChainSpecificToken)), hex.EncodeToString(maxBridgedAmountForTokenInt.Bytes())})
	require.NoError(testSetup, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("safe set max bridge amount for token tx executed", "hash", hash, "status", txResult.Status)

	// multi-transfer set max bridge amount for token
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		testSetup.mvxMultisigAddress,
		zeroValue,
		multiTransferEsdtSetMaxBridgedAmountForToken,
		[]string{hex.EncodeToString([]byte(testSetup.mvxChainSpecificToken)), hex.EncodeToString(maxBridgedAmountForTokenInt.Bytes())})
	require.NoError(testSetup, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("multi-transfer set max bridge amount for token tx executed", "hash", hash, "status", txResult.Status)
}

func getTokenNameFromResult(tb testing.TB, txResult data.TransactionOnNetwork) string {
	resultData := txResult.ScResults[0].Data
	splittedData := strings.Split(resultData, "@")
	if len(splittedData) < 2 {
		require.Fail(tb, fmt.Sprintf("received invalid data received while issuing: %s", resultData))
	}

	newUniversalTokenBytes, err := hex.DecodeString(splittedData[1])
	require.NoError(tb, err)

	return string(newUniversalTokenBytes)
}

func getHexAddress(tb testing.TB, bech32Address string) string {
	address, err := data.NewAddressFromBech32String(bech32Address)
	require.NoError(tb, err)

	return hex.EncodeToString(address.AddressBytes())
}

func getHexBool(input bool) string {
	if input {
		return hexTrue
	}

	return hexFalse
}

func (testSetup *simulatedSetup) checkESDTBalance(
	address sdkCore.AddressHandler,
	token string,
	expectedBalance string,
	checkResult bool,
) bool {
	balance, err := testSetup.mvxChainSimulator.GetESDTBalance(testSetup.testContext, address, token)
	require.NoError(testSetup, err)

	if checkResult {
		require.Equal(testSetup, expectedBalance, balance)
	}

	return expectedBalance == balance
}

func (testSetup *simulatedSetup) checkETHStatus(receiver common.Address, expectedBalance uint64) bool {
	balance, err := testSetup.ethGenericTokenContract.BalanceOf(nil, receiver)
	require.NoError(testSetup, err)

	return balance.Uint64() == expectedBalance
}

func (testSetup *simulatedSetup) sendMVXToEthTransaction(value []byte) string {
	// unwrap token
	paramsUnwrap := []string{
		hex.EncodeToString([]byte(testSetup.mvxUniversalToken)),
		hex.EncodeToString(value),
		hex.EncodeToString([]byte(unwrapToken)),
		hex.EncodeToString([]byte(testSetup.mvxChainSpecificToken)),
	}

	hash, err := testSetup.mvxChainSimulator.ScCall(testSetup.testContext, testSetup.mvxReceiverKeys.pk, testSetup.mvxReceiverKeys.sk, testSetup.mvxWrapperAddress, zeroValue, esdtTransfer, paramsUnwrap)
	require.NoError(testSetup, err)
	txResult, err := testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("unwrap transaction sent", "hash", hash, "token", testSetup.mvxUniversalToken, "status", txResult.Status)

	// send tx to safe contract
	params := []string{
		hex.EncodeToString([]byte(testSetup.mvxChainSpecificToken)),
		hex.EncodeToString(value),
		hex.EncodeToString([]byte(createTransactionParam)),
		hex.EncodeToString(testSetup.ethOwnerAddress.Bytes()),
	}

	hash, err = testSetup.mvxChainSimulator.ScCall(testSetup.testContext, testSetup.mvxReceiverKeys.pk, testSetup.mvxReceiverKeys.sk, testSetup.mvxSafeAddress, zeroValue, esdtTransfer, params)
	require.NoError(testSetup, err)

	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("MVX->ETH transaction sent", "hash", hash, "status", txResult.Status)

	return hash
}

func (testSetup *simulatedSetup) submitAggregatorBatch() {
	timestamp, err := testSetup.mvxChainSimulator.GetBlockchainTimeStamp(testSetup.testContext)
	require.Nil(testSetup, err)
	require.Greater(testSetup, timestamp, uint64(0), "something went wrong and the chain simulator returned 0 for the current timestamp")

	timestampAsBigInt := big.NewInt(0).SetUint64(timestamp)

	hash, err := testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		testSetup.mvxAggregatorAddress,
		zeroValue,
		submitBatch,
		[]string{hex.EncodeToString([]byte(gwei)), hex.EncodeToString([]byte(chainSpecificTokenTicker)), hex.EncodeToString(timestampAsBigInt.Bytes()), hex.EncodeToString(feeInt.Bytes()), numOfDecimalsChainSpecific})
	require.NoError(testSetup, err)
	txResult, err := testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("submit aggregator batch tx executed", "hash", hash, "submitter", testSetup.mvxOwnerKeys.pk, "status", txResult.Status)
}

func (testSetup *simulatedSetup) createEthereumSimulatorAndDeployContracts(
	ethDepositorAddr common.Address,
	isMintBurn bool,
	isNative bool,
) {
	addr := map[common.Address]ethCore.GenesisAccount{
		testSetup.ethOwnerAddress: {Balance: new(big.Int).Lsh(big.NewInt(1), 100)},
		ethDepositorAddr:          {Balance: new(big.Int).Lsh(big.NewInt(1), 100)},
	}
	for _, relayerKeys := range testSetup.relayersKeys {
		addr[relayerKeys.ethAddress] = ethCore.GenesisAccount{Balance: new(big.Int).Lsh(big.NewInt(1), 100)}
	}
	alloc := ethCore.GenesisAlloc(addr)
	testSetup.simulatedETHChain = backends.NewSimulatedBackend(alloc, ethSimulatedGasLimit)

	testSetup.simulatedETHChainWrapper = integrationTests.NewSimulatedETHChainWrapper(testSetup.simulatedETHChain)
	testSetup.ethChainID, _ = testSetup.simulatedETHChainWrapper.ChainID(testSetup.testContext)

	// deploy safe
	testSetup.ethSafeAddress = testSetup.deployETHContract(erc20SafeABI, erc20SafeBytecode)
	ethSafeContract, err := contract.NewERC20Safe(testSetup.ethSafeAddress, testSetup.simulatedETHChain)
	require.NoError(testSetup, err)
	testSetup.ethSafeContract = ethSafeContract

	// deploy bridge
	ethRelayersAddresses := make([]common.Address, 0, len(testSetup.relayersKeys))
	for _, relayerKeys := range testSetup.relayersKeys {
		ethRelayersAddresses = append(ethRelayersAddresses, relayerKeys.ethAddress)
	}
	quorumInt, _ := big.NewInt(0).SetString(quorum, 10)
	testSetup.ethBridgeAddress = testSetup.deployETHContract(bridgeABI, bridgeBytecode, ethRelayersAddresses, quorumInt, testSetup.ethSafeAddress)
	testSetup.ethBridgeContract, err = contract.NewBridge(testSetup.ethBridgeAddress, testSetup.simulatedETHChain)
	require.NoError(testSetup, err)

	// set bridge on safe
	auth, _ := bind.NewKeyedTransactorWithChainID(ethOwnerSK, testSetup.ethChainID)
	tx, err := ethSafeContract.SetBridge(auth, testSetup.ethBridgeAddress)
	require.NoError(testSetup, err)
	testSetup.simulatedETHChain.Commit()
	testSetup.checkEthTxResult(tx.Hash())

	// deploy exec-proxy
	testSetup.ethSCProxyAddress = testSetup.deployETHContract(scExecProxyABI, scExecProxyBytecode, testSetup.ethSafeAddress)
	scProxyContract, err := contract.NewSCExecProxy(testSetup.ethSCProxyAddress, testSetup.simulatedETHChain)
	require.NoError(testSetup, err)
	testSetup.ethSCProxyContract = scProxyContract

	// deploy generic eth token
	ethGenericTokenAddress := testSetup.deployETHContract(genericERC20ABI, genericERC20Bytecode, ethTokenName, ethTokenSymbol)
	testSetup.ethGenericTokenAddress = ethGenericTokenAddress
	ethGenericTokenContract, err := contract.NewGenericERC20(ethGenericTokenAddress, testSetup.simulatedETHChain)
	require.NoError(testSetup, err)
	testSetup.ethGenericTokenContract = ethGenericTokenContract

	// mint generic token
	auth, _ = bind.NewKeyedTransactorWithChainID(ethDepositorSK, testSetup.ethChainID)
	tx, err = ethGenericTokenContract.Mint(auth, ethDepositorAddr, mintAmount)
	require.NoError(testSetup, err)
	testSetup.simulatedETHChain.Commit()
	testSetup.checkEthTxResult(tx.Hash())
	testSetup.checkETHStatus(ethDepositorAddr, mintAmount.Uint64())

	// whitelist eth token
	auth, _ = bind.NewKeyedTransactorWithChainID(ethOwnerSK, testSetup.ethChainID)
	tx, err = ethSafeContract.WhitelistToken(auth, ethGenericTokenAddress, big.NewInt(ethMinAmountAllowedToTransfer), big.NewInt(ethMaxAmountAllowedToTransfer), isMintBurn, isNative)
	require.NoError(testSetup, err)
	testSetup.simulatedETHChain.Commit()
	testSetup.checkEthTxResult(tx.Hash())

	// unpause bridge contract
	auth, _ = bind.NewKeyedTransactorWithChainID(ethOwnerSK, testSetup.ethChainID)
	tx, err = testSetup.ethBridgeContract.Unpause(auth)
	require.NoError(testSetup, err)
	testSetup.simulatedETHChain.Commit()
	testSetup.checkEthTxResult(tx.Hash())

	// unpause safe contract
	auth, _ = bind.NewKeyedTransactorWithChainID(ethOwnerSK, testSetup.ethChainID)
	tx, err = ethSafeContract.Unpause(auth)
	require.NoError(testSetup, err)
	testSetup.simulatedETHChain.Commit()
	testSetup.checkEthTxResult(tx.Hash())
}

func (testSetup *simulatedSetup) deployETHContract(
	abiFile string,
	bytecodeFile string,
	params ...interface{},
) common.Address {
	abiBytes, err := os.ReadFile(abiFile)
	require.NoError(testSetup, err)
	parsed, err := abi.JSON(bytes.NewReader(abiBytes))
	require.NoError(testSetup, err)

	contractBytes, err := os.ReadFile(bytecodeFile)
	require.NoError(testSetup, err)

	contractAuth, _ := bind.NewKeyedTransactorWithChainID(ethOwnerSK, testSetup.ethChainID)
	contractAddress, tx, _, err := bind.DeployContract(contractAuth, parsed, common.FromHex(converters.TrimWhiteSpaceCharacters(string(contractBytes))), testSetup.simulatedETHChain, params...)
	require.NoError(testSetup, err)
	testSetup.simulatedETHChain.Commit()

	testSetup.checkEthTxResult(tx.Hash())

	log.Info("deployed eth contract", "from file", bytecodeFile, "address", contractAddress.Hex())

	return contractAddress
}

func (testSetup *simulatedSetup) createBatch(direction batchProcessor.Direction) {
	if direction == batchProcessor.ToMultiversX {
		// create a pending batch on ethereum
		testSetup.createBatchOnEthereum()
	}
	if direction == batchProcessor.FromMultiversX {
		// create a pending batch on multiversx
		testSetup.createBatchOnMultiversX()
	}
}

func (testSetup *simulatedSetup) createBatchOnEthereum() {
	// add allowance for the sender
	auth, _ := bind.NewKeyedTransactorWithChainID(ethDepositorSK, testSetup.ethChainID)

	if len(testSetup.ethSCCallMethod) > 0 {
		tx, err := testSetup.ethGenericTokenContract.Approve(auth, testSetup.ethSCProxyAddress, mintAmount)
		require.NoError(testSetup, err)
		testSetup.simulatedETHChain.Commit()
		testSetup.checkEthTxResult(tx.Hash())

		codec := parsers.MultiversxCodec{}
		callData := parsers.CallData{
			Type:      parsers.DataPresentProtocolMarker,
			Function:  testSetup.ethSCCallMethod,
			GasLimit:  testSetup.ethSCCallGasLimit,
			Arguments: testSetup.ethSCCallArguments,
		}

		buff := codec.EncodeCallData(callData)

		tx, err = testSetup.ethSCProxyContract.Deposit(
			auth,
			testSetup.ethGenericTokenAddress,
			mintAmount,
			testSetup.mvxTestCallerAddress.AddressSlice(),
			string(buff),
		)
		require.NoError(testSetup, err)
		testSetup.simulatedETHChain.Commit()
		testSetup.checkEthTxResult(tx.Hash())
	} else {
		tx, err := testSetup.ethGenericTokenContract.Approve(auth, testSetup.ethSafeAddress, mintAmount)
		require.NoError(testSetup, err)
		testSetup.simulatedETHChain.Commit()
		testSetup.checkEthTxResult(tx.Hash())

		// deposit on ETH safe as a simple transfer
		auth, _ = bind.NewKeyedTransactorWithChainID(ethDepositorSK, testSetup.ethChainID)
		tx, err = testSetup.ethSafeContract.Deposit(auth, testSetup.ethGenericTokenAddress, mintAmount, testSetup.mvxReceiverAddress.AddressSlice())
		require.NoError(testSetup, err)
		testSetup.simulatedETHChain.Commit()
		testSetup.checkEthTxResult(tx.Hash())
	}

	// wait until batch is settled
	batchSettleLimit, _ := testSetup.ethSafeContract.BatchSettleLimit(nil)
	for i := uint8(0); i < batchSettleLimit+1; i++ {
		testSetup.simulatedETHChain.Commit()
	}
}

func (testSetup *simulatedSetup) createBatchOnMultiversX() {
	// create a pending batch on multiversx
	valueToSendFromMVX := big.NewInt(0).Div(mintAmount, big.NewInt(2))

	// mint erc20 token into eth safe
	auth, _ := bind.NewKeyedTransactorWithChainID(ethDepositorSK, testSetup.ethChainID)
	tx, err := testSetup.ethGenericTokenContract.Mint(auth, testSetup.ethSafeAddress, valueToSendFromMVX)
	require.NoError(testSetup, err)
	testSetup.simulatedETHChain.Commit()
	testSetup.checkEthTxResult(tx.Hash())
	testSetup.checkETHStatus(testSetup.ethSafeAddress, mintAmount.Uint64())

	// transfer to sender tx
	hash, err := testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		testSetup.mvxReceiverKeys.pk,
		zeroValue,
		esdtTransfer,
		[]string{hex.EncodeToString([]byte(testSetup.mvxChainSpecificToken)), hex.EncodeToString(valueToSendFromMVX.Bytes())})
	require.NoError(testSetup, err)
	txResult, err := testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("transfer to sender tx executed", "hash", hash, "status", txResult.Status)

	// send tx to safe contract
	params := []string{
		hex.EncodeToString([]byte(testSetup.mvxChainSpecificToken)),
		hex.EncodeToString(valueToSendFromMVX.Bytes()),
		hex.EncodeToString([]byte(createTransactionParam)),
		hex.EncodeToString(testSetup.ethOwnerAddress.Bytes()),
	}
	hash, err = testSetup.mvxChainSimulator.ScCall(testSetup.testContext, testSetup.mvxReceiverKeys.pk, testSetup.mvxReceiverKeys.sk, testSetup.mvxSafeAddress, zeroValue, esdtTransfer, params)
	require.NoError(testSetup, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup, err)

	log.Info("MVX->ETH transaction sent", "hash", hash, "status", txResult.Status)
}

func (testSetup *simulatedSetup) checkEthTxResult(hash common.Hash) {
	receipt, err := testSetup.simulatedETHChain.TransactionReceipt(testSetup.testContext, hash)
	require.NoError(testSetup, err)
	require.Equal(testSetup, ethStatusSuccess, receipt.Status)
}
