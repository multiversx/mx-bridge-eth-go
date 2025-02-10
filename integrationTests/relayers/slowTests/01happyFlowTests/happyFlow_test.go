//go:build slow

package happyFlowTests

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/mock"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

func TestRelayersShouldExecuteTransfers(t *testing.T) {
	t.Run("lock-unlock tokens", func(t *testing.T) {
		_ = slowTests.NewTestEnvironmentWithChainSimulatorAndTokens(
			t,
			make(chan error),
			slowTests.GenerateTestUSDCToken(),
			slowTests.GenerateTestMEMEToken(),
		)
	})
	t.Run("mint-burn tokens", func(t *testing.T) {
		_ = slowTests.NewTestEnvironmentWithChainSimulatorAndTokens(
			t,
			make(chan error),
			slowTests.GenerateTestEUROCToken(),
			slowTests.GenerateTestMEXToken(),
			slowTests.GenerateTestTADAToken(),
		)
	})
	t.Run("lock-unlock tokens with arguments on SC call", func(t *testing.T) {
		dummyAddress := strings.Repeat("2", 32)
		dummyUint64 := string([]byte{37})

		callData := slowTests.CreateScCallData("callPayableWithParams", 50000000, dummyUint64, dummyAddress)

		usdcToken := slowTests.GenerateTestUSDCToken()
		usdcToken.TestOperations[2].MvxSCCallData = callData

		memeToken := slowTests.GenerateTestMEMEToken()
		memeToken.TestOperations[2].MvxSCCallData = callData

		testSetup := slowTests.NewTestEnvironmentWithChainSimulatorAndTokens(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)

		testSetup.TestCallPayableWithParamsWasCalled(
			37,
			usdcToken.AbstractTokenIdentifier,
			memeToken.AbstractTokenIdentifier,
		)
	})
	t.Run("mint-burn tokens with arguments on SC call", func(t *testing.T) {
		dummyAddress := strings.Repeat("2", 32)
		dummyUint64 := string([]byte{37})

		callData := slowTests.CreateScCallData("callPayableWithParams", 50000000, dummyUint64, dummyAddress)

		eurocToken := slowTests.GenerateTestEUROCToken()
		eurocToken.TestOperations[2].MvxSCCallData = callData

		mexToken := slowTests.GenerateTestMEXToken()
		mexToken.TestOperations[2].MvxSCCallData = callData

		tadaToken := slowTests.GenerateTestTADAToken()
		tadaToken.TestOperations[2].MvxSCCallData = callData

		testSetup := slowTests.NewTestEnvironmentWithChainSimulatorAndTokens(
			t,
			make(chan error),
			eurocToken,
			mexToken,
			tadaToken,
		)

		testSetup.TestCallPayableWithParamsWasCalled(
			37,
			eurocToken.AbstractTokenIdentifier,
			mexToken.AbstractTokenIdentifier,
			tadaToken.AbstractTokenIdentifier,
		)
	})
	t.Run("on a valid setup, errors should not be caught", func(t *testing.T) {
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

		_ = slowTests.NewTestEnvironmentWithChainSimulatorAndTokens(
			t,
			stopChan,
			slowTests.GenerateTestMEMEToken(),
		)
	})
	t.Run("lock-unlock tokens with init supply", func(t *testing.T) {
		usdcInitialValue := big.NewInt(100000)
		usdcToken := slowTests.GenerateTestUSDCToken()
		usdcToken.InitialSupplyValue = usdcInitialValue.String()
		usdcToken.MintBurnChecks.MvxSafeMintValue.Add(usdcToken.MintBurnChecks.MvxSafeMintValue, usdcInitialValue)

		memeInitialValue := big.NewInt(200000)
		memeToken := slowTests.GenerateTestMEMEToken()
		memeToken.InitialSupplyValue = memeInitialValue.String()
		memeToken.MintBurnChecks.EthSafeMintValue.Add(memeToken.MintBurnChecks.EthSafeMintValue, memeInitialValue)

		_ = slowTests.NewTestEnvironmentWithChainSimulatorAndTokens(
			t,
			make(chan error),
			usdcToken,
			memeToken,
		)
	})
	t.Run("mint-burn tokens with init supply", func(t *testing.T) {
		eurocInitialValue := big.NewInt(100010)
		eurocToken := slowTests.GenerateTestEUROCToken()
		eurocToken.InitialSupplyValue = eurocInitialValue.String()
		eurocToken.MintBurnChecks.MvxSafeMintValue.Add(eurocToken.MintBurnChecks.MvxSafeMintValue, eurocInitialValue)
		eurocToken.MintBurnChecks.EthSafeBurnValue.Add(eurocToken.MintBurnChecks.EthSafeBurnValue, eurocInitialValue)

		mexInitialValue := big.NewInt(300000)
		mexToken := slowTests.GenerateTestMEXToken()
		mexToken.InitialSupplyValue = mexInitialValue.String()
		mexToken.MintBurnChecks.MvxSafeBurnValue.Add(mexToken.MintBurnChecks.MvxSafeBurnValue, mexInitialValue)
		mexToken.MintBurnChecks.EthSafeMintValue.Add(mexToken.MintBurnChecks.EthSafeMintValue, mexInitialValue)

		tadaInitialValue := big.NewInt(300000)
		tadaToken := slowTests.GenerateTestTADAToken()
		tadaToken.InitialSupplyValue = tadaInitialValue.String()
		tadaToken.MintBurnChecks.MvxSafeBurnValue.Add(tadaToken.MintBurnChecks.MvxSafeBurnValue, tadaInitialValue)
		tadaToken.MintBurnChecks.EthSafeMintValue.Add(tadaToken.MintBurnChecks.EthSafeMintValue, tadaInitialValue)

		_ = slowTests.NewTestEnvironmentWithChainSimulatorAndTokens(
			t,
			make(chan error),
			eurocToken,
			mexToken,
			tadaToken,
		)
	})
}
