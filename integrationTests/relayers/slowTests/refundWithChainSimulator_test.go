//go:build slow

// To run these slow tests, simply add the slow tag on the go test command. Also, provide a chain simulator instance on the 8085 port
// example: go test -tags slow

package slowTests

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
)

func TestRelayersShouldExecuteTransfersWithRefund(t *testing.T) {
	usdcToken := GenerateTestUSDCToken()
	usdcToken.TestOperations[2].MvxSCCallMethod = "unknownFunction"
	usdcToken.TestOperations[2].MvxFaultySCCall = true
	usdcToken.EthTestAddrExtraBalance = big.NewInt(-5000 + 2500 - 50 - 7000 + 300 - 50 - 1000 + 950) // -(eth->mvx) + (mvx->eth) - fees + revert after bad SC call
	usdcToken.ESDTSafeExtraBalance = big.NewInt(150)                                                 // extra is just for the fees for the 2 transfers mvx->eth and the failed eth->mvx that needed refund

	memeToken := GenerateTestMEMEToken()
	memeToken.TestOperations[2].MvxSCCallMethod = "unknownFunction"
	memeToken.TestOperations[2].MvxFaultySCCall = true

	testRelayersWithChainSimulatorAndTokensAndRefund(
		t,
		make(chan error),
		usdcToken,
		memeToken,
	)
}

func testRelayersWithChainSimulatorAndTokensAndRefund(tb testing.TB, manualStopChan chan error, tokens ...framework.TestTokenParams) {
	startsFromEthFlow, startsFromMvXFlow := createFlowsBasedOnToken(tb, tokens...)

	setupFunc := func(tb testing.TB, setup *framework.TestSetup) {
		startsFromMvXFlow.setup = setup
		startsFromEthFlow.setup = setup

		setup.IssueAndConfigureTokens(tokens...)
		setup.MultiversxHandler.CheckForZeroBalanceOnReceivers(setup.Ctx, tokens...)
		if len(startsFromEthFlow.tokens) > 0 {
			setup.EthereumHandler.CreateBatchOnEthereum(setup.Ctx, setup.MultiversxHandler.TestCallerAddress, startsFromEthFlow.tokens...)
		}
		if len(startsFromMvXFlow.tokens) > 0 {
			setup.CreateBatchOnMultiversX(startsFromMvXFlow.tokens...)
		}
	}

	processFunc := func(tb testing.TB, setup *framework.TestSetup) bool {
		if startsFromEthFlow.process() && startsFromMvXFlow.process() && startsFromMvXFlow.areTokensFullyRefunded() {
			return true
		}

		// commit blocks in order to execute incoming txs from relayers
		setup.EthereumHandler.SimulatedChain.Commit()
		setup.ChainSimulator.GenerateBlocks(setup.Ctx, 1)

		return false
	}

	testRelayersWithChainSimulator(tb,
		setupFunc,
		processFunc,
		manualStopChan,
	)
}