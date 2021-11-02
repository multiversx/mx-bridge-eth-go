package relayers

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests"
	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelayersShouldExecuteTransferFromElrondToEth(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numTransactions := 2
	senders, receivers, tokens, values, tickers := createTransactions(numTransactions)

	ethereumChainMock := mock.NewEthereumChainMock()
	ethereumChainMock.SetQuorum(3)
	expectedStatuses := []byte{bridge.Executed, bridge.Rejected}
	ethereumChainMock.GetStatusesAfterExecutionHandler = func() []byte {
		return expectedStatuses
	}
	elrondChainMock := mock.NewElrondChainMock()
	deposits := make([]mock.ElrondDeposit, 0)
	for i := 0; i < len(senders); i++ {
		deposits = append(deposits, mock.ElrondDeposit{
			From:         senders[i],
			To:           receivers[i],
			TokenAddress: tokens[i],
			Amount:       values[i],
		})
		elrondChainMock.AddTokensPair(tokens[i], tickers[i])
	}
	pendingBatch := mock.ElrondPendingBatch{
		Nonce:                  big.NewInt(1),
		Timestamp:              big.NewInt(0),
		LastUpdatedBlockNumber: big.NewInt(0),
		ElrondDeposits:         deposits,
		Status:                 0,
	}

	elrondChainMock.SetPendingBatch(&pendingBatch)

	numRelayers := 3
	relayers := make([]*relay.Relay, 0, numRelayers)
	defer func() {
		for _, r := range relayers {
			_ = r.Stop()
		}
	}()

	messengers := integrationTests.CreateLinkedMessengers(numRelayers)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1200000)
	defer cancel()
	elrondChainMock.ProcessFinishedHandler = func() {
		cancel()
	}

	for i := 0; i < numRelayers; i++ {
		argsRelay := mock.CreateMockRelayArgs("elrond <-> eth", i, messengers[i], elrondChainMock, ethereumChainMock)
		r, err := relay.NewRelay(argsRelay)
		require.Nil(t, err)

		elrondChainMock.AddRelayer(r.ElrondAddress())
		ethereumChainMock.AddRelayer(r.EthereumAddress())

		go func() {
			err = r.Start(ctx)
			integrationTests.Log.LogIfError(err)
			require.Nil(t, err)
		}()

		relayers = append(relayers, r)
	}

	<-ctx.Done()

	transactions := elrondChainMock.GetAllSentTransactions()
	assert.Equal(t, 1, len(transactions))
	assert.Nil(t, elrondChainMock.ProposedTransfer())
	assert.Nil(t, elrondChainMock.PerformedActionID())

	transfer := ethereumChainMock.GetLastProposedTransfer()
	require.NotNil(t, transfer)

	// if ExecuteTransfer got executed -> len(transfer.Amounts) == len(transfer.Tokens) == len(transfer.Recipients)
	require.Equal(t, numTransactions, len(transfer.Amounts))

	for i := 0; i < len(transfer.Amounts); i++ {
		assert.Equal(t, receivers[i], transfer.Recipients[i])
		assert.Equal(t, tokens[i], transfer.Tokens[i])
		assert.Equal(t, values[i], transfer.Amounts[i])
	}
}

func createTransactions(n int) ([]erdgoCore.AddressHandler, []common.Address, []common.Address, []*big.Int, []string) {
	tokens := make([]common.Address, 0)
	tickers := make([]string, 0)
	values := make([]*big.Int, 0)
	senders := make([]erdgoCore.AddressHandler, 0)
	receivers := make([]common.Address, 0)
	for i := 0; i < n; i++ {
		sender, receiver, token, value, ticker := createTransaction(i)
		tokens = append(tokens, token)
		tickers = append(tickers, ticker)
		values = append(values, value)
		senders = append(senders, sender)
		receivers = append(receivers, receiver)
	}
	return senders, receivers, tokens, values, tickers
}
func createTransaction(index int) (erdgoCore.AddressHandler, common.Address, common.Address, *big.Int, string) {
	tokenErc20 := integrationTests.CreateRandomEthereumAddress()
	ticker := fmt.Sprintf("tck-00000%d", index+1)
	value := big.NewInt(int64(index))
	from := integrationTests.CreateRandomElrondAddress()
	dest := integrationTests.CreateRandomEthereumAddress()
	return from, dest, tokenErc20, value, ticker
}
