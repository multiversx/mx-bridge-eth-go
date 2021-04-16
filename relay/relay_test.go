package relay

import (
	"context"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"math/big"
	"reflect"
	"testing"
	"time"
)

// implements interface
var (
	_ = Startable(&Relay{})
)

type testSafe struct {
	channel                bridge.SafeTxChan
	lastBridgedTransaction *bridge.DepositTransaction
}

func (c *testSafe) GetTransactions(context.Context, *big.Int, bridge.SafeTxChan) {}

func (c *testSafe) Bridge(tx *bridge.DepositTransaction) (string, error) {
	c.lastBridgedTransaction = tx

	return "", nil
}

func TestWillBridgeToElrond(t *testing.T) {
	ethChannel := make(bridge.SafeTxChan)
	defer close(ethChannel)
	ethSafe := &testSafe{ethChannel, nil}
	elrondSafe := &testSafe{}
	relay := Relay{
		ethSafe:       ethSafe,
		elrondSafe:    elrondSafe,
		ethChannel:    ethChannel,
		elrondChannel: nil,
	}

	transaction := &bridge.DepositTransaction{
		Hash:         "hash",
		From:         "someone",
		TokenAddress: "erc20 address",
		Amount:       big.NewInt(42),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go func() { ethChannel <- transaction }()
	relay.Start(ctx)

	if !reflect.DeepEqual(elrondSafe.lastBridgedTransaction, transaction) {
		t.Errorf("Expected transaction: %v to be bridged to elrond, but %v was actually bridged", transaction, elrondSafe.lastBridgedTransaction)
	}
}
