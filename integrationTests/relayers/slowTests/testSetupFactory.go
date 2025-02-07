package slowTests

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
	"github.com/stretchr/testify/require"
)

const (
	timeout = time.Minute * 15
)

// NewTestEnvironmentWithChainSimulatorAndTokens creates a new test environment with chain simulator and tokens
func NewTestEnvironmentWithChainSimulatorAndTokens(tb testing.TB, manualStopChan chan error, tokens ...framework.TestTokenParams) *framework.TestSetup {
	flows := CreateFlowsBasedOnToken(tb, tokens...)

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

		if allFlowsFinished {
			for _, flow := range flows {
				setup.TestWithdrawTotalFeesOnEthereumForTokens(flow.Tokens...)
			}

			return true
		}

		// commit blocks in order to execute incoming txs from relayers
		setup.EthereumHandler.SimulatedChain.Commit()
		setup.ChainSimulator.GenerateBlocks(setup.Ctx, 1)
		require.LessOrEqual(tb, setup.ScCallerModuleInstance.GetNumSentTransaction(), setup.GetNumScCallsOperations())

		return false
	}

	return NewTestEnvironmentWithChainSimulator(tb,
		setupFunc,
		processFunc,
		manualStopChan,
	)
}

// CreateFlowsBasedOnToken will split the provided tokens in categories and wrap them in test flow instances
func CreateFlowsBasedOnToken(tb testing.TB, tokens ...framework.TestTokenParams) []*TestFlow {
	startsFromEthFlow := &TestFlow{
		FlowType:                     StartFromEthereumFlow,
		TB:                           tb,
		Tokens:                       make([]framework.TestTokenParams, 0, len(tokens)),
		MessageAfterFirstHalfBridge:  "Ethereum->MultiversX transfer finished, now sending back to Ethereum...",
		MessageAfterSecondHalfBridge: "MultiversX<->Ethereum from Ethereum transfers done",
	}
	startsFromEthFlow.HandlerAfterFirstHalfBridge = func(flow *TestFlow) {
		flow.Setup.SendFromMultiversxToEthereum(flow.Setup.BobKeys, flow.Setup.CharlieKeys, flow.Tokens...)
	}
	startsFromEthFlow.HandlerToStartFirstBridge = func(flow *TestFlow) {
		if len(flow.Tokens) == 0 {
			return
		}

		flow.Setup.CreateBatchOnEthereum(flow.Setup.MultiversxHandler.CalleeScAddress, startsFromEthFlow.Tokens...)
	}

	startsFromMvXFlow := &TestFlow{
		FlowType:                     StartFromMultiversXFlow,
		TB:                           tb,
		Tokens:                       make([]framework.TestTokenParams, 0, len(tokens)),
		MessageAfterFirstHalfBridge:  "MultiversX->Ethereum transfer finished, now sending back to MultiversX...",
		MessageAfterSecondHalfBridge: "MultiversX<->Ethereum from MultiversX transfers done",
	}
	startsFromMvXFlow.HandlerAfterFirstHalfBridge = func(flow *TestFlow) {
		flow.Setup.SendFromEthereumToMultiversX(flow.Setup.BobKeys, flow.Setup.CharlieKeys, flow.Setup.MultiversxHandler.CalleeScAddress, flow.Tokens...)
	}
	startsFromMvXFlow.HandlerToStartFirstBridge = func(flow *TestFlow) {
		if len(flow.Tokens) == 0 {
			return
		}

		flow.Setup.CreateBatchOnMultiversX(startsFromMvXFlow.Tokens...)
	}

	// split the tokens from where should the bridge start
	for _, token := range tokens {
		if token.IsNativeOnEth {
			startsFromEthFlow.Tokens = append(startsFromEthFlow.Tokens, token)
			continue
		}
		if token.IsNativeOnMvX {
			startsFromMvXFlow.Tokens = append(startsFromMvXFlow.Tokens, token)
			continue
		}
		require.Fail(tb, "invalid setup, found a token that is not native on any chain", "abstract identifier", token.AbstractTokenIdentifier)
	}

	return []*TestFlow{startsFromEthFlow, startsFromMvXFlow}
}

// NewTestEnvironmentWithChainSimulator creates a new test environment with chain simulator and general setup & process functions
func NewTestEnvironmentWithChainSimulator(tb testing.TB,
	setupFunc func(tb testing.TB, setup *framework.TestSetup),
	processLoopFunc func(tb testing.TB, setup *framework.TestSetup) bool,
	stopChan chan error,
) *framework.TestSetup {
	defer func() {
		r := recover()
		if r != nil {
			require.Fail(tb, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	testSetup := framework.NewTestSetup(tb)
	log.Info(fmt.Sprintf(framework.LogStepMarker, "calling setupFunc"))
	setupFunc(tb, testSetup)

	testSetup.StartRelayersAndScModule()
	defer testSetup.Close()

	log.Info(fmt.Sprintf(framework.LogStepMarker, "running and continously call processLoopFunc"))
	interrupt := make(chan os.Signal, 1)
	for {
		select {
		case <-interrupt:
			require.Fail(tb, "signal interrupted")
			return testSetup
		case <-time.After(timeout):
			require.Fail(tb, "time out")
			return testSetup
		case err := <-stopChan:
			require.Nil(tb, err)
			return testSetup
		default:
			testDone := processLoopFunc(tb, testSetup)
			if testDone {
				return testSetup
			}
		}
	}
}

// NewTestEnvironmentWithChainSimulatorAndTokensAndRefund creates a new test environment with chain simulator and tokens prepared for refunds
func NewTestEnvironmentWithChainSimulatorAndTokensAndRefund(tb testing.TB, manualStopChan chan error, tokens ...framework.TestTokenParams) {
	flows := CreateFlowsBasedOnToken(tb, tokens...)

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
		require.LessOrEqual(tb, setup.ScCallerModuleInstance.GetNumSentTransaction(), setup.GetNumScCallsOperations())

		return allFlowsFinished
	}

	_ = NewTestEnvironmentWithChainSimulator(tb,
		setupFunc,
		processFunc,
		manualStopChan,
	)
}
