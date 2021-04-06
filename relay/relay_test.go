package relay

import (
	"context"
	"github.com/ElrondNetwork/elrond-eth-bridge/safe"
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
	channel                safe.SafeTxChan
	lastBridgedTransaction *safe.DepositTransaction
}

func (c *testSafe) GetTransactions(context.Context, *big.Int, safe.SafeTxChan) {}

func (c *testSafe) Bridge(tx *safe.DepositTransaction) (string, error) {
	c.lastBridgedTransaction = tx

	return "", nil
}

func TestWillBridgeToElrond(t *testing.T) {
	ethChannel := make(safe.SafeTxChan)
	defer close(ethChannel)
	ethSafe := &testSafe{ethChannel, nil}
	elrondSafe := &testSafe{}
	bridge := Relay{
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
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go func() { ethChannel <- transaction }()
	bridge.Start(ctx)

	if !reflect.DeepEqual(elrondSafe.lastBridgedTransaction, transaction) {
		t.Errorf("Expected transaction: %v to be bridged to elrond, but %v was actually bridged", transaction, elrondSafe.lastBridgedTransaction)
	}
}
