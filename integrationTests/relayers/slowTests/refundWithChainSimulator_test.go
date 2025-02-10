//go:build slow

// To run these slow tests, simply add the slow tag on the go test command. Also, provide a chain simulator instance on the 8085 port
// example: go test -tags slow

package slowTests

import (
	"math/big"
	"strings"
	"testing"

	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
	"github.com/multiversx/mx-sdk-go/data"
)

func TestRelayersShouldExecuteTransfersWithRefund(t *testing.T) {
	t.Run("uninitialized contract should refund", func(t *testing.T) {
		callData := CreateScCallData("claim", 50000000)
		uninitializedSCAddressBytes, _ := data.NewAddressFromBech32String("erd1qqqqqqqqqqqqqpgqcc69ts8409p3h77q5chsaqz57y6hugvc4fvs64k74v")
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		usdcToken.TestOperations[2].InvalidReceiver = uninitializedSCAddressBytes.AddressBytes()
		ApplyUSDCRefundBalances(&usdcToken)

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		memeToken.TestOperations[2].InvalidReceiver = uninitializedSCAddressBytes.AddressBytes()
		ApplyMEMERefundBalances(&memeToken)

		tadaToken := GenerateTestTADAToken()
		tadaToken.TestOperations[2].MvxSCCallData = callData
		tadaToken.TestOperations[2].MvxFaultySCCall = true
		tadaToken.TestOperations[2].InvalidReceiver = uninitializedSCAddressBytes.AddressBytes()
		ApplyTADARefundBalances(&tadaToken)

		eurocToken := GenerateTestEUROCToken()
		eurocToken.TestOperations[2].MvxSCCallData = callData
		eurocToken.TestOperations[2].MvxFaultySCCall = true
		eurocToken.TestOperations[2].InvalidReceiver = uninitializedSCAddressBytes.AddressBytes()
		ApplyEUROCRefundBalances(&eurocToken)

		mexToken := GenerateTestMEXToken()
		mexToken.TestOperations[2].MvxSCCallData = callData
		mexToken.TestOperations[2].MvxFaultySCCall = true
		mexToken.TestOperations[2].InvalidReceiver = uninitializedSCAddressBytes.AddressBytes()
		ApplyMEXRefundBalances(&mexToken)

		NewTestEnvironmentWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
			tadaToken,
			eurocToken,
			mexToken,
		)
	})
	t.Run("0 gas limit should refund", func(t *testing.T) {
		callData := CreateScCallData("callPayable", 0)
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		ApplyUSDCRefundBalances(&usdcToken)

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEMERefundBalances(&memeToken)

		tadaToken := GenerateTestTADAToken()
		tadaToken.TestOperations[2].MvxSCCallData = callData
		tadaToken.TestOperations[2].MvxFaultySCCall = true
		ApplyTADARefundBalances(&tadaToken)

		eurocToken := GenerateTestEUROCToken()
		eurocToken.TestOperations[2].MvxSCCallData = callData
		eurocToken.TestOperations[2].MvxFaultySCCall = true
		ApplyEUROCRefundBalances(&eurocToken)

		mexToken := GenerateTestMEXToken()
		mexToken.TestOperations[2].MvxSCCallData = callData
		mexToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEXRefundBalances(&mexToken)

		NewTestEnvironmentWithChainSimulatorAndTokensAndRefund(
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
		callData := CreateScCallData("callPayable", 2000)
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		ApplyUSDCRefundBalances(&usdcToken)

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEMERefundBalances(&memeToken)

		tadaToken := GenerateTestTADAToken()
		tadaToken.TestOperations[2].MvxSCCallData = callData
		tadaToken.TestOperations[2].MvxFaultySCCall = true
		ApplyTADARefundBalances(&tadaToken)

		eurocToken := GenerateTestEUROCToken()
		eurocToken.TestOperations[2].MvxSCCallData = callData
		eurocToken.TestOperations[2].MvxFaultySCCall = true
		ApplyEUROCRefundBalances(&eurocToken)

		mexToken := GenerateTestMEXToken()
		mexToken.TestOperations[2].MvxSCCallData = callData
		mexToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEXRefundBalances(&mexToken)

		NewTestEnvironmentWithChainSimulatorAndTokensAndRefund(
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
		callData := CreateScCallData("callPayable", 610000000)
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		ApplyUSDCRefundBalances(&usdcToken)

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEMERefundBalances(&memeToken)

		tadaToken := GenerateTestTADAToken()
		tadaToken.TestOperations[2].MvxSCCallData = callData
		tadaToken.TestOperations[2].MvxFaultySCCall = true
		ApplyTADARefundBalances(&tadaToken)

		eurocToken := GenerateTestEUROCToken()
		eurocToken.TestOperations[2].MvxSCCallData = callData
		eurocToken.TestOperations[2].MvxFaultySCCall = true
		ApplyEUROCRefundBalances(&eurocToken)

		mexToken := GenerateTestMEXToken()
		mexToken.TestOperations[2].MvxSCCallData = callData
		mexToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEXRefundBalances(&mexToken)

		NewTestEnvironmentWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
			tadaToken,
			eurocToken,
			mexToken,
		)
	})
	t.Run("extra parameter should refund", func(t *testing.T) {
		callData := CreateScCallData("callPayable", 50000000, "extra parameter")
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		ApplyUSDCRefundBalances(&usdcToken)

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEMERefundBalances(&memeToken)

		tadaToken := GenerateTestTADAToken()
		tadaToken.TestOperations[2].MvxSCCallData = callData
		tadaToken.TestOperations[2].MvxFaultySCCall = true
		ApplyTADARefundBalances(&tadaToken)

		eurocToken := GenerateTestEUROCToken()
		eurocToken.TestOperations[2].MvxSCCallData = callData
		eurocToken.TestOperations[2].MvxFaultySCCall = true
		ApplyEUROCRefundBalances(&eurocToken)

		mexToken := GenerateTestMEXToken()
		mexToken.TestOperations[2].MvxSCCallData = callData
		mexToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEXRefundBalances(&mexToken)

		NewTestEnvironmentWithChainSimulatorAndTokensAndRefund(
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
		callData := CreateScCallData("callPayableWithParams", 50000000)
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		ApplyUSDCRefundBalances(&usdcToken)

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEMERefundBalances(&memeToken)

		tadaToken := GenerateTestTADAToken()
		tadaToken.TestOperations[2].MvxSCCallData = callData
		tadaToken.TestOperations[2].MvxFaultySCCall = true
		ApplyTADARefundBalances(&tadaToken)

		eurocToken := GenerateTestEUROCToken()
		eurocToken.TestOperations[2].MvxSCCallData = callData
		eurocToken.TestOperations[2].MvxFaultySCCall = true
		ApplyEUROCRefundBalances(&eurocToken)

		mexToken := GenerateTestMEXToken()
		mexToken.TestOperations[2].MvxSCCallData = callData
		mexToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEXRefundBalances(&mexToken)

		NewTestEnvironmentWithChainSimulatorAndTokensAndRefund(
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
		callData := CreateScCallData("callPayableWithParams", 50000000, string([]byte{37}))
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		ApplyUSDCRefundBalances(&usdcToken)

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEMERefundBalances(&memeToken)

		tadaToken := GenerateTestTADAToken()
		tadaToken.TestOperations[2].MvxSCCallData = callData
		tadaToken.TestOperations[2].MvxFaultySCCall = true
		ApplyTADARefundBalances(&tadaToken)

		eurocToken := GenerateTestEUROCToken()
		eurocToken.TestOperations[2].MvxSCCallData = callData
		eurocToken.TestOperations[2].MvxFaultySCCall = true
		ApplyEUROCRefundBalances(&eurocToken)

		mexToken := GenerateTestMEXToken()
		mexToken.TestOperations[2].MvxSCCallData = callData
		mexToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEXRefundBalances(&mexToken)

		NewTestEnvironmentWithChainSimulatorAndTokensAndRefund(
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

		callData := CreateScCallData("callPayableWithParams", 50000000, malformedUint64String, dummyAddress)
		usdcToken := GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData
		usdcToken.TestOperations[2].MvxFaultySCCall = true
		ApplyUSDCRefundBalances(&usdcToken)

		memeToken := GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData
		memeToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEMERefundBalances(&memeToken)

		tadaToken := GenerateTestTADAToken()
		tadaToken.TestOperations[2].MvxSCCallData = callData
		tadaToken.TestOperations[2].MvxFaultySCCall = true
		ApplyTADARefundBalances(&tadaToken)

		eurocToken := GenerateTestEUROCToken()
		eurocToken.TestOperations[2].MvxSCCallData = callData
		eurocToken.TestOperations[2].MvxFaultySCCall = true
		ApplyEUROCRefundBalances(&eurocToken)

		mexToken := GenerateTestMEXToken()
		mexToken.TestOperations[2].MvxSCCallData = callData
		mexToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEXRefundBalances(&mexToken)

		NewTestEnvironmentWithChainSimulatorAndTokensAndRefund(
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
		callData := CreateScCallData("callPayableWithParams", 50000000)
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

		tadaToken := GenerateTestTADAToken()
		tadaToken.TestOperations[2].MvxSCCallData = callData
		tadaToken.TestOperations[2].MvxFaultySCCall = true
		ApplyTADARefundBalances(&tadaToken)

		eurocToken := GenerateTestEUROCToken()
		eurocToken.TestOperations[2].MvxSCCallData = callData
		eurocToken.TestOperations[2].MvxFaultySCCall = true
		ApplyEUROCRefundBalances(&eurocToken)

		mexToken := GenerateTestMEXToken()
		mexToken.TestOperations[2].MvxSCCallData = callData
		mexToken.TestOperations[2].MvxFaultySCCall = true
		ApplyMEXRefundBalances(&mexToken)

		NewTestEnvironmentWithChainSimulatorAndTokensAndRefund(
			t,
			make(chan error),
			usdcToken,
			memeToken,
			tadaToken,
			eurocToken,
			mexToken,
		)
	})
	t.Run("frozen tokens for receiver should refund", func(t *testing.T) {
		usdcToken := GenerateTestUSDCToken()
		usdcToken.IssueTokenParams.IsFrozen = true
		usdcToken.TestOperations = []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(3000),
				ValueToSendFromMvX:   nil,
			},
		}
		usdcToken.DeltaBalances = map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-3000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(3000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
			},
			framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-3000 + 2950),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(3000 - 2950),
					OnMvx:    big.NewInt(50),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
			},
		}
		usdcToken.MintBurnChecks = &framework.MintBurnBalances{
			MvxTotalUniversalMint:     big.NewInt(3000),
			MvxTotalChainSpecificMint: big.NewInt(3000),
			MvxTotalUniversalBurn:     big.NewInt(3000),
			MvxTotalChainSpecificBurn: big.NewInt(3000 - 50),
			MvxSafeMintValue:          big.NewInt(3000),
			MvxSafeBurnValue:          big.NewInt(3000 - 50),

			EthSafeMintValue: big.NewInt(0),
			EthSafeBurnValue: big.NewInt(0),
		}
		usdcToken.SpecialChecks = &framework.SpecialBalanceChecks{
			WrapperDeltaLiquidityCheck: big.NewInt(0),
		}

		memeToken := GenerateTestMEMEToken()
		memeToken.IssueTokenParams.IsFrozen = true
		memeToken.TestOperations = []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(1300),
				ValueToSendFromMvX:   big.NewInt(1800),
			},
		}
		memeToken.DeltaBalances = map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(-1800),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(1800 - 51),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(1800),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
			},
			framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(-1800),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(1800 - 51 - 1300 + 1300 - 51),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(1800 - 1300 + 1300),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
			},
		}
		memeToken.MintBurnChecks = &framework.MintBurnBalances{
			MvxTotalUniversalMint:     big.NewInt(0),
			MvxTotalChainSpecificMint: big.NewInt(0),
			MvxTotalUniversalBurn:     big.NewInt(0),
			MvxTotalChainSpecificBurn: big.NewInt(0),
			MvxSafeMintValue:          big.NewInt(0),
			MvxSafeBurnValue:          big.NewInt(0),

			EthSafeMintValue: big.NewInt(1800 - 51 + 1300 - 51),
			EthSafeBurnValue: big.NewInt(1300),
		}
		memeToken.SpecialChecks = &framework.SpecialBalanceChecks{
			WrapperDeltaLiquidityCheck: big.NewInt(0),
		}

		tadaToken := GenerateTestTADAToken()
		tadaToken.IssueTokenParams.IsFrozen = true
		tadaToken.TestOperations = []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(700),
				ValueToSendFromMvX:   big.NewInt(4700),
			},
		}
		tadaToken.DeltaBalances = map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(-4700),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(4700 - 57),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(57),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(-4700),
					MvxToken: framework.ChainSpecificToken,
				},
			},
			framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(-4700),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(4700 - 57 - 700 + 700 - 57),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(57 + 57),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(-4700 + 700 - 700),
					MvxToken: framework.ChainSpecificToken,
				},
			},
		}
		tadaToken.MintBurnChecks = &framework.MintBurnBalances{
			MvxTotalUniversalMint:     big.NewInt(700),
			MvxTotalChainSpecificMint: big.NewInt(700),
			MvxTotalUniversalBurn:     big.NewInt(4700 + 700),
			MvxTotalChainSpecificBurn: big.NewInt(4700 - 57 + 700 - 57),
			MvxSafeMintValue:          big.NewInt(700),
			MvxSafeBurnValue:          big.NewInt(4700 - 57 + 700 - 57),

			EthSafeMintValue: big.NewInt(4700 - 57 + 700 - 57),
			EthSafeBurnValue: big.NewInt(700),
		}
		tadaToken.SpecialChecks = &framework.SpecialBalanceChecks{
			WrapperDeltaLiquidityCheck: big.NewInt(-4700),
		}

		eurocToken := GenerateTestEUROCToken()
		eurocToken.IssueTokenParams.IsFrozen = true
		eurocToken.TestOperations = []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(2000),
				ValueToSendFromMvX:   nil,
			},
			{
				ValueToTransferToMvx: big.NewInt(1500),
				ValueToSendFromMvX:   nil,
				MvxSCCallData:        CreateScCallData("callPayable", 50000000),
			},
		}
		eurocToken.DeltaBalances = map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-2000 - 1500),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(1500),
					MvxToken: framework.UniversalToken,
				},
			},
			framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-2000 - 1500 + 1948),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(52),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(1500),
					MvxToken: framework.UniversalToken,
				},
			},
		}
		eurocToken.MintBurnChecks = &framework.MintBurnBalances{
			MvxTotalUniversalMint:     big.NewInt(2000 + 1500),
			MvxTotalChainSpecificMint: big.NewInt(0),
			MvxTotalUniversalBurn:     big.NewInt(2000 - 52),
			MvxTotalChainSpecificBurn: big.NewInt(0),
			MvxSafeMintValue:          big.NewInt(2000 + 1500),
			MvxSafeBurnValue:          big.NewInt(2000 - 52),

			EthSafeMintValue: big.NewInt(2000 - 52),
			EthSafeBurnValue: big.NewInt(2000 + 1500),
		}
		eurocToken.SpecialChecks = &framework.SpecialBalanceChecks{
			WrapperDeltaLiquidityCheck: big.NewInt(0),
		}

		mexToken := GenerateTestMEXToken()
		mexToken.IssueTokenParams.IsFrozen = true
		mexToken.TestOperations = []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(6500),
				ValueToSendFromMvX:   big.NewInt(9000),
			},
		}
		mexToken.DeltaBalances = map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(-9000),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(9000 - 53),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(53),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
			},
			framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(-9000),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(9000 - 53 - 6500 + 6500 - 53),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(53 + 53),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
			},
		}
		mexToken.MintBurnChecks = &framework.MintBurnBalances{
			MvxTotalUniversalMint:     big.NewInt(6500),
			MvxTotalChainSpecificMint: big.NewInt(0),
			MvxTotalUniversalBurn:     big.NewInt(9000 - 53 + 6500 - 53),
			MvxTotalChainSpecificBurn: big.NewInt(0),
			MvxSafeMintValue:          big.NewInt(6500),
			MvxSafeBurnValue:          big.NewInt(9000 - 53 + 6500 - 53),

			EthSafeMintValue: big.NewInt(9000 - 53 + 6500 - 53),
			EthSafeBurnValue: big.NewInt(6500),
		}
		mexToken.SpecialChecks = &framework.SpecialBalanceChecks{
			WrapperDeltaLiquidityCheck: big.NewInt(0),
		}

		NewTestEnvironmentWithChainSimulatorAndTokensAndRefund(
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
