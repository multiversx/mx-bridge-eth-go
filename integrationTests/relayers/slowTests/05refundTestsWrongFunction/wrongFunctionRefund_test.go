//go:build slow

package refundTestsWrongFunction

import (
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
)

func TestRelayersShouldExecuteTransfersWithRefundForWrongFunction(t *testing.T) {
	t.Run("unknown function should refund", func(t *testing.T) {
		callData := slowTests.CreateScCallData("unknownFunction", 50000000)
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
	t.Run("not a valid function name should refund", func(t *testing.T) {
		callData := slowTests.CreateScCallData("=", 50000000)
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
	t.Run("empty function with no args should refund", func(t *testing.T) {
		callData := slowTests.CreateScCallData("", 50000000)
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
	t.Run("empty function with args should refund", func(t *testing.T) {
		dummyAddress := strings.Repeat("2", 32)
		callData := slowTests.CreateScCallData("", 50000000, dummyAddress)
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
	t.Run("built-in function should refund", func(t *testing.T) {
		callData := slowTests.CreateScCallData("SaveKeyValue", 50000000, "6b657930", "76616c756530")
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
	t.Run("not a valid function name and small gas limit should refund", func(t *testing.T) {
		callData := slowTests.CreateScCallData("=", 1)

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
