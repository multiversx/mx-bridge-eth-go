package relayers

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/factory"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/mock"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("integrationTests/relayers")
var zero = big.NewInt(0)

func asyncCancelCall(cancelHandler func(), delay time.Duration) {
	go func() {
		time.Sleep(delay)
		cancelHandler()
	}()
}

func TestRelayersShouldExecuteSimpleTransfersFromMultiversXToEth(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numTransactions := 2
	deposits, tokensAddresses, erc20Map := createTransactions(numTransactions)

	tokens, availableBalances := availableTokensMapToSlices(erc20Map)
	safeContractEthAddress := testsCommon.CreateRandomEthereumAddress()
	erc20ContractsHolder := createMockErc20ContractsHolder(tokens, safeContractEthAddress, availableBalances)

	numRelayers := 3
	ethereumChainMock := mock.NewEthereumChainMock()
	ethereumChainMock.SetQuorum(numRelayers)
	expectedStatuses := []byte{clients.Executed, clients.Rejected}
	ethereumChainMock.GetStatusesAfterExecutionHandler = func() []byte {
		return expectedStatuses
	}
	multiversXChainMock := mock.NewMultiversXChainMock()
	for i := 0; i < len(deposits); i++ {
		ethereumChainMock.UpdateNativeTokens(tokensAddresses[i], false)
		ethereumChainMock.UpdateMintBurnTokens(tokensAddresses[i], true)
		ethereumChainMock.UpdateMintBalances(tokensAddresses[i], zero)
		ethereumChainMock.UpdateBurnBalances(tokensAddresses[i], zero)

		multiversXChainMock.AddTokensPair(tokensAddresses[i], deposits[i].Ticker, true, true, zero, zero, deposits[i].Amount)
	}
	pendingBatch := mock.MultiversXPendingBatch{
		Nonce:              big.NewInt(1),
		MultiversXDeposits: deposits,
	}

	multiversXChainMock.SetPendingBatch(&pendingBatch)
	multiversXChainMock.SetQuorum(numRelayers)

	relayers := make([]bridgeComponents, 0, numRelayers)
	defer closeRelayers(relayers)

	messengers := integrationTests.CreateLinkedMessengers(numRelayers)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1200)
	defer cancel()
	multiversXChainMock.ProcessFinishedHandler = func() {
		log.Info("multiversXChainMock.ProcessFinishedHandler called")
		asyncCancelCall(cancel, time.Second*5)
	}

	for i := 0; i < numRelayers; i++ {
		argsBridgeComponents := createMockBridgeComponentsArgs(i, messengers[i], multiversXChainMock, ethereumChainMock)
		argsBridgeComponents.Configs.GeneralConfig.Eth.SafeContractAddress = safeContractEthAddress.Hex()
		argsBridgeComponents.Erc20ContractsHolder = erc20ContractsHolder
		relayer, err := factory.NewEthMultiversXBridgeComponents(argsBridgeComponents)
		require.Nil(t, err)

		multiversXChainMock.AddRelayer(relayer.MultiversXRelayerAddress())
		ethereumChainMock.AddRelayer(relayer.EthereumRelayerAddress())

		go func() {
			err = relayer.Start()
			integrationTests.Log.LogIfError(err)
			require.Nil(t, err)
		}()

		relayers = append(relayers, relayer)
	}

	<-ctx.Done()

	// let all transactions propagate
	time.Sleep(time.Second * 5)

	checkTestStatus(t, multiversXChainMock, ethereumChainMock, numTransactions, deposits, tokensAddresses)
}

func TestRelayersShouldExecuteTransfersFromMultiversXToEthIfTransactionsAppearInBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("simple tokens transfers", func(t *testing.T) {
		testRelayersShouldExecuteTransfersFromMultiversXToEthIfTransactionsAppearInBatch(t, false)
	})
	t.Run("native tokens transfers", func(t *testing.T) {
		testRelayersShouldExecuteTransfersFromMultiversXToEthIfTransactionsAppearInBatch(t, true)
	})
}

func testRelayersShouldExecuteTransfersFromMultiversXToEthIfTransactionsAppearInBatch(t *testing.T, withNativeTokens bool) {
	numTransactions := 2
	deposits, tokensAddresses, erc20Map := createTransactions(numTransactions)

	safeContractEthAddress := testsCommon.CreateRandomEthereumAddress()
	tokens, availableBalances := availableTokensMapToSlices(erc20Map)
	erc20ContractsHolder := createMockErc20ContractsHolder(tokens, safeContractEthAddress, availableBalances)

	numRelayers := 3
	ethereumChainMock := mock.NewEthereumChainMock()
	ethereumChainMock.SetQuorum(numRelayers)
	expectedStatuses := []byte{clients.Executed, clients.Rejected}
	ethereumChainMock.GetStatusesAfterExecutionHandler = func() []byte {
		return expectedStatuses
	}
	multiversXChainMock := mock.NewMultiversXChainMock()
	for i := 0; i < len(deposits); i++ {
		nativeBalanceValue := big.NewInt(int64(i * 1000))

		if !withNativeTokens {
			ethereumChainMock.UpdateNativeTokens(tokensAddresses[i], true)
			ethereumChainMock.UpdateMintBurnTokens(tokensAddresses[i], false)
			ethereumChainMock.UpdateTotalBalances(tokensAddresses[i], nativeBalanceValue)

			multiversXChainMock.AddTokensPair(tokensAddresses[i], deposits[i].Ticker, withNativeTokens, true, zero, nativeBalanceValue, nativeBalanceValue)
		} else {
			ethereumChainMock.UpdateNativeTokens(tokensAddresses[i], false)
			ethereumChainMock.UpdateMintBurnTokens(tokensAddresses[i], true)
			ethereumChainMock.UpdateBurnBalances(tokensAddresses[i], zero)
			ethereumChainMock.UpdateMintBalances(tokensAddresses[i], zero)

			multiversXChainMock.AddTokensPair(tokensAddresses[i], deposits[i].Ticker, withNativeTokens, true, zero, zero, nativeBalanceValue)
		}
	}
	pendingBatch := mock.MultiversXPendingBatch{
		Nonce:              big.NewInt(1),
		MultiversXDeposits: deposits,
	}
	multiversXChainMock.SetPendingBatch(&pendingBatch)
	multiversXChainMock.SetQuorum(numRelayers)

	ethereumChainMock.ProposeMultiTransferEsdtBatchCalled = func() {
		deposit := deposits[0]

		multiversXChainMock.AddDepositToCurrentBatch(deposit)
	}

	relayers := make([]bridgeComponents, 0, numRelayers)
	defer closeRelayers(relayers)

	messengers := integrationTests.CreateLinkedMessengers(numRelayers)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1200)
	defer cancel()
	multiversXChainMock.ProcessFinishedHandler = func() {
		log.Info("multiversXChainMock.ProcessFinishedHandler called")
		asyncCancelCall(cancel, time.Second*5)
	}

	for i := 0; i < numRelayers; i++ {
		argsBridgeComponents := createMockBridgeComponentsArgs(i, messengers[i], multiversXChainMock, ethereumChainMock)
		argsBridgeComponents.Configs.GeneralConfig.Eth.SafeContractAddress = safeContractEthAddress.Hex()
		argsBridgeComponents.Erc20ContractsHolder = erc20ContractsHolder
		relayer, err := factory.NewEthMultiversXBridgeComponents(argsBridgeComponents)
		require.Nil(t, err)

		multiversXChainMock.AddRelayer(relayer.MultiversXRelayerAddress())
		ethereumChainMock.AddRelayer(relayer.EthereumRelayerAddress())

		go func() {
			err = relayer.Start()
			integrationTests.Log.LogIfError(err)
			require.Nil(t, err)
		}()

		relayers = append(relayers, relayer)
	}

	<-ctx.Done()

	// let all transactions propagate
	time.Sleep(time.Second * 5)

	checkTestStatus(t, multiversXChainMock, ethereumChainMock, numTransactions, deposits, tokensAddresses)
}

func createTransactions(n int) ([]mock.MultiversXDeposit, []common.Address, map[common.Address]*big.Int) {
	tokensAddresses := make([]common.Address, 0, n)
	deposits := make([]mock.MultiversXDeposit, 0, n)
	erc20 := make(map[common.Address]*big.Int)
	for i := 0; i < n; i++ {
		deposit, tokenAddress := createTransaction(i)
		tokensAddresses = append(tokensAddresses, tokenAddress)
		deposits = append(deposits, deposit)

		val, found := erc20[tokenAddress]
		if !found {
			val = big.NewInt(0)
			erc20[tokenAddress] = val
		}
		val.Add(val, deposit.Amount)
	}

	return deposits, tokensAddresses, erc20
}

func createTransaction(index int) (mock.MultiversXDeposit, common.Address) {
	tokenAddress := testsCommon.CreateRandomEthereumAddress()

	return mock.MultiversXDeposit{
		From:   testsCommon.CreateRandomMultiversXAddress(),
		To:     testsCommon.CreateRandomEthereumAddress(),
		Ticker: fmt.Sprintf("tck-00000%d", index+1),
		Amount: big.NewInt(int64(index * 1000)),
	}, tokenAddress
}

func closeRelayers(relayers []bridgeComponents) {
	for _, r := range relayers {
		_ = r.Close()
	}
}

func checkTestStatus(
	t *testing.T,
	multiversXChainMock *mock.MultiversXChainMock,
	ethereumChainMock *mock.EthereumChainMock,
	numTransactions int,
	deposits []mock.MultiversXDeposit,
	tokensAddresses []common.Address,
) {
	transactions := multiversXChainMock.GetAllSentTransactions(context.Background())
	assert.Equal(t, 5, len(transactions))
	assert.Nil(t, multiversXChainMock.ProposedTransfer())
	assert.NotNil(t, multiversXChainMock.PerformedActionID())

	transfer := ethereumChainMock.GetLastProposedTransfer()
	require.NotNil(t, transfer)

	require.Equal(t, numTransactions, len(transfer.Amounts))

	for i := 0; i < len(transfer.Amounts); i++ {
		assert.Equal(t, deposits[i].To, transfer.Recipients[i])
		assert.Equal(t, tokensAddresses[i], transfer.Tokens[i])
		assert.Equal(t, deposits[i].Amount, transfer.Amounts[i])
	}
}
