//go:build slow

package slowTests

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"testing"

	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
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
		CallData: prependLenAndDataMarker(createScCallData("callPayable", 500000000)),
	}

	ethTx2 := framework.EthTransaction{
		From:     bytes.Repeat([]byte{0x02}, 20),
		To:       setup.MultiversxHandler.CalleeScAddress.AddressBytes(),
		TokenID:  token.MvxUniversalTokenTicker,
		Amount:   big.NewInt(2000),
		Nonce:    2,
		CallData: prependLenAndDataMarker(createScCallData("callPayable", 500000000)),
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

	// check transactions status
	setup.CheckTransactionStatus(tx1Hash, 0)
	setup.CheckTransactionStatus(tx2Hash, 1)
}

func prependLenAndDataMarker(input []byte) []byte {
	buff32 := make([]byte, bridgeCore.Uint32ArgBytes)
	binary.BigEndian.PutUint32(buff32, uint32(len(input)))

	prefix := append([]byte{bridgeCore.DataPresentProtocolMarker}, buff32...)

	return append(prefix, input...)
}
