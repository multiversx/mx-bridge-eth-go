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
		},
		DeltaBalances: map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-5000 - 7000 - 1000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(5000 + 7000),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(5000 + 7000 + 1000),
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
					OnEth:    big.NewInt(-5000 - 7000 - 1000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(2500 + 6700),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(2450 + 250),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(5000 + 7000 + 1000 - 2450 - 250),
					OnMvx:    big.NewInt(50 + 50),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(1000),
					MvxToken: framework.UniversalToken,
				},
			},
		},
	}
}

// ApplyUSDCRefundBalances will apply the refund balances on the involved entities for the USDC token
func ApplyUSDCRefundBalances(token *framework.TestTokenParams) {
	// extra is just for the fees for the 2 transfers mvx->eth and the failed eth->mvx that needed refund
	token.DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50 + 50 + 50)
	// we need to subtract the refunded value from the Ethereum Safe contract
	token.DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnEth = big.NewInt(5000 + 7000 + 1000 - 2450 - 250 - 950)
	// Alice will get her tokens back from the refund
	token.DeltaBalances[framework.SecondHalfBridge][framework.Alice].OnEth = big.NewInt(-5000 - 7000 - 1000 + 950)
	// no funds remain in the test caller SC
	token.DeltaBalances[framework.SecondHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(0)
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
			HasChainSpecificToken:            true,
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
		},
		DeltaBalances: map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(-4000 - 6000 - 2000),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(4000 - 50 + 6000 - 50 + 2000 - 50),
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
			},
			framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(-4000 - 6000 - 2000),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(4000 - 50 - 2400 + 6000 - 50 - 200 + 2000 - 50 - 1000),
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
					OnMvx:    big.NewInt(4000 - 2400 + 6000 - 200 + 2000 - 1000),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(1000),
					MvxToken: framework.UniversalToken,
				},
			},
		},
	}
}

// ApplyMEMERefundBalances will apply the refund balances on the involved entities for the MEME token
func ApplyMEMERefundBalances(token *framework.TestTokenParams) {
	// we need to add the 1000 MEME tokens as the third bridge was done that include the refund on the Ethereum side
	token.DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(4000 - 2400 + 6000 - 200 + 2000 - 1000 + 1000)
	// Bob will get his tokens back from the refund
	token.DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnEth = big.NewInt(4000 - 50 - 2400 + 6000 - 50 - 200 + 2000 - 50 - 1000 + 950)
	// no funds remain in the test caller SC
	token.DeltaBalances[framework.SecondHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(0)
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
		},
		DeltaBalances: map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-5010 - 7010 - 1010),
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
			},
			framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-5010 - 7010 - 1010),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(5010 - 2510 + 7010 - 310),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(2510 - 50 + 310 - 50),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(50 + 50),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(1010),
					MvxToken: framework.UniversalToken,
				},
			},
		},
	}
}

// GenerateTestMEXToken will generate a test EUROC token
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
		},
		DeltaBalances: map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(-4010 - 6010 - 2010),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(4010 - 50 + 6010 - 50 + 2010 - 50),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(50 + 50 + 50),
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
					OnMvx:    big.NewInt(-4010 - 6010 - 2010),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(4010 - 50 - 2410 + 6010 - 50 - 210 + 2010 - 50 - 1010),
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
					OnMvx:    big.NewInt(50 + 50 + 50),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(1010),
					MvxToken: framework.UniversalToken,
				},
			},
		},
	}
}

// ApplyMEXRefundBalances will apply the refund balances on the involved entities for the MEX token
func ApplyMEXRefundBalances(token *framework.TestTokenParams) {
	// 3 normal swaps + the refund one
	token.DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50 + 50 + 50 + 50)
	// Bob will get his tokens back from the refund
	token.DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnEth = big.NewInt(4010 - 50 - 2410 + 6010 - 50 - 210 + 2010 - 50 - 1010 + 960)
	// no funds remain in the test caller SC
	token.DeltaBalances[framework.SecondHalfBridge][framework.CalledTestSC].OnMvx = big.NewInt(0)
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
