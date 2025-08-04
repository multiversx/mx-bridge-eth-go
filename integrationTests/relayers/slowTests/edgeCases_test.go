//go:build slow

package slowTests

import (
	"context"
	"errors"
	"fmt"
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
	usdcToken.ESDTSafeExtraBalance = big.NewInt(50)
	usdcToken.PeerChainTestAddrExtraBalance = big.NewInt(-5000 - 5000 + 200 - 50)

	_ = testRelayersWithChainSimulatorAndTokensForSimultaneousSwaps(
		t,
		stopChan,
		usdcToken,
	)
}

func testRelayersWithChainSimulatorAndTokensForSimultaneousSwaps(tb testing.TB, manualStopChan chan error, tokens ...framework.TestTokenParams) *framework.TestSetup {
	startsFromEthFlow := &startsFromEthereumEdgecaseFlow{
		TB:     tb,
		tokens: tokens,
	}

	setupFunc := func(tb testing.TB, setup *framework.TestSetup) {
		startsFromEthFlow.setup = setup

		setup.IssueAndConfigureTokens(tokens...)
		setup.MultiversxHandler.CheckForZeroBalanceOnReceivers(setup.Ctx, tokens...)
		setup.PeerChainHandler.CreateBatchOnPeerChain(setup.Ctx, setup.MultiversxHandler.TestCallerAddress, startsFromEthFlow.tokens...)
	}

	processFunc := func(tb testing.TB, setup *framework.TestSetup) bool {
		if startsFromEthFlow.process() {
			setup.TestWithdrawTotalFeesOnPeerChainForTokens(startsFromEthFlow.tokens...)

			return true
		}

		switch handler := setup.PeerChainHandler.(type) {
		case *framework.EthereumHandler:
			handler.SimulatedChain.Commit()
		case *framework.SuiHandler:
			panic("sui chain simulator not yet implemented")
		default:
			panic(fmt.Sprintf("unsupported peer chain handler type: %T", handler))
		}
		setup.ChainSimulator.GenerateBlocks(setup.Ctx, 1)
		require.LessOrEqual(tb, setup.ScCallerModuleInstance.GetNumSentTransaction(), setup.GetNumScCallsOperations())

		return false
	}

	chainType := tokens[0].ChainType

	return testRelayersWithChainSimulator(tb,
		setupFunc,
		processFunc,
		manualStopChan,
		chainType,
	)
}
