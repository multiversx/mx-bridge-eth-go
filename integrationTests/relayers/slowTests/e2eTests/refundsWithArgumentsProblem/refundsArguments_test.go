package refundsWithArgumentsProblem

import (
	"strings"
	"testing"

	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/e2eTests"
)

func TestRefundWithExtraArgument(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	callData := e2eTests.CreateScCallData("callPayable", 50000000, "extra parameter")
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

func TestRefundWithNoArguments(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	callData := e2eTests.CreateScCallData("callPayableWithParams", 50000000)
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

func TestRefundWithWrongNumberOfArguments(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	callData := e2eTests.CreateScCallData("callPayableWithParams", 50000000, string([]byte{37}))
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

func TestRefundWithAnArgumentThatIsNotUint64(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	malformedUint64String := string([]byte{37, 36, 35, 34, 33, 32, 31, 32, 33}) // 9 bytes instead of 8
	dummyAddress := strings.Repeat("2", 32)

	callData := e2eTests.CreateScCallData("callPayableWithParams", 50000000, malformedUint64String, dummyAddress)
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

func TestRefundWithWrongArgumentsEncoding(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	callData := e2eTests.CreateScCallData("callPayableWithParams", 50000000)
	// the last byte is the data missing marker, we will replace that
	callData[len(callData)-1] = bridgeCore.DataPresentProtocolMarker
	// add garbage data
	callData = append(callData, []byte{5, 4, 55}...)

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
