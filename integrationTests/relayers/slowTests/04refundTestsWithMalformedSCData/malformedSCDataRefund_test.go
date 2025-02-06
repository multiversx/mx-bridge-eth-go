//go:build slow

package refundTestsWithMalformedSCData

import (
	"crypto/rand"
	"math/big"
	"strings"
	"testing"

	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
)

func TestRelayersShouldExecuteTransfersWithRefundForMalformedSCData(t *testing.T) {
	t.Run("unknown marker and malformed SC call data should refund", func(t *testing.T) {
		callData := []byte{5, 4, 55}
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
	t.Run("malformed SC call data should refund", func(t *testing.T) {
		callData := []byte{bridgeCore.DataPresentProtocolMarker, 4, 55}
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
	t.Run("wrong deposit with empty sc call data should refund", func(t *testing.T) {
		usdcToken := slowTests.GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = nil
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		usdcToken.TestOperations[2].MvxForceSCCall = true
		slowTests.ApplyUSDCRefundBalances(&usdcToken)

		memeToken := slowTests.GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = nil
		memeToken.TestOperations[2].MvxFaultySCCall = true
		memeToken.TestOperations[2].MvxForceSCCall = true
		slowTests.ApplyMEMERefundBalances(&memeToken)

		tadaToken := slowTests.GenerateTestTADAToken()
		tadaToken.TestOperations[2].MvxSCCallData = nil
		tadaToken.TestOperations[2].MvxFaultySCCall = true
		tadaToken.TestOperations[2].MvxForceSCCall = true
		slowTests.ApplyTADARefundBalances(&tadaToken)

		eurocToken := slowTests.GenerateTestEUROCToken()
		eurocToken.TestOperations[2].MvxSCCallData = nil
		eurocToken.TestOperations[2].MvxFaultySCCall = true
		eurocToken.TestOperations[2].MvxForceSCCall = true
		slowTests.ApplyEUROCRefundBalances(&eurocToken)

		mexToken := slowTests.GenerateTestMEXToken()
		mexToken.TestOperations[2].MvxSCCallData = nil
		mexToken.TestOperations[2].MvxFaultySCCall = true
		mexToken.TestOperations[2].MvxForceSCCall = true
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
	t.Run("input too short on encoding should refund", func(t *testing.T) {
		dummyAddress := strings.Repeat("2", 32)
		dummyUint64 := string([]byte{37})

		callData := slowTests.CreateScCallData("callPayableWithParams", 50000000, dummyUint64, dummyAddress)
		callData[3] += 60 // we simulate that the buffer was longer, but it was trimmed

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
	t.Run("working with a large buffer, should refund", func(t *testing.T) {
		ethereumLimit := 130727
		buff := make([]byte, ethereumLimit)
		_, _ = rand.Read(buff)
		callData := slowTests.CreateScCallData("callPayableWithBuff", 100000000, string(buff))

		usdcToken := slowTests.GenerateTestUSDCToken()
		usdcToken.TestOperations = []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(500),
				ValueToSendFromMvX:   nil,
				MvxSCCallData:        callData,
				MvxFaultySCCall:      true,
			},
			{
				ValueToTransferToMvx: big.NewInt(600),
				ValueToSendFromMvX:   nil,
				MvxSCCallData:        callData,
				MvxFaultySCCall:      true,
			},
		}
		usdcToken.DeltaBalances = map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-500 - 600),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(500 + 600),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(500 + 600),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
			},
			framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-500 + 450 - 600 + 550),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(50 + 50),
					OnMvx:    big.NewInt(50 + 50),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
			},
		}
		usdcToken.MintBurnChecks = &framework.MintBurnBalances{
			MvxTotalUniversalMint:     big.NewInt(500 + 600),
			MvxTotalChainSpecificMint: big.NewInt(500 + 600),
			MvxTotalUniversalBurn:     big.NewInt(500 + 600),
			MvxTotalChainSpecificBurn: big.NewInt(450 + 550),
			MvxSafeMintValue:          big.NewInt(500 + 600),
			MvxSafeBurnValue:          big.NewInt(450 + 550),

			EthSafeMintValue: big.NewInt(0),
			EthSafeBurnValue: big.NewInt(0),
		}
		usdcToken.SpecialChecks.WrapperDeltaLiquidityCheck = big.NewInt(0)

		slowTests.NewTestEnvironmentWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
		)
	})
}
