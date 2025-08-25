package slowTests

import (
	"context"
	"errors"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/mock"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func TestRelayersShouldExecuteTransfersSui(t *testing.T) {
	suiUSDC := GenerateTestUSDCSuiToken()
	suiUSDC.ChainType = currentChainType

	xmnToken := GenerateTestXMNToken()
	xmnToken.ChainType = currentChainType
	xmnToken.InitialSupplyValue = "1000000000"

	_ = testRelayersWithChainSimulatorAndTokens(
		t,
		make(chan error),
		suiUSDC,
		xmnToken,
	)
}

func TestRelayerShouldExecuteTransfersAndNotCatchErrorsSui(t *testing.T) {
	errorString := "ERROR"
	mockLogObserver := mock.NewMockLogObserver(errorString)
	err := logger.AddLogObserver(mockLogObserver, &logger.PlainFormatter{})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, logger.RemoveLogObserver(mockLogObserver))
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopChan := make(chan error, 1000) // ensure sufficient error buffer

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-mockLogObserver.LogFoundChan():
				stopChan <- errors.New("logger should have not caught errors")
			}
		}
	}()

	usdcToken := GenerateTestUSDCSuiToken()
	usdcToken.ChainType = currentChainType

	xmnToken := GenerateTestXMNToken()
	xmnToken.ChainType = currentChainType

	_ = testRelayersWithChainSimulatorAndTokens(
		t,
		stopChan,
		usdcToken,
		xmnToken,
	)
}

func createSuiBadToken() framework.TestTokenParams {
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "BAD",
			NumOfDecimalsUniversal:           6,
			NumOfDecimalsChainSpecific:       6,
			MvxUniversalTokenTicker:          "BAD",
			MvxChainSpecificTokenTicker:      "SUIBAD",
			MvxUniversalTokenDisplayName:     "WrappedBAD",
			MvxChainSpecificTokenDisplayName: "SuiWrappedBAD",
			ValueToMintOnMvx:                 "10000000000",
			PeerChainTokenName:               "SUITOKEN",
			PeerChainTokenSymbol:             "SUIT",
			ValueToMintOnPeerChain:           "10000000000",
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
		ESDTSafeExtraBalance:          big.NewInt(0),
		PeerChainTestAddrExtraBalance: big.NewInt(0),
	}
}

func TestRelayersShouldNotExecuteTransfersSui(t *testing.T) {
	t.Run("IsNativeOnPeerChain = true, IsMintBurnOnPeerChain = false, isNativeOnMvX = true, isMintBurnOnMvX = false", func(t *testing.T) {
		badToken := createSuiBadToken()
		badToken.IsNativeOnPeerChain = true
		badToken.IsMintBurnOnPeerChain = false
		badToken.IsNativeOnMvX = true
		badToken.IsMintBurnOnMvX = false
		badToken.HasChainSpecificToken = true
		badToken.ChainType = currentChainType

		expectedStringInLogs := "error = invalid setup isNativeOnEthereum = true, isNativeOnMultiversX = true"
		testRelayersShouldNotExecuteTransfers(t, expectedStringInLogs, badToken)
	})
	t.Run("IsNativeOnPeerChain = true, IsMintBurnOnPeerChain = false, isNativeOnMvX = true, isMintBurnOnMvX = true", func(t *testing.T) {
		badToken := createSuiBadToken()
		badToken.IsNativeOnPeerChain = true
		badToken.IsMintBurnOnPeerChain = false
		badToken.IsNativeOnMvX = true
		badToken.IsMintBurnOnMvX = true
		badToken.HasChainSpecificToken = false
		badToken.ChainType = currentChainType

		expectedStringInLogs := "error = invalid setup isNativeOnEthereum = true, isNativeOnMultiversX = true"
		testRelayersShouldNotExecuteTransfers(t, expectedStringInLogs, badToken)
	})
}
