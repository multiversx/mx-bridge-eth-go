//go:build slow

package slowTests

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/mock"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

func TestRelayerShouldExecuteSimultaneousSwapsAndNotCatchErrors(t *testing.T) {
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

	usdcToken := GenerateTestUSDCToken()
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
		TotalUniversalMint:     big.NewInt(5000 + 5000),
		TotalChainSpecificMint: big.NewInt(5000 + 5000),
		TotalUniversalBurn:     big.NewInt(200),
		TotalChainSpecificBurn: big.NewInt(200 - 50),
		SafeMintValue:          big.NewInt(5000 + 5000),
		SafeBurnValue:          big.NewInt(200 - 50),
	}
	usdcToken.SpecialChecks.WrapperDeltaLiquidityCheck = big.NewInt(5000 + 5000 - 200)

	_ = testRelayersWithChainSimulatorAndTokensForSimultaneousSwaps(
		t,
		stopChan,
		usdcToken,
	)
}

func testRelayersWithChainSimulatorAndTokensForSimultaneousSwaps(tb testing.TB, manualStopChan chan error, tokens ...framework.TestTokenParams) *framework.TestSetup {
	startsFromEthFlow := &testFlow{
		TB:                           tb,
		tokens:                       tokens,
		messageAfterFirstHalfBridge:  "Ethereum->MultiversX transfer finished, now sending back to Ethereum & another round from Ethereum...",
		messageAfterSecondHalfBridge: "MultiversX<->Ethereum from Ethereum transfers done",
	}
	startsFromEthFlow.handlerAfterFirstHalfBridge = func(flow *testFlow) {
		flow.setup.SendFromMultiversxToEthereum(flow.setup.BobKeys, flow.setup.AliceKeys, flow.tokens...)
		flow.setup.SendFromEthereumToMultiversX(flow.setup.AliceKeys, flow.setup.BobKeys, flow.setup.MultiversxHandler.CalleeScAddress, flow.tokens...)
	}

	setupFunc := func(tb testing.TB, setup *framework.TestSetup) {
		startsFromEthFlow.setup = setup

		setup.IssueAndConfigureTokens(tokens...)
		setup.MultiversxHandler.CheckForZeroBalanceOnReceivers(setup.Ctx, tokens...)
		setup.CreateBatchOnEthereum(setup.MultiversxHandler.CalleeScAddress, startsFromEthFlow.tokens...)
	}

	processFunc := func(tb testing.TB, setup *framework.TestSetup) bool {
		if startsFromEthFlow.process() {
			setup.TestWithdrawTotalFeesOnEthereumForTokens(startsFromEthFlow.tokens...)

			return true
		}

		setup.EthereumHandler.SimulatedChain.Commit()
		setup.ChainSimulator.GenerateBlocks(setup.Ctx, 1)
		require.LessOrEqual(tb, setup.ScCallerModuleInstance.GetNumSentTransaction(), setup.GetNumScCallsOperations())

		return false
	}

	return testRelayersWithChainSimulator(tb,
		setupFunc,
		processFunc,
		manualStopChan,
	)
}
