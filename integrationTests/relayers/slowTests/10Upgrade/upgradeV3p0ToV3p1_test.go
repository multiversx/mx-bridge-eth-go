// TODO: fix this test
package upgrade

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
)

const testMarker = "==================================== %s ===================================="

func TestComplexScenarioWithMissingTransferRole(t *testing.T) {
	tokens := []framework.TestTokenParams{
		GenerateTestUSDCToken(),
		//GenerateTestMEMEToken(),
		//GenerateTestTADAToken(),
		//GenerateTestEUROCToken(),
		//GenerateTestMEXToken(),
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

	for _, flow := range flows {
		flow.HandlerToStartFirstBridge(flow)
	}

	startsFromEthFlow := flows[0]
	// startsFromMvxFlow := flows[1]

	{
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.Alice].OnEth = big.NewInt(2 * (-5000 - 7000 - 1000))
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.Bob].OnMvx = big.NewInt(2*(5000+7000) - 2500)
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.Charlie].OnEth = big.NewInt(2500 - 50 + 300 - 50)
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnEth = big.NewInt(2*(5000+7000+1000) - 2450 - 250)
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50 + 50)
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(2 * 1000)
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(2*(5000+7000+1000) - 2500 - 300)

		startsFromEthFlow.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.Alice].OnEth = big.NewInt(2 * (-5000 - 7000 - 1000))
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnMvx = big.NewInt(2 * (5000 + 7000 - 2500))
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.Charlie].OnEth = big.NewInt(2 * (2500 - 50 + 300 - 50))
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnEth = big.NewInt(2 * (5000 + 7000 + 1000 - 2450 - 250))
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(2 * (50 + 50))
		startsFromEthFlow.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(2 * (5000 + 7000 + 1000 - 2500 - 300))

		startsFromEthFlow.Tokens[0].MintBurnChecks.MvxTotalUniversalMint = big.NewInt(5000 + 5000)
		startsFromEthFlow.Tokens[0].MintBurnChecks.MvxTotalChainSpecificMint = big.NewInt(5000 + 5000)
		startsFromEthFlow.Tokens[0].MintBurnChecks.MvxTotalUniversalBurn = big.NewInt(2500 + 2500)
		startsFromEthFlow.Tokens[0].MintBurnChecks.MvxTotalChainSpecificBurn = big.NewInt(2500 - 50 + 2500 - 50)
		startsFromEthFlow.Tokens[0].MintBurnChecks.MvxSafeMintValue = big.NewInt(5000 + 5000)
		startsFromEthFlow.Tokens[0].MintBurnChecks.MvxSafeBurnValue = big.NewInt(2500 - 50 + 2500 - 50)

		startsFromEthFlow.Tokens[0].SpecialChecks.WrapperDeltaLiquidityCheck = big.NewInt(5000 - 2500 + 5000 - 2500)
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
