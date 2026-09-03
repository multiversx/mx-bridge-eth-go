//go:build slow

package ESDTtransferIssues

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

const testMarker = "==================================== %s ===================================="
const performActionFunction = "performAction"
const numAcceptedFailedPerformActionCalls = 10

var log = logger.GetOrCreate("transferIssuesTestsLog")

func TestComplexScenarioWithMissingTransferRole(t *testing.T) {
	okToken := GenerateOKToken()
	notOkTransferRoleToken := GenerateNotOKTransferRoleToken()
	notOkFrozenSafeToken := GenerateNotOKFrozenSafeToken()
	notOkFrozenMultiTransferToken := GenerateNotOKFrozenMultiTransferToken()
	failedTransactionNotifier := framework.NewFailedTransactionsNotifier(
		performActionFunction,
		numAcceptedFailedPerformActionCalls,
	)

	testFlowEthToMvx := &slowTests.TestFlow{
		TB:       t,
		FlowType: slowTests.StartFromEthereumFlow,
		Tokens: []framework.TestTokenParams{
			okToken,
			notOkTransferRoleToken,
			notOkFrozenSafeToken,
			notOkFrozenMultiTransferToken,
		},
		MessageAfterFirstHalfBridge:  "Ethereum->MultiversX transfer finished, now sending back to Ethereum...",
		MessageAfterSecondHalfBridge: "MultiversX<->Ethereum from Ethereum transfers done",
	}

	testFlowEthToMvx.HandlerAfterFirstHalfBridge = func(flow *slowTests.TestFlow) {
		flow.Setup.SendFromMultiversxToEthereum(flow.Setup.BobKeys, flow.Setup.CharlieKeys, flow.Tokens...)
	}
	testFlowEthToMvx.HandlerToStartFirstBridge = func(flow *slowTests.TestFlow) {
		if len(flow.Tokens) == 0 {
			return
		}

		flow.Setup.CreateBatchOnEthereum(flow.Setup.MultiversxHandler.CalleeScAddress, testFlowEthToMvx.Tokens...)
	}

	testStep := 0
	processFunc := func(tb testing.TB, setup *framework.TestSetup) bool {
		if testStep == 0 {
			setNextCompletionChecker(testFlowEthToMvx, testStep, failedTransactionNotifier)
			testStep++
		}

		isFinished := testFlowEthToMvx.Process()
		if isFinished {
			completeFinish := setNextCompletionChecker(testFlowEthToMvx, testStep, failedTransactionNotifier)
			testStep++
			if completeFinish {
				return true
			}
		}

		// commit blocks in order to execute incoming txs from relayers
		setup.EthereumHandler.SimulatedChain.Commit()
		setup.ChainSimulator.GenerateBlocks(setup.Ctx, 1)
		return false
	}

	_ = executeComplexScenario(t, testFlowEthToMvx, processFunc, failedTransactionNotifier)
}

func setNextCompletionChecker(
	testFlowEthToMvx *slowTests.TestFlow,
	testStep int,
	failedTransactionNotifier FailedTransactionNotifier,
) bool {
	testFlowEthToMvx.FirstHalfBridgeDone = false
	testFlowEthToMvx.SecondHalfBridgeDone = false

	switch testStep {
	case 0:
		step0PrepareInitialStep(testFlowEthToMvx)
		return false
	case 1:
		step1PrepareWrongTokensSetup(testFlowEthToMvx, failedTransactionNotifier)
		return false
	case 2:
		step2ExecuteRefundForBlacklistedToken(testFlowEthToMvx, failedTransactionNotifier)
		return false
	case 3:
		step3RemoveBlacklistToken(testFlowEthToMvx)
		return false
	default:
		return true
	}
}

func step0PrepareInitialStep(testFlowEthToMvx *slowTests.TestFlow) {
	log.Info(fmt.Sprintf(testMarker, "Starting step 0 - swaps that work"))

	testFlowEthToMvx.HandlerToStartFirstBridge(testFlowEthToMvx)
	// balance tests are the default ones (the ones defined when setting the token)
}

func step1PrepareWrongTokensSetup(
	testFlowEthToMvx *slowTests.TestFlow,
	failedTransactionNotifier FailedTransactionNotifier,
) {
	log.Info(fmt.Sprintf(testMarker, "Starting step 1 - swaps that do not work (bridge lock) because some tokens have wrong setup (frozen, no transfer role, etc.)"))

	testFlowEthToMvx.Setup.SendFromEthereumToMultiversX(
		testFlowEthToMvx.Setup.AliceKeys,
		testFlowEthToMvx.Setup.BobKeys,
		testFlowEthToMvx.Setup.MultiversxHandler.CalleeScAddress,
		testFlowEthToMvx.Tokens...,
	)

	badTransferRoleToken := testFlowEthToMvx.Tokens[1]
	// we are setting the transfer role only for token[1] so all transfers of that token will fail
	testFlowEthToMvx.Setup.MultiversxHandler.SetTransferRolesForToken(
		testFlowEthToMvx.Setup.Ctx,
		badTransferRoleToken.IssueTokenParams,
		testFlowEthToMvx.Setup.MultiversxHandler.CalleeScAddress, // a random address, just to activate the transfer role feat
	)

	//we are freezing Safe contract for token[2]
	badFrozenSafeToken := testFlowEthToMvx.Tokens[2]
	testFlowEthToMvx.Setup.MultiversxHandler.FreezeTokenForAddresses(
		testFlowEthToMvx.Setup.Ctx,
		badFrozenSafeToken.IssueTokenParams,
		testFlowEthToMvx.Setup.MultiversxHandler.SafeAddress,
	)

	//we are freezing Multi-transfer contract for token[3]
	badFrozenMultiTransferToken := testFlowEthToMvx.Tokens[3]
	testFlowEthToMvx.Setup.MultiversxHandler.FreezeTokenForAddresses(
		testFlowEthToMvx.Setup.Ctx,
		badFrozenMultiTransferToken.IssueTokenParams,
		testFlowEthToMvx.Setup.MultiversxHandler.MultiTransferAddress,
	)

	failedTransactionNotifier.ClearNotifierHandlers()
	failedTransactionNotifier.ClearInternalStateData()

	failedTransactionNotifier.RegisterHandler(func(txData string, numCalls int) {
		log.Warn("Bridge locked, performAction calls failures detected",
			"perform action call", txData,
			"num calls", numCalls)

		// blacklist the token with missing transfer role
		testFlowEthToMvx.Setup.MultiversxHandler.BlacklistToken(
			testFlowEthToMvx.Setup.Ctx,
			badTransferRoleToken.IssueTokenParams,
		)

		// blacklist the token with frozen safe contract
		testFlowEthToMvx.Setup.MultiversxHandler.BlacklistToken(
			testFlowEthToMvx.Setup.Ctx,
			badFrozenSafeToken.IssueTokenParams,
		)

		// blacklist the token with frozen multi transfer contract
		testFlowEthToMvx.Setup.MultiversxHandler.BlacklistToken(
			testFlowEthToMvx.Setup.Ctx,
			badFrozenMultiTransferToken.IssueTokenParams,
		)
	})

	{
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.Alice].OnEth = big.NewInt(-5000 - 5000)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.Bob].OnMvx = big.NewInt(5000 - 2500 + 5000)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.Charlie].OnEth = big.NewInt(2500 - 50)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnEth = big.NewInt(5000 - 2450 + 5000)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(5000 - 2500 + 5000)

		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.Alice].OnEth = big.NewInt(-5000 - 5000)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnMvx = big.NewInt(5000 - 2500 + 5000 - 2500)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.Charlie].OnEth = big.NewInt(2500 - 50 + 2500 - 50)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnEth = big.NewInt(5000 - 2450 + 5000 - 2450)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50 + 50)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(5000 - 2500 + 5000 - 2500)

		testFlowEthToMvx.Tokens[0].MintBurnChecks.MvxTotalUniversalMint = big.NewInt(5000 + 5000)
		testFlowEthToMvx.Tokens[0].MintBurnChecks.MvxTotalChainSpecificMint = big.NewInt(5000 + 5000)
		testFlowEthToMvx.Tokens[0].MintBurnChecks.MvxTotalUniversalBurn = big.NewInt(2500 + 2500)
		testFlowEthToMvx.Tokens[0].MintBurnChecks.MvxTotalChainSpecificBurn = big.NewInt(2500 - 50 + 2500 - 50)
		testFlowEthToMvx.Tokens[0].MintBurnChecks.MvxSafeMintValue = big.NewInt(5000 + 5000)
		testFlowEthToMvx.Tokens[0].MintBurnChecks.MvxSafeBurnValue = big.NewInt(2500 - 50 + 2500 - 50)

		testFlowEthToMvx.Tokens[0].SpecialChecks.WrapperDeltaLiquidityCheck = big.NewInt(5000 - 2500 + 5000 - 2500)
	}
	{
		testFlowEthToMvx.Tokens[1].TestOperations[0].ValueToSendFromMvX = nil //do not attempt sending to Ethereum, as the transaction will fail

		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.Alice].OnEth = big.NewInt(-4000 - 4000)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.Bob].OnMvx = big.NewInt(4000 - 1500) // Bob does not receive tokens
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.Charlie].OnEth = big.NewInt(1500 - 50)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnEth = big.NewInt(4000 - 1450 + 4000)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(4000 - 1500)

		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.Alice].OnEth = big.NewInt(-4000 - 4000)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnMvx = big.NewInt(4000 - 1500)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.Charlie].OnEth = big.NewInt(1500 - 50)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnEth = big.NewInt(4000 - 1450 + 4000)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(4000 - 1500)

		testFlowEthToMvx.Tokens[1].MintBurnChecks.MvxTotalUniversalMint = big.NewInt(4000)
		testFlowEthToMvx.Tokens[1].MintBurnChecks.MvxTotalChainSpecificMint = big.NewInt(4000)
		testFlowEthToMvx.Tokens[1].MintBurnChecks.MvxTotalUniversalBurn = big.NewInt(1500)
		testFlowEthToMvx.Tokens[1].MintBurnChecks.MvxTotalChainSpecificBurn = big.NewInt(1500 - 50)
		testFlowEthToMvx.Tokens[1].MintBurnChecks.MvxSafeMintValue = big.NewInt(4000)
		testFlowEthToMvx.Tokens[1].MintBurnChecks.MvxSafeBurnValue = big.NewInt(1500 - 50)

		testFlowEthToMvx.Tokens[1].SpecialChecks.WrapperDeltaLiquidityCheck = big.NewInt(4000 - 1500)
	}
	{
		testFlowEthToMvx.Tokens[2].TestOperations[0].ValueToSendFromMvX = nil //do not attempt sending to Ethereum, as the transaction will fail

		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.FirstHalfBridge][framework.Alice].OnEth = big.NewInt(-6000 - 6000)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.FirstHalfBridge][framework.Bob].OnMvx = big.NewInt(6000 - 4500) // Bob does not receive tokens
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.FirstHalfBridge][framework.Charlie].OnEth = big.NewInt(4500 - 50)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnEth = big.NewInt(6000 - 4450 + 6000)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50)

		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.SecondHalfBridge][framework.Alice].OnEth = big.NewInt(-6000 - 6000)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnMvx = big.NewInt(6000 - 4500)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.SecondHalfBridge][framework.Charlie].OnEth = big.NewInt(4500 - 50)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnEth = big.NewInt(6000 - 4450 + 6000)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50)

		testFlowEthToMvx.Tokens[2].MintBurnChecks.MvxTotalUniversalMint = big.NewInt(6000)
		testFlowEthToMvx.Tokens[2].MintBurnChecks.MvxTotalChainSpecificMint = big.NewInt(6000)
		testFlowEthToMvx.Tokens[2].MintBurnChecks.MvxTotalUniversalBurn = big.NewInt(4500 - 50)
		testFlowEthToMvx.Tokens[2].MintBurnChecks.MvxTotalChainSpecificBurn = big.NewInt(4500 - 50)
		testFlowEthToMvx.Tokens[2].MintBurnChecks.MvxSafeMintValue = big.NewInt(6000)
		testFlowEthToMvx.Tokens[2].MintBurnChecks.MvxSafeBurnValue = big.NewInt(4500 - 50)
	}
	{
		testFlowEthToMvx.Tokens[3].TestOperations[0].ValueToSendFromMvX = nil //do not attempt sending to Ethereum, as the transaction will fail

		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.FirstHalfBridge][framework.Alice].OnEth = big.NewInt(-7000 - 7000)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.FirstHalfBridge][framework.Bob].OnMvx = big.NewInt(7000 - 5500) // Bob does not receive tokens
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.FirstHalfBridge][framework.Charlie].OnEth = big.NewInt(5500 - 50)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnEth = big.NewInt(7000 - 5450 + 7000)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.FirstHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(7000 - 5500)

		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.SecondHalfBridge][framework.Alice].OnEth = big.NewInt(-7000 - 7000)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnMvx = big.NewInt(7000 - 5500)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.SecondHalfBridge][framework.Charlie].OnEth = big.NewInt(5500 - 50)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnEth = big.NewInt(7000 - 5450 + 7000)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.SecondHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(7000 - 5500)

		testFlowEthToMvx.Tokens[3].MintBurnChecks.MvxTotalUniversalMint = big.NewInt(7000)
		testFlowEthToMvx.Tokens[3].MintBurnChecks.MvxTotalChainSpecificMint = big.NewInt(7000)
		testFlowEthToMvx.Tokens[3].MintBurnChecks.MvxTotalUniversalBurn = big.NewInt(5500)
		testFlowEthToMvx.Tokens[3].MintBurnChecks.MvxTotalChainSpecificBurn = big.NewInt(5500 - 50)
		testFlowEthToMvx.Tokens[3].MintBurnChecks.MvxSafeMintValue = big.NewInt(7000)
		testFlowEthToMvx.Tokens[3].MintBurnChecks.MvxSafeBurnValue = big.NewInt(5500 - 50)

		testFlowEthToMvx.Tokens[3].SpecialChecks.WrapperDeltaLiquidityCheck = big.NewInt(7000 - 5500)
	}
}

func step2ExecuteRefundForBlacklistedToken(
	testFlowEthToMvx *slowTests.TestFlow,
	failedTransactionNotifier FailedTransactionNotifier,
) {
	log.Info(fmt.Sprintf(testMarker, "Starting step 2 - execute the refunds for the blacklisted tokens"))

	failedTransactionNotifier.ClearNotifierHandlers()
	failedTransactionNotifier.ClearInternalStateData()

	buff := testFlowEthToMvx.Setup.MultiversxHandler.GetTransactionForBlacklistTokens(
		testFlowEthToMvx.Setup.Ctx,
	)
	// 6 elements should be returned:
	//     the tuple (tx_id, eth_tx) for transfer role token
	//     the tuple (tx_id, eth_tx) for frozen safe contract
	//     the tuple (tx_id, eth_tx) for frozen multi transfer contract
	require.Equal(testFlowEthToMvx.TB, 6, len(buff))

	for i := 0; i < len(buff); i += 2 {
		txID := big.NewInt(0).SetBytes(buff[i]).Uint64()
		log.Info("Found a refundable transaction due to a blacklisted token", "tx ID", txID)

		txStatus := testFlowEthToMvx.Setup.MultiversxHandler.RefundTransactionForBlacklistTokens(
			testFlowEthToMvx.Setup.Ctx,
			txID,
		)
		require.Equal(testFlowEthToMvx.TB, transaction.TxStatusSuccess, txStatus)

		// making a call again will make the transaction fail
		txStatus = testFlowEthToMvx.Setup.MultiversxHandler.RefundTransactionForBlacklistTokens(
			testFlowEthToMvx.Setup.Ctx,
			txID,
		)
		require.Equal(testFlowEthToMvx.TB, transaction.TxStatusFail, txStatus)
	}

	{
		// no transfers
		testFlowEthToMvx.Tokens[0].TestOperations[0].ValueToSendFromMvX = nil
		testFlowEthToMvx.Tokens[0].TestOperations[0].ValueToTransferToMvx = nil

		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.Alice].OnEth = big.NewInt(-5000 - 5000)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.Bob].OnMvx = big.NewInt(5000 - 2500 + 5000 - 2500)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.Charlie].OnEth = big.NewInt(2500 - 50 + 2500 - 50)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnEth = big.NewInt(5000 - 2450 + 5000 - 2450)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50 + 50)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(5000 - 2500 + 5000 - 2500)
	}
	{
		// no transfers
		testFlowEthToMvx.Tokens[1].TestOperations[0].ValueToSendFromMvX = nil
		testFlowEthToMvx.Tokens[1].TestOperations[0].ValueToTransferToMvx = nil

		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.Alice].OnEth = big.NewInt(-4000 - 4000)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.Bob].OnMvx = big.NewInt(4000 - 1500)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.Charlie].OnEth = big.NewInt(1500 - 50)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnEth = big.NewInt(4000 - 1450 + 4000)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(4000 - 1500)

		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.Alice].OnEth = big.NewInt(-4000 - 4000 + 4000) // refunded without fees
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnMvx = big.NewInt(4000 - 1500)           // Bob does not receive tokens
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.Charlie].OnEth = big.NewInt(1500 - 50)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnEth = big.NewInt(4000 - 1450 + 4000 - 4000)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(4000 - 1500)

		testFlowEthToMvx.Tokens[1].MintBurnChecks.MvxTotalUniversalMint = big.NewInt(4000)
		testFlowEthToMvx.Tokens[1].MintBurnChecks.MvxTotalChainSpecificMint = big.NewInt(4000)
		testFlowEthToMvx.Tokens[1].MintBurnChecks.MvxTotalUniversalBurn = big.NewInt(1500)
		testFlowEthToMvx.Tokens[1].MintBurnChecks.MvxTotalChainSpecificBurn = big.NewInt(1500 - 50)
		testFlowEthToMvx.Tokens[1].MintBurnChecks.MvxSafeMintValue = big.NewInt(4000)
		testFlowEthToMvx.Tokens[1].MintBurnChecks.MvxSafeBurnValue = big.NewInt(1500 - 50)

		testFlowEthToMvx.Tokens[1].SpecialChecks.WrapperDeltaLiquidityCheck = big.NewInt(4000 - 1500)
	}
	{
		// no transfers
		testFlowEthToMvx.Tokens[2].TestOperations[0].ValueToSendFromMvX = nil
		testFlowEthToMvx.Tokens[2].TestOperations[0].ValueToTransferToMvx = nil

		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.FirstHalfBridge][framework.Alice].OnEth = big.NewInt(-6000 - 6000)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.FirstHalfBridge][framework.Bob].OnMvx = big.NewInt(6000 - 4500)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.FirstHalfBridge][framework.Charlie].OnEth = big.NewInt(4500 - 50)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnEth = big.NewInt(6000 - 4450 + 6000)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50)

		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.SecondHalfBridge][framework.Alice].OnEth = big.NewInt(-6000 - 6000 + 6000) // refunded without fees
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnMvx = big.NewInt(6000 - 4500)           // Bob does not receive tokens
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.SecondHalfBridge][framework.Charlie].OnEth = big.NewInt(4500 - 50)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnEth = big.NewInt(6000 - 4450)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50)

		testFlowEthToMvx.Tokens[2].MintBurnChecks.MvxTotalUniversalMint = big.NewInt(6000)
		testFlowEthToMvx.Tokens[2].MintBurnChecks.MvxTotalChainSpecificMint = big.NewInt(6000)
		testFlowEthToMvx.Tokens[2].MintBurnChecks.MvxTotalUniversalBurn = big.NewInt(4500 - 50)
		testFlowEthToMvx.Tokens[2].MintBurnChecks.MvxTotalChainSpecificBurn = big.NewInt(4500 - 50)
		testFlowEthToMvx.Tokens[2].MintBurnChecks.MvxSafeMintValue = big.NewInt(6000)
		testFlowEthToMvx.Tokens[2].MintBurnChecks.MvxSafeBurnValue = big.NewInt(4500 - 50)
	}
	{
		// no transfers
		testFlowEthToMvx.Tokens[3].TestOperations[0].ValueToSendFromMvX = nil
		testFlowEthToMvx.Tokens[3].TestOperations[0].ValueToTransferToMvx = nil

		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.FirstHalfBridge][framework.Alice].OnEth = big.NewInt(-7000 - 7000)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.FirstHalfBridge][framework.Bob].OnMvx = big.NewInt(7000 - 5500)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.FirstHalfBridge][framework.Charlie].OnEth = big.NewInt(5500 - 50)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnEth = big.NewInt(7000 - 5450 + 7000)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.FirstHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(7000 - 5500)

		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.SecondHalfBridge][framework.Alice].OnEth = big.NewInt(-7000 - 7000 + 7000) // refunded without fees
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnMvx = big.NewInt(7000 - 5500)           // Bob does not receive tokens
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.SecondHalfBridge][framework.Charlie].OnEth = big.NewInt(5500 - 50)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnEth = big.NewInt(7000 - 5450 + 7000 - 7000)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.SecondHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(7000 - 5500)

		testFlowEthToMvx.Tokens[3].MintBurnChecks.MvxTotalUniversalMint = big.NewInt(7000)
		testFlowEthToMvx.Tokens[3].MintBurnChecks.MvxTotalChainSpecificMint = big.NewInt(7000)
		testFlowEthToMvx.Tokens[3].MintBurnChecks.MvxTotalUniversalBurn = big.NewInt(5500)
		testFlowEthToMvx.Tokens[3].MintBurnChecks.MvxTotalChainSpecificBurn = big.NewInt(5500 - 50)
		testFlowEthToMvx.Tokens[3].MintBurnChecks.MvxSafeMintValue = big.NewInt(7000)
		testFlowEthToMvx.Tokens[3].MintBurnChecks.MvxSafeBurnValue = big.NewInt(5500 - 50)

		testFlowEthToMvx.Tokens[3].SpecialChecks.WrapperDeltaLiquidityCheck = big.NewInt(7000 - 5500)
	}
}

func step3RemoveBlacklistToken(testFlowEthToMvx *slowTests.TestFlow) {
	log.Info(fmt.Sprintf(testMarker, "Starting step 3 - remove blacklisted tokens, grant transfer role, do un-freez, etc & redo the swaps"))

	// grant transfer role
	testFlowEthToMvx.Setup.MultiversxHandler.SetTransferRolesForToken(
		testFlowEthToMvx.Setup.Ctx,
		testFlowEthToMvx.Tokens[1].IssueTokenParams,
		testFlowEthToMvx.Setup.MultiversxHandler.WrapperAddress,
		testFlowEthToMvx.Setup.MultiversxHandler.SafeAddress,
		testFlowEthToMvx.Setup.MultiversxHandler.MultiTransferAddress,
		testFlowEthToMvx.Setup.MultiversxHandler.ScProxyAddress,
	)

	//unfreeze the Safe contract
	testFlowEthToMvx.Setup.MultiversxHandler.UnFreezeTokenForAddresses(
		testFlowEthToMvx.Setup.Ctx,
		testFlowEthToMvx.Tokens[2].IssueTokenParams,
		testFlowEthToMvx.Setup.MultiversxHandler.SafeAddress,
	)

	//unfreeze the Multi-transfer contract
	testFlowEthToMvx.Setup.MultiversxHandler.UnFreezeTokenForAddresses(
		testFlowEthToMvx.Setup.Ctx,
		testFlowEthToMvx.Tokens[3].IssueTokenParams,
		testFlowEthToMvx.Setup.MultiversxHandler.MultiTransferAddress,
	)

	// remove blacklisted tokens
	testFlowEthToMvx.Setup.MultiversxHandler.RemoveBlacklistedToken(
		testFlowEthToMvx.Setup.Ctx,
		testFlowEthToMvx.Tokens[1].IssueTokenParams,
	)
	testFlowEthToMvx.Setup.MultiversxHandler.RemoveBlacklistedToken(
		testFlowEthToMvx.Setup.Ctx,
		testFlowEthToMvx.Tokens[2].IssueTokenParams,
	)
	testFlowEthToMvx.Setup.MultiversxHandler.RemoveBlacklistedToken(
		testFlowEthToMvx.Setup.Ctx,
		testFlowEthToMvx.Tokens[3].IssueTokenParams,
	)

	{
		testFlowEthToMvx.Tokens[0].TestOperations[0].ValueToSendFromMvX = big.NewInt(1400)
		testFlowEthToMvx.Tokens[0].TestOperations[0].ValueToTransferToMvx = big.NewInt(2800)

		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.Alice].OnEth = big.NewInt(-5000 - 5000 - 2800)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.Bob].OnMvx = big.NewInt(5000 - 2500 + 5000 - 2500 + 2800)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.Charlie].OnEth = big.NewInt(2500 - 50 + 2500 - 50)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnEth = big.NewInt(5000 - 2450 + 5000 - 2450 + 2800)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50 + 50)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.FirstHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(5000 - 2500 + 5000 - 2500 + 2800)

		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.Alice].OnEth = big.NewInt(-5000 - 5000 - 2800)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnMvx = big.NewInt(5000 - 2500 + 5000 - 2500 + 2800 - 1400)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.Charlie].OnEth = big.NewInt(2500 - 50 + 2500 - 50 + 1400 - 50)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnEth = big.NewInt(5000 - 2450 + 5000 - 2450 + 2800 - 1350)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50 + 50 + 50)
		testFlowEthToMvx.Tokens[0].DeltaBalances[framework.SecondHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(5000 - 2500 + 5000 - 2500 + 2800 - 1400)

		testFlowEthToMvx.Tokens[0].MintBurnChecks.MvxTotalUniversalMint = big.NewInt(5000 + 5000 + 2800)
		testFlowEthToMvx.Tokens[0].MintBurnChecks.MvxTotalChainSpecificMint = big.NewInt(5000 + 5000 + 2800)
		testFlowEthToMvx.Tokens[0].MintBurnChecks.MvxTotalUniversalBurn = big.NewInt(2500 + 2500 + 1400)
		testFlowEthToMvx.Tokens[0].MintBurnChecks.MvxTotalChainSpecificBurn = big.NewInt(2500 - 50 + 2500 - 50 + 1400 - 50)
		testFlowEthToMvx.Tokens[0].MintBurnChecks.MvxSafeMintValue = big.NewInt(5000 + 5000 + 2800)
		testFlowEthToMvx.Tokens[0].MintBurnChecks.MvxSafeBurnValue = big.NewInt(2500 - 50 + 2500 - 50 + 1400 - 50)

		testFlowEthToMvx.Tokens[0].SpecialChecks.WrapperDeltaLiquidityCheck = big.NewInt(5000 - 2500 + 5000 - 2500 + 2800 - 1400)
	}
	{
		testFlowEthToMvx.Tokens[1].TestOperations[0].ValueToSendFromMvX = big.NewInt(1900)
		testFlowEthToMvx.Tokens[1].TestOperations[0].ValueToTransferToMvx = big.NewInt(3600)

		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.Alice].OnEth = big.NewInt(-4000 - 4000 + 4000 - 3600)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.Bob].OnMvx = big.NewInt(4000 - 1500 + 3600)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.Charlie].OnEth = big.NewInt(1500 - 50)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnEth = big.NewInt(4000 - 1450 + 4000 - 4000 + 3600)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.FirstHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(4000 - 1500 + 3600)

		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.Alice].OnEth = big.NewInt(-4000 - 4000 + 4000 - 3600)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnMvx = big.NewInt(4000 - 1500 + 3600 - 1900)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.Charlie].OnEth = big.NewInt(1500 - 50 + 1900 - 50)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnEth = big.NewInt(4000 - 1450 + 4000 - 4000 + 3600 - 1850)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50 + 50)
		testFlowEthToMvx.Tokens[1].DeltaBalances[framework.SecondHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(4000 - 1500 + 3600 - 1900)

		testFlowEthToMvx.Tokens[1].MintBurnChecks.MvxTotalUniversalMint = big.NewInt(4000 + 3600)
		testFlowEthToMvx.Tokens[1].MintBurnChecks.MvxTotalChainSpecificMint = big.NewInt(4000 + 3600)
		testFlowEthToMvx.Tokens[1].MintBurnChecks.MvxTotalUniversalBurn = big.NewInt(1500 + 1900)
		testFlowEthToMvx.Tokens[1].MintBurnChecks.MvxTotalChainSpecificBurn = big.NewInt(1500 - 50 + 1900 - 50)
		testFlowEthToMvx.Tokens[1].MintBurnChecks.MvxSafeMintValue = big.NewInt(4000 + 3600)
		testFlowEthToMvx.Tokens[1].MintBurnChecks.MvxSafeBurnValue = big.NewInt(1500 - 50 + 1900 - 50)

		testFlowEthToMvx.Tokens[1].SpecialChecks.WrapperDeltaLiquidityCheck = big.NewInt(4000 - 1500 + 3600 - 1900)
	}
	{
		testFlowEthToMvx.Tokens[2].TestOperations[0].ValueToSendFromMvX = big.NewInt(2100)
		testFlowEthToMvx.Tokens[2].TestOperations[0].ValueToTransferToMvx = big.NewInt(3800)

		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.FirstHalfBridge][framework.Alice].OnEth = big.NewInt(-6000 - 6000 + 6000 - 3800)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.FirstHalfBridge][framework.Bob].OnMvx = big.NewInt(6000 - 4500 + 3800)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.FirstHalfBridge][framework.Charlie].OnEth = big.NewInt(4500 - 50)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnEth = big.NewInt(6000 - 4450 + 6000 - 6000 + 3800)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50)

		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.SecondHalfBridge][framework.Alice].OnEth = big.NewInt(-6000 - 6000 + 6000 - 3800)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnMvx = big.NewInt(6000 - 4500 + 3800 - 2100)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.SecondHalfBridge][framework.Charlie].OnEth = big.NewInt(4500 - 50 + 2100 - 50)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnEth = big.NewInt(6000 - 4450 + 6000 - 6000 + 3800 - 2050)
		testFlowEthToMvx.Tokens[2].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50 + 50)

		testFlowEthToMvx.Tokens[2].MintBurnChecks.MvxTotalUniversalMint = big.NewInt(6000 + 3800)
		testFlowEthToMvx.Tokens[2].MintBurnChecks.MvxTotalChainSpecificMint = big.NewInt(6000 + 3800)
		testFlowEthToMvx.Tokens[2].MintBurnChecks.MvxTotalUniversalBurn = big.NewInt(4500 - 50 + 2100 - 50)
		testFlowEthToMvx.Tokens[2].MintBurnChecks.MvxTotalChainSpecificBurn = big.NewInt(4500 - 50 + 2100 - 50)
		testFlowEthToMvx.Tokens[2].MintBurnChecks.MvxSafeMintValue = big.NewInt(6000 + 3800)
		testFlowEthToMvx.Tokens[2].MintBurnChecks.MvxSafeBurnValue = big.NewInt(4500 - 50 + 2100 - 50)
	}
	{
		testFlowEthToMvx.Tokens[3].TestOperations[0].ValueToSendFromMvX = big.NewInt(3100)
		testFlowEthToMvx.Tokens[3].TestOperations[0].ValueToTransferToMvx = big.NewInt(8800)

		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.FirstHalfBridge][framework.Alice].OnEth = big.NewInt(-7000 - 7000 + 7000 - 8800)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.FirstHalfBridge][framework.Bob].OnMvx = big.NewInt(7000 - 5500 + 8800)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.FirstHalfBridge][framework.Charlie].OnEth = big.NewInt(5500 - 50)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnEth = big.NewInt(7000 - 5450 + 7000 - 7000 + 8800)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.FirstHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.FirstHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(7000 - 5500 + 8800)

		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.SecondHalfBridge][framework.Alice].OnEth = big.NewInt(-7000 - 7000 + 7000 - 8800)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.SecondHalfBridge][framework.Bob].OnMvx = big.NewInt(7000 - 5500 + 8800 - 3100)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.SecondHalfBridge][framework.Charlie].OnEth = big.NewInt(5500 - 50 + 3100 - 50)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnEth = big.NewInt(7000 - 5450 + 7000 - 7000 + 8800 - 3050)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.SecondHalfBridge][framework.SafeSC].OnMvx = big.NewInt(50 + 50)
		testFlowEthToMvx.Tokens[3].DeltaBalances[framework.SecondHalfBridge][framework.WrapperSC].OnMvx = big.NewInt(7000 - 5500 + 8800 - 3100)

		testFlowEthToMvx.Tokens[3].MintBurnChecks.MvxTotalUniversalMint = big.NewInt(7000 + 8800)
		testFlowEthToMvx.Tokens[3].MintBurnChecks.MvxTotalChainSpecificMint = big.NewInt(7000 + 8800)
		testFlowEthToMvx.Tokens[3].MintBurnChecks.MvxTotalUniversalBurn = big.NewInt(5500 + 3100)
		testFlowEthToMvx.Tokens[3].MintBurnChecks.MvxTotalChainSpecificBurn = big.NewInt(5500 - 50 + 3100 - 50)
		testFlowEthToMvx.Tokens[3].MintBurnChecks.MvxSafeMintValue = big.NewInt(7000 + 8800)
		testFlowEthToMvx.Tokens[3].MintBurnChecks.MvxSafeBurnValue = big.NewInt(5500 - 50 + 3100 - 50)

		testFlowEthToMvx.Tokens[3].SpecialChecks.WrapperDeltaLiquidityCheck = big.NewInt(7000 - 5500 + 8800 - 3100)
	}

	testFlowEthToMvx.Setup.SendFromEthereumToMultiversX(
		testFlowEthToMvx.Setup.AliceKeys,
		testFlowEthToMvx.Setup.BobKeys,
		testFlowEthToMvx.Setup.MultiversxHandler.CalleeScAddress,
		testFlowEthToMvx.Tokens...,
	)
}

func executeComplexScenario(
	tb testing.TB,
	testFlow *slowTests.TestFlow,
	processFunc func(tb testing.TB, setup *framework.TestSetup) bool,
	failedTransactionNotifier FailedTransactionNotifier,
) *framework.TestSetup {
	setupFunc := func(tb testing.TB, setup *framework.TestSetup) {
		setup.ProxyWrapperInstance.RegisterBeforeTransactionSendHandler(failedTransactionNotifier.BeforeSendingTransaction)

		testFlow.Setup = setup

		setup.IssueAndConfigureTokens(testFlow.Tokens...)
		setup.MultiversxHandler.CheckForZeroBalanceOnReceivers(setup.Ctx, testFlow.Tokens...)
	}

	return slowTests.NewTestEnvironmentWithChainSimulator(
		tb,
		setupFunc,
		processFunc,
		make(chan error),
		framework.ContractsVersion3p1,
	)
}

// GenerateOKToken will generate a test OK token
func GenerateOKToken() framework.TestTokenParams {
	// OK is ethNative = true, ethMintBurn = false, mvxNative = false, mvxMintBurn = true
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "TKOK",
			NumOfDecimalsUniversal:           6,
			NumOfDecimalsChainSpecific:       6,
			MvxUniversalTokenTicker:          "TKOK",
			MvxChainSpecificTokenTicker:      "ETHTKOK",
			MvxUniversalTokenDisplayName:     "WrappedTKOK",
			MvxChainSpecificTokenDisplayName: "EthereumWrappedTKOK",
			MvxToEthFee:                      big.NewInt(50),
			ValueToMintOnMvx:                 "10000000000",
			IsMintBurnOnMvX:                  true,
			IsNativeOnMvX:                    false,
			HasChainSpecificToken:            true,
			EthTokenName:                     "EthTKOK",
			EthTokenSymbol:                   "TKOK",
			ValueToMintOnEth:                 "10000000000",
			IsMintBurnOnEth:                  false,
			IsNativeOnEth:                    true,
			MultipleSpendings:                big.NewInt(100), // ensure enough tokens to Alice
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(5000),
				ValueToSendFromMvX:   big.NewInt(2500),
			},
		},
		DeltaBalances: map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
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
				framework.Charlie: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(5000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(5000),
					MvxToken: framework.ChainSpecificToken,
				},
			},
			framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-5000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(5000 - 2500),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(2500 - 50),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(5000 - 2450),
					OnMvx:    big.NewInt(50),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(5000 - 2500),
					MvxToken: framework.ChainSpecificToken,
				},
			},
		},
		MintBurnChecks: &framework.MintBurnBalances{
			MvxTotalUniversalMint:     big.NewInt(5000),
			MvxTotalChainSpecificMint: big.NewInt(5000),
			MvxTotalUniversalBurn:     big.NewInt(2500),
			MvxTotalChainSpecificBurn: big.NewInt(2500 - 50),
			MvxSafeMintValue:          big.NewInt(5000),
			MvxSafeBurnValue:          big.NewInt(2500 - 50),

			EthSafeMintValue: big.NewInt(0),
			EthSafeBurnValue: big.NewInt(0),
		},
		SpecialChecks: &framework.SpecialBalanceChecks{
			WrapperDeltaLiquidityCheck: big.NewInt(5000 - 2500),
		},
	}
}

// GenerateNotOKTransferRoleToken will generate a test Not-OK token for transfer role tests
func GenerateNotOKTransferRoleToken() framework.TestTokenParams {
	// OK is ethNative = true, ethMintBurn = false, mvxNative = false, mvxMintBurn = true
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "NOKTF",
			NumOfDecimalsUniversal:           6,
			NumOfDecimalsChainSpecific:       6,
			MvxUniversalTokenTicker:          "NOKTF",
			MvxChainSpecificTokenTicker:      "ETHNOKTF",
			MvxUniversalTokenDisplayName:     "WrappedNOKTF",
			MvxChainSpecificTokenDisplayName: "EthereumWrappedNOKTF",
			MvxToEthFee:                      big.NewInt(50),
			ValueToMintOnMvx:                 "10000000000",
			IsMintBurnOnMvX:                  true,
			IsNativeOnMvX:                    false,
			HasChainSpecificToken:            true,
			EthTokenName:                     "EthNOKTF",
			EthTokenSymbol:                   "NOKTF",
			ValueToMintOnEth:                 "10000000000",
			IsMintBurnOnEth:                  false,
			IsNativeOnEth:                    true,
			MultipleSpendings:                big.NewInt(100), // ensure enough tokens to Alice
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(4000),
				ValueToSendFromMvX:   big.NewInt(1500),
			},
		},
		DeltaBalances: map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-4000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(4000),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(4000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(4000),
					MvxToken: framework.ChainSpecificToken,
				},
			},
			framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-4000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(4000 - 1500),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(1500 - 50),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(4000 - 1450),
					OnMvx:    big.NewInt(50),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(4000 - 1500),
					MvxToken: framework.ChainSpecificToken,
				},
			},
		},
		MintBurnChecks: &framework.MintBurnBalances{
			MvxTotalUniversalMint:     big.NewInt(4000),
			MvxTotalChainSpecificMint: big.NewInt(4000),
			MvxTotalUniversalBurn:     big.NewInt(1500),
			MvxTotalChainSpecificBurn: big.NewInt(1500 - 50),
			MvxSafeMintValue:          big.NewInt(4000),
			MvxSafeBurnValue:          big.NewInt(1500 - 50),

			EthSafeMintValue: big.NewInt(0),
			EthSafeBurnValue: big.NewInt(0),
		},
		SpecialChecks: &framework.SpecialBalanceChecks{
			WrapperDeltaLiquidityCheck: big.NewInt(4000 - 1500),
		},
	}
}

// GenerateNotOKFrozenSafeToken will generate a test Not-OK token for Safe frozen tests
func GenerateNotOKFrozenSafeToken() framework.TestTokenParams {
	// OK is ethNative = true, ethMintBurn = false, mvxNative = false, mvxMintBurn = true
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "NOKFS",
			NumOfDecimalsUniversal:           6,
			NumOfDecimalsChainSpecific:       6,
			MvxUniversalTokenTicker:          "NOKFS",
			MvxChainSpecificTokenTicker:      "ETHNOKFS",
			MvxUniversalTokenDisplayName:     "WrappedNOKFS",
			MvxChainSpecificTokenDisplayName: "EthereumWrappedNOKFS",
			MvxToEthFee:                      big.NewInt(50),
			ValueToMintOnMvx:                 "10000000000",
			IsMintBurnOnMvX:                  true,
			IsNativeOnMvX:                    false,
			HasChainSpecificToken:            false,
			EthTokenName:                     "EthNOKFS",
			EthTokenSymbol:                   "NOKFS",
			ValueToMintOnEth:                 "10000000000",
			IsMintBurnOnEth:                  false,
			IsNativeOnEth:                    true,
			MultipleSpendings:                big.NewInt(100), // ensure enough tokens to Alice
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(6000),
				ValueToSendFromMvX:   big.NewInt(4500),
			},
		},
		DeltaBalances: map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-6000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(6000),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(6000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
			},
			framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-6000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(6000 - 4500),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(4500 - 50),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(6000 - 4450),
					OnMvx:    big.NewInt(50),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
			},
		},
		MintBurnChecks: &framework.MintBurnBalances{
			MvxTotalUniversalMint:     big.NewInt(6000),
			MvxTotalChainSpecificMint: big.NewInt(6000),
			MvxTotalUniversalBurn:     big.NewInt(4500 - 50),
			MvxTotalChainSpecificBurn: big.NewInt(4500 - 50),
			MvxSafeMintValue:          big.NewInt(6000),
			MvxSafeBurnValue:          big.NewInt(4500 - 50),

			EthSafeMintValue: big.NewInt(0),
			EthSafeBurnValue: big.NewInt(0),
		},
		SpecialChecks: &framework.SpecialBalanceChecks{
			WrapperDeltaLiquidityCheck: big.NewInt(0),
		},
	}
}

// GenerateNotOKFrozenMultiTransferToken will generate a test Not-OK token for frozen multi-transfer tests
func GenerateNotOKFrozenMultiTransferToken() framework.TestTokenParams {
	// OK is ethNative = true, ethMintBurn = false, mvxNative = false, mvxMintBurn = true
	return framework.TestTokenParams{
		IssueTokenParams: framework.IssueTokenParams{
			AbstractTokenIdentifier:          "NOKFMT",
			NumOfDecimalsUniversal:           6,
			NumOfDecimalsChainSpecific:       6,
			MvxUniversalTokenTicker:          "NOKFMT",
			MvxChainSpecificTokenTicker:      "ENOKFMT",
			MvxUniversalTokenDisplayName:     "WrappedNOKFMT",
			MvxChainSpecificTokenDisplayName: "EthWrappedNOKFMT",
			MvxToEthFee:                      big.NewInt(50),
			ValueToMintOnMvx:                 "10000000000",
			IsMintBurnOnMvX:                  true,
			IsNativeOnMvX:                    false,
			HasChainSpecificToken:            true,
			EthTokenName:                     "EthNOKFMT",
			EthTokenSymbol:                   "NOKFMT",
			ValueToMintOnEth:                 "10000000000",
			IsMintBurnOnEth:                  false,
			IsNativeOnEth:                    true,
			MultipleSpendings:                big.NewInt(100), // ensure enough tokens to Alice
		},
		TestOperations: []framework.TokenOperations{
			{
				ValueToTransferToMvx: big.NewInt(7000),
				ValueToSendFromMvX:   big.NewInt(5500),
			},
		},
		DeltaBalances: map[framework.HalfBridgeIdentifier]framework.DeltaBalancesOnKeys{
			framework.FirstHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-7000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(7000),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(7000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(7000),
					MvxToken: framework.ChainSpecificToken,
				},
			},
			framework.SecondHalfBridge: map[string]*framework.DeltaBalanceHolder{
				framework.Alice: {
					OnEth:    big.NewInt(-7000),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.Bob: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(7000 - 5500),
					MvxToken: framework.UniversalToken,
				},
				framework.Charlie: {
					OnEth:    big.NewInt(5500 - 50),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.SafeSC: {
					OnEth:    big.NewInt(7000 - 5450),
					OnMvx:    big.NewInt(50),
					MvxToken: framework.ChainSpecificToken,
				},
				framework.CalledTestSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(0),
					MvxToken: framework.UniversalToken,
				},
				framework.WrapperSC: {
					OnEth:    big.NewInt(0),
					OnMvx:    big.NewInt(7000 - 5500),
					MvxToken: framework.ChainSpecificToken,
				},
			},
		},
		MintBurnChecks: &framework.MintBurnBalances{
			MvxTotalUniversalMint:     big.NewInt(7000),
			MvxTotalChainSpecificMint: big.NewInt(7000),
			MvxTotalUniversalBurn:     big.NewInt(5500),
			MvxTotalChainSpecificBurn: big.NewInt(5500 - 50),
			MvxSafeMintValue:          big.NewInt(7000),
			MvxSafeBurnValue:          big.NewInt(5500 - 50),

			EthSafeMintValue: big.NewInt(0),
			EthSafeBurnValue: big.NewInt(0),
		},
		SpecialChecks: &framework.SpecialBalanceChecks{
			WrapperDeltaLiquidityCheck: big.NewInt(7000 - 5500),
		},
	}
}
