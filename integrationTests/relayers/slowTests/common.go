//go:build slow

package slowTests

import (
	"math/big"

	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
	"github.com/multiversx/mx-bridge-eth-go/parsers"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var (
	log = logger.GetOrCreate("integrationTests/relayers/slowTests")
)

// GenerateTestUSDCToken will generate a test USDC token
func GenerateTestUSDCToken() framework.TestTokenParams {
	// USDC is peerChainNative = true, peerChainMintBurn = false, mvxNative = false, mvxMintBurn = true
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
			HasChainSpecificToken:            true,
			PeerChainTokenName:               "EthUSDC",
			PeerChainTokenSymbol:             "USDC",
			ValueToMintOnPeerChain:           "10000000000",
			IsMintBurnOnPeerChain:            false,
			IsNativeOnPeerChain:              true,
			PeerChainType:                    framework.ChainTypeEthereum,
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
		},
		ESDTSafeExtraBalance:          big.NewInt(100),                                        // extra is just for the fees for the 2 transfers mvx->peerChain
		PeerChainTestAddrExtraBalance: big.NewInt(-5000 + 2500 - 50 - 7000 + 300 - 50 - 1000), // -(peerChain->mvx) + (mvx->peerChain) - fees
	}
}

// GenerateTestMEMEToken will generate a test MEME token
func GenerateTestMEMEToken() framework.TestTokenParams {
	//MEME is peerChainNative = false, peerChainMintBurn = true, mvxNative = true, mvxMintBurn = false
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
			HasChainSpecificToken:            true,
			PeerChainTokenName:               "EthMEME",
			PeerChainTokenSymbol:             "MEME",
			ValueToMintOnPeerChain:           "10000000000",
			IsMintBurnOnPeerChain:            true,
			IsNativeOnPeerChain:              false,
			PeerChainType:                    framework.ChainTypeEthereum,
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
		},
		ESDTSafeExtraBalance:          big.NewInt(4000 + 6000 + 2000), // everything is locked in the safe esdt contract
		PeerChainTestAddrExtraBalance: big.NewInt(4000 - 50 + 6000 - 50 + 2000 - 50),
	}
}

// GenerateTestEUROCToken will generate a test EUROC token
func GenerateTestEUROCToken() framework.TestTokenParams {
	//EUROC is peerChainNative = true, peerChainMintBurn = true, mvxNative = false, mvxMintBurn = true
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "EUROC",
			NumOfDecimalsUniversal:           6,
			NumOfDecimalsChainSpecific:       6,
			MvxUniversalTokenTicker:          "EUROC",
			MvxChainSpecificTokenTicker:      "EUROC",
			MvxUniversalTokenDisplayName:     "TestEUROC",
			MvxChainSpecificTokenDisplayName: "TestEUROC",
			ValueToMintOnMvx:                 "10000000000",
			IsMintBurnOnMvX:                  true,
			IsNativeOnMvX:                    false,
			HasChainSpecificToken:            false,
			PeerChainTokenName:               "EthEuroC",
			PeerChainTokenSymbol:             "EUROC",
			ValueToMintOnPeerChain:           "10000000000",
			IsMintBurnOnPeerChain:            true,
			IsNativeOnPeerChain:              true,
			PeerChainType:                    framework.ChainTypeEthereum,
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
		},
		ESDTSafeExtraBalance:          big.NewInt(100),                                        // extra is just for the fees for the 2 transfers mvx->peerChain
		PeerChainTestAddrExtraBalance: big.NewInt(-5010 + 2510 - 50 - 7010 + 310 - 50 - 1010), // -(peerChain->mvx) + (mvx->peerChain) - fees
	}
}

// GenerateTestMEXToken will generate a test EUROC token
func GenerateTestMEXToken() framework.TestTokenParams {
	//MEX is peerChainNative = false, peerChainMintBurn = true, mvxNative = true, mvxMintBurn = true
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "MEX",
			NumOfDecimalsUniversal:           2,
			NumOfDecimalsChainSpecific:       2,
			MvxUniversalTokenTicker:          "MEX",
			MvxChainSpecificTokenTicker:      "MEX",
			MvxUniversalTokenDisplayName:     "TestMEX",
			MvxChainSpecificTokenDisplayName: "TestMEX",
			ValueToMintOnMvx:                 "10000000000",
			IsMintBurnOnMvX:                  true,
			IsNativeOnMvX:                    true,
			HasChainSpecificToken:            false,
			PeerChainTokenName:               "EthMex",
			PeerChainTokenSymbol:             "MEX",
			ValueToMintOnPeerChain:           "10000000000",
			IsMintBurnOnPeerChain:            true,
			IsNativeOnPeerChain:              false,
			PeerChainType:                    framework.ChainTypeEthereum,
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
		},
		ESDTSafeExtraBalance:          big.NewInt(150), // just the fees should be collected in ESDT safe
		PeerChainTestAddrExtraBalance: big.NewInt(4010 - 50 + 6010 - 50 + 2010 - 50),
	}
}

func GenerateTestSuiUSDCToken() framework.TestTokenParams {
	// SuiUSDC Sui is peerChainNative = true, peerChainMintBurn = false, mvxNative = false, mvxMintBurn = true
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "USDC",
			NumOfDecimalsUniversal:           6,
			NumOfDecimalsChainSpecific:       6,
			MvxUniversalTokenTicker:          "USDC",
			MvxChainSpecificTokenTicker:      "SUIUSDC",
			MvxUniversalTokenDisplayName:     "WrappedUSDC",
			MvxChainSpecificTokenDisplayName: "SuiWrappedUSDC",
			ValueToMintOnMvx:                 "10000000000",
			IsMintBurnOnMvX:                  true,
			IsNativeOnMvX:                    false,
			HasChainSpecificToken:            true,
			PeerChainTokenName:               "SuiUSDC",
			PeerChainTokenSymbol:             "USDC",
			ValueToMintOnPeerChain:           "10000000000",
			IsMintBurnOnPeerChain:            false,
			IsNativeOnPeerChain:              true,
			PeerChainType:                    framework.ChainTypeSui,
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
		},
		ESDTSafeExtraBalance:          big.NewInt(100),                                 // extra is just for the fees for the 2 transfers mvx->peerChain
		PeerChainTestAddrExtraBalance: big.NewInt(-5000 + 2500 - 50 - 7000 + 300 - 50), // -(peerChain->mvx) + (mvx->peerChain) - fees
	}
}

func GenerateTestWALToken() framework.TestTokenParams {
	// WAL Sui is peerChainNative = true, peerChainMintBurn = false, mvxNative = false, mvxMintBurn = true
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "WAL",
			NumOfDecimalsUniversal:           6,
			NumOfDecimalsChainSpecific:       6,
			MvxUniversalTokenTicker:          "WAL",
			MvxChainSpecificTokenTicker:      "SUIWAL",
			MvxUniversalTokenDisplayName:     "WrappedWAL",
			MvxChainSpecificTokenDisplayName: "SuiWrappedWAL",
			ValueToMintOnMvx:                 "10000000000",
			IsMintBurnOnMvX:                  true,
			IsNativeOnMvX:                    false,
			HasChainSpecificToken:            false,
			PeerChainTokenName:               "Walrus",
			PeerChainTokenSymbol:             "WAL",
			ValueToMintOnPeerChain:           "10000000000",
			IsMintBurnOnPeerChain:            false,
			IsNativeOnPeerChain:              true,
			PeerChainType:                    framework.ChainTypeSui,
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(7300),
				ValueToSendFromMvX:   big.NewInt(6150),
			},
			{
				ValueToTransferToMvx: big.NewInt(1900),
				ValueToSendFromMvX:   big.NewInt(1280),
			},
		},
		ESDTSafeExtraBalance:          big.NewInt(100),                                  // extra is just for the fees for the 2 transfers mvx->peerChain
		PeerChainTestAddrExtraBalance: big.NewInt(-7300 + 6150 - 50 - 1900 + 1280 - 50), // -(peerChain->mvx) + (mvx->peerChain) - fees
	}
}

func GenerateTestLKXMNToken() framework.TestTokenParams {
	// LKXMN is peerChainNative = true, peerChainMintBurn = true, mvxNative = false, mvxMintBurn = true
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "LKXMN",
			NumOfDecimalsUniversal:           6,
			NumOfDecimalsChainSpecific:       6,
			MvxUniversalTokenTicker:          "LKXMN",
			MvxChainSpecificTokenTicker:      "SUILKXMN",
			MvxUniversalTokenDisplayName:     "WrappedLKXMN",
			MvxChainSpecificTokenDisplayName: "SuiWrappedLKXMN",
			ValueToMintOnMvx:                 "10000000000",
			IsMintBurnOnMvX:                  true,
			IsNativeOnMvX:                    false,
			HasChainSpecificToken:            false,
			PeerChainTokenName:               "xMoney",
			PeerChainTokenSymbol:             "LKXMN",
			ValueToMintOnPeerChain:           "0",
			IsMintBurnOnPeerChain:            true,
			IsNativeOnPeerChain:              true,
			PeerChainType:                    framework.ChainTypeSui,
			IsLocked:                         true,
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToMvx: nil,
				ValueToSendFromMvX:   big.NewInt(1550),
			},
			{
				ValueToTransferToMvx: nil,
				ValueToSendFromMvX:   big.NewInt(4650),
			},
		},
		ESDTSafeExtraBalance:          big.NewInt(100),                   // extra is just for the fees for the 2 transfers mvx->peerChain
		PeerChainTestAddrExtraBalance: big.NewInt(1550 - 50 + 4650 - 50), // -(peerChain->mvx) + (mvx->peerChain) - fees
	}
}

func createScCallData(function string, gasLimit uint64, args ...string) []byte {
	codec := testsCommon.TestMultiversXCodec{}
	callData := parsers.CallData{
		Type:      bridgeCore.DataPresentProtocolMarker,
		Function:  function,
		GasLimit:  gasLimit,
		Arguments: args,
	}

	return codec.EncodeCallDataStrict(callData)
}
