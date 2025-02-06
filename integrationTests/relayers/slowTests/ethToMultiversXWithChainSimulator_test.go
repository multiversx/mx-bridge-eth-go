//go:build slow

// To run these slow tests, simply add the slow tag on the go test command. Also, provide a chain simulator instance on the 8085 port
// example: go test -tags slow

package slowTests

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/mock"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	timeout                   = time.Minute * 15
	projectedShardForTestKeys = byte(2)
)

func TestRelayersShouldExecuteTransfers(t *testing.T) {
	_ = testRelayersWithChainSimulatorAndTokens(
		t,
		make(chan error),
		GenerateTestUSDCToken(),
		GenerateTestMEMEToken(),
	)
}

func TestRelayersShouldExecuteTransfersWithMintBurnTokens(t *testing.T) {
	_ = testRelayersWithChainSimulatorAndTokens(
		t,
		make(chan error),
		GenerateTestEUROCToken(),
		GenerateTestMEXToken(),
		GenerateTestTADAToken(),
	)
}

func TestRelayersShouldNotExecuteTransfersWithNonWhitelistedTokens(t *testing.T) {
	_ = testRelayersWithChainSimulatorAndTokens(
		t,
		make(chan error),
		GenerateUnlistedTokenFromEth(),
		GenerateUnlistedTokenFromMvx(),
	)
}

func TestRelayersShouldExecuteTransfersWithSCCallsWithArguments(t *testing.T) {
	dummyAddress := strings.Repeat("2", 32)
	dummyUint64 := string([]byte{37})

	callData := createScCallData("callPayableWithParams", 50000000, dummyUint64, dummyAddress)

	usdcToken := GenerateTestUSDCToken()
	usdcToken.TestOperations[2].MvxSCCallData = callData

	memeToken := GenerateTestMEMEToken()
	memeToken.TestOperations[2].MvxSCCallData = callData

	testSetup := testRelayersWithChainSimulatorAndTokens(
		t,
		make(chan error),
		usdcToken,
		memeToken,
	)

	testCallPayableWithParamsWasCalled(
		testSetup,
		37,
		usdcToken.AbstractTokenIdentifier,
		memeToken.AbstractTokenIdentifier,
	)
}

func TestRelayersShouldExecuteTransfersWithSCCallsWithArgumentsWithMintBurnTokens(t *testing.T) {
	dummyAddress := strings.Repeat("2", 32)
	dummyUint64 := string([]byte{37})

	callData := createScCallData("callPayableWithParams", 50000000, dummyUint64, dummyAddress)

	eurocToken := GenerateTestEUROCToken()
	eurocToken.TestOperations[2].MvxSCCallData = callData

	mexToken := GenerateTestMEXToken()
	mexToken.TestOperations[2].MvxSCCallData = callData

	tadaToken := GenerateTestTADAToken()
	tadaToken.TestOperations[2].MvxSCCallData = callData

	testSetup := testRelayersWithChainSimulatorAndTokens(
		t,
		make(chan error),
		eurocToken,
		mexToken,
		tadaToken,
	)

	testCallPayableWithParamsWasCalled(
		testSetup,
		37,
		eurocToken.AbstractTokenIdentifier,
		mexToken.AbstractTokenIdentifier,
		tadaToken.AbstractTokenIdentifier,
	)
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

	_ = testRelayersWithChainSimulatorAndTokens(
		t,
		stopChan,
		GenerateTestMEMEToken(),
	)
}

func TestRelayersShouldExecuteTransfersWithInitSupply(t *testing.T) {
	usdcInitialValue := big.NewInt(100000)
	usdcToken := GenerateTestUSDCToken()
	usdcToken.InitialSupplyValue = usdcInitialValue.String()
	usdcToken.MintBurnChecks.MvxSafeMintValue.Add(usdcToken.MintBurnChecks.MvxSafeMintValue, usdcInitialValue)

	memeInitialValue := big.NewInt(200000)
	memeToken := GenerateTestMEMEToken()
	memeToken.InitialSupplyValue = memeInitialValue.String()
	memeToken.MintBurnChecks.EthSafeMintValue.Add(memeToken.MintBurnChecks.EthSafeMintValue, memeInitialValue)

	_ = testRelayersWithChainSimulatorAndTokens(
		t,
		make(chan error),
		usdcToken,
		memeToken,
	)
}

func TestRelayersShouldExecuteTransfersWithInitSupplyMintBurn(t *testing.T) {
	eurocInitialValue := big.NewInt(100010)
	eurocToken := GenerateTestEUROCToken()
	eurocToken.InitialSupplyValue = eurocInitialValue.String()
	eurocToken.MintBurnChecks.MvxSafeMintValue.Add(eurocToken.MintBurnChecks.MvxSafeMintValue, eurocInitialValue)
	eurocToken.MintBurnChecks.EthSafeBurnValue.Add(eurocToken.MintBurnChecks.EthSafeBurnValue, eurocInitialValue)

	mexInitialValue := big.NewInt(300000)
	mexToken := GenerateTestMEXToken()
	mexToken.InitialSupplyValue = mexInitialValue.String()
	mexToken.MintBurnChecks.MvxSafeBurnValue.Add(mexToken.MintBurnChecks.MvxSafeBurnValue, mexInitialValue)
	mexToken.MintBurnChecks.EthSafeMintValue.Add(mexToken.MintBurnChecks.EthSafeMintValue, mexInitialValue)

	tadaInitialValue := big.NewInt(300000)
	tadaToken := GenerateTestTADAToken()
	tadaToken.InitialSupplyValue = tadaInitialValue.String()
	tadaToken.MintBurnChecks.MvxSafeBurnValue.Add(tadaToken.MintBurnChecks.MvxSafeBurnValue, tadaInitialValue)
	tadaToken.MintBurnChecks.EthSafeMintValue.Add(tadaToken.MintBurnChecks.EthSafeMintValue, tadaInitialValue)

	_ = testRelayersWithChainSimulatorAndTokens(
		t,
		make(chan error),
		eurocToken,
		mexToken,
		tadaToken,
	)
}

func TestRelayersShouldExecuteTransfersWithTransferRoleTokens(t *testing.T) {
	usdcToken := GenerateTestUSDCToken()
	usdcToken.IssueTokenParams.HasTransferRole = true
	usdcToken.IssueTokenParams.GrantRoleToAllAddresses = true
	usdcToken.TestOperations = []framework.TokenOperations{
		{
			ValueToTransferToMvx: big.NewInt(3000),
			ValueToSendFromMvX:   nil,
		},
		{
			ValueToTransferToMvx: big.NewInt(5050),
			ValueToSendFromMvX:   nil,
			MvxSCCallData:        createScCallData("callPayable", 50000000),
		},
	}
	usdcToken.DeltaBalances = map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
		framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
			framework.Alice: {
				OnEth:    big.NewInt(-3000 - 5050),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.UniversalToken,
			},
			framework.Bob: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(3000),
				MvxToken: framework.UniversalToken,
			},
			framework.Charlie: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.UniversalToken,
			},
			framework.SafeSC: {
				OnEth:    big.NewInt(3000 + 5050),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.ChainSpecificToken,
			},
			framework.CalledTestSC: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(5050),
				MvxToken: framework.UniversalToken,
			},
			framework.WrapperSC: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(3000 + 5050),
				MvxToken: framework.ChainSpecificToken,
			},
		},
		framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
			framework.Alice: {
				OnEth:    big.NewInt(-3000 - 5050),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.UniversalToken,
			},
			framework.Bob: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(3000),
				MvxToken: framework.UniversalToken,
			},
			framework.Charlie: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.UniversalToken,
			},
			framework.SafeSC: {
				OnEth:    big.NewInt(3000 + 5050),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.ChainSpecificToken,
			},
			framework.CalledTestSC: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(5050),
				MvxToken: framework.UniversalToken,
			},
			framework.WrapperSC: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(3000 + 5050),
				MvxToken: framework.ChainSpecificToken,
			},
		},
	}
	usdcToken.MintBurnChecks = &framework.MintBurnBalances{
		MvxTotalUniversalMint:     big.NewInt(3000 + 5050),
		MvxTotalChainSpecificMint: big.NewInt(3000 + 5050),
		MvxTotalUniversalBurn:     big.NewInt(0),
		MvxTotalChainSpecificBurn: big.NewInt(0),
		MvxSafeMintValue:          big.NewInt(3000 + 5050),
		MvxSafeBurnValue:          big.NewInt(0),

		EthSafeMintValue: big.NewInt(0),
		EthSafeBurnValue: big.NewInt(0),
	}
	usdcToken.SpecialChecks = &framework.SpecialBalanceChecks{
		WrapperDeltaLiquidityCheck: big.NewInt(3000 + 5050),
	}

	eurocToken := GenerateTestEUROCToken()
	eurocToken.IssueTokenParams.HasTransferRole = true
	eurocToken.IssueTokenParams.GrantRoleToAllAddresses = true
	eurocToken.TestOperations = []framework.TokenOperations{
		{
			ValueToTransferToMvx: big.NewInt(2000),
			ValueToSendFromMvX:   nil,
		},
		{
			ValueToTransferToMvx: big.NewInt(1500),
			ValueToSendFromMvX:   nil,
			MvxSCCallData:        createScCallData("callPayable", 50000000),
		},
	}
	eurocToken.DeltaBalances = map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
		framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
			framework.Alice: {
				OnEth:    big.NewInt(-2000 - 1500),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.UniversalToken,
			},
			framework.Bob: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(2000),
				MvxToken: framework.UniversalToken,
			},
			framework.Charlie: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.UniversalToken,
			},
			framework.SafeSC: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.ChainSpecificToken,
			},
			framework.CalledTestSC: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(1500),
				MvxToken: framework.UniversalToken,
			},
		},
		framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
			framework.Alice: {
				OnEth:    big.NewInt(-2000 - 1500),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.UniversalToken,
			},
			framework.Bob: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(2000),
				MvxToken: framework.UniversalToken,
			},
			framework.Charlie: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.UniversalToken,
			},
			framework.SafeSC: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.ChainSpecificToken,
			},
			framework.CalledTestSC: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(1500),
				MvxToken: framework.UniversalToken,
			},
		},
	}
	eurocToken.MintBurnChecks = &framework.MintBurnBalances{
		MvxTotalUniversalMint:     big.NewInt(2000 + 1500),
		MvxTotalChainSpecificMint: big.NewInt(0),
		MvxTotalUniversalBurn:     big.NewInt(0),
		MvxTotalChainSpecificBurn: big.NewInt(0),
		MvxSafeMintValue:          big.NewInt(2000 + 1500),
		MvxSafeBurnValue:          big.NewInt(0),

		EthSafeMintValue: big.NewInt(0),
		EthSafeBurnValue: big.NewInt(2000 + 1500),
	}
	eurocToken.SpecialChecks = &framework.SpecialBalanceChecks{
		WrapperDeltaLiquidityCheck: big.NewInt(0),
	}

	testRelayersWithChainSimulatorAndTokens(
		t,
		make(chan error),
		usdcToken,
		eurocToken,
	)
}

func testRelayersWithChainSimulatorAndTokens(tb testing.TB, manualStopChan chan error, tokens ...framework.TestTokenParams) *framework.TestSetup {
	flows := createFlowsBasedOnToken(tb, tokens...)

	setupFunc := func(tb testing.TB, setup *framework.TestSetup) {
		for _, flow := range flows {
			flow.setup = setup
		}

		setup.IssueAndConfigureTokens(tokens...)
		setup.MultiversxHandler.CheckForZeroBalanceOnReceivers(setup.Ctx, tokens...)
		for _, flow := range flows {
			flow.handlerToStartFirstBridge(flow)
		}
	}

	processFunc := func(tb testing.TB, setup *framework.TestSetup) bool {
		allFlowsFinished := true
		for _, flow := range flows {
			allFlowsFinished = allFlowsFinished && flow.process()
		}

		if allFlowsFinished {
			for _, flow := range flows {
				setup.TestWithdrawTotalFeesOnEthereumForTokens(flow.tokens...)
			}

			return true
		}

		// commit blocks in order to execute incoming txs from relayers
		setup.EthereumHandler.SimulatedChain.Commit()
		setup.ChainSimulator.GenerateBlocks(setup.Ctx, 1)
		require.LessOrEqual(tb, setup.ScCallerModuleInstance.GetNumSentTransaction(), setup.GetNumScCallsOperations())

		return false
	}

	return testRelayersWithChainSimulator(tb,
		setupFunc,
		processFunc,
		manualStopChan,
	)
}

func createFlowsBasedOnToken(tb testing.TB, tokens ...framework.TestTokenParams) []*testFlow {
	startsFromEthFlow := &testFlow{
		flowType:                     startFromEthereumFlow,
		TB:                           tb,
		tokens:                       make([]framework.TestTokenParams, 0, len(tokens)),
		messageAfterFirstHalfBridge:  "Ethereum->MultiversX transfer finished, now sending back to Ethereum...",
		messageAfterSecondHalfBridge: "MultiversX<->Ethereum from Ethereum transfers done",
	}
	startsFromEthFlow.handlerAfterFirstHalfBridge = func(flow *testFlow) {
		flow.setup.SendFromMultiversxToEthereum(flow.setup.BobKeys, flow.setup.CharlieKeys, flow.tokens...)
	}
	startsFromEthFlow.handlerToStartFirstBridge = func(flow *testFlow) {
		if len(flow.tokens) == 0 {
			return
		}

		flow.setup.CreateBatchOnEthereum(flow.setup.MultiversxHandler.CalleeScAddress, startsFromEthFlow.tokens...)
	}

	startsFromMvXFlow := &testFlow{
		flowType:                     startFromMultiversXFlow,
		TB:                           tb,
		tokens:                       make([]framework.TestTokenParams, 0, len(tokens)),
		messageAfterFirstHalfBridge:  "MultiversX->Ethereum transfer finished, now sending back to MultiversX...",
		messageAfterSecondHalfBridge: "MultiversX<->Ethereum from MultiversX transfers done",
	}
	startsFromMvXFlow.handlerAfterFirstHalfBridge = func(flow *testFlow) {
		flow.setup.SendFromEthereumToMultiversX(flow.setup.BobKeys, flow.setup.CharlieKeys, flow.setup.MultiversxHandler.CalleeScAddress, flow.tokens...)
	}
	startsFromMvXFlow.handlerToStartFirstBridge = func(flow *testFlow) {
		if len(flow.tokens) == 0 {
			return
		}

		flow.setup.CreateBatchOnMultiversX(startsFromMvXFlow.tokens...)
	}

	// split the tokens from where should the bridge start
	for _, token := range tokens {
		if token.IsNativeOnEth {
			startsFromEthFlow.tokens = append(startsFromEthFlow.tokens, token)
			continue
		}
		if token.IsNativeOnMvX {
			startsFromMvXFlow.tokens = append(startsFromMvXFlow.tokens, token)
			continue
		}
		require.Fail(tb, "invalid setup, found a token that is not native on any chain", "abstract identifier", token.AbstractTokenIdentifier)
	}

	return []*testFlow{startsFromEthFlow, startsFromMvXFlow}
}

func testRelayersWithChainSimulator(tb testing.TB,
	setupFunc func(tb testing.TB, setup *framework.TestSetup),
	processLoopFunc func(tb testing.TB, setup *framework.TestSetup) bool,
	stopChan chan error,
) *framework.TestSetup {
	defer func() {
		r := recover()
		if r != nil {
			require.Fail(tb, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	testSetup := framework.NewTestSetup(tb)
	log.Info(fmt.Sprintf(framework.LogStepMarker, "calling setupFunc"))
	setupFunc(tb, testSetup)

	testSetup.StartRelayersAndScModule()
	defer testSetup.Close()

	log.Info(fmt.Sprintf(framework.LogStepMarker, "running and continously call processLoopFunc"))
	interrupt := make(chan os.Signal, 1)
	for {
		select {
		case <-interrupt:
			require.Fail(tb, "signal interrupted")
			return testSetup
		case <-time.After(timeout):
			require.Fail(tb, "time out")
			return testSetup
		case err := <-stopChan:
			require.Nil(tb, err)
			return testSetup
		default:
			testDone := processLoopFunc(tb, testSetup)
			if testDone {
				return testSetup
			}
		}
	}
}

func createBadToken() framework.TestTokenParams {
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "BAD",
			NumOfDecimalsUniversal:           6,
			NumOfDecimalsChainSpecific:       6,
			MvxUniversalTokenTicker:          "BAD",
			MvxChainSpecificTokenTicker:      "ETHBAD",
			MvxUniversalTokenDisplayName:     "WrappedBAD",
			MvxChainSpecificTokenDisplayName: "EthereumWrappedBAD",
			MvxToEthFee:                      big.NewInt(50),
			ValueToMintOnMvx:                 "10000000000",
			EthTokenName:                     "ETHTOKEN",
			EthTokenSymbol:                   "ETHT",
			ValueToMintOnEth:                 "10000000000",
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(5000),
				ValueToSendFromMvX:   big.NewInt(2500),
			},
			{
				ValueToTransferToMvx: big.NewInt(7000),
				ValueToSendFromMvX:   big.NewInt(300),
			},
			{
				ValueToTransferToMvx: big.NewInt(1000),
				ValueToSendFromMvX:   nil,
				MvxSCCallData:        createScCallData("callPayable", 50000000),
			},
		},
		DeltaBalances: map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-5000 - 7000 - 1000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(5000 + 7000),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(5000 + 7000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
			},
		},
		MintBurnChecks: &framework.MintBurnBalances{
			MvxTotalUniversalMint:     big.NewInt(0),
			MvxTotalChainSpecificMint: big.NewInt(0),
			MvxTotalUniversalBurn:     big.NewInt(0),
			MvxTotalChainSpecificBurn: big.NewInt(0),
			MvxSafeMintValue:          big.NewInt(0),
			MvxSafeBurnValue:          big.NewInt(0),
		},
	}
}

func TestRelayersShouldNotExecuteTransfers(t *testing.T) {
	t.Run("isNativeOnEth = true, isMintBurnOnEth = false, isNativeOnMvX = true, isMintBurnOnMvX = false", func(t *testing.T) {
		badToken := createBadToken()
		badToken.IsNativeOnEth = true
		badToken.IsMintBurnOnEth = false
		badToken.IsNativeOnMvX = true
		badToken.IsMintBurnOnMvX = false
		badToken.HasChainSpecificToken = true

		expectedStringInLogs := "error = invalid setup isNativeOnEthereum = true, isNativeOnMultiversX = true"
		testRelayersShouldNotExecuteTransfers(t, expectedStringInLogs, badToken)
	})
	t.Run("isNativeOnEth = true, isMintBurnOnEth = false, isNativeOnMvX = true, isMintBurnOnMvX = true", func(t *testing.T) {
		badToken := createBadToken()
		badToken.IsNativeOnEth = true
		badToken.IsMintBurnOnEth = false
		badToken.IsNativeOnMvX = true
		badToken.IsMintBurnOnMvX = true
		badToken.HasChainSpecificToken = false

		expectedStringInLogs := "error = invalid setup isNativeOnEthereum = true, isNativeOnMultiversX = true"
		testRelayersShouldNotExecuteTransfers(t, expectedStringInLogs, badToken)
	})
	t.Run("isNativeOnEth = true, isMintBurnOnEth = true, isNativeOnMvX = true, isMintBurnOnMvX = false", func(t *testing.T) {
		badToken := createBadToken()
		badToken.IsNativeOnEth = true
		badToken.IsMintBurnOnEth = true
		badToken.IsNativeOnMvX = true
		badToken.IsMintBurnOnMvX = false
		badToken.HasChainSpecificToken = true

		testEthContractsShouldError(t, badToken)
	})
	t.Run("isNativeOnEth = false, isMintBurnOnEth = true, isNativeOnMvX = false, isMintBurnOnMvX = true", func(t *testing.T) {
		badToken := createBadToken()
		badToken.IsNativeOnEth = false
		badToken.IsMintBurnOnEth = true
		badToken.IsNativeOnMvX = false
		badToken.IsMintBurnOnMvX = true
		badToken.HasChainSpecificToken = true

		testEthContractsShouldError(t, badToken)
	})
}

func testRelayersShouldNotExecuteTransfers(
	tb testing.TB,
	expectedStringInLogs string,
	tokens ...framework.TestTokenParams,
) {
	flows := createFlowsBasedOnToken(tb, tokens...)

	setupFunc := func(tb testing.TB, setup *framework.TestSetup) {
		for _, flow := range flows {
			flow.setup = setup
		}

		setup.IssueAndConfigureTokens(tokens...)
		setup.MultiversxHandler.CheckForZeroBalanceOnReceivers(setup.Ctx, tokens...)

		for _, flow := range flows {
			flow.handlerToStartFirstBridge(flow)
		}
	}

	processFunc := func(tb testing.TB, setup *framework.TestSetup) bool {
		allFlowsFinished := true
		for _, flow := range flows {
			allFlowsFinished = allFlowsFinished && flow.process()
		}

		// commit blocks in order to execute incoming txs from relayers
		setup.EthereumHandler.SimulatedChain.Commit()
		setup.ChainSimulator.GenerateBlocks(setup.Ctx, 1)

		return allFlowsFinished
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
	numOfErrorsToWait := numOfTimesToRepeatErrorForRelayer * framework.NumRelayers

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

	_ = testRelayersWithChainSimulator(tb, setupFunc, processFunc, stopChan)
}

func testEthContractsShouldError(tb testing.TB, testToken framework.TestTokenParams) {
	setupFunc := func(tb testing.TB, setup *framework.TestSetup) {
		setup.IssueAndConfigureTokens(testToken)

		token := setup.GetTokenData(testToken.AbstractTokenIdentifier)
		require.NotNil(tb, token)

		valueToMintOnEth, ok := big.NewInt(0).SetString(testToken.ValueToMintOnEth, 10)
		require.True(tb, ok)

		receiverKeys := framework.GenerateMvxPrivatePublicKey(tb, projectedShardForTestKeys)
		auth, _ := bind.NewKeyedTransactorWithChainID(setup.DepositorKeys.EthSK, setup.EthereumHandler.ChainID)
		_, err := setup.EthereumHandler.SafeContract.Deposit(auth, token.EthErc20Address, valueToMintOnEth, receiverKeys.MvxAddress.AddressSlice())
		require.Error(tb, err)
	}

	processFunc := func(tb testing.TB, setup *framework.TestSetup) bool {
		time.Sleep(time.Second) // allow go routines to start
		return true
	}

	_ = testRelayersWithChainSimulator(tb,
		setupFunc,
		processFunc,
		make(chan error),
	)
}

func testCallPayableWithParamsWasCalled(testSetup *framework.TestSetup, value uint64, tokens ...string) {
	if len(tokens) == 0 {
		return
	}

	universalTokens := make([]string, 0, len(tokens))
	for _, identifier := range tokens {
		tkData := testSetup.TokensRegistry.GetTokenData(identifier)
		universalTokens = append(universalTokens, tkData.MvxUniversalToken)
	}

	vmRequest := &data.VmValueRequest{
		Address:  testSetup.MultiversxHandler.CalleeScAddress.Bech32(),
		FuncName: "getCalledDataParams",
	}

	vmResponse, err := testSetup.ChainSimulator.Proxy().ExecuteVMQuery(context.Background(), vmRequest)
	require.Nil(testSetup, err)

	returnedData := vmResponse.Data.ReturnData
	require.Equal(testSetup, len(tokens), len(returnedData))

	mapUniversalTokens := make(map[string]int)
	for _, tokenIdentifier := range universalTokens {
		mapUniversalTokens[tokenIdentifier] = 0
	}

	for _, buff := range returnedData {
		parsedValue, parsedToken := processCalledDataParams(buff)
		assert.Equal(testSetup, value, parsedValue)
		mapUniversalTokens[parsedToken]++
	}

	assert.Equal(testSetup, len(tokens), len(mapUniversalTokens))
	for _, numTokens := range mapUniversalTokens {
		assert.Equal(testSetup, 1, numTokens)
	}
}

func processCalledDataParams(buff []byte) (uint64, string) {
	valBuff := buff[:8]
	value := binary.BigEndian.Uint64(valBuff)

	buff = buff[8+32:] // trim the nonce and the address
	tokenLenBuff := buff[:4]
	tokenLen := binary.BigEndian.Uint32(tokenLenBuff)
	buff = buff[4:] // trim the length of the token string

	token := string(buff[:tokenLen])

	return value, token
}
