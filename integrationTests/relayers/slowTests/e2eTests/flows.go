package e2eTests

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/mock"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const timeout = time.Minute * 15

// TestRelayersWithChainSimulatorAndTokens creates a new test setup and a running process
func TestRelayersWithChainSimulatorAndTokens(tb testing.TB, manualStopChan chan error, tokens ...framework.TestTokenParams) *framework.TestSetup {
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

	return TestRelayersWithChainSimulator(tb,
		setupFunc,
		processFunc,
		manualStopChan,
	)
}

// CreateFlowsBasedOnToken splits the tokens by their native chain. Will put those 2 sets in 2 test flows instances
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

// TestRelayersWithChainSimulator triggers the execution of the test
func TestRelayersWithChainSimulator(tb testing.TB,
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

// TestCallPayableWithParamsWasCalled creates a new test setup and a running process
func TestCallPayableWithParamsWasCalled(testSetup *framework.TestSetup, value uint64, tokens ...string) {
	if len(tokens) == 0 {
		return
	}

	universalTokens := make([]string, 0, len(tokens))
	for _, identifier := range tokens {
		tkData := testSetup.TokensRegistry.GetTokenData(identifier)
		universalTokens = append(universalTokens, tkData.MvxUniversalToken)
	}

	vmRequest := &data.VmValueRequest{
		Address:  testSetup.MultiversxHandler.CalleeScAddress.Bech32(),
		FuncName: "getCalledDataParams",
	}

	vmResponse, err := testSetup.ChainSimulator.Proxy().ExecuteVMQuery(context.Background(), vmRequest)
	require.Nil(testSetup, err)

	returnedData := vmResponse.Data.ReturnData
	require.Equal(testSetup, len(tokens), len(returnedData))

	mapUniversalTokens := make(map[string]int)
	for _, tokenIdentifier := range universalTokens {
		mapUniversalTokens[tokenIdentifier] = 0
	}

	for _, buff := range returnedData {
		parsedValue, parsedToken := ProcessCalledDataParams(buff)
		assert.Equal(testSetup, value, parsedValue)
		mapUniversalTokens[parsedToken]++
	}

	assert.Equal(testSetup, len(tokens), len(mapUniversalTokens))
	for _, numTokens := range mapUniversalTokens {
		assert.Equal(testSetup, 1, numTokens)
	}
}

// TestRelayersShouldNotExecuteTransfers creates a new test setup and a running process
func TestRelayersShouldNotExecuteTransfers(
	tb testing.TB,
	expectedStringInLogs string,
	tokens ...framework.TestTokenParams,
) {
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

	_ = TestRelayersWithChainSimulator(tb, setupFunc, processFunc, stopChan)
}

// TestEthContractsShouldError creates a new test setup and a running process
func TestEthContractsShouldError(tb testing.TB, testToken framework.TestTokenParams) {
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

	_ = TestRelayersWithChainSimulator(tb,
		setupFunc,
		processFunc,
		make(chan error),
	)
}

// TestRelayersWithChainSimulatorAndTokensAndRefund creates a new test setup and a running process
func TestRelayersWithChainSimulatorAndTokensAndRefund(tb testing.TB, manualStopChan chan error, tokens ...framework.TestTokenParams) {
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

	_ = TestRelayersWithChainSimulator(tb,
		setupFunc,
		processFunc,
		manualStopChan,
	)
}

// ProcessCalledDataParams gets the token from the provided buffer
func ProcessCalledDataParams(buff []byte) (uint64, string) {
	valBuff := buff[:8]
	value := binary.BigEndian.Uint64(valBuff)

	buff = buff[8+32:] // trim the nonce and the address
	tokenLenBuff := buff[:4]
	tokenLen := binary.BigEndian.Uint32(tokenLenBuff)
	buff = buff[4:] // trim the length of the token string

	token := string(buff[:tokenLen])

	return value, token
}
