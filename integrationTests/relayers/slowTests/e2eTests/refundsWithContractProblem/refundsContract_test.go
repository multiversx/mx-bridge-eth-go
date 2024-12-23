package refundsWithContractProblem

import (
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/e2eTests"
	"github.com/multiversx/mx-sdk-go/data"
)

func TestRefundWithUninitializedContract(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	callData := e2eTests.CreateScCallData("claim", 50000000)
	uninitializedSCAddressBytes, _ := data.NewAddressFromBech32String("erd1qqqqqqqqqqqqqpgqcc69ts8409p3h77q5chsaqz57y6hugvc4fvs64k74v")
	usdcToken := e2eTests.GenerateTestUSDCToken()
	usdcToken.TestOperations[2].MvxSCCallData = callData
	usdcToken.TestOperations[2].MvxFaultySCCall = true
	usdcToken.TestOperations[2].InvalidReceiver = uninitializedSCAddressBytes.AddressBytes()
	e2eTests.ApplyUSDCRefundBalances(&usdcToken)

	memeToken := e2eTests.GenerateTestMEMEToken()
	memeToken.TestOperations[2].MvxSCCallData = callData
	memeToken.TestOperations[2].MvxFaultySCCall = true
	memeToken.TestOperations[2].InvalidReceiver = uninitializedSCAddressBytes.AddressBytes()
	e2eTests.ApplyMEMERefundBalances(&memeToken)

	eurocToken := e2eTests.GenerateTestEUROCToken()
	eurocToken.TestOperations[2].MvxSCCallData = callData
	eurocToken.TestOperations[2].MvxFaultySCCall = true
	eurocToken.TestOperations[2].InvalidReceiver = uninitializedSCAddressBytes.AddressBytes()
	e2eTests.ApplyEUROCRefundBalances(&eurocToken)

	mexToken := e2eTests.GenerateTestMEXToken()
	mexToken.TestOperations[2].MvxSCCallData = callData
	mexToken.TestOperations[2].MvxFaultySCCall = true
	mexToken.TestOperations[2].InvalidReceiver = uninitializedSCAddressBytes.AddressBytes()
	e2eTests.ApplyMEXRefundBalances(&mexToken)

	e2eTests.TestRelayersWithChainSimulatorAndTokensAndRefund(
		t,
		make(chan error),
		usdcToken,
		memeToken,
		eurocToken,
		mexToken,
	)
}
