//go:test parallel 1

package edgeCases

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/mock"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/e2eTests"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

func TestRelayerShouldExecuteSimultaneousSwapsAndNotCatchErrors(t *testing.T) {
	if e2eTests.ShouldSkipTest() {
		t.Skip("skipping this test because the .env file is not found in the slowTests directory")
	}

	errorString := "ERROR"
	mockLogObserver := mock.NewMockLogObserver(errorString, "got invalid action ID")
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

	usdcToken := e2eTests.GenerateTestUSDCToken()
	usdcToken.MultipleSpendings = big.NewInt(2)
	usdcToken.TestOperations = []framework.TokenOperations{
		{
			ValueToTransferToMvx: big.NewInt(5000),
			ValueToSendFromMvX:   big.NewInt(200),
			MvxSCCallData:        nil,
			MvxFaultySCCall:      false,
			MvxForceSCCall:       false,
		},
	}
	usdcToken.DeltaBalances = map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
		framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
			framework.Alice: {
				OnEth:    big.NewInt(-5000),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.UniversalToken,
			},
			framework.Bob: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(5000),
				MvxToken: framework.UniversalToken,
			},
			framework.SafeSC: {
				OnEth:    big.NewInt(5000),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.ChainSpecificToken,
			},
			framework.WrapperSC: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(5000),
				MvxToken: framework.ChainSpecificToken,
			},
		},
		framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
			framework.Alice: {
				OnEth:    big.NewInt(-5000 - 5000 + 150),
				OnMvx:    big.NewInt(0),
				MvxToken: framework.UniversalToken,
			},
			framework.Bob: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(5000 + 4800),
				MvxToken: framework.UniversalToken,
			},
			framework.SafeSC: {
				OnEth:    big.NewInt(5000 + 5000 - 150),
				OnMvx:    big.NewInt(50),
				MvxToken: framework.ChainSpecificToken,
			},
			framework.WrapperSC: {
				OnEth:    big.NewInt(0),
				OnMvx:    big.NewInt(5000 + 5000 - 200),
				MvxToken: framework.ChainSpecificToken,
			},
		},
	}
	usdcToken.MintBurnChecks = &framework.MintBurnBalances{
		MvxTotalUniversalMint:     big.NewInt(5000 + 5000),
		MvxTotalChainSpecificMint: big.NewInt(5000 + 5000),
		MvxTotalUniversalBurn:     big.NewInt(200),
		MvxTotalChainSpecificBurn: big.NewInt(200 - 50),
		MvxSafeMintValue:          big.NewInt(5000 + 5000),
		MvxSafeBurnValue:          big.NewInt(200 - 50),

		EthSafeMintValue: big.NewInt(0),
		EthSafeBurnValue: big.NewInt(0),
	}
	usdcToken.SpecialChecks.WrapperDeltaLiquidityCheck = big.NewInt(5000 + 5000 - 200)

	_ = testRelayersWithChainSimulatorAndTokensForSimultaneousSwaps(
		t,
		stopChan,
		usdcToken,
	)
}

func testRelayersWithChainSimulatorAndTokensForSimultaneousSwaps(tb testing.TB, manualStopChan chan error, tokens ...framework.TestTokenParams) *framework.TestSetup {
	startsFromEthFlow := &e2eTests.TestFlow{
		TB:                           tb,
		Tokens:                       tokens,
		MessageAfterFirstHalfBridge:  "Ethereum->MultiversX transfer finished, now sending back to Ethereum & another round from Ethereum...",
		MessageAfterSecondHalfBridge: "MultiversX<->Ethereum from Ethereum transfers done",
	}
	startsFromEthFlow.HandlerAfterFirstHalfBridge = func(flow *e2eTests.TestFlow) {
		flow.Setup.SendFromMultiversxToEthereum(flow.Setup.BobKeys, flow.Setup.AliceKeys, flow.Tokens...)
		flow.Setup.SendFromEthereumToMultiversX(flow.Setup.AliceKeys, flow.Setup.BobKeys, flow.Setup.MultiversxHandler.CalleeScAddress, flow.Tokens...)
	}

	setupFunc := func(tb testing.TB, setup *framework.TestSetup) {
		startsFromEthFlow.Setup = setup

		setup.IssueAndConfigureTokens(tokens...)
		setup.MultiversxHandler.CheckForZeroBalanceOnReceivers(setup.Ctx, tokens...)
		setup.CreateBatchOnEthereum(setup.MultiversxHandler.CalleeScAddress, startsFromEthFlow.Tokens...)
	}

	processFunc := func(tb testing.TB, setup *framework.TestSetup) bool {
		if startsFromEthFlow.Process() {
			setup.TestWithdrawTotalFeesOnEthereumForTokens(startsFromEthFlow.Tokens...)

			return true
		}

		setup.EthereumHandler.SimulatedChain.Commit()
		setup.ChainSimulator.GenerateBlocks(setup.Ctx, 1)
		require.LessOrEqual(tb, setup.ScCallerModuleInstance.GetNumSentTransaction(), setup.GetNumScCallsOperations())

		return false
	}

	return e2eTests.TestRelayersWithChainSimulator(tb,
		setupFunc,
		processFunc,
		manualStopChan,
	)
}
