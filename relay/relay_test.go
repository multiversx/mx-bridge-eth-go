package relay

import (
	"context"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"testing"
)

// implements interface
var (
	_ = Startable(&Relay{})
)

type testBridge struct{}

func (b *testBridge) GetPendingDepositTransaction(context.Context) *bridge.DepositTransaction {
	return nil
}

func (b *testBridge) Propose(*bridge.DepositTransaction) {}

func (b *testBridge) WasProposed(*bridge.DepositTransaction) bool {
	return false
}

func (b *testBridge) WasExecuted(*bridge.DepositTransaction) bool {
	return false
}

func (b *testBridge) Sign(*bridge.DepositTransaction) {}

func (b *testBridge) Execute(*bridge.DepositTransaction) (string, error) {
	return "", nil
}

func (b *testBridge) SignersCount(*bridge.DepositTransaction) uint {
	return 0
}

func TestWillBridgeToElrond(t *testing.T) {
	_ = Relay{
		elrondBridge: &testBridge{},
		ethBridge:    &testBridge{},
	}
}
