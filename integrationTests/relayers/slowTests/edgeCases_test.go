package slowTests

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/mock"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

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

func TestRelayersShouldExecuteTransfersForEdgeCases(t *testing.T) {
	callData := []byte{5, 4, 55}
	usdcToken := GenerateOneOperationToken()
	usdcToken.TestOperations[0].MvxSCCallData = callData
	usdcToken.TestOperations[0].MvxFaultySCCall = true

	t.Run("increasing aggregation fee before wrong SC call should stop refund", func(t *testing.T) {
		testRelayersWithChainSimulatorAndTokensForChangedAggregationFee(
			t,
			make(chan error),
			usdcToken,
		)
	})

	t.Run("decreasing max bridge amount on Safe before wrong SC call should stop refund", func(t *testing.T) {
		testRelayersWithChainSimulatorAndTokensForChangedMaxBridgeAmount(
			t,
			make(chan error),
			usdcToken,
		)
	})
}

func testRelayersWithChainSimulatorAndTokensForChangedAggregationFee(tb testing.TB, manualStopChan chan error, tokens ...framework.TestTokenParams) *framework.TestSetup {
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

			setup.ProxyWrapperInstance.RegisterBeforeTransactionSendHandler(func(tx *transaction.FrontendTransaction) {
				if tx.Sender == setup.MultiversxHandler.SCExecutorKeys.MvxAddress.Bech32() {
					if len(tokens) == 0 {
						return
					}
					setup.MultiversxHandler.SubmitAggregatorBatch(setup.Ctx, tokens[0].IssueTokenParams, big.NewInt(1200))
				}
			})
		}

		allFlowsFinished := true
		for _, flow := range flows {
			allFlowsFinished = allFlowsFinished && flow.process()
		}

		// commit blocks in order to execute incoming txs from relayers
		setup.EthereumHandler.SimulatedChain.Commit()
		setup.ChainSimulator.GenerateBlocks(setup.Ctx, 1)
		require.LessOrEqual(tb, setup.ScCallerModuleInstance.GetNumSentTransaction(), setup.GetNumScCallsOperations())

		return allFlowsFinished
	}

	return testRelayersWithChainSimulator(tb,
		setupFunc,
		processFunc,
		manualStopChan,
	)
}

func testRelayersWithChainSimulatorAndTokensForChangedMaxBridgeAmount(tb testing.TB, manualStopChan chan error, tokens ...framework.TestTokenParams) *framework.TestSetup {
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

			setup.ProxyWrapperInstance.RegisterBeforeTransactionSendHandler(func(tx *transaction.FrontendTransaction) {
				if tx.Sender == setup.MultiversxHandler.SCExecutorKeys.MvxAddress.Bech32() {
					if len(tokens) == 0 {
						return
					}
					setup.MultiversxHandler.SetMaxBridgeAmountOnSafe(setup.Ctx, tokens[0].IssueTokenParams, "800")
				}
			})
		}

		allFlowsFinished := true
		for _, flow := range flows {
			allFlowsFinished = allFlowsFinished && flow.process()
		}

		// commit blocks in order to execute incoming txs from relayers
		setup.EthereumHandler.SimulatedChain.Commit()
		setup.ChainSimulator.GenerateBlocks(setup.Ctx, 1)
		require.LessOrEqual(tb, setup.ScCallerModuleInstance.GetNumSentTransaction(), setup.GetNumScCallsOperations())

		return allFlowsFinished
	}

	return testRelayersWithChainSimulator(tb,
		setupFunc,
		processFunc,
		manualStopChan,
	)
}
