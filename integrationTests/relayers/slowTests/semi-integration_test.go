//go:build integration

package slowTests

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/require"
)

func TestBridgeProxyExecuteTwiceSameDeposit(t *testing.T) {
	setup := framework.NewSetup(t)

	// generate and issue token
	token := GenerateUnlistedTokenFromMvx()
	setup.IssueToken(token)

	ethTx1 := framework.EthTransaction{
		From:     bytes.Repeat([]byte{0x01}, 20),
		To:       setup.MultiversxHandler.CalleeScAddress.AddressBytes(),
		TokenID:  token.MvxUniversalTokenTicker,
		Amount:   big.NewInt(100),
		Nonce:    1,
		CallData: prependLenAndDataMarker(CreateScCallData("callPayable", 500000000)),
	}

	ethTx2 := framework.EthTransaction{
		From:     bytes.Repeat([]byte{0x02}, 20),
		To:       setup.MultiversxHandler.CalleeScAddress.AddressBytes(),
		TokenID:  token.MvxUniversalTokenTicker,
		Amount:   big.NewInt(2000),
		Nonce:    2,
		CallData: prependLenAndDataMarker(CreateScCallData("callPayable", 500000000)),
	}

	// deposit txs in bridge proxy
	tokenData := setup.TokensRegistry.GetTokenData(token.AbstractTokenIdentifier)
	setup.MultiversxHandler.CallDepositOnBridgeProxy(setup.Ctx, ethTx1, 0, tokenData)
	setup.MultiversxHandler.CallDepositOnBridgeProxy(setup.Ctx, ethTx2, 0, tokenData)

	// make 2 execute calls for same deposit
	tx1Hash := setup.MultiversxHandler.ExecuteDepositWithoutGenerateBlocks(setup.Ctx, 1, 0)
	tx2Hash := setup.MultiversxHandler.ExecuteDepositWithoutGenerateBlocks(setup.Ctx, 1, 1)

	// generate blocks
	setup.ChainSimulator.GenerateBlocks(setup.Ctx, 10)

	txResult1, txStatus1 := setup.ChainSimulator.GetTransactionResultWithoutGenerateBlocks(setup.Ctx, tx1Hash)
	// the first transaction will succeed
	require.Equal(setup, transaction.TxStatusSuccess, txStatus1, fmt.Sprintf("tx hash: %s,\n tx: %s", tx1Hash, stringifyTxResult(txResult1)))

	txResult2, txStatus2 := setup.ChainSimulator.GetTransactionResultWithoutGenerateBlocks(setup.Ctx, tx2Hash)
	// the second transaction will fail
	require.Equal(setup, transaction.TxStatusFail, txStatus2, fmt.Sprintf("tx hash: %s,\n tx: %s", tx2Hash, stringifyTxResult(txResult2)))
}

func stringifyTxResult(txResult *data.TransactionOnNetwork) string {
	jsonData, _ := json.MarshalIndent(txResult, "", "  ")

	return string(jsonData)
}

func prependLenAndDataMarker(input []byte) []byte {
	buff32 := make([]byte, bridgeCore.Uint32ArgBytes)
	binary.BigEndian.PutUint32(buff32, uint32(len(input)))

	prefix := append([]byte{bridgeCore.DataPresentProtocolMarker}, buff32...)

	return append(prefix, input...)
}
