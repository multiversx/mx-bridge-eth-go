//go:build slow

// To run these slow tests, simply add the slow tag on the go test command. Also, provide a chain simulator instance on the 8085 port
// example: go test -tags slow

package relayers

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
}

type blockchainClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	ChainID(ctx context.Context) (*big.Int, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	FilterLogs(ctx context.Context, q goEthereum.FilterQuery) ([]types.Log, error)
}

type keysHolder struct {
	pk         string
	sk         []byte
	ethSK      *ecdsa.PrivateKey
	ethAddress common.Address
}

func TestRelayersShouldExecuteTransfers(t *testing.T) {
	t.Run("ETH->MVX and back, ethNative = true, ethMintBurn = false, mvxNative = false, mvxMintBurn = true", testRelayersShouldExecuteTransfersEthToMVX(
		argSimulatedSetup{
			mvxHexIsMintBurn: hexTrue,
			mvxHexIsNative:   hexFalse,
			ethIsMintBurn:    false,
			ethIsNative:      true,
		},
	))
	t.Run("MVX->ETH, ethNative = false, ethMintBurn = true, mvxNative = true, mvxMintBurn = false", testRelayersShouldExecuteTransfersMVXToETH(
		argSimulatedSetup{
			mvxHexIsMintBurn: hexFalse,
			mvxHexIsNative:   hexTrue,
			ethIsMintBurn:    true,
			ethIsNative:      false,
		},
	))
}

func testRelayersShouldExecuteTransfersEthToMVX(argsSimulatedSetup argSimulatedSetup) func(t *testing.T) {
	return func(t *testing.T) {
		defer func() {
			r := recover()
			if r != nil {
				require.Fail(t, "should have not panicked")
			}
		}()

		argsSimulatedSetup.t = t
		testSetup := prepareSimulatedSetup(argsSimulatedSetup)
		defer testSetup.close(t)

		checkESDTBalance(t, testSetup.testContext, testSetup.mvxChainSimulator, testSetup.mvxReceiverAddress, testSetup.mvxUniversalToken, "0", true)

		createBatch(t, testSetup, batchProcessor.ToMultiversX)

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
			case <-time.After(time.Minute * 15):
				require.Fail(t, "time out")
				return
			default:
				isTransferDoneFromETH := checkESDTBalance(t, testSetup.testContext, testSetup.mvxChainSimulator, testSetup.mvxReceiverAddress, testSetup.mvxUniversalToken, mintAmount.String(), false)
				if !ethToMVXDone && isTransferDoneFromETH {
					ethToMVXDone = true
					log.Info("ETH->MVX transfer finished, now sending back to ETH...")

					sendMVXToEthTransaction(
						t,
						testSetup.testContext,
						testSetup.mvxChainSimulator,
						valueToSendFromMVX.Bytes(),
						testSetup.mvxUniversalToken,
						testSetup.mvxChainSpecificToken,
						testSetup.mvxReceiverKeys,
						testSetup.mvxSafeAddress,
						testSetup.mvxWrapperAddress,
						testSetup.ethOwnerAddress.Bytes(),
					)
				}

				isTransferDoneFromMVX := checkETHStatus(t, testSetup.ethGenericTokenContract, testSetup.ethOwnerAddress, expectedFinalValueOnETH.Uint64())
				safeSavedFee := checkESDTBalance(t, testSetup.testContext, testSetup.mvxChainSimulator, safeAddr, testSetup.mvxChainSpecificToken, expectedFinalValueOnMVXSafe.String(), false)
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
}

func testRelayersShouldExecuteTransfersMVXToETH(argsSimulatedSetup argSimulatedSetup) func(t *testing.T) {
	return func(t *testing.T) {
		defer func() {
			r := recover()
			if r != nil {
				require.Fail(t, "should have not panicked")
			}
		}()

		argsSimulatedSetup.t = t
		testSetup := prepareSimulatedSetup(argsSimulatedSetup)
		defer testSetup.close(t)

		checkESDTBalance(t, testSetup.testContext, testSetup.mvxChainSimulator, testSetup.mvxReceiverAddress, testSetup.mvxUniversalToken, "0", true)

		createBatch(t, testSetup, batchProcessor.FromMultiversX)

		// wait for signal interrupt or time out
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

		safeAddr, err := data.NewAddressFromBech32String(testSetup.mvxSafeAddress)
		require.NoError(t, err)

		// send half of the amount back to ETH
		valueSentFromETH := big.NewInt(0).Div(mintAmount, big.NewInt(2))
		initialSafeValue, err := testSetup.mvxChainSimulator.GetESDTBalance(testSetup.testContext, safeAddr, testSetup.mvxChainSpecificToken)
		require.NoError(t, err)
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
				isTransferDoneFromMVX := checkETHStatus(t, testSetup.ethGenericTokenContract, testSetup.ethOwnerAddress, expectedFinalValueOnETH.Uint64())
				safeSavedFunds := checkESDTBalance(t, testSetup.testContext, testSetup.mvxChainSimulator, safeAddr, testSetup.mvxChainSpecificToken, expectedFinalValueOnMVXSafe.String(), false)
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
}

func TestRelayersShouldNotExecuteTransfers(t *testing.T) {
	t.Run("ETH->MVX, ethNative = true, ethMintBurn = false, mvxNative = true, mvxMintBurn = false", testRelayersShouldNotExecuteTransfers(
		argSimulatedSetup{
			mvxHexIsMintBurn: hexFalse,
			mvxHexIsNative:   hexTrue,
			ethIsMintBurn:    false,
			ethIsNative:      true,
		},
		"error = invalid setup isNativeOnEthereum = true, isNativeOnMultiversX = true",
		batchProcessor.ToMultiversX,
	))
	t.Run("ETH->MVX, ethNative = true, ethMintBurn = false, mvxNative = true, mvxMintBurn = true", testRelayersShouldNotExecuteTransfers(
		argSimulatedSetup{
			mvxHexIsMintBurn: hexTrue,
			mvxHexIsNative:   hexTrue,
			ethIsMintBurn:    false,
			ethIsNative:      true,
		},
		"error = invalid setup isNativeOnEthereum = true, isNativeOnMultiversX = true",
		batchProcessor.ToMultiversX,
	))
	t.Run("ETH->MVX, ethNative = true, ethMintBurn = true, mvxNative = true, mvxMintBurn = false", testEthContractsShouldError(
		argSimulatedSetup{
			mvxHexIsMintBurn: hexFalse,
			mvxHexIsNative:   hexTrue,
			ethIsMintBurn:    true,
			ethIsNative:      true,
		},
	))
	t.Run("ETH->MVX, ethNative = true, ethMintBurn = true, mvxNative = true, mvxMintBurn = true", testEthContractsShouldError(
		argSimulatedSetup{
			mvxHexIsMintBurn: hexTrue,
			mvxHexIsNative:   hexTrue,
			ethIsMintBurn:    true,
			ethIsNative:      true,
		},
	))
	t.Run("ETH->MVX, ethNative = false, ethMintBurn = true, mvxNative = false, mvxMintBurn = true", testEthContractsShouldError(
		argSimulatedSetup{
			mvxHexIsMintBurn: hexTrue,
			mvxHexIsNative:   hexFalse,
			ethIsMintBurn:    true,
			ethIsNative:      false,
		},
	))
	t.Run("MVX->ETH, ethNative = true, ethMintBurn = false, mvxNative = true, mvxMintBurn = false", testRelayersShouldNotExecuteTransfers(
		argSimulatedSetup{
			mvxHexIsMintBurn: hexFalse,
			mvxHexIsNative:   hexTrue,
			ethIsMintBurn:    false,
			ethIsNative:      true,
		},
		"error = invalid setup isNativeOnEthereum = true, isNativeOnMultiversX = true",
		batchProcessor.FromMultiversX,
	))
}

func testRelayersShouldNotExecuteTransfers(
	argsSimulatedSetup argSimulatedSetup,
	expectedStringInLogs string,
	direction batchProcessor.Direction,
) func(t *testing.T) {
	return func(t *testing.T) {
		defer func() {
			r := recover()
			if r != nil {
				require.Fail(t, "should have not panicked")
			}
		}()

		argsSimulatedSetup.t = t
		testSetup := prepareSimulatedSetup(argsSimulatedSetup)
		defer testSetup.close(t)

		checkESDTBalance(t, testSetup.testContext, testSetup.mvxChainSimulator, testSetup.mvxReceiverAddress, testSetup.mvxUniversalToken, "0", true)

		createBatch(t, testSetup, direction)

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
					checkESDTBalance(t, testSetup.testContext, testSetup.mvxChainSimulator, testSetup.mvxReceiverAddress, testSetup.mvxUniversalToken, "0", true)

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
}

func testEthContractsShouldError(argsSimulatedSetup argSimulatedSetup) func(t *testing.T) {
	return func(t *testing.T) {
		defer func() {
			r := recover()
			if r != nil {
				require.Fail(t, "should have not panicked")
			}
		}()

		// create a test context
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		tempDir := t.TempDir()

		relayersKeys := getRelayersKeys(t, tempDir)

		receiverKeys := generateMvxPrivatePublicKey(t)
		mvxReceiverAddress, err := data.NewAddressFromBech32String(receiverKeys.pk)
		require.NoError(t, err)

		ethOwnerAddr := crypto.PubkeyToAddress(ethOwnerSK.PublicKey)
		ethDepositorAddr := crypto.PubkeyToAddress(ethDepositorSK.PublicKey)

		// create ethereum simulator
		simulatedETHChain, simulatedETHChainWrapper, ethSafeContract, ethSafeAddress, _, _, ethGenericTokenContract, ethGenericTokenAddress := createEthereumSimulatorAndDeployContracts(t, ctx, relayersKeys, ethOwnerAddr, ethDepositorAddr, argsSimulatedSetup.ethIsMintBurn, argsSimulatedSetup.ethIsNative)

		ethChainID, _ := simulatedETHChainWrapper.ChainID(ctx)

		// add allowance for the sender
		auth, _ := bind.NewKeyedTransactorWithChainID(ethDepositorSK, ethChainID)
		tx, err := ethGenericTokenContract.Approve(auth, ethSafeAddress, mintAmount)
		require.NoError(t, err)
		simulatedETHChain.Commit()
		checkEthTxResult(t, ctx, simulatedETHChain, tx.Hash())

		// deposit on ETH safe should fail due to bad setup
		auth, _ = bind.NewKeyedTransactorWithChainID(ethDepositorSK, ethChainID)
		_, err = ethSafeContract.Deposit(auth, ethGenericTokenAddress, mintAmount, mvxReceiverAddress.AddressSlice())
		require.Error(t, err)
	}
}

type simulatedSetup struct {
	testContextCancel       func()
	testContext             context.Context
	simulatedETHChain       *backends.SimulatedBackend
	mvxChainSimulator       chainSimulatorWrapper
	relayers                []bridgeComponents
	mvxUniversalToken       string
	mvxChainSpecificToken   string
	mvxReceiverAddress      sdkCore.AddressHandler
	mvxReceiverKeys         keysHolder
	mvxSafeAddress          string
	mvxWrapperAddress       string
	mvxOwnerKeys            keysHolder
	ethOwnerAddress         common.Address
	ethGenericTokenAddress  common.Address
	ethGenericTokenContract *contract.GenericERC20
	ethChainID              *big.Int
	ethSafeAddress          common.Address
	ethSafeContract         *contract.ERC20Safe
	workingDir              string
}

type argSimulatedSetup struct {
	t                *testing.T
	mvxHexIsMintBurn string
	mvxHexIsNative   string
	ethIsMintBurn    bool
	ethIsNative      bool
}

func prepareSimulatedSetup(args argSimulatedSetup) *simulatedSetup {
	testSetup := &simulatedSetup{}

	// create a test context
	ctx, cancel := context.WithCancel(context.Background())
	testSetup.testContextCancel = cancel
	testSetup.testContext = ctx

	testSetup.workingDir = args.t.TempDir()

	relayersKeys := getRelayersKeys(args.t, testSetup.workingDir)

	// generate the mvx receiver keys
	receiverKeys := generateMvxPrivatePublicKey(args.t)
	receiverAddress, err := data.NewAddressFromBech32String(receiverKeys.pk)
	require.NoError(args.t, err)

	testSetup.mvxReceiverKeys = receiverKeys
	testSetup.mvxReceiverAddress = receiverAddress

	ethOwnerAddr := crypto.PubkeyToAddress(ethOwnerSK.PublicKey)
	testSetup.ethOwnerAddress = ethOwnerAddr
	ethDepositorAddr := crypto.PubkeyToAddress(ethDepositorSK.PublicKey)

	// create ethereum simulator
	simulatedETHChain, simulatedETHChainWrapper, ethSafeContract, ethSafeAddress, ethBridgeContract, _, ethGenericTokenContract, ethGenericTokenAddress := createEthereumSimulatorAndDeployContracts(args.t, ctx, relayersKeys, ethOwnerAddr, ethDepositorAddr, args.ethIsMintBurn, args.ethIsNative)
	testSetup.simulatedETHChain = simulatedETHChain
	testSetup.ethSafeAddress = ethSafeAddress
	testSetup.ethSafeContract = ethSafeContract
	testSetup.ethGenericTokenAddress = ethGenericTokenAddress
	testSetup.ethGenericTokenContract = ethGenericTokenContract

	testSetup.ethChainID, _ = simulatedETHChainWrapper.ChainID(ctx)

	// generate the mvx owner keys
	ownerKeys := generateMvxPrivatePublicKey(args.t)
	testSetup.mvxOwnerKeys = ownerKeys

	erc20ContractsHolder, err := ethereum.NewErc20SafeContractsHolder(ethereum.ArgsErc20SafeContractsHolder{
		EthClient:              simulatedETHChain,
		EthClientStatusHandler: &testsCommon.StatusHandlerStub{},
	})
	require.NoError(args.t, err)

	ethChainWrapper, err := wrappers.NewEthereumChainWrapper(wrappers.ArgsEthereumChainWrapper{
		StatusHandler:    &testsCommon.StatusHandlerStub{},
		MultiSigContract: ethBridgeContract,
		SafeContract:     ethSafeContract,
		BlockchainClient: simulatedETHChainWrapper,
	})
	require.NoError(args.t, err)

	mvxChainSimulator := startChainSimulatorWrapper(args.t)
	mvxChainSimulator.GenerateBlocksUntilEpochReached(testSetup.testContext, 1)
	testSetup.mvxChainSimulator = mvxChainSimulator

	// deploy all contracts and execute all txs needed
	safeAddress, multisigAddress, wrapperAddress, aggregatorAddress := executeContractsTxs(args.t, ctx, mvxChainSimulator, relayersKeys, ownerKeys, receiverKeys)
	testSetup.mvxSafeAddress = safeAddress
	testSetup.mvxWrapperAddress = wrapperAddress

	// issue and whitelist token
	newUniversalToken, newChainSpecificToken := issueAndWhitelistToken(args.t, ctx, mvxChainSimulator, ownerKeys, wrapperAddress, safeAddress, multisigAddress, aggregatorAddress, hex.EncodeToString(ethGenericTokenAddress.Bytes()), args.mvxHexIsMintBurn, args.mvxHexIsNative)
	testSetup.mvxUniversalToken = newUniversalToken
	testSetup.mvxChainSpecificToken = newChainSpecificToken

	// start relayers
	relayers := startRelayers(args.t, testSetup.workingDir, mvxChainSimulator, ethChainWrapper, ethSafeAddress, erc20ContractsHolder, safeAddress, multisigAddress)
	testSetup.relayers = relayers

	return testSetup
}

func (testSetup *simulatedSetup) close(t *testing.T) {
	closeRelayers(testSetup.relayers)

	require.NoError(t, testSetup.simulatedETHChain.Close())

	testSetup.testContextCancel()
}

func getRelayersKeys(t *testing.T, workingDir string) []keysHolder {
	relayersKeys := make([]keysHolder, 0, numRelayers)
	for i := 0; i < numRelayers; i++ {
		relayerKeys := generateMvxPrivatePublicKey(t)
		log.Info("generated relayer", "index", i, "address", relayerKeys.pk)

		relayerETHSKBytes, err := os.ReadFile(fmt.Sprintf(relayerETHKeyPathFormat, i))
		require.Nil(t, err)
		relayerKeys.ethSK, err = crypto.HexToECDSA(string(relayerETHSKBytes))
		require.Nil(t, err)
		relayerKeys.ethAddress = crypto.PubkeyToAddress(relayerKeys.ethSK.PublicKey)

		relayersKeys = append(relayersKeys, relayerKeys)

		saveRelayerKey(t, workingDir, i, relayerKeys)
	}

	return relayersKeys
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
	mvxChainSimulator chainSimulatorWrapper,
	ethereumChain ethereum.ClientWrapper,
	safeContractEthAddress common.Address,
	erc20ContractsHolder ethereum.Erc20ContractsHolder,
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
			ClientWrapper:                 ethereumChain,
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
	hexIsMintBurn string,
	hexIsNative string,
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

	// transfer to wrapper sc
	initialMintValue := valueToMintInt.Div(valueToMintInt, big.NewInt(3))
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		wrapperAddress,
		zeroValue,
		esdtTransfer,
		[]string{hex.EncodeToString([]byte(newChainSpecificToken)), hex.EncodeToString(initialMintValue.Bytes()), hex.EncodeToString([]byte(depositLiquidity))})
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("transfer to wrapper sc tx executed", "hash", hash, "status", txResult.Status)

	// transfer to safe sc
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		safeAddress,
		zeroValue,
		esdtTransfer,
		[]string{hex.EncodeToString([]byte(newChainSpecificToken)), hex.EncodeToString(initialMintValue.Bytes())})
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("transfer to safe sc tx executed", "hash", hash, "status", txResult.Status)

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
	statusBefore, _ := mvxChainSimulator.Proxy().GetNetworkStatus(ctx, 0)
	hash, err = mvxChainSimulator.ScCall(
		ctx,
		ownerKeys.pk,
		ownerKeys.sk,
		multisigAddress,
		zeroValue,
		esdtSafeAddTokenToWhitelist,
		[]string{hex.EncodeToString([]byte(newChainSpecificToken)), hex.EncodeToString([]byte(chainSpecificTokenTicker)), hexIsMintBurn, hexIsNative})
	require.NoError(t, err)
	txResult, err = mvxChainSimulator.GetTransactionResult(ctx, hash)
	require.NoError(t, err)

	log.Info("whitelist token tx executed", "hash", hash, "status", txResult.Status)

	// submit aggregator batch
	statusAfter, _ := mvxChainSimulator.Proxy().GetNetworkStatus(ctx, 0)
	networkConfig, err := mvxChainSimulator.Proxy().GetNetworkConfig(ctx)
	require.NoError(t, err)
	roundDurationInSeconds := networkConfig.RoundDuration / 1000
	timePassed := (statusAfter.CurrentRound - statusBefore.CurrentRound - 1) * uint64(roundDurationInSeconds)
	currentChainTimestamp := txResult.Timestamp + timePassed
	submitAggregatorBatch(t, ctx, mvxChainSimulator, aggregatorAddress, ownerKeys, currentChainTimestamp)

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

func checkETHStatus(t *testing.T, ethGenericTokenContract *contract.GenericERC20, receiver common.Address, expectedBalance uint64) bool {
	balance, err := ethGenericTokenContract.BalanceOf(nil, receiver)
	require.NoError(t, err)

	return balance.Uint64() == expectedBalance
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

func submitAggregatorBatch(
	t *testing.T,
	ctx context.Context,
	mvxChainSimulator chainSimulatorWrapper,
	aggregatorAddress string,
	ownerKeys keysHolder,
	currentChainTimestamp uint64,
) {
	timestamp := big.NewInt(0).SetUint64(currentChainTimestamp)

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

func createEthereumSimulatorAndDeployContracts(
	t *testing.T,
	ctx context.Context,
	relayersKeys []keysHolder,
	ethOwnerAddr common.Address,
	ethDepositorAddr common.Address,
	isMintBurn bool,
	isNative bool,
) (*backends.SimulatedBackend, blockchainClient, *contract.ERC20Safe, common.Address, *contract.Bridge, common.Address, *contract.GenericERC20, common.Address) {
	addr := map[common.Address]ethCore.GenesisAccount{
		ethOwnerAddr:     {Balance: new(big.Int).Lsh(big.NewInt(1), 100)},
		ethDepositorAddr: {Balance: new(big.Int).Lsh(big.NewInt(1), 100)},
	}
	for _, relayerKeys := range relayersKeys {
		addr[relayerKeys.ethAddress] = ethCore.GenesisAccount{Balance: new(big.Int).Lsh(big.NewInt(1), 100)}
	}
	alloc := ethCore.GenesisAlloc(addr)
	simulatedETHChain := backends.NewSimulatedBackend(alloc, ethSimulatedGasLimit)

	simulatedETHChainWrapper := integrationTests.NewSimulatedETHChainWrapper(simulatedETHChain)
	ethChainID, _ := simulatedETHChainWrapper.ChainID(ctx)

	// deploy safe
	ethSafeAddress := deployETHContract(t, ctx, simulatedETHChain, ethChainID, erc20SafeABI, erc20SafeBytecode)
	ethSafeContract, err := contract.NewERC20Safe(ethSafeAddress, simulatedETHChain)
	require.NoError(t, err)

	// deploy bridge
	ethRelayersAddresses := make([]common.Address, 0, len(relayersKeys))
	for _, relayerKeys := range relayersKeys {
		ethRelayersAddresses = append(ethRelayersAddresses, relayerKeys.ethAddress)
	}
	quorumInt, _ := big.NewInt(0).SetString(quorum, 10)
	ethBridgeAddress := deployETHContract(t, ctx, simulatedETHChain, ethChainID, bridgeABI, bridgeBytecode, ethRelayersAddresses, quorumInt, ethSafeAddress)
	ethBridgeContract, err := contract.NewBridge(ethBridgeAddress, simulatedETHChain)
	require.NoError(t, err)

	// set bridge on safe
	auth, _ := bind.NewKeyedTransactorWithChainID(ethOwnerSK, ethChainID)
	tx, err := ethSafeContract.SetBridge(auth, ethBridgeAddress)
	require.NoError(t, err)
	simulatedETHChain.Commit()
	checkEthTxResult(t, ctx, simulatedETHChain, tx.Hash())

	// deploy exec-proxy
	ethExecProxyAddress := deployETHContract(t, ctx, simulatedETHChain, ethChainID, scExecProxyABI, scExecProxyBytecode, ethSafeAddress)
	_, err = contract.NewSCExecProxy(ethExecProxyAddress, simulatedETHChain)
	require.NoError(t, err)

	// deploy generic eth token
	ethGenericTokenAddress := deployETHContract(t, ctx, simulatedETHChain, ethChainID, genericERC20ABI, genericERC20Bytecode, ethTokenName, ethTokenSymbol)
	ethGenericTokenContract, err := contract.NewGenericERC20(ethGenericTokenAddress, simulatedETHChain)
	require.NoError(t, err)

	// mint generic token
	auth, _ = bind.NewKeyedTransactorWithChainID(ethDepositorSK, ethChainID)
	tx, err = ethGenericTokenContract.Mint(auth, ethDepositorAddr, mintAmount)
	require.NoError(t, err)
	simulatedETHChain.Commit()
	checkEthTxResult(t, ctx, simulatedETHChain, tx.Hash())
	checkETHStatus(t, ethGenericTokenContract, ethDepositorAddr, mintAmount.Uint64())

	// whitelist eth token
	auth, _ = bind.NewKeyedTransactorWithChainID(ethOwnerSK, ethChainID)
	tx, err = ethSafeContract.WhitelistToken(auth, ethGenericTokenAddress, big.NewInt(ethMinAmountAllowedToTransfer), big.NewInt(ethMaxAmountAllowedToTransfer), isMintBurn, isNative)
	require.NoError(t, err)
	simulatedETHChain.Commit()
	checkEthTxResult(t, ctx, simulatedETHChain, tx.Hash())

	// unpause bridge contract
	auth, _ = bind.NewKeyedTransactorWithChainID(ethOwnerSK, ethChainID)
	tx, err = ethBridgeContract.Unpause(auth)
	require.NoError(t, err)
	simulatedETHChain.Commit()
	checkEthTxResult(t, ctx, simulatedETHChain, tx.Hash())

	// unpause safe contract
	auth, _ = bind.NewKeyedTransactorWithChainID(ethOwnerSK, ethChainID)
	tx, err = ethSafeContract.Unpause(auth)
	require.NoError(t, err)
	simulatedETHChain.Commit()
	checkEthTxResult(t, ctx, simulatedETHChain, tx.Hash())

	return simulatedETHChain, simulatedETHChainWrapper, ethSafeContract, ethSafeAddress, ethBridgeContract, ethBridgeAddress, ethGenericTokenContract, ethGenericTokenAddress
}

func deployETHContract(
	t *testing.T,
	ctx context.Context,
	simulatedETHChain *backends.SimulatedBackend,
	chainID *big.Int,
	abiFile string,
	bytecodeFile string,
	params ...interface{},
) common.Address {
	abiBytes, err := os.ReadFile(abiFile)
	require.NoError(t, err)
	parsed, err := abi.JSON(bytes.NewReader(abiBytes))
	require.NoError(t, err)

	contractBytes, err := os.ReadFile(bytecodeFile)
	require.NoError(t, err)

	contractAuth, _ := bind.NewKeyedTransactorWithChainID(ethOwnerSK, chainID)
	contractAddress, tx, _, err := bind.DeployContract(contractAuth, parsed, common.FromHex(converters.TrimWhiteSpaceCharacters(string(contractBytes))), simulatedETHChain, params...)
	require.NoError(t, err)
	simulatedETHChain.Commit()

	checkEthTxResult(t, ctx, simulatedETHChain, tx.Hash())

	log.Info("deployed eth contract", "from file", bytecodeFile, "address", contractAddress.Hex())

	return contractAddress
}

func createBatch(
	t *testing.T,
	testSetup *simulatedSetup,
	direction batchProcessor.Direction,
) {
	if direction == batchProcessor.ToMultiversX {
		// create a pending batch on ethereum
		createBatchOnEthereum(t, testSetup)
	}
	if direction == batchProcessor.FromMultiversX {
		// create a pending batch on multiversx
		createBatchOnMultiversX(t, testSetup)
	}
}

func createBatchOnEthereum(t *testing.T, testSetup *simulatedSetup) {
	// add allowance for the sender
	auth, _ := bind.NewKeyedTransactorWithChainID(ethDepositorSK, testSetup.ethChainID)
	tx, err := testSetup.ethGenericTokenContract.Approve(auth, testSetup.ethSafeAddress, mintAmount)
	require.NoError(t, err)
	testSetup.simulatedETHChain.Commit()
	checkEthTxResult(t, testSetup.testContext, testSetup.simulatedETHChain, tx.Hash())

	// deposit on ETH safe
	auth, _ = bind.NewKeyedTransactorWithChainID(ethDepositorSK, testSetup.ethChainID)
	tx, err = testSetup.ethSafeContract.Deposit(auth, testSetup.ethGenericTokenAddress, mintAmount, testSetup.mvxReceiverAddress.AddressSlice())
	require.NoError(t, err)
	testSetup.simulatedETHChain.Commit()
	checkEthTxResult(t, testSetup.testContext, testSetup.simulatedETHChain, tx.Hash())

	// wait until batch is settled
	batchSettleLimit, _ := testSetup.ethSafeContract.BatchSettleLimit(nil)
	for i := uint8(0); i < batchSettleLimit+1; i++ {
		testSetup.simulatedETHChain.Commit()
	}
}

func createBatchOnMultiversX(t *testing.T, testSetup *simulatedSetup) {
	// create a pending batch on multiversx
	valueToSendFromMVX := big.NewInt(0).Div(mintAmount, big.NewInt(2))

	// mint erc20 token into eth safe
	auth, _ := bind.NewKeyedTransactorWithChainID(ethDepositorSK, testSetup.ethChainID)
	tx, err := testSetup.ethGenericTokenContract.Mint(auth, testSetup.ethSafeAddress, valueToSendFromMVX)
	require.NoError(t, err)
	testSetup.simulatedETHChain.Commit()
	checkEthTxResult(t, testSetup.testContext, testSetup.simulatedETHChain, tx.Hash())
	checkETHStatus(t, testSetup.ethGenericTokenContract, testSetup.ethSafeAddress, mintAmount.Uint64())

	// transfer to sender tx
	hash, err := testSetup.mvxChainSimulator.ScCall(
		testSetup.testContext,
		testSetup.mvxOwnerKeys.pk,
		testSetup.mvxOwnerKeys.sk,
		testSetup.mvxReceiverKeys.pk,
		zeroValue,
		esdtTransfer,
		[]string{hex.EncodeToString([]byte(testSetup.mvxChainSpecificToken)), hex.EncodeToString(valueToSendFromMVX.Bytes())})
	require.NoError(t, err)
	txResult, err := testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(t, err)

	log.Info("transfer to sender tx executed", "hash", hash, "status", txResult.Status)

	// send tx to safe contract
	params := []string{
		hex.EncodeToString([]byte(testSetup.mvxChainSpecificToken)),
		hex.EncodeToString(valueToSendFromMVX.Bytes()),
		hex.EncodeToString([]byte(createTransactionParam)),
		hex.EncodeToString(testSetup.ethOwnerAddress.Bytes()),
	}
	hash, err = testSetup.mvxChainSimulator.ScCall(testSetup.testContext, testSetup.mvxReceiverKeys.pk, testSetup.mvxReceiverKeys.sk, testSetup.mvxSafeAddress, zeroValue, esdtTransfer, params)
	require.NoError(t, err)
	txResult, err = testSetup.mvxChainSimulator.GetTransactionResult(testSetup.testContext, hash)
	require.NoError(t, err)

	log.Info("MVX->ETH transaction sent", "hash", hash, "status", txResult.Status)
}

func checkEthTxResult(t *testing.T, ctx context.Context, simulatedETHChain *backends.SimulatedBackend, hash common.Hash) {
	receipt, err := simulatedETHChain.TransactionReceipt(ctx, hash)
	require.NoError(t, err)
	require.Equal(t, ethStatusSuccess, receipt.Status)
}
