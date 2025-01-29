//go:build slow

package slowTests

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
	"github.com/multiversx/mx-bridge-eth-go/parsers"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var (
	log            = logger.GetOrCreate("integrationTests/relayers/slowTests")
	mvxZeroAddress = bytes.Repeat([]byte{0x00}, 32)
	ethZeroAddress = common.Address{}
)

// GenerateTestUSDCToken will generate a test USDC token
func GenerateTestUSDCToken() framework.TestTokenParams {
	// USDC is ethNative = true, ethMintBurn = false, mvxNative = false, mvxMintBurn = true
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "USDC",
			NumOfDecimalsUniversal:           6,
			NumOfDecimalsChainSpecific:       6,
			MvxUniversalTokenTicker:          "USDC",
			MvxChainSpecificTokenTicker:      "ETHUSDC",
			MvxUniversalTokenDisplayName:     "WrappedUSDC",
			MvxChainSpecificTokenDisplayName: "EthereumWrappedUSDC",
			MvxToEthFee:                      big.NewInt(50),
			ValueToMintOnMvx:                 "10000000000",
			IsMintBurnOnMvX:                  true,
			IsNativeOnMvX:                    false,
			HasChainSpecificToken:            true,
			EthTokenName:                     "EthUSDC",
			EthTokenSymbol:                   "USDC",
			ValueToMintOnEth:                 "10000000000",
			IsMintBurnOnEth:                  false,
			IsNativeOnEth:                    true,
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(5000),
				ValueToSendFromMvX:   big.NewInt(2500),
			},
			{
				ValueToTransferToMvx: big.NewInt(7000),
				ValueToSendFromMvX:   big.NewInt(300),
			},
			{
				ValueToTransferToMvx: big.NewInt(1000),
				ValueToSendFromMvX:   nil,
				MvxSCCallData:        createScCallData("callPayable", 50000000),
			},
			{
				ValueToTransferToMvx: big.NewInt(20),
				ValueToSendFromMvX:   nil,
				IsFaultyDeposit:      true,
			},
			{
				ValueToTransferToMvx: big.NewInt(900),
				ValueToSendFromMvX:   nil,
				InvalidReceiver:      mvxZeroAddress,
			},
			{
				ValueToTransferToMvx: nil,
				ValueToSendFromMvX:   big.NewInt(730),
				InvalidReceiver:      ethZeroAddress,
				IsFaultyDeposit:      true,
			},
		},
		DeltaBalances: map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-5000 - 7000 - 1000 - 900),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(5000 + 7000),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(5000 + 7000 + 1000 + 900),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(1000),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(5000 + 7000 + 1000),
					MvxToken: framework.ChainSpecificToken,
				},
			},
			framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-5000 - 7000 - 1000 - 900 + 850), // 850 is the refund value
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(5000 - 2500 + 7000 - 300),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(2500 - 50 + 300 - 50),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(5000 + 7000 + 1000 + 900 - 2450 - 250 - 850),
					OnMvx:    big.NewInt(50 + 50 + 50),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(1000),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(5000 + 7000 + 1000 - 2500 - 300),
					MvxToken: framework.ChainSpecificToken,
				},
			},
		},
		MintBurnChecks: &framework.MintBurnBalances{
			MvxTotalUniversalMint:     big.NewInt(5000 + 7000 + 1000),
			MvxTotalChainSpecificMint: big.NewInt(5000 + 7000 + 1000 + 900),
			MvxTotalUniversalBurn:     big.NewInt(2500 + 300),
			MvxTotalChainSpecificBurn: big.NewInt(2500 - 50 + 300 - 50 + 900 - 50),
			MvxSafeMintValue:          big.NewInt(5000 + 7000 + 1000 + 900),
			MvxSafeBurnValue:          big.NewInt(2500 - 50 + 300 - 50 + 900 - 50),

			EthSafeMintValue: big.NewInt(0),
			EthSafeBurnValue: big.NewInt(0),
		},
		SpecialChecks: &framework.SpecialBalanceChecks{
			WrapperDeltaLiquidityCheck: big.NewInt(5000 + 7000 + 1000 - 2500 - 300),
		},
	}
}

// ApplyUSDCRefundBalances will apply the refund balances on the involved entities for the USDC token
func ApplyUSDCRefundBalances(token *framework.TestTokenParams) {
	// called test SC will have 0 balance since eth->mvx transfer failed
	token.DeltaBalances[framework.FirstHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(0)
	// extra is just for the fees for the 2 transfers mvx->eth and the failed eth->mvx that needed refund
	token.DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50 + 50 + 50 + 50)
	// we need to subtract the refunded value from the Ethereum Safe contract
	token.DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnEth = big.NewInt(5000 + 7000 + 1000 + 900 - 2450 - 250 - 950 - 850)
	// Alice will get her tokens back from the refund
	token.DeltaBalances[framework.SecondHalfBridge][framework.Alice].OnEth = big.NewInt(-5000 - 7000 - 1000 - 900 + 950 + 850)
	// no funds remain in the called test SC
	token.DeltaBalances[framework.SecondHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(0)
	// we need to subtract the refunded value from the wrapper contract
	token.DeltaBalances[framework.SecondHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(5000 + 7000 + 1000 - 2500 - 300 - 1000)

	token.MintBurnChecks.MvxTotalChainSpecificBurn = big.NewInt(2500 - 50 + 300 - 50 + 1000 - 50 + 900 - 50)
	token.MintBurnChecks.MvxTotalUniversalBurn = big.NewInt(2500 + 300 + 1000)
	token.MintBurnChecks.MvxSafeBurnValue = big.NewInt(2500 - 50 + 300 - 50 + 1000 - 50 + 900 - 50)

	token.SpecialChecks.WrapperDeltaLiquidityCheck = big.NewInt(5000 + 7000 + 1000 - 2500 - 300 - 1000)
}

// GenerateTestMEMEToken will generate a test MEME token
func GenerateTestMEMEToken() framework.TestTokenParams {
	//MEME is ethNative = false, ethMintBurn = true, mvxNative = true, mvxMintBurn = false
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "MEME",
			NumOfDecimalsUniversal:           1,
			NumOfDecimalsChainSpecific:       1,
			MvxUniversalTokenTicker:          "MEME",
			MvxChainSpecificTokenTicker:      "ETHMEME",
			MvxUniversalTokenDisplayName:     "WrappedMEME",
			MvxChainSpecificTokenDisplayName: "EthereumWrappedMEME",
			MvxToEthFee:                      big.NewInt(51),
			ValueToMintOnMvx:                 "10000000000",
			IsMintBurnOnMvX:                  false,
			IsNativeOnMvX:                    true,
			HasChainSpecificToken:            false,
			EthTokenName:                     "EthMEME",
			EthTokenSymbol:                   "MEME",
			ValueToMintOnEth:                 "10000000000",
			IsMintBurnOnEth:                  true,
			IsNativeOnEth:                    false,
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(2400),
				ValueToSendFromMvX:   big.NewInt(4000),
			},
			{
				ValueToTransferToMvx: big.NewInt(200),
				ValueToSendFromMvX:   big.NewInt(6000),
			},
			{
				ValueToTransferToMvx: big.NewInt(1000),
				ValueToSendFromMvX:   big.NewInt(2000),
				MvxSCCallData:        createScCallData("callPayable", 50000000),
			},
			{
				ValueToTransferToMvx: nil,
				ValueToSendFromMvX:   big.NewInt(38),
				IsFaultyDeposit:      true,
			},
			{
				ValueToTransferToMvx: nil,
				ValueToSendFromMvX:   big.NewInt(420),
				InvalidReceiver:      ethZeroAddress,
				IsFaultyDeposit:      true,
			},
			{
				ValueToTransferToMvx: big.NewInt(1300),
				ValueToSendFromMvX:   nil,
				InvalidReceiver:      mvxZeroAddress,
			},
		},
		DeltaBalances: map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(-4000 - 6000 - 2000),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(4000 - 51 + 6000 - 51 + 2000 - 51),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(4000 + 6000 + 2000),
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
					OnMvx:    big.NewInt(-4000 - 6000 - 2000),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(4000 - 51 - 2400 + 6000 - 51 - 200 + 2000 - 51 - 1000 - 1300 + 1249),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(2400 + 200),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(4000 - 2400 + 6000 - 200 + 2000 - 1000 - 1300 + 1300),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(1000),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
			},
		},
		MintBurnChecks: &framework.MintBurnBalances{
			MvxTotalUniversalMint:     big.NewInt(0),
			MvxTotalChainSpecificMint: big.NewInt(0),
			MvxTotalUniversalBurn:     big.NewInt(0),
			MvxTotalChainSpecificBurn: big.NewInt(0),
			MvxSafeMintValue:          big.NewInt(0),
			MvxSafeBurnValue:          big.NewInt(0),

			EthSafeMintValue: big.NewInt(4000 - 51 + 6000 - 51 + 2000 - 51 + 1300 - 51),
			EthSafeBurnValue: big.NewInt(2400 + 200 + 1000 + 1300),
		},
		SpecialChecks: &framework.SpecialBalanceChecks{
			WrapperDeltaLiquidityCheck: big.NewInt(0),
		},
	}
}

// ApplyMEMERefundBalances will apply the refund balances on the involved entities for the MEME token
func ApplyMEMERefundBalances(token *framework.TestTokenParams) {
	// we need to add the 1000 MEME tokens as the third bridge was done that include the refund on the Ethereum side
	token.DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(4000 - 2400 + 6000 - 200 + 2000 - 1300 + 1300 - 1000 + 1000)
	// Bob will get his tokens back from the refund
	token.DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnEth = big.NewInt(4000 - 51 - 2400 + 6000 - 51 - 200 + 2000 - 51 - 1300 + 1249 - 1000 + 949)
	// no funds remain in the test caller SC
	token.DeltaBalances[framework.SecondHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(0)

	token.MintBurnChecks.EthSafeMintValue = big.NewInt(4000 - 51 + 6000 - 51 + 2000 - 51 + 1300 - 51 + 1000 - 51)
}

// GenerateTestTADAToken will generate a test MEME token
func GenerateTestTADAToken() framework.TestTokenParams {
	//TADA is ethNative = false, ethMintBurn = true, mvxNative = true, mvxMintBurn = false, hasChainSpecificToken = true
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "TADA",
			NumOfDecimalsUniversal:           1,
			NumOfDecimalsChainSpecific:       1,
			MvxUniversalTokenTicker:          "TADA",
			MvxChainSpecificTokenTicker:      "ETHTADA",
			MvxUniversalTokenDisplayName:     "WrappedTADA",
			MvxChainSpecificTokenDisplayName: "EthereumWrappedTADA",
			ValueToMintOnMvx:                 "10000000000",
			MvxToEthFee:                      big.NewInt(57),
			IsMintBurnOnMvX:                  true,
			IsNativeOnMvX:                    true,
			HasChainSpecificToken:            true,
			EthTokenName:                     "EthTADA",
			EthTokenSymbol:                   "TADA",
			ValueToMintOnEth:                 "10000000000",
			IsMintBurnOnEth:                  true,
			IsNativeOnEth:                    false,
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(3100),
				ValueToSendFromMvX:   big.NewInt(5980),
			},
			{
				ValueToTransferToMvx: big.NewInt(800),
				ValueToSendFromMvX:   big.NewInt(2300),
			},
			{
				ValueToTransferToMvx: big.NewInt(2000),
				ValueToSendFromMvX:   big.NewInt(4000),
				MvxSCCallData:        createScCallData("callPayable", 50000000),
			},
			{
				ValueToTransferToMvx: nil,
				ValueToSendFromMvX:   big.NewInt(29),
				IsFaultyDeposit:      true,
			},
			{
				ValueToTransferToMvx: nil,
				ValueToSendFromMvX:   big.NewInt(670),
				InvalidReceiver:      ethZeroAddress,
				IsFaultyDeposit:      true,
			},
			{
				ValueToTransferToMvx: big.NewInt(1900),
				ValueToSendFromMvX:   nil,
				InvalidReceiver:      mvxZeroAddress,
			},
		},
		DeltaBalances: map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(-5980 - 2300 - 4000),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(5980 - 57 + 2300 - 57 + 4000 - 57),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(57 + 57 + 57),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(-5980 - 2300 - 4000),
					MvxToken: framework.ChainSpecificToken,
				},
			},
			framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(-5980 - 2300 - 4000),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(5980 - 57 - 3100 + 2300 - 57 - 800 + 4000 - 57 - 2000 - 1900 + 1843),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(3100 + 800),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(57 + 57 + 57 + 57),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(2000),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(-5980 - 2300 - 4000 + 3100 + 800 + 2000),
					MvxToken: framework.ChainSpecificToken,
				},
			},
		},
		MintBurnChecks: &framework.MintBurnBalances{
			MvxTotalUniversalMint:     big.NewInt(3100 + 800 + 2000),
			MvxTotalChainSpecificMint: big.NewInt(3100 + 800 + 2000 + 1900),
			MvxTotalUniversalBurn:     big.NewInt(5980 + 2300 + 4000),
			MvxTotalChainSpecificBurn: big.NewInt(5980 - 57 + 2300 - 57 + 4000 - 57 + 1900 - 57),
			MvxSafeMintValue:          big.NewInt(3100 + 800 + 2000 + 1900),
			MvxSafeBurnValue:          big.NewInt(5980 - 57 + 2300 - 57 + 4000 - 57 + 1900 - 57),

			EthSafeMintValue: big.NewInt(5980 - 57 + 2300 - 57 + 4000 - 57 + 1900 - 57),
			EthSafeBurnValue: big.NewInt(3100 + 800 + 2000 + 1900),
		},
		SpecialChecks: &framework.SpecialBalanceChecks{
			WrapperDeltaLiquidityCheck: big.NewInt(-5980 - 2300 - 4000 + 3100 + 800 + 2000),
		},
	}
}

// ApplyTADARefundBalances will apply the refund balances on the involved entities for the MEME token
func ApplyTADARefundBalances(token *framework.TestTokenParams) {
	// we need to add the 1000 MEME tokens as the third bridge was done that include the refund on the Ethereum side
	token.DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(57 + 57 + 57 + 57 + 57)
	// Bob will get his tokens back from the refund
	token.DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnEth = big.NewInt(5980 - 57 - 3100 + 2300 - 57 - 800 + 4000 - 57 - 1900 + 1843 - 2000 + 1943)
	// no funds remain in the test caller SC
	token.DeltaBalances[framework.SecondHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(0)
	// we need to subtract the refunded value from the wrapper contract
	token.DeltaBalances[framework.SecondHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(-5980 - 2300 - 4000 + 3100 + 800 + 2000 - 2000)

	token.MintBurnChecks.MvxTotalChainSpecificBurn = big.NewInt(5980 - 57 + 2300 - 57 + 4000 - 57 + 2000 - 57 + 1900 - 57)
	token.MintBurnChecks.MvxTotalUniversalBurn = big.NewInt(5980 + 2300 + 4000 + 2000)
	token.MintBurnChecks.EthSafeMintValue = big.NewInt(5980 - 57 + 2300 - 57 + 4000 - 57 + 2000 - 57 + 1900 - 57)
	token.MintBurnChecks.MvxSafeBurnValue = big.NewInt(5980 - 57 + 2300 - 57 + 4000 - 57 + 2000 - 57 + 1900 - 57)

	token.SpecialChecks.WrapperDeltaLiquidityCheck = big.NewInt(-5980 - 2300 - 4000 + 3100 + 800 + 2000 - 2000)
}

// GenerateTestEUROCToken will generate a test EUROC token
func GenerateTestEUROCToken() framework.TestTokenParams {
	//EUROC is ethNative = true, ethMintBurn = true, mvxNative = false, mvxMintBurn = true
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "EUROC",
			NumOfDecimalsUniversal:           6,
			NumOfDecimalsChainSpecific:       6,
			MvxUniversalTokenTicker:          "EUROC",
			MvxChainSpecificTokenTicker:      "EUROC",
			MvxUniversalTokenDisplayName:     "TestEUROC",
			MvxChainSpecificTokenDisplayName: "TestEUROC",
			MvxToEthFee:                      big.NewInt(52),
			ValueToMintOnMvx:                 "10000000000",
			IsMintBurnOnMvX:                  true,
			IsNativeOnMvX:                    false,
			HasChainSpecificToken:            false,
			EthTokenName:                     "EthEuroC",
			EthTokenSymbol:                   "EUROC",
			ValueToMintOnEth:                 "10000000000",
			IsMintBurnOnEth:                  true,
			IsNativeOnEth:                    true,
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(5010),
				ValueToSendFromMvX:   big.NewInt(2510),
			},
			{
				ValueToTransferToMvx: big.NewInt(7010),
				ValueToSendFromMvX:   big.NewInt(310),
			},
			{
				ValueToTransferToMvx: big.NewInt(1010),
				ValueToSendFromMvX:   nil,
				MvxSCCallData:        createScCallData("callPayable", 50000000),
			},
			{
				ValueToTransferToMvx: big.NewInt(24),
				ValueToSendFromMvX:   nil,
				IsFaultyDeposit:      true,
			},
			{
				ValueToTransferToMvx: big.NewInt(700),
				ValueToSendFromMvX:   nil,
				InvalidReceiver:      mvxZeroAddress,
			},
			{
				ValueToTransferToMvx: nil,
				ValueToSendFromMvX:   big.NewInt(853),
				InvalidReceiver:      ethZeroAddress,
				IsFaultyDeposit:      true,
			},
		},
		DeltaBalances: map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-5010 - 7010 - 1010 - 700),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(5010 + 7010),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(1010),
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
					OnEth:    big.NewInt(-5010 - 7010 - 1010 - 700 + 648), // 648 is the refund value
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(5010 - 2510 + 7010 - 310),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(2510 - 52 + 310 - 52),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(52 + 52 + 52),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(1010),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
			},
		},
		MintBurnChecks: &framework.MintBurnBalances{
			MvxTotalUniversalMint:     big.NewInt(5010 + 7010 + 1010 + 700),
			MvxTotalChainSpecificMint: big.NewInt(0),
			MvxTotalUniversalBurn:     big.NewInt(2510 - 52 + 310 - 52 + 700 - 52),
			MvxTotalChainSpecificBurn: big.NewInt(0),
			MvxSafeMintValue:          big.NewInt(5010 + 7010 + 1010 + 700),
			MvxSafeBurnValue:          big.NewInt(2510 - 52 + 310 - 52 + 700 - 52),

			EthSafeMintValue: big.NewInt(2510 - 52 + 310 - 52 + 648),
			EthSafeBurnValue: big.NewInt(5010 + 7010 + 1010 + 700),
		},
		SpecialChecks: &framework.SpecialBalanceChecks{
			WrapperDeltaLiquidityCheck: big.NewInt(0),
		},
	}
}

// ApplyEUROCRefundBalances will apply the refund balances on the involved entities for the EUROC token
func ApplyEUROCRefundBalances(token *framework.TestTokenParams) {
	// called test SC will have 0 balance since eth->mvx transfer failed
	token.DeltaBalances[framework.FirstHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(0)
	// extra is just for the fees for the 2 transfers mvx->eth and the failed eth->mvx that needed refund
	token.DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(52 + 52 + 52 + 52)
	// Alice will get her tokens back from the refund
	token.DeltaBalances[framework.SecondHalfBridge][framework.Alice].OnEth = big.NewInt(-5010 - 7010 - 1010 - 700 + 958 + 648)
	// no funds remain in the called test SC
	token.DeltaBalances[framework.SecondHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(0)

	token.MintBurnChecks.MvxTotalUniversalBurn = big.NewInt(2510 - 52 + 310 - 52 + 700 - 52 + 1010 - 52)
	token.MintBurnChecks.MvxSafeBurnValue = big.NewInt(2510 - 52 + 310 - 52 + 700 - 52 + 1010 - 52)
	token.MintBurnChecks.EthSafeMintValue = big.NewInt(2510 - 52 + 310 - 52 + 648 + 1010 - 52)
}

// GenerateTestMEXToken will generate a test MEX token
func GenerateTestMEXToken() framework.TestTokenParams {
	//MEX is ethNative = false, ethMintBurn = true, mvxNative = true, mvxMintBurn = true
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "MEX",
			NumOfDecimalsUniversal:           2,
			NumOfDecimalsChainSpecific:       2,
			MvxUniversalTokenTicker:          "MEX",
			MvxChainSpecificTokenTicker:      "MEX",
			MvxUniversalTokenDisplayName:     "TestMEX",
			MvxChainSpecificTokenDisplayName: "TestMEX",
			MvxToEthFee:                      big.NewInt(53),
			ValueToMintOnMvx:                 "10000000000",
			IsMintBurnOnMvX:                  true,
			IsNativeOnMvX:                    true,
			HasChainSpecificToken:            false,
			EthTokenName:                     "EthMex",
			EthTokenSymbol:                   "MEX",
			ValueToMintOnEth:                 "10000000000",
			IsMintBurnOnEth:                  true,
			IsNativeOnEth:                    false,
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(2410),
				ValueToSendFromMvX:   big.NewInt(4010),
			},
			{
				ValueToTransferToMvx: big.NewInt(210),
				ValueToSendFromMvX:   big.NewInt(6010),
			},
			{
				ValueToTransferToMvx: big.NewInt(1010),
				ValueToSendFromMvX:   big.NewInt(2010),
				MvxSCCallData:        createScCallData("callPayable", 50000000),
			},
			{
				ValueToTransferToMvx: big.NewInt(10),
				ValueToSendFromMvX:   nil,
				IsFaultyDeposit:      true,
			},
			{
				ValueToTransferToMvx: nil,
				ValueToSendFromMvX:   big.NewInt(500),
				InvalidReceiver:      ethZeroAddress,
				IsFaultyDeposit:      true,
			},
			{
				ValueToTransferToMvx: big.NewInt(3000),
				ValueToSendFromMvX:   nil,
				InvalidReceiver:      mvxZeroAddress,
			},
		},
		DeltaBalances: map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(-4010 - 6010 - 2010),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(4010 - 53 + 6010 - 53 + 2010 - 53),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(53 + 53 + 53),
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
					OnMvx:    big.NewInt(-4010 - 6010 - 2010),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(4010 - 53 - 2410 + 6010 - 53 - 210 + 2010 - 53 - 1010 - 3000 + 2947),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(2410 + 210),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(53 + 53 + 53 + 53),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(1010),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
			},
		},
		MintBurnChecks: &framework.MintBurnBalances{
			MvxTotalUniversalMint:     big.NewInt(2410 + 210 + 1010 + 3000),
			MvxTotalChainSpecificMint: big.NewInt(0),
			MvxTotalUniversalBurn:     big.NewInt(4010 - 53 + 6010 - 53 + 2010 - 53 + 3000 - 53),
			MvxTotalChainSpecificBurn: big.NewInt(0),
			MvxSafeMintValue:          big.NewInt(2410 + 210 + 1010 + 3000),
			MvxSafeBurnValue:          big.NewInt(4010 - 53 + 6010 - 53 + 2010 - 53 + 3000 - 53),

			EthSafeMintValue: big.NewInt(4010 - 53 + 6010 - 53 + 2010 - 53 + 3000 - 53),
			EthSafeBurnValue: big.NewInt(2410 + 210 + 1010 + 3000),
		},
		SpecialChecks: &framework.SpecialBalanceChecks{
			WrapperDeltaLiquidityCheck: big.NewInt(0),
		},
	}
}

// ApplyMEXRefundBalances will apply the refund balances on the involved entities for the MEX token
func ApplyMEXRefundBalances(token *framework.TestTokenParams) {
	// 3 normal swaps + the refund one
	token.DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(53 + 53 + 53 + 53 + 53)
	// Bob will get his tokens back from the refund
	token.DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnEth = big.NewInt(4010 - 53 - 2410 + 6010 - 53 - 210 + 2010 - 53 - 1010 + 957 - 3000 + 2947)
	// no funds remain in the test caller SC
	token.DeltaBalances[framework.SecondHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(0)

	token.MintBurnChecks.MvxTotalUniversalBurn = big.NewInt(4010 - 53 + 6010 - 53 + 2010 - 53 + 3000 - 53 + 1010 - 53)
	token.MintBurnChecks.MvxSafeBurnValue = big.NewInt(4010 - 53 + 6010 - 53 + 2010 - 53 + 3000 - 53 + 1010 - 53)
	token.MintBurnChecks.EthSafeMintValue = big.NewInt(4010 - 53 + 6010 - 53 + 2010 - 53 + 3000 - 53 + 1010 - 53)
}

// GenerateUnlistedTokenFromEth will generate an unlisted token on Eth
func GenerateUnlistedTokenFromEth() framework.TestTokenParams {
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "ULTKE",
			NumOfDecimalsUniversal:           6,
			NumOfDecimalsChainSpecific:       6,
			MvxUniversalTokenTicker:          "ULTKE",
			MvxChainSpecificTokenTicker:      "ULTKE",
			MvxUniversalTokenDisplayName:     "TestULTKE",
			MvxChainSpecificTokenDisplayName: "TestULTKE",
			ValueToMintOnMvx:                 "10000000000",
			MvxToEthFee:                      big.NewInt(54),
			IsMintBurnOnMvX:                  true,
			IsNativeOnMvX:                    false,
			HasChainSpecificToken:            false,
			EthTokenName:                     "EthULTKE",
			EthTokenSymbol:                   "ULTKE",
			ValueToMintOnEth:                 "10000000000",
			IsMintBurnOnEth:                  true,
			IsNativeOnEth:                    true,
			PreventWhitelist:                 true,
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(5010),
				ValueToSendFromMvX:   nil,
				IsFaultyDeposit:      true,
			},
			{
				ValueToTransferToMvx: big.NewInt(1010),
				ValueToSendFromMvX:   nil,
				MvxSCCallData:        createScCallData("callPayable", 50000000),
				IsFaultyDeposit:      true,
			},
		},
		DeltaBalances: map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(0),
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
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
			},
			framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(0),
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
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
			},
		},
		MintBurnChecks: &framework.MintBurnBalances{
			MvxTotalUniversalMint:     big.NewInt(0),
			MvxTotalChainSpecificMint: big.NewInt(0),
			MvxTotalUniversalBurn:     big.NewInt(0),
			MvxTotalChainSpecificBurn: big.NewInt(0),
			MvxSafeMintValue:          big.NewInt(0),
			MvxSafeBurnValue:          big.NewInt(0),

			EthSafeBurnValue: big.NewInt(0),
			EthSafeMintValue: big.NewInt(0),
		},
		SpecialChecks: &framework.SpecialBalanceChecks{
			WrapperDeltaLiquidityCheck: big.NewInt(0),
		},
	}
}

// GenerateUnlistedTokenFromMvx will generate an unlisted token on Mvx
func GenerateUnlistedTokenFromMvx() framework.TestTokenParams {
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "ULTKM",
			NumOfDecimalsUniversal:           2,
			NumOfDecimalsChainSpecific:       2,
			MvxUniversalTokenTicker:          "ULTKM",
			MvxChainSpecificTokenTicker:      "ULTKM",
			MvxUniversalTokenDisplayName:     "TestULTKM",
			MvxChainSpecificTokenDisplayName: "TestULTKM",
			MvxToEthFee:                      big.NewInt(55),
			ValueToMintOnMvx:                 "10000000000",
			IsMintBurnOnMvX:                  true,
			IsNativeOnMvX:                    true,
			HasChainSpecificToken:            false,
			EthTokenName:                     "EthULTKM",
			EthTokenSymbol:                   "ULTKM",
			ValueToMintOnEth:                 "10000000000",
			IsMintBurnOnEth:                  true,
			IsNativeOnEth:                    false,
			PreventWhitelist:                 true,
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToMvx: nil,
				ValueToSendFromMvX:   big.NewInt(4010),
			},
			{
				ValueToTransferToMvx: nil,
				ValueToSendFromMvX:   big.NewInt(2010),
				MvxSCCallData:        createScCallData("callPayable", 50000000),
			},
		},
		DeltaBalances: map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(0),
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
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
			},
			framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(0),
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
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
			},
		},
		MintBurnChecks: &framework.MintBurnBalances{
			MvxTotalUniversalMint:     big.NewInt(0),
			MvxTotalChainSpecificMint: big.NewInt(0),
			MvxTotalUniversalBurn:     big.NewInt(0),
			MvxTotalChainSpecificBurn: big.NewInt(0),
			MvxSafeMintValue:          big.NewInt(0),
			MvxSafeBurnValue:          big.NewInt(0),

			EthSafeMintValue: big.NewInt(0),
			EthSafeBurnValue: big.NewInt(0),
		},
		SpecialChecks: &framework.SpecialBalanceChecks{
			WrapperDeltaLiquidityCheck: big.NewInt(0),
		},
	}
}

// GenerateFrozenToken will generate a token that will be frozen
func GenerateFrozenToken() framework.TestTokenParams {
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "FROZEN",
			NumOfDecimalsUniversal:           2,
			NumOfDecimalsChainSpecific:       2,
			MvxUniversalTokenTicker:          "FROZEN",
			MvxChainSpecificTokenTicker:      "FROZEN",
			MvxUniversalTokenDisplayName:     "TestFROZEN",
			MvxChainSpecificTokenDisplayName: "TestFROZEN",
			ValueToMintOnMvx:                 "10000000000",
			IsMintBurnOnMvX:                  true,
			IsNativeOnMvX:                    false,
			HasChainSpecificToken:            false,
			EthTokenName:                     "EthFROZEN",
			EthTokenSymbol:                   "FROZEN",
			ValueToMintOnEth:                 "10000000000",
			IsMintBurnOnEth:                  true,
			IsNativeOnEth:                    true,
			IsFrozen:                         true,
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(2000),
				ValueToSendFromMvX:   nil,
			},
			{
				ValueToTransferToMvx: big.NewInt(1500),
				ValueToSendFromMvX:   nil,
				MvxSCCallData:        createScCallData("callPayable", 50000000),
			},
		},
		DeltaBalances: map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
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
					OnEth:    big.NewInt(-2000 - 1500 + 1950),
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
					OnMvx:    big.NewInt(50),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
			},
		},
	}
}

func createScCallData(function string, gasLimit uint64, args ...string) []byte {
	codec := parsers.MultiversxCodec{}
	callData := bridgeCore.CallData{
		Type:      bridgeCore.DataPresentProtocolMarker,
		Function:  function,
		GasLimit:  gasLimit,
		Arguments: args,
	}

	buff := codec.EncodeCallDataStrict(callData)
	log.Info("working with SC call data", "buff", buff)

	return buff
}
