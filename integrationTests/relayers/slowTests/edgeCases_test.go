//go:build slow

package slowTests

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/mock"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

const scCallsDeltaLimit = 10

func TestRelayerShouldExecuteSimultaneousSwapsAndNotCatchErrors(t *testing.T) {
	errorString := "ERROR"
	mockLogObserver := mock.NewMockLogObserver(errorString, "got invalid action ID")
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

	usdcToken := GenerateTestUSDCToken()
	usdcToken.MultipleSpendings = big.NewInt(2)
	usdcToken.TestOperations = []framework.TokenOperations{
		{
			ValueToTransferToMvx: big.NewInt(5000),
			ValueToSendFromMvX:   big.NewInt(200),
			MvxSCCallData:        nil,
			MvxFaultySCCall:      false,
			MvxForceSCCall:       false,
		},
	}
	usdcToken.DeltaBalances = map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
		framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
			framework.Alice: {
				OnEth:    big.NewInt(-5000),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.UniversalToken,
			},
			framework.Bob: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(5000),
				MvxToken: framework.UniversalToken,
			},
			framework.SafeSC: {
				OnEth:    big.NewInt(5000),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.ChainSpecificToken,
			},
			framework.WrapperSC: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(5000),
				MvxToken: framework.ChainSpecificToken,
			},
		},
		framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
			framework.Alice: {
				OnEth:    big.NewInt(-5000 - 5000 + 150),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.UniversalToken,
			},
			framework.Bob: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(5000 + 4800),
				MvxToken: framework.UniversalToken,
			},
			framework.SafeSC: {
				OnEth:    big.NewInt(5000 + 5000 - 150),
				OnMvx:    big.NewInt(50),
				MvxToken: framework.ChainSpecificToken,
			},
			framework.WrapperSC: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(5000 + 5000 - 200),
				MvxToken: framework.ChainSpecificToken,
			},
		},
	}
	usdcToken.MintBurnChecks = &framework.MintBurnBalances{
		MvxTotalUniversalMint:     big.NewInt(5000 + 5000),
		MvxTotalChainSpecificMint: big.NewInt(5000 + 5000),
		MvxTotalUniversalBurn:     big.NewInt(200),
		MvxTotalChainSpecificBurn: big.NewInt(200 - 50),
		MvxSafeMintValue:          big.NewInt(5000 + 5000),
		MvxSafeBurnValue:          big.NewInt(200 - 50),

		EthSafeMintValue: big.NewInt(0),
		EthSafeBurnValue: big.NewInt(0),
	}
	usdcToken.SpecialChecks.WrapperDeltaLiquidityCheck = big.NewInt(5000 + 5000 - 200)

	_ = testRelayersWithChainSimulatorAndTokensForSimultaneousSwaps(
		t,
		stopChan,
		usdcToken,
	)
}

func testRelayersWithChainSimulatorAndTokensForSimultaneousSwaps(tb testing.TB, manualStopChan chan error, tokens ...framework.TestTokenParams) *framework.TestSetup {
	startsFromEthFlow := &testFlow{
		TB:                           tb,
		tokens:                       tokens,
		messageAfterFirstHalfBridge:  "Ethereum->MultiversX transfer finished, now sending back to Ethereum & another round from Ethereum...",
		messageAfterSecondHalfBridge: "MultiversX<->Ethereum from Ethereum transfers done",
	}
	startsFromEthFlow.handlerAfterFirstHalfBridge = func(flow *testFlow) {
		flow.setup.SendFromMultiversxToEthereum(flow.setup.BobKeys, flow.setup.AliceKeys, flow.tokens...)
		flow.setup.SendFromEthereumToMultiversX(flow.setup.AliceKeys, flow.setup.BobKeys, flow.setup.MultiversxHandler.CalleeScAddress, flow.tokens...)
	}

	setupFunc := func(tb testing.TB, setup *framework.TestSetup) {
		startsFromEthFlow.setup = setup

		setup.IssueAndConfigureTokens(tokens...)
		setup.MultiversxHandler.CheckForZeroBalanceOnReceivers(setup.Ctx, tokens...)
		setup.CreateBatchOnEthereum(setup.MultiversxHandler.CalleeScAddress, startsFromEthFlow.tokens...)
	}

	processFunc := func(tb testing.TB, setup *framework.TestSetup) bool {
		if startsFromEthFlow.process() {
			setup.TestWithdrawTotalFeesOnEthereumForTokens(startsFromEthFlow.tokens...)

			return true
		}

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

func TestRelayerShouldExecuteMultipleSwapsWithLargeData(t *testing.T) {
	usdcToken := GenerateTestUSDCToken()

	numTxs := int64(16)
	maxLimitWithForScCalls := 984
	buff := make([]byte, maxLimitWithForScCalls)
	_, _ = rand.Read(buff)
	scCallData := createScCallData("callPayableWithBuff", 100000000, string(buff))

	usdcToken.TestOperations = make([]framework.TokenOperations, 0, numTxs)
	for i := 0; i < int(numTxs); i++ {
		tokenOperation := framework.TokenOperations{
			ValueToTransferToMvx: big.NewInt(50),
			ValueToSendFromMvX:   nil,
			MvxSCCallData:        scCallData,
			MvxFaultySCCall:      true,
		}
		usdcToken.TestOperations = append(usdcToken.TestOperations, tokenOperation)
	}
	usdcToken.DeltaBalances = map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
		framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
			framework.Alice: {
				OnEth:    big.NewInt(-50 * numTxs),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.UniversalToken,
			},
			framework.SafeSC: {
				OnEth:    big.NewInt(50 * numTxs),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.ChainSpecificToken,
			},
			framework.WrapperSC: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(50 * numTxs),
				MvxToken: framework.ChainSpecificToken,
			},
			framework.CalledTestSC: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(50 * numTxs),
				MvxToken: framework.UniversalToken,
			},
		},
		framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
			framework.Alice: {
				OnEth:    big.NewInt(-50 * numTxs),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.UniversalToken,
			},
			framework.SafeSC: {
				OnEth:    big.NewInt(50 * numTxs),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.ChainSpecificToken,
			},
			framework.WrapperSC: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(50 * numTxs),
				MvxToken: framework.ChainSpecificToken,
			},
			framework.CalledTestSC: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(50 * numTxs),
				MvxToken: framework.UniversalToken,
			},
		},
	}
	usdcToken.MintBurnChecks = &framework.MintBurnBalances{
		MvxTotalUniversalMint:     big.NewInt(50 * numTxs),
		MvxTotalChainSpecificMint: big.NewInt(50 * numTxs),
		MvxTotalUniversalBurn:     big.NewInt(0),
		MvxTotalChainSpecificBurn: big.NewInt(0),
		MvxSafeMintValue:          big.NewInt(50 * numTxs),
		MvxSafeBurnValue:          big.NewInt(0),

		EthSafeMintValue: big.NewInt(0),
		EthSafeBurnValue: big.NewInt(0),
	}
	usdcToken.SpecialChecks.WrapperDeltaLiquidityCheck = big.NewInt(50 * numTxs)

	_ = testRelayersWithChainSimulatorAndTokensWithMultipleSwapsAndLargeScCalls(
		t,
		make(chan error),
		usdcToken,
	)
}

func testRelayersWithChainSimulatorAndTokensWithMultipleSwapsAndLargeScCalls(tb testing.TB, manualStopChan chan error, tokens ...framework.TestTokenParams) *framework.TestSetup {
	flows := createFlowsBasedOnToken(tb, tokens...)

	setupFunc := func(tb testing.TB, setup *framework.TestSetup) {
		for _, flow := range flows {
			flow.setup = setup
		}

		setup.EthereumHandler.SetBatchSize(setup.Ctx, 100)

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

func TestRelayersShouldExecuteTransfersForEdgeCases(t *testing.T) {
	badCallData := []byte{5, 4, 55}
	testToken := framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "TEST",
			NumOfDecimalsUniversal:           6,
			NumOfDecimalsChainSpecific:       6,
			MvxUniversalTokenTicker:          "TEST",
			MvxChainSpecificTokenTicker:      "ONETEST",
			MvxUniversalTokenDisplayName:     "WrappedTEST",
			MvxChainSpecificTokenDisplayName: "EthereumWrappedTEST",
			ValueToMintOnMvx:                 "10000000000",
			IsMintBurnOnMvX:                  true,
			IsNativeOnMvX:                    false,
			HasChainSpecificToken:            true,
			EthTokenName:                     "EthTEST",
			EthTokenSymbol:                   "TEST",
			ValueToMintOnEth:                 "10000000000",
			IsMintBurnOnEth:                  false,
			IsNativeOnEth:                    true,
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(1000),
				ValueToSendFromMvX:   nil,
				MvxSCCallData:        badCallData,
				MvxFaultySCCall:      true,
			},
		},
		DeltaBalances: map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-1000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(1000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(1000),
					MvxToken: framework.ChainSpecificToken,
				},
			},
			framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-1000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(1000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(1000),
					MvxToken: framework.ChainSpecificToken,
				},
			},
		},
		MintBurnChecks: &framework.MintBurnBalances{
			MvxTotalUniversalMint:     big.NewInt(1000),
			MvxTotalChainSpecificMint: big.NewInt(1000),
			MvxTotalUniversalBurn:     big.NewInt(0),
			MvxTotalChainSpecificBurn: big.NewInt(0),
			MvxSafeMintValue:          big.NewInt(1000),
			MvxSafeBurnValue:          big.NewInt(0),

			EthSafeMintValue: big.NewInt(0),
			EthSafeBurnValue: big.NewInt(0),
		},
		SpecialChecks: &framework.SpecialBalanceChecks{
			WrapperDeltaLiquidityCheck: big.NewInt(1000),
		},
	}

	t.Run("increasing aggregation fee before wrong SC call should stop refund", func(t *testing.T) {
		handler := func(setup *framework.TestSetup, tokens []framework.TestTokenParams) {
			setup.ProxyWrapperInstance.RegisterBeforeTransactionSendHandler(func(tx *transaction.FrontendTransaction) {
				if tx.Sender == setup.MultiversxHandler.SCExecutorKeys.MvxAddress.Bech32() && len(tokens) > 0 {
					setup.MultiversxHandler.SubmitAggregatorBatch(setup.Ctx, tokens[0].IssueTokenParams, big.NewInt(1200))
				}
			})
		}

		testRelayersWithChainSimulatorAndTokensForDynamicPriceChange(
			t,
			make(chan error),
			handler,
			testToken,
		)
	})

	t.Run("decreasing max bridge amount on Safe before wrong SC call should stop refund", func(t *testing.T) {
		handler := func(setup *framework.TestSetup, tokens []framework.TestTokenParams) {
			setup.ProxyWrapperInstance.RegisterBeforeTransactionSendHandler(func(tx *transaction.FrontendTransaction) {
				if tx.Sender == setup.MultiversxHandler.SCExecutorKeys.MvxAddress.Bech32() && len(tokens) > 0 {
					setup.MultiversxHandler.SetMaxBridgeAmountOnSafe(setup.Ctx, tokens[0].IssueTokenParams, "800")
				}
			})
		}

		testRelayersWithChainSimulatorAndTokensForDynamicPriceChange(
			t,
			make(chan error),
			handler,
			testToken,
		)
	})
}

func testRelayersWithChainSimulatorAndTokensForDynamicPriceChange(
	tb testing.TB,
	manualStopChan chan error,
	beforeTransactionHandler func(setup *framework.TestSetup, tokens []framework.TestTokenParams),
	tokens ...framework.TestTokenParams,
) *framework.TestSetup {
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

	firstProcessRun := true
	processFunc := func(tb testing.TB, setup *framework.TestSetup) bool {
		if firstProcessRun {
			firstProcessRun = false

			if beforeTransactionHandler != nil {
				beforeTransactionHandler(setup, tokens)
			}
		}

		allFlowsFinished := true
		for _, flow := range flows {
			allFlowsFinished = allFlowsFinished && flow.process()
		}

		// commit blocks in order to execute incoming txs from relayers
		setup.EthereumHandler.SimulatedChain.Commit()
		setup.ChainSimulator.GenerateBlocks(setup.Ctx, 1)
		scCallsLimitReached := int32(setup.ScCallerModuleInstance.GetNumSentTransaction()-setup.GetNumScCallsOperations()) >= scCallsDeltaLimit

		return allFlowsFinished && scCallsLimitReached
	}

	return testRelayersWithChainSimulator(tb,
		setupFunc,
		processFunc,
		manualStopChan,
	)
}
