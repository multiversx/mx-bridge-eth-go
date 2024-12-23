package happyScenarios

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/mock"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/e2eTests"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

func TestRelayersShouldExecuteTransfers(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	_ = e2eTests.TestRelayersWithChainSimulatorAndTokens(
		t,
		make(chan error),
		e2eTests.GenerateTestUSDCToken(),
		e2eTests.GenerateTestMEMEToken(),
	)
}

func TestRelayersShouldExecuteTransfersWithMintBurnTokens(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	_ = e2eTests.TestRelayersWithChainSimulatorAndTokens(
		t,
		make(chan error),
		e2eTests.GenerateTestEUROCToken(),
		e2eTests.GenerateTestMEXToken(),
	)
}

func TestRelayersShouldExecuteTransfersWithSCCallsWithArguments(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	dummyAddress := strings.Repeat("2", 32)
	dummyUint64 := string([]byte{37})

	callData := e2eTests.CreateScCallData("callPayableWithParams", 50000000, dummyUint64, dummyAddress)

	usdcToken := e2eTests.GenerateTestUSDCToken()
	usdcToken.TestOperations[2].MvxSCCallData = callData

	memeToken := e2eTests.GenerateTestMEMEToken()
	memeToken.TestOperations[2].MvxSCCallData = callData

	testSetup := e2eTests.TestRelayersWithChainSimulatorAndTokens(
		t,
		make(chan error),
		usdcToken,
		memeToken,
	)

	e2eTests.TestCallPayableWithParamsWasCalled(
		testSetup,
		37,
		usdcToken.AbstractTokenIdentifier,
		memeToken.AbstractTokenIdentifier,
	)
}

func TestRelayersShouldExecuteTransfersWithSCCallsWithArgumentsWithMintBurnTokens(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	dummyAddress := strings.Repeat("2", 32)
	dummyUint64 := string([]byte{37})

	callData := e2eTests.CreateScCallData("callPayableWithParams", 50000000, dummyUint64, dummyAddress)

	eurocToken := e2eTests.GenerateTestEUROCToken()
	eurocToken.TestOperations[2].MvxSCCallData = callData

	mexToken := e2eTests.GenerateTestMEXToken()
	mexToken.TestOperations[2].MvxSCCallData = callData

	testSetup := e2eTests.TestRelayersWithChainSimulatorAndTokens(
		t,
		make(chan error),
		eurocToken,
		mexToken,
	)

	e2eTests.TestCallPayableWithParamsWasCalled(
		testSetup,
		37,
		eurocToken.AbstractTokenIdentifier,
		mexToken.AbstractTokenIdentifier,
	)
}

func TestRelayerShouldExecuteTransfersAndNotCatchErrors(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

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

	_ = e2eTests.TestRelayersWithChainSimulatorAndTokens(
		t,
		stopChan,
		e2eTests.GenerateTestMEMEToken(),
	)
}

func TestRelayersShouldExecuteTransfersWithInitSupply(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	usdcInitialValue := big.NewInt(100000)
	usdcToken := e2eTests.GenerateTestUSDCToken()
	usdcToken.InitialSupplyValue = usdcInitialValue.String()
	usdcToken.MintBurnChecks.MvxSafeMintValue.Add(usdcToken.MintBurnChecks.MvxSafeMintValue, usdcInitialValue)

	memeInitialValue := big.NewInt(200000)
	memeToken := e2eTests.GenerateTestMEMEToken()
	memeToken.InitialSupplyValue = memeInitialValue.String()
	memeToken.MintBurnChecks.EthSafeMintValue.Add(memeToken.MintBurnChecks.EthSafeMintValue, memeInitialValue)

	_ = e2eTests.TestRelayersWithChainSimulatorAndTokens(
		t,
		make(chan error),
		usdcToken,
		memeToken,
	)
}

func TestRelayersShouldExecuteTransfersWithInitSupplyMintBurn(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	eurocInitialValue := big.NewInt(100010)
	eurocToken := e2eTests.GenerateTestEUROCToken()
	eurocToken.InitialSupplyValue = eurocInitialValue.String()
	eurocToken.MintBurnChecks.MvxSafeMintValue.Add(eurocToken.MintBurnChecks.MvxSafeMintValue, eurocInitialValue)
	eurocToken.MintBurnChecks.EthSafeBurnValue.Add(eurocToken.MintBurnChecks.EthSafeBurnValue, eurocInitialValue)

	mexInitialValue := big.NewInt(300000)
	mexToken := e2eTests.GenerateTestMEXToken()
	mexToken.InitialSupplyValue = mexInitialValue.String()
	mexToken.MintBurnChecks.MvxSafeBurnValue.Add(mexToken.MintBurnChecks.MvxSafeBurnValue, mexInitialValue)
	mexToken.MintBurnChecks.EthSafeMintValue.Add(mexToken.MintBurnChecks.EthSafeMintValue, mexInitialValue)

	_ = e2eTests.TestRelayersWithChainSimulatorAndTokens(
		t,
		make(chan error),
		eurocToken,
		mexToken,
	)
}
