//go:build slow

package setupErrorTests

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/log"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/mock"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

const projectedShardForTestKeys = byte(2)

func TestRelayersShouldNotExecuteTransfers(t *testing.T) {
	t.Run("non-whitelisted tokens", func(t *testing.T) {
		_ = slowTests.NewTestEnvironmentWithChainSimulatorAndTokens(
			t,
			make(chan error),
			slowTests.GenerateUnlistedTokenFromEth(),
			slowTests.GenerateUnlistedTokenFromMvx(),
		)
	})
	t.Run("bad setup", func(t *testing.T) {
		t.Run("isNativeOnEth = true, isMintBurnOnEth = false, isNativeOnMvX = true, isMintBurnOnMvX = false", func(t *testing.T) {
			badToken := createBadToken()
			badToken.IsNativeOnEth = true
			badToken.IsMintBurnOnEth = false
			badToken.IsNativeOnMvX = true
			badToken.IsMintBurnOnMvX = false
			badToken.HasChainSpecificToken = true

			expectedStringInLogs := "error = invalid setup isNativeOnEthereum = true, isNativeOnMultiversX = true"
			testRelayersShouldNotExecuteTransfers(t, expectedStringInLogs, badToken)
		})
		t.Run("isNativeOnEth = true, isMintBurnOnEth = false, isNativeOnMvX = true, isMintBurnOnMvX = true", func(t *testing.T) {
			badToken := createBadToken()
			badToken.IsNativeOnEth = true
			badToken.IsMintBurnOnEth = false
			badToken.IsNativeOnMvX = true
			badToken.IsMintBurnOnMvX = true
			badToken.HasChainSpecificToken = false

			expectedStringInLogs := "error = invalid setup isNativeOnEthereum = true, isNativeOnMultiversX = true"
			testRelayersShouldNotExecuteTransfers(t, expectedStringInLogs, badToken)
		})
		t.Run("isNativeOnEth = true, isMintBurnOnEth = true, isNativeOnMvX = true, isMintBurnOnMvX = false", func(t *testing.T) {
			badToken := createBadToken()
			badToken.IsNativeOnEth = true
			badToken.IsMintBurnOnEth = true
			badToken.IsNativeOnMvX = true
			badToken.IsMintBurnOnMvX = false
			badToken.HasChainSpecificToken = true

			testEthContractsShouldError(t, badToken)
		})
		t.Run("isNativeOnEth = false, isMintBurnOnEth = true, isNativeOnMvX = false, isMintBurnOnMvX = true", func(t *testing.T) {
			badToken := createBadToken()
			badToken.IsNativeOnEth = false
			badToken.IsMintBurnOnEth = true
			badToken.IsNativeOnMvX = false
			badToken.IsMintBurnOnMvX = true
			badToken.HasChainSpecificToken = true

			testEthContractsShouldError(t, badToken)
		})
	})
}

func createBadToken() framework.TestTokenParams {
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "BAD",
			NumOfDecimalsUniversal:           6,
			NumOfDecimalsChainSpecific:       6,
			MvxUniversalTokenTicker:          "BAD",
			MvxChainSpecificTokenTicker:      "ETHBAD",
			MvxUniversalTokenDisplayName:     "WrappedBAD",
			MvxChainSpecificTokenDisplayName: "EthereumWrappedBAD",
			MvxToEthFee:                      big.NewInt(50),
			ValueToMintOnMvx:                 "10000000000",
			EthTokenName:                     "ETHTOKEN",
			EthTokenSymbol:                   "ETHT",
			ValueToMintOnEth:                 "10000000000",
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
				MvxSCCallData:        slowTests.CreateScCallData("callPayable", 50000000),
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
					OnEth:    big.NewInt(5000 + 7000),
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
		},
	}
}

func testRelayersShouldNotExecuteTransfers(
	tb testing.TB,
	expectedStringInLogs string,
	tokens ...framework.TestTokenParams,
) {
	flows := slowTests.CreateFlowsBasedOnToken(tb, tokens...)

	setupFunc := func(tb testing.TB, setup *framework.TestSetup) {
		for _, flow := range flows {
			flow.Setup = setup
		}

		setup.IssueAndConfigureTokens(tokens...)
		setup.MultiversxHandler.CheckForZeroBalanceOnReceivers(setup.Ctx, tokens...)

		for _, flow := range flows {
			flow.HandlerToStartFirstBridge(flow)
		}
	}

	processFunc := func(tb testing.TB, setup *framework.TestSetup) bool {
		allFlowsFinished := true
		for _, flow := range flows {
			allFlowsFinished = allFlowsFinished && flow.Process()
		}

		// commit blocks in order to execute incoming txs from relayers
		setup.EthereumHandler.SimulatedChain.Commit()
		setup.ChainSimulator.GenerateBlocks(setup.Ctx, 1)

		return allFlowsFinished
	}

	// start a mocked log observer that is looking for a specific relayer error
	chanCnt := 0
	mockLogObserver := mock.NewMockLogObserver(expectedStringInLogs)
	err := logger.AddLogObserver(mockLogObserver, &logger.PlainFormatter{})
	require.NoError(tb, err)
	defer func() {
		require.NoError(tb, logger.RemoveLogObserver(mockLogObserver))
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	numOfTimesToRepeatErrorForRelayer := 10
	numOfErrorsToWait := numOfTimesToRepeatErrorForRelayer * framework.NumRelayers

	stopChan := make(chan error, 1)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-mockLogObserver.LogFoundChan():
				chanCnt++
				if chanCnt >= numOfErrorsToWait {
					log.Info(fmt.Sprintf("test passed, relayers are stuck, expected string `%s` found in all relayers' logs for %d times", expectedStringInLogs, numOfErrorsToWait))
					stopChan <- nil
					return
				}
			}
		}
	}()

	_ = slowTests.NewTestEnvironmentWithChainSimulator(tb, setupFunc, processFunc, stopChan)
}

func testEthContractsShouldError(tb testing.TB, testToken framework.TestTokenParams) {
	setupFunc := func(tb testing.TB, setup *framework.TestSetup) {
		setup.IssueAndConfigureTokens(testToken)

		token := setup.GetTokenData(testToken.AbstractTokenIdentifier)
		require.NotNil(tb, token)

		valueToMintOnEth, ok := big.NewInt(0).SetString(testToken.ValueToMintOnEth, 10)
		require.True(tb, ok)

		receiverKeys := framework.GenerateMvxPrivatePublicKey(tb, projectedShardForTestKeys)
		auth, _ := bind.NewKeyedTransactorWithChainID(setup.DepositorKeys.EthSK, setup.EthereumHandler.ChainID)
		_, err := setup.EthereumHandler.SafeContract.Deposit(auth, token.EthErc20Address, valueToMintOnEth, receiverKeys.MvxAddress.AddressSlice())
		require.Error(tb, err)
	}

	processFunc := func(tb testing.TB, setup *framework.TestSetup) bool {
		time.Sleep(time.Second) // allow go routines to start
		return true
	}

	_ = slowTests.NewTestEnvironmentWithChainSimulator(
		tb,
		setupFunc,
		processFunc,
		make(chan error),
	)
}
