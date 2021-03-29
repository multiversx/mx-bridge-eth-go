package bridge

import (
	"context"
	"github.com/ElrondNetwork/elrond-eth-bridge/safe"
	"math/big"
	"testing"
)

// implements interface
var (
	_ = Startable(&Bridge{})
)

type testSafe struct {
	channel                safe.SafeTxChan
	lastBridgedTransaction *safe.DepositTransaction
}

func (c *testSafe) GetTransactions(context.Context, *big.Int, safe.SafeTxChan) {}

func (c *testSafe) Bridge(tx *safe.DepositTransaction) {
	c.lastBridgedTransaction = tx
}

func TestWillBridgeToElrond(t *testing.T) {
	ethChannel := make(safe.SafeTxChan)
	defer close(ethChannel)
	ethSafe := &testSafe{ethChannel, nil}
	elrondSafe := &testSafe{}
	bridge := Bridge{
		ethSafe:       ethSafe,
		elrondSafe:    elrondSafe,
		ethChannel:    ethChannel,
		elrondChannel: nil,
	}

	transaction := &safe.DepositTransaction{
		Hash:         "hash",
		From:         "someone",
		TokenAddress: "erc20 address",
		Amount:       big.NewInt(42),
	}

	go func() { ethChannel <- transaction }()
	bridge.Start(context.Background())
	bridge.Monitor()

	if elrondSafe.lastBridgedTransaction != transaction {
		t.Errorf("Expected transaction: %v to be bridged to elrond", transaction)
	}
}
