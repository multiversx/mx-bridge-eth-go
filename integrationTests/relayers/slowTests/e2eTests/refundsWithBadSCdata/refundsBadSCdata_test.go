package refundsWithBadSCdata

import (
	"testing"

	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/e2eTests"
)

func TestRefundWithUnknownMarkerAndMalformedSCcallData(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	callData := []byte{5, 4, 55}
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

func TestRefundWithMalformedSCcallData(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	callData := []byte{bridgeCore.DataPresentProtocolMarker, 4, 55}
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

func TestRefundWithEmptySCCallData(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	usdcToken := e2eTests.GenerateTestUSDCToken()
	usdcToken.TestOperations[2].MvxSCCallData = nil
	usdcToken.TestOperations[2].MvxFaultySCCall = true
	usdcToken.TestOperations[2].MvxForceSCCall = true
	e2eTests.ApplyUSDCRefundBalances(&usdcToken)

	memeToken := e2eTests.GenerateTestMEMEToken()
	memeToken.TestOperations[2].MvxSCCallData = nil
	memeToken.TestOperations[2].MvxFaultySCCall = true
	memeToken.TestOperations[2].MvxForceSCCall = true
	e2eTests.ApplyMEMERefundBalances(&memeToken)

	eurocToken := e2eTests.GenerateTestEUROCToken()
	eurocToken.TestOperations[2].MvxSCCallData = nil
	eurocToken.TestOperations[2].MvxFaultySCCall = true
	eurocToken.TestOperations[2].MvxForceSCCall = true
	e2eTests.ApplyEUROCRefundBalances(&eurocToken)

	mexToken := e2eTests.GenerateTestMEXToken()
	mexToken.TestOperations[2].MvxSCCallData = nil
	mexToken.TestOperations[2].MvxFaultySCCall = true
	mexToken.TestOperations[2].MvxForceSCCall = true
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
