package refundsWithGasLimitProblem

import (
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/e2eTests"
)

func TestRefundWith0GasLimit(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	callData := e2eTests.CreateScCallData("callPayable", 0)
	usdcToken := e2eTests.GenerateTestUSDCToken()
	usdcToken.TestOperations[2].MvxSCCallData = callData
	usdcToken.TestOperations[2].MvxFaultySCCall = true
	e2eTests.ApplyUSDCRefundBalances(&usdcToken)

	memeToken := e2eTests.GenerateTestMEMEToken()
	memeToken.TestOperations[2].MvxSCCallData = callData
	memeToken.TestOperations[2].MvxFaultySCCall = true
	e2eTests.ApplyMEMERefundBalances(&memeToken)

	eurocToken := e2eTests.GenerateTestEUROCToken()
	eurocToken.TestOperations[2].MvxSCCallData = callData
	eurocToken.TestOperations[2].MvxFaultySCCall = true
	e2eTests.ApplyEUROCRefundBalances(&eurocToken)

	mexToken := e2eTests.GenerateTestMEXToken()
	mexToken.TestOperations[2].MvxSCCallData = callData
	mexToken.TestOperations[2].MvxFaultySCCall = true
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

func TestRefundWithSmallGasLimit(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	callData := e2eTests.CreateScCallData("callPayable", 2000)
	usdcToken := e2eTests.GenerateTestUSDCToken()
	usdcToken.TestOperations[2].MvxSCCallData = callData
	usdcToken.TestOperations[2].MvxFaultySCCall = true
	e2eTests.ApplyUSDCRefundBalances(&usdcToken)

	memeToken := e2eTests.GenerateTestMEMEToken()
	memeToken.TestOperations[2].MvxSCCallData = callData
	memeToken.TestOperations[2].MvxFaultySCCall = true
	e2eTests.ApplyMEMERefundBalances(&memeToken)

	eurocToken := e2eTests.GenerateTestEUROCToken()
	eurocToken.TestOperations[2].MvxSCCallData = callData
	eurocToken.TestOperations[2].MvxFaultySCCall = true
	e2eTests.ApplyEUROCRefundBalances(&eurocToken)

	mexToken := e2eTests.GenerateTestMEXToken()
	mexToken.TestOperations[2].MvxSCCallData = callData
	mexToken.TestOperations[2].MvxFaultySCCall = true
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

func TestRefundWithHighGasLimit(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	callData := e2eTests.CreateScCallData("callPayable", 610000000)
	usdcToken := e2eTests.GenerateTestUSDCToken()
	usdcToken.TestOperations[2].MvxSCCallData = callData
	usdcToken.TestOperations[2].MvxFaultySCCall = true
	e2eTests.ApplyUSDCRefundBalances(&usdcToken)

	memeToken := e2eTests.GenerateTestMEMEToken()
	memeToken.TestOperations[2].MvxSCCallData = callData
	memeToken.TestOperations[2].MvxFaultySCCall = true
	e2eTests.ApplyMEMERefundBalances(&memeToken)

	eurocToken := e2eTests.GenerateTestEUROCToken()
	eurocToken.TestOperations[2].MvxSCCallData = callData
	eurocToken.TestOperations[2].MvxFaultySCCall = true
	e2eTests.ApplyEUROCRefundBalances(&eurocToken)

	mexToken := e2eTests.GenerateTestMEXToken()
	mexToken.TestOperations[2].MvxSCCallData = callData
	mexToken.TestOperations[2].MvxFaultySCCall = true
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
