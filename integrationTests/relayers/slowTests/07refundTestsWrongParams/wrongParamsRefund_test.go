//go:build slow

package refundTestsWrongParams

import (
	"strings"
	"testing"

	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests"
)

func TestRelayersShouldExecuteTransfersWithRefundForWrongParameters(t *testing.T) {
	t.Run("extra parameter should refund", func(t *testing.T) {
		callData := slowTests.CreateScCallData("callPayable", 50000000, "extra parameter")
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
	t.Run("no arguments should refund", func(t *testing.T) {
		callData := slowTests.CreateScCallData("callPayableWithParams", 50000000)
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
	t.Run("wrong number of arguments should refund", func(t *testing.T) {
		callData := slowTests.CreateScCallData("callPayableWithParams", 50000000, string([]byte{37}))
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
	t.Run("not an uint64 argument should refund", func(t *testing.T) {
		malformedUint64String := string([]byte{37, 36, 35, 34, 33, 32, 31, 32, 33}) // 9 bytes instead of 8
		dummyAddress := strings.Repeat("2", 32)

		callData := slowTests.CreateScCallData("callPayableWithParams", 50000000, malformedUint64String, dummyAddress)
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
	t.Run("wrong arguments encoding should refund", func(t *testing.T) {
		callData := slowTests.CreateScCallData("callPayableWithParams", 50000000)
		// the last byte is the data missing marker, we will replace that
		callData[len(callData)-1] = bridgeCore.DataPresentProtocolMarker
		// add garbage data
		callData = append(callData, []byte{5, 4, 55}...)

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
