//go:build slow

package refundTestsWrongGasLimit

import (
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests"
)

func TestRelayersShouldExecuteTransfersWithRefundForWrongGasLimit(t *testing.T) {
	t.Run("0 gas limit should refund", func(t *testing.T) {
		callData := slowTests.CreateScCallData("callPayable", 0)
		usdcToken := slowTests.GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		slowTests.ApplyUSDCRefundBalances(&usdcToken)

		memeToken := slowTests.GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		slowTests.ApplyMEMERefundBalances(&memeToken)

		tadaToken := slowTests.GenerateTestTADAToken()
		tadaToken.TestOperations[2].MvxSCCallData = callData
		tadaToken.TestOperations[2].MvxFaultySCCall = true
		slowTests.ApplyTADARefundBalances(&tadaToken)

		eurocToken := slowTests.GenerateTestEUROCToken()
		eurocToken.TestOperations[2].MvxSCCallData = callData
		eurocToken.TestOperations[2].MvxFaultySCCall = true
		slowTests.ApplyEUROCRefundBalances(&eurocToken)

		mexToken := slowTests.GenerateTestMEXToken()
		mexToken.TestOperations[2].MvxSCCallData = callData
		mexToken.TestOperations[2].MvxFaultySCCall = true
		slowTests.ApplyMEXRefundBalances(&mexToken)

		slowTests.NewTestEnvironmentWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
			tadaToken,
			eurocToken,
			mexToken,
		)
	})
	t.Run("small gas limit should refund", func(t *testing.T) {
		callData := slowTests.CreateScCallData("callPayable", 2000)
		usdcToken := slowTests.GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		slowTests.ApplyUSDCRefundBalances(&usdcToken)

		memeToken := slowTests.GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		slowTests.ApplyMEMERefundBalances(&memeToken)

		tadaToken := slowTests.GenerateTestTADAToken()
		tadaToken.TestOperations[2].MvxSCCallData = callData
		tadaToken.TestOperations[2].MvxFaultySCCall = true
		slowTests.ApplyTADARefundBalances(&tadaToken)

		eurocToken := slowTests.GenerateTestEUROCToken()
		eurocToken.TestOperations[2].MvxSCCallData = callData
		eurocToken.TestOperations[2].MvxFaultySCCall = true
		slowTests.ApplyEUROCRefundBalances(&eurocToken)

		mexToken := slowTests.GenerateTestMEXToken()
		mexToken.TestOperations[2].MvxSCCallData = callData
		mexToken.TestOperations[2].MvxFaultySCCall = true
		slowTests.ApplyMEXRefundBalances(&mexToken)

		slowTests.NewTestEnvironmentWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
			tadaToken,
			eurocToken,
			mexToken,
		)
	})
	t.Run("high gas limit should refund", func(t *testing.T) {
		callData := slowTests.CreateScCallData("callPayable", 610000000)
		usdcToken := slowTests.GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		slowTests.ApplyUSDCRefundBalances(&usdcToken)

		memeToken := slowTests.GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		slowTests.ApplyMEMERefundBalances(&memeToken)

		tadaToken := slowTests.GenerateTestTADAToken()
		tadaToken.TestOperations[2].MvxSCCallData = callData
		tadaToken.TestOperations[2].MvxFaultySCCall = true
		slowTests.ApplyTADARefundBalances(&tadaToken)

		eurocToken := slowTests.GenerateTestEUROCToken()
		eurocToken.TestOperations[2].MvxSCCallData = callData
		eurocToken.TestOperations[2].MvxFaultySCCall = true
		slowTests.ApplyEUROCRefundBalances(&eurocToken)

		mexToken := slowTests.GenerateTestMEXToken()
		mexToken.TestOperations[2].MvxSCCallData = callData
		mexToken.TestOperations[2].MvxFaultySCCall = true
		slowTests.ApplyMEXRefundBalances(&mexToken)

		slowTests.NewTestEnvironmentWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
			tadaToken,
			eurocToken,
			mexToken,
		)
	})
}
