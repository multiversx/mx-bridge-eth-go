// To run these slow tests, simply add the slow tag on the go test command. Also, provide a chain simulator instance on the 8085 port
// example: go test -tags slow

package relayers_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
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

	goEthereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	ethCore "github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/wrappers"
	"github.com/multiversx/mx-bridge-eth-go/clients/multiversx"
	"github.com/multiversx/mx-bridge-eth-go/config"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	"github.com/multiversx/mx-bridge-eth-go/core/converters"
	"github.com/multiversx/mx-bridge-eth-go/factory"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/mock"
	testsRelayers "github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers"
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

var addressPubkeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, "erd")

const (
	numRelayers                                  = 3
	safeContract                                 = "testdata/contracts/mvx/esdt-safe.wasm"
	multisigContract                             = "testdata/contracts/mvx/multisig.wasm"
	multiTransferContract                        = "testdata/contracts/mvx/multi-transfer-esdt.wasm"
	bridgeProxyContract                          = "testdata/contracts/mvx/bridge-proxy.wasm"
	aggregatorContract                           = "testdata/contracts/mvx/aggregator.wasm"
	wrapperContract                              = "testdata/contracts/mvx/bridged-tokens-wrapper.wasm"
	bridgeABI                                    = "testdata/contracts/eth/bridgeABI.json"
	bridgeBytecode                               = "testdata/contracts/eth/bridgeBytecode.hex"
	erc20SafeABI                                 = "testdata/contracts/eth/erc20SafeABI.json"
	erc20SafeBytecode                            = "testdata/contracts/eth/erc20SafeBytecode.hex"
	genericERC20ABI                              = "testdata/contracts/eth/genericERC20ABI.json"
	genericERC20Bytecode                         = "testdata/contracts/eth/genericERC20Bytecode.hex"
	scExecProxyABI                               = "testdata/contracts/eth/scExecProxyABI.json"
	scExecProxyBytecode                          = "testdata/contracts/eth/scExecProxyBytecode.hex"
	minRelayerStake                              = "10000000000000000000" // 10egld
	slashAmount                                  = "00"
	quorum                                       = "03"
	relayerPemPathFormat                         = "multiversx%d.pem"
	relayerETHKeyPathFormat                      = "testdata/ethereum%d.sk"
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
	hexTrue                                      = "01"
	hexFalse                                     = "00"
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
	fee                                          = "50"
	maxBridgedAmountForToken                     = "500000"
	createTransactionParam                       = "createTransaction"
	unwrapToken                                  = "unwrapToken"
	setPairDecimals                              = "setPairDecimals"
	ethStatusSuccess                             = uint64(1)
	ethTokenName                                 = "ETHTOKEN"
	ethTokenSymbol                               = "ETHT"
	ethMinAmountAllowedToTransfer                = 25
	ethMaxAmountAllowedToTransfer                = 500000
	ethSimulatedGasLimit                         = 9000000
	timeout                                      = time.Minute * 15
)

var (
	ethOwnerSK, _     = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	ethDepositorSK, _ = crypto.HexToECDSA("9bb971db41e3815a669a71c3f1bcb24e0b81f21e04bf11faa7a34b9b40e7cfb1")
	mintAmount        = big.NewInt(20000)
	feeInt, _         = big.NewInt(0).SetString(fee, 10)
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
	GetBlockchainTimeStamp(ctx context.Context) (uint64, error)
}

type blockchainClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	ChainID(ctx context.Context) (*big.Int, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	FilterLogs(ctx context.Context, q goEthereum.FilterQuery) ([]types.Log, error)
}

type bridgeComponents interface {
	MultiversXRelayerAddress() sdkCore.AddressHandler
	EthereumRelayerAddress() common.Address
	Start() error
	Close() error
}

type keysHolder struct {
	pk         string
	sk         []byte
	ethSK      *ecdsa.PrivateKey
	ethAddress common.Address
}

func TestRelayersShouldExecuteTransfers(t *testing.T) {
	t.Run("ETH->MVX and back, ethNative = true, ethMintBurn = false, mvxNative = false, mvxMintBurn = true", func(t *testing.T) {
		args := argSimulatedSetup{
			mvxHexIsMintBurn: true,
			mvxHexIsNative:   false,
			ethIsMintBurn:    false,
			ethIsNative:      true,
		}
		testRelayersShouldExecuteTransfersEthToMVX(t, args)
	})
	t.Run("MVX->ETH, ethNative = false, ethMintBurn = true, mvxNative = true, mvxMintBurn = false", func(t *testing.T) {
		args := argSimulatedSetup{
			mvxHexIsMintBurn: false,
			mvxHexIsNative:   true,
			ethIsMintBurn:    true,
			ethIsNative:      false,
		}
		testRelayersShouldExecuteTransfersMVXToETH(t, args)
	})
}

func testRelayersShouldExecuteTransfersEthToMVX(t *testing.T, argsSimulatedSetup argSimulatedSetup) {
	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, "should have not panicked")
		}
	}()

	argsSimulatedSetup.t = t
	testSetup := prepareSimulatedSetup(argsSimulatedSetup)
	defer testSetup.close()

	testSetup.checkESDTBalance(testSetup.mvxReceiverAddress, testSetup.mvxUniversalToken, "0", true)

	testSetup.createBatch(batchProcessor.ToMultiversX)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	ethToMVXDone := false
	mvxToETHDone := false

	safeAddr, err := data.NewAddressFromBech32String(testSetup.mvxSafeAddress)
	require.NoError(t, err)

	// send half of the amount back to ETH
	valueToSendFromMVX := big.NewInt(0).Div(mintAmount, big.NewInt(2))
	initialSafeValue, err := testSetup.mvxChainSimulator.GetESDTBalance(testSetup.testContext, safeAddr, testSetup.mvxChainSpecificToken)
	require.NoError(t, err)
	initialSafeValueInt, _ := big.NewInt(0).SetString(initialSafeValue, 10)
	expectedFinalValueOnMVXSafe := initialSafeValueInt.Add(initialSafeValueInt, feeInt)
	expectedFinalValueOnETH := big.NewInt(0).Sub(valueToSendFromMVX, feeInt)
	for {
		select {
		case <-interrupt:
			require.Fail(t, "signal interrupted")
			return
		case <-time.After(timeout):
			require.Fail(t, "time out")
			return
		default:
			isTransferDoneFromETH := testSetup.checkESDTBalance(testSetup.mvxReceiverAddress, testSetup.mvxUniversalToken, mintAmount.String(), false)
			if !ethToMVXDone && isTransferDoneFromETH {
				ethToMVXDone = true
				log.Info("ETH->MVX transfer finished, now sending back to ETH...")

				testSetup.sendMVXToEthTransaction(valueToSendFromMVX.Bytes())
			}

			isTransferDoneFromMVX := testSetup.checkETHStatus(testSetup.ethOwnerAddress, expectedFinalValueOnETH.Uint64())
			safeSavedFee := testSetup.checkESDTBalance(safeAddr, testSetup.mvxChainSpecificToken, expectedFinalValueOnMVXSafe.String(), false)
			if !mvxToETHDone && isTransferDoneFromMVX && safeSavedFee {
				mvxToETHDone = true
			}

			if ethToMVXDone && mvxToETHDone {
				log.Info("MVX<->ETH transfers done")
				return
			}

			// commit blocks in order to execute incoming txs from relayers
			testSetup.simulatedETHChain.Commit()

			testSetup.mvxChainSimulator.GenerateBlocks(testSetup.testContext, 1)

		case <-interrupt:
			require.Fail(t, "signal interrupted")
			return
		case <-time.After(time.Minute * 15):
			require.Fail(t, "time out")
			return
		}
	}
}

func testRelayersShouldExecuteTransfersMVXToETH(t *testing.T, argsSimulatedSetup argSimulatedSetup) {
	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, "should have not panicked")
		}
	}()

	argsSimulatedSetup.t = t
	testSetup := prepareSimulatedSetup(argsSimulatedSetup)
	defer testSetup.close()

	testSetup.checkESDTBalance(testSetup.mvxReceiverAddress, testSetup.mvxUniversalToken, "0", true)

	safeAddr, err := data.NewAddressFromBech32String(testSetup.mvxSafeAddress)
	require.NoError(t, err)

	initialSafeValue, err := testSetup.mvxChainSimulator.GetESDTBalance(testSetup.testContext, safeAddr, testSetup.mvxChainSpecificToken)
	require.NoError(t, err)

	testSetup.createBatch(batchProcessor.FromMultiversX)

	// wait for signal interrupt or time out
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	// send half of the amount back to ETH
	valueSentFromETH := big.NewInt(0).Div(mintAmount, big.NewInt(2))
	initialSafeValueInt, _ := big.NewInt(0).SetString(initialSafeValue, 10)
	expectedFinalValueOnMVXSafe := initialSafeValueInt.Add(initialSafeValueInt, valueSentFromETH)
	expectedFinalValueOnETH := big.NewInt(0).Sub(valueSentFromETH, feeInt)
	expectedFinalValueOnETH = expectedFinalValueOnETH.Mul(expectedFinalValueOnETH, big.NewInt(1000000))
	for {
		select {
		case <-interrupt:
			require.Fail(t, "signal interrupted")
			return
		case <-time.After(time.Minute * 15):
			require.Fail(t, "time out")
			return
		default:
			isTransferDoneFromMVX := testSetup.checkETHStatus(testSetup.ethOwnerAddress, expectedFinalValueOnETH.Uint64())
			safeSavedFunds := testSetup.checkESDTBalance(safeAddr, testSetup.mvxChainSpecificToken, expectedFinalValueOnMVXSafe.String(), false)
			if isTransferDoneFromMVX && safeSavedFunds {
				log.Info("MVX->ETH transfer finished")

				return
			}

			// commit blocks in order to execute incoming txs from relayers
			testSetup.simulatedETHChain.Commit()

			testSetup.mvxChainSimulator.GenerateBlocks(testSetup.testContext, 1)
		}
	}
}

func TestRelayersShouldNotExecuteTransfers(t *testing.T) {
	t.Run("ETH->MVX, ethNative = true, ethMintBurn = false, mvxNative = true, mvxMintBurn = false", func(t *testing.T) {
		args := argSimulatedSetup{
			mvxHexIsMintBurn: false,
			mvxHexIsNative:   true,
			ethIsMintBurn:    false,
			ethIsNative:      true,
		}
		expectedStringInLogs := "error = invalid setup isNativeOnEthereum = true, isNativeOnMultiversX = true"
		testRelayersShouldNotExecuteTransfers(t, args, expectedStringInLogs, batchProcessor.ToMultiversX)
	})
	t.Run("ETH->MVX, ethNative = true, ethMintBurn = false, mvxNative = true, mvxMintBurn = true", func(t *testing.T) {
		args := argSimulatedSetup{
			mvxHexIsMintBurn: true,
			mvxHexIsNative:   true,
			ethIsMintBurn:    false,
			ethIsNative:      true,
		}
		expectedStringInLogs := "error = invalid setup isNativeOnEthereum = true, isNativeOnMultiversX = true"
		testRelayersShouldNotExecuteTransfers(t, args, expectedStringInLogs, batchProcessor.ToMultiversX)
	})
	t.Run("ETH->MVX, ethNative = true, ethMintBurn = true, mvxNative = true, mvxMintBurn = false", func(t *testing.T) {
		args := argSimulatedSetup{
			mvxHexIsMintBurn: false,
			mvxHexIsNative:   true,
			ethIsMintBurn:    true,
			ethIsNative:      true,
		}
		testEthContractsShouldError(t, args)
	})
	t.Run("ETH->MVX, ethNative = true, ethMintBurn = true, mvxNative = true, mvxMintBurn = true", func(t *testing.T) {
		args := argSimulatedSetup{
			mvxHexIsMintBurn: true,
			mvxHexIsNative:   true,
			ethIsMintBurn:    true,
			ethIsNative:      true,
		}
		testEthContractsShouldError(t, args)
	})
	t.Run("ETH->MVX, ethNative = false, ethMintBurn = true, mvxNative = false, mvxMintBurn = true", func(t *testing.T) {
		args := argSimulatedSetup{
			mvxHexIsMintBurn: true,
			mvxHexIsNative:   false,
			ethIsMintBurn:    true,
			ethIsNative:      false,
		}
		testEthContractsShouldError(t, args)
	})
	t.Run("MVX->ETH, ethNative = true, ethMintBurn = false, mvxNative = true, mvxMintBurn = false", func(t *testing.T) {
		args := argSimulatedSetup{
			mvxHexIsMintBurn: false,
			mvxHexIsNative:   true,
			ethIsMintBurn:    false,
			ethIsNative:      true,
		}
		expectedStringInLogs := "error = invalid setup isNativeOnEthereum = true, isNativeOnMultiversX = true"
		testRelayersShouldNotExecuteTransfers(t, args, expectedStringInLogs, batchProcessor.FromMultiversX)
	})
}

func testRelayersShouldNotExecuteTransfers(
	t *testing.T,
	argsSimulatedSetup argSimulatedSetup,
	expectedStringInLogs string,
	direction batchProcessor.Direction,
) {
	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, "should have not panicked")
		}
	}()

	argsSimulatedSetup.t = t
	testSetup := prepareSimulatedSetup(argsSimulatedSetup)
	defer testSetup.close()

	testSetup.checkESDTBalance(testSetup.mvxReceiverAddress, testSetup.mvxUniversalToken, "0", true)

	testSetup.createBatch(direction)

	// start a mocked log observer that is looking for a specific relayer error
	chanCnt := 0
	mockLogObserver := mock.NewMockLogObserver(expectedStringInLogs)
	err := logger.AddLogObserver(mockLogObserver, &logger.PlainFormatter{})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, logger.RemoveLogObserver(mockLogObserver))
	}()

	numOfTimesToRepeatErrorForRelayer := 10
	numOfErrorsToWait := numOfTimesToRepeatErrorForRelayer * numRelayers

	// wait for signal interrupt or time out
	roundDuration := time.Second
	roundTimer := time.NewTimer(roundDuration)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	for {
		roundTimer.Reset(roundDuration)
		select {
		case <-interrupt:
			require.Fail(t, "signal interrupted")
			return
		case <-time.After(time.Minute * 15):
			require.Fail(t, "time out")
			return
		case <-mockLogObserver.LogFoundChan():
			chanCnt++
			if chanCnt >= numOfErrorsToWait {
				testSetup.checkESDTBalance(testSetup.mvxReceiverAddress, testSetup.mvxUniversalToken, "0", true)

				log.Info(fmt.Sprintf("test passed, relayers are stuck, expected string `%s` found in all relayers' logs for %d times", expectedStringInLogs, numOfErrorsToWait))

				return
			}
		case <-roundTimer.C:
			// commit blocks
			testSetup.simulatedETHChain.Commit()

			testSetup.mvxChainSimulator.GenerateBlocks(testSetup.testContext, 1)
		}
	}
}

func testEthContractsShouldError(t *testing.T, argsSimulatedSetup argSimulatedSetup) {
	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, "should have not panicked")
		}
	}()

	testSetup := &simulatedSetup{}
	testSetup.t = t

	// create a test context
	testSetup.testContext, testSetup.testContextCancel = context.WithCancel(context.Background())

	testSetup.workingDir = t.TempDir()

	testSetup.getRelayersKeys()

	receiverKeys := generateMvxPrivatePublicKey(t)
	mvxReceiverAddress, err := data.NewAddressFromBech32String(receiverKeys.pk)
	require.NoError(t, err)

	testSetup.ethOwnerAddress = crypto.PubkeyToAddress(ethOwnerSK.PublicKey)
	ethDepositorAddr := crypto.PubkeyToAddress(ethDepositorSK.PublicKey)

	// create ethereum simulator
	testSetup.createEthereumSimulatorAndDeployContracts(ethDepositorAddr, argsSimulatedSetup.ethIsMintBurn, argsSimulatedSetup.ethIsNative)

	// add allowance for the sender
	auth, _ := bind.NewKeyedTransactorWithChainID(ethDepositorSK, testSetup.ethChainID)
	tx, err := testSetup.ethGenericTokenContract.Approve(auth, testSetup.ethSafeAddress, mintAmount)
	require.NoError(t, err)
	testSetup.simulatedETHChain.Commit()
	testSetup.checkEthTxResult(tx.Hash())

	// deposit on ETH safe should fail due to bad setup
	auth, _ = bind.NewKeyedTransactorWithChainID(ethDepositorSK, testSetup.ethChainID)
	_, err = testSetup.ethSafeContract.Deposit(auth, testSetup.ethGenericTokenAddress, mintAmount, mvxReceiverAddress.AddressSlice())
	require.Error(t, err)
}

type simulatedSetup struct {
	t                        *testing.T
	testContextCancel        func()
	testContext              context.Context
	simulatedETHChain        *backends.SimulatedBackend
	simulatedETHChainWrapper blockchainClient
	mvxChainSimulator        chainSimulatorWrapper
	relayers                 []bridgeComponents
	relayersKeys             []keysHolder
	mvxUniversalToken        string
	mvxChainSpecificToken    string
	mvxReceiverAddress       sdkCore.AddressHandler
	mvxReceiverKeys          keysHolder
	mvxSafeAddress           string
	mvxWrapperAddress        string
	mvxMultisigAddress       string
	mvxAggregatorAddress     string
	mvxOwnerKeys             keysHolder
	ethOwnerAddress          common.Address
	ethGenericTokenAddress   common.Address
	ethGenericTokenContract  *contract.GenericERC20
	ethChainID               *big.Int
	ethSafeAddress           common.Address
	ethSafeContract          *contract.ERC20Safe
	ethBridgeAddress         common.Address
	ethBridgeContract        *contract.Bridge
	workingDir               string
}

type argSimulatedSetup struct {
	t                *testing.T
	mvxHexIsMintBurn bool
	mvxHexIsNative   bool
	ethIsMintBurn    bool
	ethIsNative      bool
}

func prepareSimulatedSetup(args argSimulatedSetup) *simulatedSetup {
	var err error
	testSetup := &simulatedSetup{}
	testSetup.t = args.t

	// create a test context
	testSetup.testContext, testSetup.testContextCancel = context.WithCancel(context.Background())

	testSetup.workingDir = args.t.TempDir()

	testSetup.getRelayersKeys()

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
	testSetup.issueAndWhitelistToken(args.mvxHexIsMintBurn, args.mvxHexIsNative)

	// start relayers
	testSetup.startRelayers(ethChainWrapper, erc20ContractsHolder)

	return testSetup
}

func (testSetup *simulatedSetup) close() {
	testSetup.closeRelayers()

	require.NoError(testSetup.t, testSetup.simulatedETHChain.Close())

	testSetup.testContextCancel()
}

func (testSetup *simulatedSetup) closeRelayers() {
	for _, r := range testSetup.relayers {
		_ = r.Close()
	}
}

func (testSetup *simulatedSetup) getRelayersKeys() {
	relayersKeys := make([]keysHolder, 0, numRelayers)
	for i := 0; i < numRelayers; i++ {
		relayerKeys := generateMvxPrivatePublicKey(testSetup.t)
		log.Info("generated relayer", "index", i, "address", relayerKeys.pk)

		relayerETHSKBytes, err := os.ReadFile(fmt.Sprintf(relayerETHKeyPathFormat, i))
		require.Nil(testSetup.t, err)
		relayerKeys.ethSK, err = crypto.HexToECDSA(string(relayerETHSKBytes))
		require.Nil(testSetup.t, err)
		relayerKeys.ethAddress = crypto.PubkeyToAddress(relayerKeys.ethSK.PublicKey)

		relayersKeys = append(relayersKeys, relayerKeys)

		saveRelayerKey(testSetup.t, testSetup.workingDir, i, relayerKeys)
	}

	testSetup.relayersKeys = relayersKeys
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

func (testSetup *simulatedSetup) startChainSimulatorWrapper() {
	// create a new working directory
	tmpDir := path.Join(testSetup.t.TempDir(), "test")
	err := os.MkdirAll(tmpDir, os.ModePerm)
	require.NoError(testSetup.t, err)

	// start the chain simulator
	args := integrationTests.ArgChainSimulatorWrapper{
		ProxyCacherExpirationSeconds: proxyCacherExpirationSeconds,
		ProxyMaxNoncesDelta:          proxyMaxNoncesDelta,
	}
	testSetup.mvxChainSimulator, err = integrationTests.CreateChainSimulatorWrapper(args)
	require.NoError(testSetup.t, err)
}

func (testSetup *simulatedSetup) startRelayers(
	ethereumChain ethereum.ClientWrapper,
	erc20ContractsHolder ethereum.Erc20ContractsHolder,
) {
	relayers := make([]bridgeComponents, 0, numRelayers)

	messengers := integrationTests.CreateLinkedMessengers(numRelayers)

	for i := 0; i < numRelayers; i++ {
		generalConfigs := testsRelayers.CreateBridgeComponentsConfig(i, testSetup.workingDir)
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
		require.Nil(testSetup.t, err)

		go func() {
			err = relayer.Start()
			integrationTests.Log.LogIfError(err)
			require.Nil(testSetup.t, err)
		}()

		relayers = append(relayers, relayer)
	}

	testSetup.relayers = relayers
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
		getHexAddress(testSetup.t, testSetup.mvxOwnerKeys.pk),
	}

	testSetup.mvxAggregatorAddress, err = testSetup.mvxChainSimulator.DeploySC(
		testSetup.testContext,
		aggregatorContract,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		aggregatorDeployParams,
	)
	require.NoError(testSetup.t, err)
	require.NotEqual(testSetup.t, emptyAddress, testSetup.mvxAggregatorAddress)

	log.Info("aggregator contract deployed", "address", testSetup.mvxAggregatorAddress)

	// deploy wrapper
	testSetup.mvxWrapperAddress, err = testSetup.mvxChainSimulator.DeploySC(
		testSetup.testContext,
		wrapperContract,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		[]string{},
	)
	require.NoError(testSetup.t, err)
	require.NotEqual(testSetup.t, emptyAddress, testSetup.mvxWrapperAddress)

	log.Info("wrapper contract deployed", "address", testSetup.mvxWrapperAddress)

	// deploy safe
	testSetup.mvxSafeAddress, err = testSetup.mvxChainSimulator.DeploySC(
		testSetup.testContext,
		safeContract,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		[]string{getHexAddress(testSetup.t, testSetup.mvxAggregatorAddress), "01"},
	)
	require.NoError(testSetup.t, err)
	require.NotEqual(testSetup.t, emptyAddress, testSetup.mvxSafeAddress)

	log.Info("safe contract deployed", "address", testSetup.mvxSafeAddress)

	// deploy multi-transfer
	multiTransferAddress, err := testSetup.mvxChainSimulator.DeploySC(
		testSetup.testContext,
		multiTransferContract,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		[]string{},
	)
	require.NoError(testSetup.t, err)
	require.NotEqual(testSetup.t, emptyAddress, multiTransferAddress)

	log.Info("multi-transfer contract deployed", "address", multiTransferAddress)

	// deploy multisig
	minRelayerStakeInt, _ := big.NewInt(0).SetString(minRelayerStake, 10)
	minRelayerStakeHex := hex.EncodeToString(minRelayerStakeInt.Bytes())
	params := []string{getHexAddress(testSetup.t, testSetup.mvxSafeAddress), getHexAddress(testSetup.t, multiTransferAddress), minRelayerStakeHex, slashAmount, quorum}
	for _, relayerKeys := range testSetup.relayersKeys {
		params = append(params, getHexAddress(testSetup.t, relayerKeys.pk))
	}
	testSetup.mvxMultisigAddress, err = testSetup.mvxChainSimulator.DeploySC(
		testSetup.testContext,
		multisigContract,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		params,
	)
	require.NoError(testSetup.t, err)
	require.NotEqual(testSetup.t, emptyAddress, testSetup.mvxMultisigAddress)

	log.Info("multisig contract deployed", "address", testSetup.mvxMultisigAddress)

	// deploy bridge proxy
	bridgeProxyAddress, err := testSetup.mvxChainSimulator.DeploySC(
		testSetup.testContext,
		bridgeProxyContract,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		[]string{getHexAddress(testSetup.t, multiTransferAddress)},
	)
	require.NoError(testSetup.t, err)
	require.NotEqual(testSetup.t, emptyAddress, bridgeProxyAddress)

	log.Info("bridge proxy contract deployed", "address", bridgeProxyAddress)

	// setBridgeProxyContractAddress
	hash, err := testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		multiTransferAddress,
		zeroValue,
		setBridgeProxyContractAddress,
		[]string{getHexAddress(testSetup.t, bridgeProxyAddress)},
	)
	require.NoError(testSetup.t, err)
	txResult, err := testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

	log.Info("setBridgeProxyContractAddress tx executed", "hash", hash, "status", txResult.Status)

	// setWrappingContractAddress
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		multiTransferAddress,
		zeroValue,
		setWrappingContractAddress,
		[]string{getHexAddress(testSetup.t, testSetup.mvxWrapperAddress)},
	)
	require.NoError(testSetup.t, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

	log.Info("setWrappingContractAddress tx executed", "hash", hash, "status", txResult.Status)

	// ChangeOwnerAddress for safe
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		testSetup.mvxSafeAddress,
		zeroValue,
		changeOwnerAddress,
		[]string{getHexAddress(testSetup.t, testSetup.mvxMultisigAddress)},
	)
	require.NoError(testSetup.t, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

	log.Info("ChangeOwnerAddress for safe tx executed", "hash", hash, "status", txResult.Status)

	// ChangeOwnerAddress for multi-transfer
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		multiTransferAddress,
		zeroValue,
		changeOwnerAddress,
		[]string{getHexAddress(testSetup.t, testSetup.mvxMultisigAddress)},
	)
	require.NoError(testSetup.t, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

	log.Info("ChangeOwnerAddress for multi-transfer tx executed", "hash", hash, "status", txResult.Status)

	// ChangeOwnerAddress for bridge proxy
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		bridgeProxyAddress,
		zeroValue,
		changeOwnerAddress,
		[]string{getHexAddress(testSetup.t, testSetup.mvxMultisigAddress)},
	)
	require.NoError(testSetup.t, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

	log.Info("ChangeOwnerAddress for bridge proxy tx executed", "hash", hash, "status", txResult.Status)

	// setMultiTransferOnEsdtSafe
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		testSetup.mvxMultisigAddress,
		zeroValue,
		setMultiTransferOnEsdtSafe,
		[]string{},
	)
	require.NoError(testSetup.t, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

	log.Info("setMultiTransferOnEsdtSafe tx executed", "hash", hash, "status", txResult.Status)

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
	require.NoError(testSetup.t, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

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
	require.NoError(testSetup.t, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

	log.Info("setPairDecimals tx executed", "hash", hash, "status", txResult.Status)

	// stake relayers on multisig
	testSetup.stakeAddressesOnContract(testSetup.mvxMultisigAddress, testSetup.relayersKeys)

	// stake relayers on price aggregator
	testSetup.stakeAddressesOnContract(testSetup.mvxAggregatorAddress, []keysHolder{testSetup.mvxOwnerKeys})

	// unpause multisig
	hash = testSetup.unpauseContract(testSetup.mvxMultisigAddress, []byte(unpause))
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)
	log.Info("unpaused multisig executed", "hash", hash, "status", txResult.Status)

	// unpause safe
	hash = testSetup.unpauseContract(testSetup.mvxMultisigAddress, []byte(unpauseEsdtSafe))
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)
	log.Info("unpaused safe executed", "hash", hash, "status", txResult.Status)

	// unpause aggregator
	hash = testSetup.unpauseContract(testSetup.mvxAggregatorAddress, []byte(unpause))
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)
	log.Info("unpaused aggregator executed", "hash", hash, "status", txResult.Status)

	// unpause wrapper
	hash = testSetup.unpauseContract(testSetup.mvxWrapperAddress, []byte(unpause))
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)
	log.Info("unpaused wrapper executed", "hash", hash, "status", txResult.Status)
}

func (testSetup *simulatedSetup) stakeAddressesOnContract(contract string, allKeys []keysHolder) {
	for _, keys := range allKeys {
		hash, err := testSetup.mvxChainSimulator.SendTx(testSetup.testContext, keys.pk, keys.sk, contract, minRelayerStake, []byte("stake"))
		require.NoError(testSetup.t, err)
		txResult, err := testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
		require.NoError(testSetup.t, err)

		log.Info(fmt.Sprintf("address %s staked on contract %s with hash %s, status %s", keys.pk, contract, hash, txResult.Status))
	}
}

func (testSetup *simulatedSetup) unpauseContract(contract string, dataField []byte) string {
	hash, err := testSetup.mvxChainSimulator.SendTx(testSetup.testContext, testSetup.mvxOwnerKeys.pk, testSetup.mvxOwnerKeys.sk, contract, zeroValue, dataField)
	require.NoError(testSetup.t, err)

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
	require.NoError(testSetup.t, err)
	txResult, err := testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)
	testSetup.mvxUniversalToken = getTokenNameFromResult(testSetup.t, *txResult)

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
	require.NoError(testSetup.t, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)
	testSetup.mvxChainSpecificToken = getTokenNameFromResult(testSetup.t, *txResult)

	log.Info("issue chain specific token tx executed", "hash", hash, "status", txResult.Status, "token", testSetup.mvxChainSpecificToken, "owner", testSetup.mvxOwnerKeys.pk)

	// set local roles bridged tokens wrapper
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		esdtSystemSCAddress,
		zeroValue,
		setSpecialRole,
		[]string{hex.EncodeToString([]byte(testSetup.mvxUniversalToken)), getHexAddress(testSetup.t, testSetup.mvxWrapperAddress), hex.EncodeToString([]byte(esdtRoleLocalMint)), hex.EncodeToString([]byte(esdtRoleLocalBurn))})
	require.NoError(testSetup.t, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

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
	require.NoError(testSetup.t, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

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
	require.NoError(testSetup.t, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

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
	require.NoError(testSetup.t, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

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
	require.NoError(testSetup.t, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

	log.Info("wrapper whitelist token tx executed", "hash", hash, "status", txResult.Status)

	// set local roles esdt safe
	hash, err = testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		esdtSystemSCAddress,
		zeroValue,
		setSpecialRole,
		[]string{hex.EncodeToString([]byte(testSetup.mvxChainSpecificToken)), getHexAddress(testSetup.t, testSetup.mvxSafeAddress), hex.EncodeToString([]byte(esdtRoleLocalMint)), hex.EncodeToString([]byte(esdtRoleLocalBurn))})
	require.NoError(testSetup.t, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

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
	require.NoError(testSetup.t, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

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
	require.NoError(testSetup.t, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

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
	require.NoError(testSetup.t, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

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
	require.NoError(testSetup.t, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

	log.Info("multi-transfer set max bridge amount for token tx executed", "hash", hash, "status", txResult.Status)
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
	require.NoError(testSetup.t, err)

	if checkResult {
		require.Equal(testSetup.t, expectedBalance, balance)
	}

	return expectedBalance == balance
}

func (testSetup *simulatedSetup) checkETHStatus(receiver common.Address, expectedBalance uint64) bool {
	balance, err := testSetup.ethGenericTokenContract.BalanceOf(nil, receiver)
	require.NoError(testSetup.t, err)

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
	require.NoError(testSetup.t, err)
	txResult, err := testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

	log.Info("unwrap transaction sent", "hash", hash, "token", testSetup.mvxUniversalToken, "status", txResult.Status)

	// send tx to safe contract
	params := []string{
		hex.EncodeToString([]byte(testSetup.mvxChainSpecificToken)),
		hex.EncodeToString(value),
		hex.EncodeToString([]byte(createTransactionParam)),
		hex.EncodeToString(testSetup.ethOwnerAddress.Bytes()),
	}

	hash, err = testSetup.mvxChainSimulator.ScCall(testSetup.testContext, testSetup.mvxReceiverKeys.pk, testSetup.mvxReceiverKeys.sk, testSetup.mvxSafeAddress, zeroValue, esdtTransfer, params)
	require.NoError(testSetup.t, err)

	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

	log.Info("MVX->ETH transaction sent", "hash", hash, "status", txResult.Status)

	return hash
}

func (testSetup *simulatedSetup) submitAggregatorBatch() {
	timestamp, err := testSetup.mvxChainSimulator.GetBlockchainTimeStamp(testSetup.testContext)
	require.Nil(testSetup.t, err)
	require.Greater(testSetup.t, timestamp, uint64(0), "something went wrong and the chain simulator returned 0 for the current timestamp")

	timestampAsBigInt := big.NewInt(0).SetUint64(timestamp)

	hash, err := testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		testSetup.mvxAggregatorAddress,
		zeroValue,
		submitBatch,
		[]string{hex.EncodeToString([]byte(gwei)), hex.EncodeToString([]byte(chainSpecificTokenTicker)), hex.EncodeToString(timestampAsBigInt.Bytes()), hex.EncodeToString(feeInt.Bytes()), numOfDecimalsChainSpecific})
	require.NoError(testSetup.t, err)
	txResult, err := testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

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
	ethSafeAddress := testSetup.deployETHContract(erc20SafeABI, erc20SafeBytecode)
	testSetup.ethSafeAddress = ethSafeAddress
	ethSafeContract, err := contract.NewERC20Safe(ethSafeAddress, testSetup.simulatedETHChain)
	require.NoError(testSetup.t, err)
	testSetup.ethSafeContract = ethSafeContract

	// deploy bridge
	ethRelayersAddresses := make([]common.Address, 0, len(testSetup.relayersKeys))
	for _, relayerKeys := range testSetup.relayersKeys {
		ethRelayersAddresses = append(ethRelayersAddresses, relayerKeys.ethAddress)
	}
	quorumInt, _ := big.NewInt(0).SetString(quorum, 10)
	testSetup.ethBridgeAddress = testSetup.deployETHContract(bridgeABI, bridgeBytecode, ethRelayersAddresses, quorumInt, ethSafeAddress)
	testSetup.ethBridgeContract, err = contract.NewBridge(testSetup.ethBridgeAddress, testSetup.simulatedETHChain)
	require.NoError(testSetup.t, err)

	// set bridge on safe
	auth, _ := bind.NewKeyedTransactorWithChainID(ethOwnerSK, testSetup.ethChainID)
	tx, err := ethSafeContract.SetBridge(auth, testSetup.ethBridgeAddress)
	require.NoError(testSetup.t, err)
	testSetup.simulatedETHChain.Commit()
	testSetup.checkEthTxResult(tx.Hash())

	// deploy exec-proxy
	ethExecProxyAddress := testSetup.deployETHContract(scExecProxyABI, scExecProxyBytecode, ethSafeAddress)
	_, err = contract.NewSCExecProxy(ethExecProxyAddress, testSetup.simulatedETHChain)
	require.NoError(testSetup.t, err)

	// deploy generic eth token
	ethGenericTokenAddress := testSetup.deployETHContract(genericERC20ABI, genericERC20Bytecode, ethTokenName, ethTokenSymbol)
	testSetup.ethGenericTokenAddress = ethGenericTokenAddress
	ethGenericTokenContract, err := contract.NewGenericERC20(ethGenericTokenAddress, testSetup.simulatedETHChain)
	require.NoError(testSetup.t, err)
	testSetup.ethGenericTokenContract = ethGenericTokenContract

	// mint generic token
	auth, _ = bind.NewKeyedTransactorWithChainID(ethDepositorSK, testSetup.ethChainID)
	tx, err = ethGenericTokenContract.Mint(auth, ethDepositorAddr, mintAmount)
	require.NoError(testSetup.t, err)
	testSetup.simulatedETHChain.Commit()
	testSetup.checkEthTxResult(tx.Hash())
	testSetup.checkETHStatus(ethDepositorAddr, mintAmount.Uint64())

	// whitelist eth token
	auth, _ = bind.NewKeyedTransactorWithChainID(ethOwnerSK, testSetup.ethChainID)
	tx, err = ethSafeContract.WhitelistToken(auth, ethGenericTokenAddress, big.NewInt(ethMinAmountAllowedToTransfer), big.NewInt(ethMaxAmountAllowedToTransfer), isMintBurn, isNative)
	require.NoError(testSetup.t, err)
	testSetup.simulatedETHChain.Commit()
	testSetup.checkEthTxResult(tx.Hash())

	// unpause bridge contract
	auth, _ = bind.NewKeyedTransactorWithChainID(ethOwnerSK, testSetup.ethChainID)
	tx, err = testSetup.ethBridgeContract.Unpause(auth)
	require.NoError(testSetup.t, err)
	testSetup.simulatedETHChain.Commit()
	testSetup.checkEthTxResult(tx.Hash())

	// unpause safe contract
	auth, _ = bind.NewKeyedTransactorWithChainID(ethOwnerSK, testSetup.ethChainID)
	tx, err = ethSafeContract.Unpause(auth)
	require.NoError(testSetup.t, err)
	testSetup.simulatedETHChain.Commit()
	testSetup.checkEthTxResult(tx.Hash())
}

func (testSetup *simulatedSetup) deployETHContract(
	abiFile string,
	bytecodeFile string,
	params ...interface{},
) common.Address {
	abiBytes, err := os.ReadFile(abiFile)
	require.NoError(testSetup.t, err)
	parsed, err := abi.JSON(bytes.NewReader(abiBytes))
	require.NoError(testSetup.t, err)

	contractBytes, err := os.ReadFile(bytecodeFile)
	require.NoError(testSetup.t, err)

	contractAuth, _ := bind.NewKeyedTransactorWithChainID(ethOwnerSK, testSetup.ethChainID)
	contractAddress, tx, _, err := bind.DeployContract(contractAuth, parsed, common.FromHex(converters.TrimWhiteSpaceCharacters(string(contractBytes))), testSetup.simulatedETHChain, params...)
	require.NoError(testSetup.t, err)
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
	tx, err := testSetup.ethGenericTokenContract.Approve(auth, testSetup.ethSafeAddress, mintAmount)
	require.NoError(testSetup.t, err)
	testSetup.simulatedETHChain.Commit()
	testSetup.checkEthTxResult(tx.Hash())

	// deposit on ETH safe
	auth, _ = bind.NewKeyedTransactorWithChainID(ethDepositorSK, testSetup.ethChainID)
	tx, err = testSetup.ethSafeContract.Deposit(auth, testSetup.ethGenericTokenAddress, mintAmount, testSetup.mvxReceiverAddress.AddressSlice())
	require.NoError(testSetup.t, err)
	testSetup.simulatedETHChain.Commit()
	testSetup.checkEthTxResult(tx.Hash())

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
	require.NoError(testSetup.t, err)
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
	require.NoError(testSetup.t, err)
	txResult, err := testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

	log.Info("transfer to sender tx executed", "hash", hash, "status", txResult.Status)

	// send tx to safe contract
	params := []string{
		hex.EncodeToString([]byte(testSetup.mvxChainSpecificToken)),
		hex.EncodeToString(valueToSendFromMVX.Bytes()),
		hex.EncodeToString([]byte(createTransactionParam)),
		hex.EncodeToString(testSetup.ethOwnerAddress.Bytes()),
	}
	hash, err = testSetup.mvxChainSimulator.ScCall(testSetup.testContext, testSetup.mvxReceiverKeys.pk, testSetup.mvxReceiverKeys.sk, testSetup.mvxSafeAddress, zeroValue, esdtTransfer, params)
	require.NoError(testSetup.t, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)

	log.Info("MVX->ETH transaction sent", "hash", hash, "status", txResult.Status)
}

func (testSetup *simulatedSetup) checkEthTxResult(hash common.Hash) {
	receipt, err := testSetup.simulatedETHChain.TransactionReceipt(testSetup.testContext, hash)
	require.NoError(testSetup.t, err)
	require.Equal(testSetup.t, ethStatusSuccess, receipt.Status)
}
