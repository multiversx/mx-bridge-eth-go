package relayers

// TODO fix these tests when the Elrond->Ethereum is finalized
//import (
//	"context"
//	"fmt"
//	"math/big"
//	"testing"
//	"time"
//
//	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
//	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth"
//	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests"
//	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests/mock"
//	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
//	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
//	mockInteractors "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/interactors"
//	"github.com/ethereum/go-ethereum/common"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//)
//
//func TestRelayersShouldExecuteTransferFromElrondToEth(t *testing.T) {
//	if testing.Short() {
//		t.Skip("this is not a short test")
//	}
//
//	numTransactions := 2
//	deposits, tokensAddresses, erc20Map := createTransactions(numTransactions)
//
//	safeContractEthAddress := testsCommon.CreateRandomEthereumAddress()
//	erc20Contracts := make(map[common.Address]eth.Erc20Contract)
//	for addr, val := range erc20Map {
//		value := big.NewInt(0).Set(val)
//		erc20Contracts[addr] = &mockInteractors.Erc20ContractStub{
//			BalanceOfCalled: func(ctx context.Context, account common.Address) (*big.Int, error) {
//				if account == safeContractEthAddress {
//					return value, nil
//				}
//
//				return big.NewInt(0), nil
//			},
//		}
//	}
//
//	ethereumChainMock := mock.NewEthereumChainMock()
//	ethereumChainMock.SetQuorum(3)
//	expectedStatuses := []byte{bridge.Executed, bridge.Rejected}
//	ethereumChainMock.GetStatusesAfterExecutionHandler = func() []byte {
//		return expectedStatuses
//	}
//	elrondChainMock := mock.NewElrondChainMock()
//	for i := 0; i < len(deposits); i++ {
//		elrondChainMock.AddTokensPair(tokensAddresses[i], deposits[i].Ticker)
//	}
//	pendingBatch := mock.ElrondPendingBatch{
//		Nonce:                  big.NewInt(1),
//		Timestamp:              big.NewInt(0),
//		LastUpdatedBlockNumber: big.NewInt(0),
//		ElrondDeposits:         deposits,
//		Status:                 0,
//	}
//
//	elrondChainMock.SetPendingBatch(&pendingBatch)
//
//	numRelayers := 3
//	relayers := make([]*relay.Relay, 0, numRelayers)
//	defer func() {
//		for _, r := range relayers {
//			_ = r.Close()
//		}
//	}()
//
//	messengers := integrationTests.CreateLinkedMessengers(numRelayers)
//
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1200)
//	defer cancel()
//	elrondChainMock.ProcessFinishedHandler = func() {
//		cancel()
//	}
//
//	for i := 0; i < numRelayers; i++ {
//		argsRelay := mock.CreateMockRelayArgs("elrond <-> eth", i, messengers[i], elrondChainMock, ethereumChainMock)
//		argsRelay.Configs.GeneralConfig.Eth.SafeContractAddress = safeContractEthAddress.Hex()
//		argsRelay.Erc20Contracts = erc20Contracts
//		r, err := relay.NewRelay(argsRelay)
//		require.Nil(t, err)
//
//		elrondChainMock.AddRelayer(r.ElrondAddress())
//		ethereumChainMock.AddRelayer(r.EthereumAddress())
//
//		go func() {
//			err = r.Start(ctx)
//			integrationTests.Log.LogIfError(err)
//			require.Nil(t, err)
//		}()
//
//		relayers = append(relayers, r)
//	}
//
//	<-ctx.Done()
//
//	transactions := elrondChainMock.GetAllSentTransactions(context.Background())
//	assert.Equal(t, 1, len(transactions))
//	assert.Nil(t, elrondChainMock.ProposedTransfer())
//	assert.Nil(t, elrondChainMock.PerformedActionID())
//
//	transfer := ethereumChainMock.GetLastProposedTransfer()
//	require.NotNil(t, transfer)
//
//	require.Equal(t, numTransactions, len(transfer.Amounts))
//
//	for i := 0; i < len(transfer.Amounts); i++ {
//		assert.Equal(t, deposits[i].To, transfer.Recipients[i])
//		assert.Equal(t, tokensAddresses[i], transfer.Tokens[i])
//		assert.Equal(t, deposits[i].Amount, transfer.Amounts[i])
//	}
//}
//
//func TestRelayersShouldExecuteTransferFromElrondToEthIfTransactionsAppearInBatch(t *testing.T) {
//	if testing.Short() {
//		t.Skip("this is not a short test")
//	}
//
//	numTransactions := 2
//	deposits, tokensAddresses, erc20Map := createTransactions(numTransactions)
//
//	safeContractEthAddress := testsCommon.CreateRandomEthereumAddress()
//	erc20Contracts := make(map[common.Address]eth.Erc20Contract)
//	for addr, val := range erc20Map {
//		value := big.NewInt(0).Set(val)
//		erc20Contracts[addr] = &mockInteractors.Erc20ContractStub{
//			BalanceOfCalled: func(ctx context.Context, account common.Address) (*big.Int, error) {
//				if account == safeContractEthAddress {
//					return value, nil
//				}
//
//				return big.NewInt(0), nil
//			},
//		}
//	}
//
//	ethereumChainMock := mock.NewEthereumChainMock()
//	ethereumChainMock.SetQuorum(3)
//	expectedStatuses := []byte{bridge.Executed, bridge.Rejected}
//	ethereumChainMock.GetStatusesAfterExecutionHandler = func() []byte {
//		return expectedStatuses
//	}
//	elrondChainMock := mock.NewElrondChainMock()
//	for i := 0; i < len(deposits); i++ {
//		elrondChainMock.AddTokensPair(tokensAddresses[i], deposits[i].Ticker)
//	}
//	pendingBatch := mock.ElrondPendingBatch{
//		Nonce:                  big.NewInt(1),
//		Timestamp:              big.NewInt(0),
//		LastUpdatedBlockNumber: big.NewInt(0),
//		ElrondDeposits:         deposits,
//		Status:                 0,
//	}
//	elrondChainMock.SetPendingBatch(&pendingBatch)
//
//	ethereumChainMock.ProposeMultiTransferEsdtBatchCalled = func() {
//		deposit := deposits[0]
//
//		elrondChainMock.AddDepositToCurrentBatch(deposit)
//	}
//
//	numRelayers := 3
//	relayers := make([]*relay.Relay, 0, numRelayers)
//	defer func() {
//		for _, r := range relayers {
//			_ = r.Close()
//		}
//	}()
//
//	messengers := integrationTests.CreateLinkedMessengers(numRelayers)
//
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1200)
//	defer cancel()
//	elrondChainMock.ProcessFinishedHandler = func() {
//		cancel()
//	}
//
//	for i := 0; i < numRelayers; i++ {
//		argsRelay := mock.CreateMockRelayArgs("elrond <-> eth", i, messengers[i], elrondChainMock, ethereumChainMock)
//		argsRelay.Configs.GeneralConfig.Eth.SafeContractAddress = safeContractEthAddress.Hex()
//		argsRelay.Erc20Contracts = erc20Contracts
//		r, err := relay.NewRelay(argsRelay)
//		require.Nil(t, err)
//
//		elrondChainMock.AddRelayer(r.ElrondAddress())
//		ethereumChainMock.AddRelayer(r.EthereumAddress())
//
//		go func() {
//			err = r.Start(ctx)
//			integrationTests.Log.LogIfError(err)
//			require.Nil(t, err)
//		}()
//
//		relayers = append(relayers, r)
//	}
//
//	<-ctx.Done()
//
//	transactions := elrondChainMock.GetAllSentTransactions(context.Background())
//	assert.Equal(t, 1, len(transactions))
//	assert.Nil(t, elrondChainMock.ProposedTransfer())
//	assert.Nil(t, elrondChainMock.PerformedActionID())
//
//	transfer := ethereumChainMock.GetLastProposedTransfer()
//	require.NotNil(t, transfer)
//
//	require.Equal(t, numTransactions, len(transfer.Amounts))
//
//	for i := 0; i < len(transfer.Amounts); i++ {
//		assert.Equal(t, deposits[i].To, transfer.Recipients[i])
//		assert.Equal(t, tokensAddresses[i], transfer.Tokens[i])
//		assert.Equal(t, deposits[i].Amount, transfer.Amounts[i])
//	}
//}
//
//func createTransactions(n int) ([]mock.ElrondDeposit, []common.Address, map[common.Address]*big.Int) {
//	tokensAddresses := make([]common.Address, 0, n)
//	deposits := make([]mock.ElrondDeposit, 0, n)
//	erc20 := make(map[common.Address]*big.Int)
//	for i := 0; i < n; i++ {
//		deposit, tokenAddress := createTransaction(i)
//		tokensAddresses = append(tokensAddresses, tokenAddress)
//		deposits = append(deposits, deposit)
//
//		val, found := erc20[tokenAddress]
//		if !found {
//			val = big.NewInt(0)
//			erc20[tokenAddress] = val
//		}
//		val.Add(val, deposit.Amount)
//	}
//
//	return deposits, tokensAddresses, erc20
//}
//
//func createTransaction(index int) (mock.ElrondDeposit, common.Address) {
//	tokenAddress := testsCommon.CreateRandomEthereumAddress()
//
//	return mock.ElrondDeposit{
//		From:   testsCommon.CreateRandomElrondAddress(),
//		To:     testsCommon.CreateRandomEthereumAddress(),
//		Ticker: fmt.Sprintf("tck-00000%d", index+1),
//		Amount: big.NewInt(int64(index)),
//	}, tokenAddress
//}
