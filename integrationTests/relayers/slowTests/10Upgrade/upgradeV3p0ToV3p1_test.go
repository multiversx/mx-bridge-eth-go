//go:build slow

package upgrade

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
)

const testMarker = "==================================== %s ===================================="

func TestComplexScenarioWithUpgrade(t *testing.T) {
	tokens := []framework.TestTokenParams{
		GenerateTestUSDCToken(),
		GenerateTestMEMEToken(),
		GenerateTestTADAToken(),
		GenerateTestEUROCToken(),
		GenerateTestMEXToken(),
	}
	flows := slowTests.CreateFlowsBasedOnToken(t, tokens...)

	testStep := 0
	processFunc := func(tb testing.TB, setup *framework.TestSetup) bool {
		if testStep == 0 {
			setNextCompletionChecker(flows, testStep, setup)
			testStep++
		}

		if areFlowsFinished(flows) {
			completeFinish := setNextCompletionChecker(flows, testStep, setup)
			testStep++
			if completeFinish {
				return true
			}
		}

		// commit blocks in order to execute incoming txs from relayers
		setup.EthereumHandler.SimulatedChain.Commit()
		setup.ChainSimulator.GenerateBlocks(setup.Ctx, 1)
		return false
	}

	_ = executeUpgradeScenario(t, processFunc, flows, tokens)
}

func areFlowsFinished(flows []*slowTests.TestFlow) bool {
	finished := true
	for _, flow := range flows {
		finished = finished && flow.Process()
	}

	return finished
}

func setNextCompletionChecker(
	flows []*slowTests.TestFlow,
	testStep int,
	setup *framework.TestSetup,
) bool {
	setExecutionStateOnStep(flows, false)

	switch testStep {
	case 0:
		step0DoSwapsOnV3p0()
		return false
	case 1:
		step1UpgradeContracts(setup, flows)
		return false
	case 2:
		step2DoSwapsOnV3p1(flows)
		return false
	default:
		return true
	}
}

func setExecutionStateOnStep(flows []*slowTests.TestFlow, state bool) {
	for _, flow := range flows {
		flow.FirstHalfBridgeDone = state
		flow.SecondHalfBridgeDone = state
	}
}

func step0DoSwapsOnV3p0() {
	log.Info(fmt.Sprintf(testMarker, "Starting step 0 - swaps that work on v3.0 contracts"))

	// balance tests are the default ones (the ones defined when setting the token)
}

func step1UpgradeContracts(setup *framework.TestSetup, flows []*slowTests.TestFlow) {
	log.Info(fmt.Sprintf(testMarker, "Starting step 1 - upgrade & setup contracts for v3.1"))

	setup.MultiversxHandler.PauseAllContracts(setup.Ctx)
	log.Info(fmt.Sprintf(testMarker, "Sub-step 1.1 - contracts paused"))

	setup.MultiversxHandler.UpgradeContractsToVersion(setup.Ctx, framework.ContractsVersion3p1)
	log.Info(fmt.Sprintf(testMarker, "Sub-step 1.2 - contracts upgraded"))

	setup.MultiversxHandler.UnPauseAllContracts(setup.Ctx)
	log.Info(fmt.Sprintf(testMarker, "Sub-step 1.3 - contracts unpaused"))

	setExecutionStateOnStep(flows, true)
}

func step2DoSwapsOnV3p1(flows []*slowTests.TestFlow) {
	log.Info(fmt.Sprintf(testMarker, "Starting step 2 - swaps that work on v3.1 contracts"))

	startsFromEthFlow := flows[0]
	startsFromMvxFlow := flows[1]

	startsFromEthFlow.Setup.SendFromEthereumToMultiversX(
		startsFromEthFlow.Setup.AliceKeys,
		startsFromEthFlow.Setup.BobKeys,
		startsFromEthFlow.Setup.MultiversxHandler.CalleeScAddress,
		startsFromEthFlow.Tokens...,
	)

	startsFromMvxFlow.Setup.SendFromMultiversxToEthereum(
		startsFromMvxFlow.Setup.AliceKeys,
		startsFromMvxFlow.Setup.BobKeys,
		startsFromMvxFlow.Tokens...,
	)

	{
		//USDC
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.Alice].OnEth = big.NewInt(2 * (-5000 - 7000 - 1000))
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.Bob].OnMvx = big.NewInt(2*(5000+7000) - 2500 - 300)
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.Charlie].OnEth = big.NewInt(2500 - 50 + 300 - 50)
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnEth = big.NewInt(2*(5000+7000+1000) - 2450 - 250)
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50 + 50)
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(2 * 1000)
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(2*(5000+7000+1000) - 2500 - 300)

		startsFromEthFlow.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.Alice].OnEth = big.NewInt(2 * (-5000 - 7000 - 1000))
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnMvx = big.NewInt(2 * (5000 + 7000 - 2500 - 300))
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.Charlie].OnEth = big.NewInt(2 * (2500 - 50 + 300 - 50))
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnEth = big.NewInt(2 * (5000 + 7000 + 1000 - 2450 - 250))
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(2 * (50 + 50))
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(2 * 1000)
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(2 * (5000 + 7000 + 1000 - 2500 - 300))

		startsFromEthFlow.Tokens[0].MintBurnChecks.MvxTotalUniversalMint = big.NewInt(2 * (5000 + 7000 + 1000))
		startsFromEthFlow.Tokens[0].MintBurnChecks.MvxTotalChainSpecificMint = big.NewInt(2 * (5000 + 7000 + 1000))
		startsFromEthFlow.Tokens[0].MintBurnChecks.MvxTotalUniversalBurn = big.NewInt(2 * (2500 + 300))
		startsFromEthFlow.Tokens[0].MintBurnChecks.MvxTotalChainSpecificBurn = big.NewInt(2 * (2500 - 50 + 300 - 50))
		startsFromEthFlow.Tokens[0].MintBurnChecks.MvxSafeMintValue = big.NewInt(2 * (5000 + 7000 + 1000))
		startsFromEthFlow.Tokens[0].MintBurnChecks.MvxSafeBurnValue = big.NewInt(2 * (2500 - 50 + 300 - 50))

		startsFromEthFlow.Tokens[0].SpecialChecks.WrapperDeltaLiquidityCheck = big.NewInt(2 * (5000 + 7000 + 1000 - 2500 - 300))
	}
	{
		//MEME
		startsFromMvxFlow.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.Alice].OnMvx = big.NewInt(2 * (-4000 - 6000 - 2000))
		startsFromMvxFlow.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.Bob].OnEth = big.NewInt(2*(4000-51+6000-51+2000-51) - 2400 - 200 - 1000)
		startsFromMvxFlow.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.Charlie].OnMvx = big.NewInt(2400 + 200)
		startsFromMvxFlow.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnMvx = big.NewInt(2*(4000+6000+2000) - 2400 - 200 - 1000)
		startsFromMvxFlow.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(1000)

		startsFromMvxFlow.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.Alice].OnMvx = big.NewInt(2 * (-4000 - 6000 - 2000))
		startsFromMvxFlow.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnEth = big.NewInt(2 * (4000 - 51 + 6000 - 51 + 2000 - 51 - 2400 - 200 - 1000))
		startsFromMvxFlow.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.Charlie].OnMvx = big.NewInt(2 * (2400 + 200))
		startsFromMvxFlow.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(2 * (4000 + 6000 + 2000 - 2400 - 200 - 1000))
		startsFromMvxFlow.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(2 * 1000)

		startsFromMvxFlow.Tokens[0].MintBurnChecks.EthSafeMintValue = big.NewInt(2 * (4000 - 51 + 6000 - 51 + 2000 - 51))
		startsFromMvxFlow.Tokens[0].MintBurnChecks.EthSafeBurnValue = big.NewInt(2 * (2400 + 200 + 1000))
	}
	{
		//TADA
		startsFromMvxFlow.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.Alice].OnMvx = big.NewInt(2 * (-5980 - 2300 - 4000))
		startsFromMvxFlow.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.Bob].OnEth = big.NewInt(2*(5980-57+2300-57+4000-57) - 3100 - 800 - 2000)
		startsFromMvxFlow.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.Charlie].OnMvx = big.NewInt(3100 + 800)
		startsFromMvxFlow.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnMvx = big.NewInt(2 * (57 + 57 + 57))
		startsFromMvxFlow.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(2000)
		startsFromMvxFlow.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(2*(-5980-2300-4000) + 3100 + 800 + 2000)

		startsFromMvxFlow.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.Alice].OnMvx = big.NewInt(2 * (-5980 - 2300 - 4000))
		startsFromMvxFlow.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnEth = big.NewInt(2 * (5980 - 57 + 2300 - 57 + 4000 - 57 - 3100 - 800 - 2000))
		startsFromMvxFlow.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.Charlie].OnMvx = big.NewInt(2 * (3100 + 800))
		startsFromMvxFlow.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(2 * (57 + 57 + 57))
		startsFromMvxFlow.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(2 * 2000)
		startsFromMvxFlow.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(2 * (-5980 - 2300 - 4000 + 3100 + 800 + 2000))

		startsFromMvxFlow.Tokens[1].MintBurnChecks.MvxTotalUniversalMint = big.NewInt(2 * (3100 + 800 + 2000))
		startsFromMvxFlow.Tokens[1].MintBurnChecks.MvxTotalChainSpecificMint = big.NewInt(2 * (3100 + 800 + 2000))
		startsFromMvxFlow.Tokens[1].MintBurnChecks.MvxTotalUniversalBurn = big.NewInt(2 * (5980 + 2300 + 4000))
		startsFromMvxFlow.Tokens[1].MintBurnChecks.MvxTotalChainSpecificBurn = big.NewInt(2 * (5980 - 57 + 2300 - 57 + 4000 - 57))
		startsFromMvxFlow.Tokens[1].MintBurnChecks.MvxSafeMintValue = big.NewInt(2 * (3100 + 800 + 2000))
		startsFromMvxFlow.Tokens[1].MintBurnChecks.MvxSafeBurnValue = big.NewInt(2 * (5980 - 57 + 2300 - 57 + 4000 - 57))

		startsFromMvxFlow.Tokens[1].MintBurnChecks.EthSafeMintValue = big.NewInt(2 * (5980 - 57 + 2300 - 57 + 4000 - 57))
		startsFromMvxFlow.Tokens[1].MintBurnChecks.EthSafeBurnValue = big.NewInt(2 * (3100 + 800 + 2000))

		startsFromMvxFlow.Tokens[1].SpecialChecks.WrapperDeltaLiquidityCheck = big.NewInt(2 * (-5980 - 2300 - 4000 + 3100 + 800 + 2000))
	}
	{
		//EUROC
		startsFromEthFlow.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.Alice].OnEth = big.NewInt(2 * (-5010 - 7010 - 1010))
		startsFromEthFlow.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.Bob].OnMvx = big.NewInt(2*(5010+7010) - 2510 - 310)
		startsFromEthFlow.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.Charlie].OnEth = big.NewInt(2510 - 52 + 310 - 52)
		startsFromEthFlow.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnMvx = big.NewInt(52 + 52)
		startsFromEthFlow.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(2 * 1010)

		startsFromEthFlow.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.Alice].OnEth = big.NewInt(2 * (-5010 - 7010 - 1010))
		startsFromEthFlow.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnMvx = big.NewInt(2 * (5010 + 7010 - 2510 - 310))
		startsFromEthFlow.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.Charlie].OnEth = big.NewInt(2 * (2510 - 52 + 310 - 52))
		startsFromEthFlow.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(2 * (52 + 52))
		startsFromEthFlow.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(2 * 1010)

		startsFromEthFlow.Tokens[1].MintBurnChecks.MvxTotalUniversalMint = big.NewInt(2 * (5010 + 7010 + 1010))
		startsFromEthFlow.Tokens[1].MintBurnChecks.MvxTotalChainSpecificMint = big.NewInt(0)
		startsFromEthFlow.Tokens[1].MintBurnChecks.MvxTotalUniversalBurn = big.NewInt(2 * (2510 - 52 + 310 - 52))
		startsFromEthFlow.Tokens[1].MintBurnChecks.MvxTotalChainSpecificBurn = big.NewInt(0)
		startsFromEthFlow.Tokens[1].MintBurnChecks.MvxSafeMintValue = big.NewInt(2 * (5010 + 7010 + 1010))
		startsFromEthFlow.Tokens[1].MintBurnChecks.MvxSafeBurnValue = big.NewInt(2 * (2510 - 52 + 310 - 52))

		startsFromEthFlow.Tokens[1].MintBurnChecks.EthSafeMintValue = big.NewInt(2 * (2510 - 52 + 310 - 52))
		startsFromEthFlow.Tokens[1].MintBurnChecks.EthSafeBurnValue = big.NewInt(2 * (5010 + 7010 + 1010))
	}
	{
		//MEX
		startsFromMvxFlow.Tokens[2].DeltaBalances[framework.FirstHalfBridge][framework.Alice].OnMvx = big.NewInt(2 * (-4010 - 6010 - 2010))
		startsFromMvxFlow.Tokens[2].DeltaBalances[framework.FirstHalfBridge][framework.Bob].OnEth = big.NewInt(2*(4010-53+6010-53+2010-53) - 2410 - 210 - 1010)
		startsFromMvxFlow.Tokens[2].DeltaBalances[framework.FirstHalfBridge][framework.Charlie].OnMvx = big.NewInt(2410 + 210)
		startsFromMvxFlow.Tokens[2].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnMvx = big.NewInt(2 * (53 + 53 + 53))
		startsFromMvxFlow.Tokens[2].DeltaBalances[framework.FirstHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(1010)

		startsFromMvxFlow.Tokens[2].DeltaBalances[framework.SecondHalfBridge][framework.Alice].OnMvx = big.NewInt(2 * (-4010 - 6010 - 2010))
		startsFromMvxFlow.Tokens[2].DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnEth = big.NewInt(2 * (4010 - 53 + 6010 - 53 + 2010 - 53 - 2410 - 210 - 1010))
		startsFromMvxFlow.Tokens[2].DeltaBalances[framework.SecondHalfBridge][framework.Charlie].OnMvx = big.NewInt(2 * (2410 + 210))
		startsFromMvxFlow.Tokens[2].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(2 * (53 + 53 + 53))
		startsFromMvxFlow.Tokens[2].DeltaBalances[framework.SecondHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(2 * 1010)

		startsFromMvxFlow.Tokens[2].MintBurnChecks.MvxTotalUniversalMint = big.NewInt(2 * (2410 + 210 + 1010))
		startsFromMvxFlow.Tokens[2].MintBurnChecks.MvxTotalChainSpecificMint = big.NewInt(0)
		startsFromMvxFlow.Tokens[2].MintBurnChecks.MvxTotalUniversalBurn = big.NewInt(2 * (4010 - 53 + 6010 - 53 + 2010 - 53))
		startsFromMvxFlow.Tokens[2].MintBurnChecks.MvxTotalChainSpecificBurn = big.NewInt(0)
		startsFromMvxFlow.Tokens[2].MintBurnChecks.MvxSafeMintValue = big.NewInt(2 * (2410 + 210 + 1010))
		startsFromMvxFlow.Tokens[2].MintBurnChecks.MvxSafeBurnValue = big.NewInt(2 * (4010 - 53 + 6010 - 53 + 2010 - 53))

		startsFromMvxFlow.Tokens[2].MintBurnChecks.EthSafeMintValue = big.NewInt(2 * (4010 - 53 + 6010 - 53 + 2010 - 53))
		startsFromMvxFlow.Tokens[2].MintBurnChecks.EthSafeBurnValue = big.NewInt(2 * (2410 + 210 + 1010))
	}
}

func executeUpgradeScenario(
	tb testing.TB,
	processFunc func(tb testing.TB, setup *framework.TestSetup) bool,
	flows []*slowTests.TestFlow,
	tokens []framework.TestTokenParams,
) *framework.TestSetup {
	setupFunc := func(tb testing.TB, setup *framework.TestSetup) {
		for _, flow := range flows {
			flow.Setup = setup
		}

		setup.IssueAndConfigureTokens(tokens...)
		setup.MultiversxHandler.CheckForZeroBalanceOnReceivers(setup.Ctx, tokens...)
		for _, flow := range flows {
			flow.HandlerToStartFirstBridge(flow)
		}
	}

	return slowTests.NewTestEnvironmentWithChainSimulator(tb,
		setupFunc,
		processFunc,
		make(chan error),
		framework.ContractsVersion3p0,
	)
}
