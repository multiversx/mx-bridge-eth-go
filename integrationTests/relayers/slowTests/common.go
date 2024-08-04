//go:build slow

package slowTests

import (
	"math/big"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
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
			ValueToMintOnMvx:                 "10000000000",
			IsMintBurnOnMvX:                  true,
			IsNativeOnMvX:                    false,
			EthTokenName:                     "ETHTOKEN",
			EthTokenSymbol:                   "ETHT",
			ValueToMintOnEth:                 "10000000000",
			IsMintBurnOnEth:                  false,
			IsNativeOnEth:                    true,
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(5000),
				ValueToSendFromMvX:   big.NewInt(2500),
				MvxSCCallMethod:      "",
				MvxSCCallGasLimit:    0,
				MvxSCCallArguments:   nil,
			},
			{
				ValueToTransferToMvx: big.NewInt(7000),
				ValueToSendFromMvX:   big.NewInt(300),
				MvxSCCallMethod:      "",
				MvxSCCallGasLimit:    0,
				MvxSCCallArguments:   nil,
			},
			{
				ValueToTransferToMvx: big.NewInt(1000),
				ValueToSendFromMvX:   nil,
				MvxSCCallMethod:      "callPayable",
				MvxSCCallGasLimit:    50000000,
				MvxSCCallArguments:   nil,
			},
		},
		ESDTSafeExtraBalance:    big.NewInt(100),                                        // extra is just for the fees for the 2 transfers mvx->eth
		EthTestAddrExtraBalance: big.NewInt(-5000 + 2500 - 50 - 7000 + 300 - 50 - 1000), // -(eth->mvx) + (mvx->eth) - fees
	}
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
			ValueToMintOnMvx:                 "10000000000",
			IsMintBurnOnMvX:                  false,
			IsNativeOnMvX:                    true,
			EthTokenName:                     "ETHMEME",
			EthTokenSymbol:                   "ETHM",
			ValueToMintOnEth:                 "10000000000",
			IsMintBurnOnEth:                  true,
			IsNativeOnEth:                    false,
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(2400),
				ValueToSendFromMvX:   big.NewInt(4000),
				MvxSCCallMethod:      "",
				MvxSCCallGasLimit:    0,
				MvxSCCallArguments:   nil,
			},
			{
				ValueToTransferToMvx: big.NewInt(200),
				ValueToSendFromMvX:   big.NewInt(6000),
				MvxSCCallMethod:      "",
				MvxSCCallGasLimit:    0,
				MvxSCCallArguments:   nil,
			},
			{
				ValueToTransferToMvx: big.NewInt(1000),
				ValueToSendFromMvX:   big.NewInt(2000),
				MvxSCCallMethod:      "callPayable",
				MvxSCCallGasLimit:    50000000,
				MvxSCCallArguments:   nil,
			},
		},
		ESDTSafeExtraBalance:    big.NewInt(4000 + 6000 + 2000), // everything is locked in the safe esdt contract
		EthTestAddrExtraBalance: big.NewInt(4000 - 50 + 6000 - 50 + 2000 - 50),
	}
}