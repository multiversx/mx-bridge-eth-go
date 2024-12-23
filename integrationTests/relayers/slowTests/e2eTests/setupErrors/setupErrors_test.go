package setupErrors

import (
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/e2eTests"
)

func TestRelayersShouldNotExecuteTransfersWithNonWhitelistedTokens(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	_ = e2eTests.TestRelayersWithChainSimulatorAndTokens(
		t,
		make(chan error),
		e2eTests.GenerateUnlistedTokenFromEth(),
		e2eTests.GenerateUnlistedTokenFromMvx(),
	)
}

func TestRelayersShouldNotExecuteTransfers(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	t.Run("isNativeOnEth = true, isMintBurnOnEth = false, isNativeOnMvX = true, isMintBurnOnMvX = false", func(t *testing.T) {
		badToken := e2eTests.CreateBadToken()
		badToken.IsNativeOnEth = true
		badToken.IsMintBurnOnEth = false
		badToken.IsNativeOnMvX = true
		badToken.IsMintBurnOnMvX = false
		badToken.HasChainSpecificToken = true

		expectedStringInLogs := "error = invalid setup isNativeOnEthereum = true, isNativeOnMultiversX = true"
		e2eTests.TestRelayersShouldNotExecuteTransfers(t, expectedStringInLogs, badToken)
	})
	t.Run("isNativeOnEth = true, isMintBurnOnEth = false, isNativeOnMvX = true, isMintBurnOnMvX = true", func(t *testing.T) {
		badToken := e2eTests.CreateBadToken()
		badToken.IsNativeOnEth = true
		badToken.IsMintBurnOnEth = false
		badToken.IsNativeOnMvX = true
		badToken.IsMintBurnOnMvX = true
		badToken.HasChainSpecificToken = false

		expectedStringInLogs := "error = invalid setup isNativeOnEthereum = true, isNativeOnMultiversX = true"
		e2eTests.TestRelayersShouldNotExecuteTransfers(t, expectedStringInLogs, badToken)
	})
	t.Run("isNativeOnEth = true, isMintBurnOnEth = true, isNativeOnMvX = true, isMintBurnOnMvX = false", func(t *testing.T) {
		badToken := e2eTests.CreateBadToken()
		badToken.IsNativeOnEth = true
		badToken.IsMintBurnOnEth = true
		badToken.IsNativeOnMvX = true
		badToken.IsMintBurnOnMvX = false
		badToken.HasChainSpecificToken = true

		e2eTests.TestEthContractsShouldError(t, badToken)
	})
	t.Run("isNativeOnEth = false, isMintBurnOnEth = true, isNativeOnMvX = false, isMintBurnOnMvX = true", func(t *testing.T) {
		badToken := e2eTests.CreateBadToken()
		badToken.IsNativeOnEth = false
		badToken.IsMintBurnOnEth = true
		badToken.IsNativeOnMvX = false
		badToken.IsMintBurnOnMvX = true
		badToken.HasChainSpecificToken = true

		e2eTests.TestEthContractsShouldError(t, badToken)
	})
}
