//go:build slow

// To run these slow tests, simply add the slow tag on the go test command. Also, provide a chain simulator instance on the 8085 port
// example: go test -tags slow

package slowTests

import (
	"strings"
	"testing"

	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
	"github.com/stretchr/testify/require"
)

func TestRelayersShouldExecuteTransfersWithRefund(t *testing.T) {
	t.Run("unknown marker and malformed SC call data should refund", func(t *testing.T) {
		callData := []byte{5, 4, 55}
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		ApplyUSDCRefundBalances(&usdcToken)

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEMERefundBalances(&memeToken)

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("unknown marker and malformed SC call data should refund with MEX", func(t *testing.T) {
		callData := []byte{5, 4, 55}
		mexToken := GenerateTestMEXToken()
		mexToken.TestOperations[2].MvxSCCallData = callData
		mexToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEXRefundBalances(&mexToken)

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			mexToken,
		)
	})
	t.Run("malformed SC call data should refund", func(t *testing.T) {
		callData := []byte{bridgeCore.DataPresentProtocolMarker, 4, 55}
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		ApplyUSDCRefundBalances(&usdcToken)

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEMERefundBalances(&memeToken)

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("unknown function should refund", func(t *testing.T) {
		callData := createScCallData("unknownFunction", 50000000)
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		ApplyUSDCRefundBalances(&usdcToken)

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEMERefundBalances(&memeToken)

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("wrong deposit with empty sc call data should refund", func(t *testing.T) {
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = nil
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		usdcToken.TestOperations[2].MvxForceSCCall = true
		ApplyUSDCRefundBalances(&usdcToken)

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = nil
		memeToken.TestOperations[2].MvxFaultySCCall = true
		memeToken.TestOperations[2].MvxForceSCCall = true
		ApplyMEMERefundBalances(&memeToken)

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("0 gas limit should refund", func(t *testing.T) {
		callData := createScCallData("callPayable", 0)
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		ApplyUSDCRefundBalances(&usdcToken)

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEMERefundBalances(&memeToken)

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("small gas limit should refund", func(t *testing.T) {
		callData := createScCallData("callPayable", 2000)
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		ApplyUSDCRefundBalances(&usdcToken)

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEMERefundBalances(&memeToken)

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("extra parameter should refund", func(t *testing.T) {
		callData := createScCallData("callPayable", 50000000, "extra parameter")
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		ApplyUSDCRefundBalances(&usdcToken)

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEMERefundBalances(&memeToken)

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("no arguments should refund", func(t *testing.T) {
		callData := createScCallData("callPayableWithParams", 50000000)
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		ApplyUSDCRefundBalances(&usdcToken)

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEMERefundBalances(&memeToken)

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("wrong number of arguments should refund", func(t *testing.T) {
		callData := createScCallData("callPayableWithParams", 50000000, string([]byte{37}))
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		ApplyUSDCRefundBalances(&usdcToken)

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEMERefundBalances(&memeToken)

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("not an uint64 argument should refund", func(t *testing.T) {
		malformedUint64String := string([]byte{37, 36, 35, 34, 33, 32, 31, 32, 33}) // 9 bytes instead of 8
		dummyAddress := strings.Repeat("2", 32)

		callData := createScCallData("callPayableWithParams", 50000000, malformedUint64String, dummyAddress)
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		ApplyUSDCRefundBalances(&usdcToken)

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEMERefundBalances(&memeToken)

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("wrong arguments encoding should refund", func(t *testing.T) {
		callData := createScCallData("callPayableWithParams", 50000000)
		// the last byte is the data missing marker, we will replace that
		callData[len(callData)-1] = bridgeCore.DataPresentProtocolMarker
		// add garbage data
		callData = append(callData, []byte{5, 4, 55}...)

		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		ApplyUSDCRefundBalances(&usdcToken)

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEMERefundBalances(&memeToken)

		testRelayersWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
}

func testRelayersWithChainSimulatorAndTokensAndRefund(tb testing.TB, manualStopChan chan error, tokens ...framework.TestTokenParams) {
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
		require.LessOrEqual(tb, setup.ScCallerModuleInstance.GetNumSentTransaction(), setup.GetNumScCallsOperations())

		return allFlowsFinished
	}

	_ = testRelayersWithChainSimulator(tb,
		setupFunc,
		processFunc,
		manualStopChan,
	)
}
