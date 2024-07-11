//go:build slow

// To run these slow tests, simply add the slow tag on the go test command. Also, provide a chain simulator instance on the 8085 port
// example: go test -tags slow

package slowTests

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/mock"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/require"
)

const (
	timeout = time.Minute * 15
)

var (
	//MEME is ethNative = false, ethMintBurn = true, mvxNative = true, mvxMintBurn = false
	memeToken = testTokenParams{
		issueTokenParams: issueTokenParams{
			abstractTokenIdentifier:          "MEME",
			numOfDecimalsUniversal:           1,
			numOfDecimalsChainSpecific:       1,
			mvxUniversalTokenTicker:          "MEME",
			mvxChainSpecificTokenTicker:      "ETHMEME",
			mvxUniversalTokenDisplayName:     "WrappedMEME",
			mvxChainSpecificTokenDisplayName: "EthereumWrappedMEME",
			valueToMintOnMvx:                 "10000000000",
			isMintBurnOnMvX:                  false,
			isNativeOnMvX:                    true,
			ethTokenName:                     "ETHMEME",
			ethTokenSymbol:                   "ETHM",
			valueToMintOnEth:                 "10000000000",
			isMintBurnOnEth:                  true,
			isNativeOnEth:                    false,
		},
		testOperations: []tokenOperations{
			{
				valueToTransferToMvx: big.NewInt(2400),
				valueToSendFromMvX:   big.NewInt(4000),
				ethSCCallMethod:      "",
				ethSCCallGasLimit:    0,
				ethSCCallArguments:   nil,
			},
			{
				valueToTransferToMvx: big.NewInt(200),
				valueToSendFromMvX:   big.NewInt(6000),
				ethSCCallMethod:      "",
				ethSCCallGasLimit:    0,
				ethSCCallArguments:   nil,
			},
			{
				valueToTransferToMvx: big.NewInt(1000),
				valueToSendFromMvX:   big.NewInt(2000),
				ethSCCallMethod:      "callPayable",
				ethSCCallGasLimit:    50000000,
				ethSCCallArguments:   nil,
			},
		},
		esdtSafeExtraBalance:    big.NewInt(4000 + 6000 + 2000), // everything is locked in the safe esdt contract
		ethTestAddrExtraBalance: big.NewInt(4000 - 50 + 6000 - 50 + 2000 - 50),
	}
)

func TestRelayersShouldExecuteTransfers(t *testing.T) {
	// USDC is ethNative = true, ethMintBurn = false, mvxNative = false, mvxMintBurn = true
	usdcToken := testTokenParams{
		issueTokenParams: issueTokenParams{
			abstractTokenIdentifier:          "USDC",
			numOfDecimalsUniversal:           6,
			numOfDecimalsChainSpecific:       6,
			mvxUniversalTokenTicker:          "USDC",
			mvxChainSpecificTokenTicker:      "ETHUSDC",
			mvxUniversalTokenDisplayName:     "WrappedUSDC",
			mvxChainSpecificTokenDisplayName: "EthereumWrappedUSDC",
			valueToMintOnMvx:                 "10000000000",
			isMintBurnOnMvX:                  true,
			isNativeOnMvX:                    false,
			ethTokenName:                     "ETHTOKEN",
			ethTokenSymbol:                   "ETHT",
			valueToMintOnEth:                 "10000000000",
			isMintBurnOnEth:                  false,
			isNativeOnEth:                    true,
		},
		testOperations: []tokenOperations{
			{
				valueToTransferToMvx: big.NewInt(5000),
				valueToSendFromMvX:   big.NewInt(2500),
				ethSCCallMethod:      "",
				ethSCCallGasLimit:    0,
				ethSCCallArguments:   nil,
			},
			{
				valueToTransferToMvx: big.NewInt(7000),
				valueToSendFromMvX:   big.NewInt(300),
				ethSCCallMethod:      "",
				ethSCCallGasLimit:    0,
				ethSCCallArguments:   nil,
			},
			{
				valueToTransferToMvx: big.NewInt(1000),
				valueToSendFromMvX:   nil,
				ethSCCallMethod:      "callPayable",
				ethSCCallGasLimit:    50000000,
				ethSCCallArguments:   nil,
			},
		},
		esdtSafeExtraBalance:    big.NewInt(100),                                        // extra is just for the fees for the 2 transfers mvx->eth
		ethTestAddrExtraBalance: big.NewInt(-5000 + 2500 - 50 - 7000 + 300 - 50 - 1000), // -(eth->mvx) + (mvx->eth) - fees
	}

	testRelayersWithChainSimulatorAndTokens(t, make(chan error), usdcToken, memeToken)
}

func TestRelayerShouldExecuteTransfersAndNotCatchErrors(t *testing.T) {
	errorString := "ERROR"
	mockLogObserver := mock.NewMockLogObserver(errorString)
	err := logger.AddLogObserver(mockLogObserver, &logger.PlainFormatter{})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, logger.RemoveLogObserver(mockLogObserver))
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopChan := make(chan error, 1000) // ensure sufficient error buffer

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-mockLogObserver.LogFoundChan():
				stopChan <- errors.New("logger should have not caught errors")
			}
		}
	}()

	testRelayersWithChainSimulatorAndTokens(t, stopChan, memeToken)
}

func testRelayersWithChainSimulatorAndTokens(tb testing.TB, manualStopChan chan error, tokens ...testTokenParams) {
	startsFromEthFlow, startsFromMvXFlow := createFlowsBasedOnToken(tb, tokens...)

	setupFunc := func(tb testing.TB, testSetup *simulatedSetup) {
		startsFromMvXFlow.testSetup = testSetup
		startsFromEthFlow.testSetup = testSetup

		testSetup.issueAndConfigureTokens(tokens...)
		testSetup.checkForZeroBalanceOnReceivers(tokens...)
		if len(startsFromEthFlow.tokens) > 0 {
			testSetup.createBatchOnEthereum(startsFromEthFlow.tokens...)
		}
		if len(startsFromMvXFlow.tokens) > 0 {
			testSetup.createBatchOnMultiversX(startsFromMvXFlow.tokens...)
		}
	}

	processFunc := func(tb testing.TB, testSetup *simulatedSetup) bool {
		if startsFromEthFlow.process() && startsFromMvXFlow.process() {
			return true
		}

		// commit blocks in order to execute incoming txs from relayers
		testSetup.simulatedETHChain.Commit()
		testSetup.mvxChainSimulator.GenerateBlocks(testSetup.testContext, 1)

		return false
	}

	testRelayersWithChainSimulator(tb,
		setupFunc,
		processFunc,
		manualStopChan,
	)
}

func createFlowsBasedOnToken(tb testing.TB, tokens ...testTokenParams) (*startsFromEthereumFlow, *startsFromMultiversXFlow) {
	startsFromEthFlow := &startsFromEthereumFlow{
		TB:     tb,
		tokens: make([]testTokenParams, 0, len(tokens)),
	}

	startsFromMvXFlow := &startsFromMultiversXFlow{
		TB:     tb,
		tokens: make([]testTokenParams, 0, len(tokens)),
	}

	// split the tokens from where should the bridge start
	for _, token := range tokens {
		if token.isNativeOnEth {
			startsFromEthFlow.tokens = append(startsFromEthFlow.tokens, token)
			continue
		}
		if token.isNativeOnMvX {
			startsFromMvXFlow.tokens = append(startsFromMvXFlow.tokens, token)
			continue
		}
		require.Fail(tb, "invalid setup, found a token that is not native on any chain", "abstract identifier", token.abstractTokenIdentifier)
	}

	return startsFromEthFlow, startsFromMvXFlow
}

func testRelayersWithChainSimulator(tb testing.TB,
	setupFunc func(tb testing.TB, testSetup *simulatedSetup),
	processLoopFunc func(tb testing.TB, testSetup *simulatedSetup) bool,
	stopChan chan error,
) {
	defer func() {
		r := recover()
		if r != nil {
			require.Fail(tb, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	testSetup := prepareSimulatedSetup(tb)
	log.Info(fmt.Sprintf(logStepMarker, "calling setupFunc"))
	setupFunc(tb, testSetup)

	testSetup.startRelayersAndScModule()
	defer testSetup.close()

	log.Info(fmt.Sprintf(logStepMarker, "running and continously call processLoopFunc"))
	interrupt := make(chan os.Signal, 1)
	for {
		select {
		case <-interrupt:
			require.Fail(tb, "signal interrupted")
			return
		case <-time.After(timeout):
			require.Fail(tb, "time out")
			return
		case err := <-stopChan:
			require.Nil(tb, err)
			return
		default:
			testDone := processLoopFunc(tb, testSetup)
			if testDone {
				return
			}
		}
	}
}

func createBadToken() testTokenParams {
	return testTokenParams{
		issueTokenParams: issueTokenParams{
			abstractTokenIdentifier:          "BAD",
			numOfDecimalsUniversal:           6,
			numOfDecimalsChainSpecific:       6,
			mvxUniversalTokenTicker:          "BAD",
			mvxChainSpecificTokenTicker:      "ETHBAD",
			mvxUniversalTokenDisplayName:     "WrappedBAD",
			mvxChainSpecificTokenDisplayName: "EthereumWrappedBAD",
			valueToMintOnMvx:                 "10000000000",
			ethTokenName:                     "ETHTOKEN",
			ethTokenSymbol:                   "ETHT",
			valueToMintOnEth:                 "10000000000",
		},
		testOperations: []tokenOperations{
			{
				valueToTransferToMvx: big.NewInt(5000),
				valueToSendFromMvX:   big.NewInt(2500),
				ethSCCallMethod:      "",
				ethSCCallGasLimit:    0,
				ethSCCallArguments:   nil,
			},
			{
				valueToTransferToMvx: big.NewInt(7000),
				valueToSendFromMvX:   big.NewInt(300),
				ethSCCallMethod:      "",
				ethSCCallGasLimit:    0,
				ethSCCallArguments:   nil,
			},
			{
				valueToTransferToMvx: big.NewInt(1000),
				valueToSendFromMvX:   nil,
				ethSCCallMethod:      "callPayable",
				ethSCCallGasLimit:    50000000,
				ethSCCallArguments:   nil,
			},
		},
		esdtSafeExtraBalance:    big.NewInt(0),
		ethTestAddrExtraBalance: big.NewInt(0),
	}
}

func TestRelayersShouldNotExecuteTransfers(t *testing.T) {
	t.Run("isNativeOnEth = true, isMintBurnOnEth = false, isNativeOnMvX = true, isMintBurnOnMvX = false", func(t *testing.T) {
		badToken := createBadToken()
		badToken.isNativeOnEth = true
		badToken.isMintBurnOnEth = false
		badToken.isNativeOnMvX = true
		badToken.isMintBurnOnMvX = false

		expectedStringInLogs := "error = invalid setup isNativeOnEthereum = true, isNativeOnMultiversX = true"
		testRelayersShouldNotExecuteTransfers(t, expectedStringInLogs, badToken)
	})
	t.Run("isNativeOnEth = true, isMintBurnOnEth = false, isNativeOnMvX = true, isMintBurnOnMvX = true", func(t *testing.T) {
		badToken := createBadToken()
		badToken.isNativeOnEth = true
		badToken.isMintBurnOnEth = false
		badToken.isNativeOnMvX = true
		badToken.isMintBurnOnMvX = true

		expectedStringInLogs := "error = invalid setup isNativeOnEthereum = true, isNativeOnMultiversX = true"
		testRelayersShouldNotExecuteTransfers(t, expectedStringInLogs, badToken)
	})
	t.Run("isNativeOnEth = true, isMintBurnOnEth = true, isNativeOnMvX = true, isMintBurnOnMvX = false", func(t *testing.T) {
		badToken := createBadToken()
		badToken.isNativeOnEth = true
		badToken.isMintBurnOnEth = true
		badToken.isNativeOnMvX = true
		badToken.isMintBurnOnMvX = false

		testEthContractsShouldError(t, badToken)
	})
	t.Run("isNativeOnEth = false, isMintBurnOnEth = true, isNativeOnMvX = false, isMintBurnOnMvX = true", func(t *testing.T) {
		badToken := createBadToken()
		badToken.isNativeOnEth = false
		badToken.isMintBurnOnEth = true
		badToken.isNativeOnMvX = false
		badToken.isMintBurnOnMvX = true

		testEthContractsShouldError(t, badToken)
	})
}

func testRelayersShouldNotExecuteTransfers(
	tb testing.TB,
	expectedStringInLogs string,
	tokens ...testTokenParams,
) {
	startsFromEthFlow, startsFromMvXFlow := createFlowsBasedOnToken(tb, tokens...)

	setupFunc := func(tb testing.TB, testSetup *simulatedSetup) {
		startsFromMvXFlow.testSetup = testSetup
		startsFromEthFlow.testSetup = testSetup

		testSetup.issueAndConfigureTokens(tokens...)
		testSetup.checkForZeroBalanceOnReceivers(tokens...)
		if len(startsFromEthFlow.tokens) > 0 {
			testSetup.createBatchOnEthereum(startsFromEthFlow.tokens...)
		}
		if len(startsFromMvXFlow.tokens) > 0 {
			testSetup.createBatchOnMultiversX(startsFromMvXFlow.tokens...)
		}
	}

	processFunc := func(tb testing.TB, testSetup *simulatedSetup) bool {
		if startsFromEthFlow.process() && startsFromMvXFlow.process() {
			return true
		}

		// commit blocks in order to execute incoming txs from relayers
		testSetup.simulatedETHChain.Commit()
		testSetup.mvxChainSimulator.GenerateBlocks(testSetup.testContext, 1)

		return false
	}

	// start a mocked log observer that is looking for a specific relayer error
	chanCnt := 0
	mockLogObserver := mock.NewMockLogObserver(expectedStringInLogs)
	err := logger.AddLogObserver(mockLogObserver, &logger.PlainFormatter{})
	require.NoError(tb, err)
	defer func() {
		require.NoError(tb, logger.RemoveLogObserver(mockLogObserver))
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	numOfTimesToRepeatErrorForRelayer := 10
	numOfErrorsToWait := numOfTimesToRepeatErrorForRelayer * numRelayers

	stopChan := make(chan error, 1)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-mockLogObserver.LogFoundChan():
				chanCnt++
				if chanCnt >= numOfErrorsToWait {
					log.Info(fmt.Sprintf("test passed, relayers are stuck, expected string `%s` found in all relayers' logs for %d times", expectedStringInLogs, numOfErrorsToWait))
					stopChan <- nil
					return
				}
			}
		}
	}()

	testRelayersWithChainSimulator(tb, setupFunc, processFunc, stopChan)
}

func testEthContractsShouldError(tb testing.TB, testToken testTokenParams) {
	setupFunc := func(tb testing.TB, testSetup *simulatedSetup) {
		testSetup.issueAndConfigureTokens(testToken)

		token := testSetup.getTokenData(testToken.abstractTokenIdentifier)
		require.NotNil(tb, token)

		valueToMintOnEth, ok := big.NewInt(0).SetString(testToken.valueToMintOnEth, 10)
		require.True(tb, ok)

		receiverKeys := generateMvxPrivatePublicKey(tb)
		mvxReceiverAddress, err := data.NewAddressFromBech32String(receiverKeys.pk)
		require.NoError(tb, err)

		auth, _ := bind.NewKeyedTransactorWithChainID(ethDepositorSK, testSetup.ethChainID)
		_, err = testSetup.ethSafeContract.Deposit(auth, token.ethErc20Address, valueToMintOnEth, mvxReceiverAddress.AddressSlice())
		require.Error(tb, err)
	}

	processFunc := func(tb testing.TB, testSetup *simulatedSetup) bool {
		time.Sleep(time.Second) // allow go routines to start
		return true
	}

	testRelayersWithChainSimulator(tb,
		setupFunc,
		processFunc,
		make(chan error),
	)
}
