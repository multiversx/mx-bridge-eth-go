package relay

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/stretchr/testify/assert"
)

func TestReadPendingTransaction(t *testing.T) {
	setLoggerLevel()
	t.Run("it will read the next pending transaction", func(t *testing.T) {
		expected := &bridge.DepositTransaction{To: "address", DepositNonce: big.NewInt(0)}
		sourceBridge := &bridgeStub{pendingTransactions: []*bridge.DepositTransaction{expected}}
		monitor := NewMonitor(sourceBridge, &bridgeStub{}, &timerStub{}, &topologyProviderStub{}, "testMonitor")

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expected, monitor.pendingTransaction)
	})
	t.Run("it will sleep and try again if there is no pending transaction", func(t *testing.T) {
		expected := &bridge.DepositTransaction{To: "address", DepositNonce: big.NewInt(0)}
		sourceBridge := &bridgeStub{pendingTransactions: []*bridge.DepositTransaction{nil, expected}}
		monitor := NewMonitor(sourceBridge, &bridgeStub{}, &timerStub{}, &topologyProviderStub{}, "testMonitor")

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expected, monitor.pendingTransaction)
		assert.GreaterOrEqual(t, sourceBridge.pendingTransactionCallIndex, 1)
	})
}

func TestPropose(t *testing.T) {
	setLoggerLevel()
	t.Run("it will propose transaction when leader", func(t *testing.T) {
		expect := &bridge.DepositTransaction{To: "address", DepositNonce: big.NewInt(0)}
		destinationBridge := &bridgeStub{}
		monitor := NewMonitor(
			&bridgeStub{pendingTransactions: []*bridge.DepositTransaction{expect}},
			destinationBridge,
			&timerStub{},
			&topologyProviderStub{peerCount: 1, amITheLeader: true},
			"testMonitor",
		)

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expect, destinationBridge.lastProposedTransaction)
	})
	t.Run("it will wait for proposal if not leader", func(t *testing.T) {
		expect := &bridge.DepositTransaction{To: "address", DepositNonce: big.NewInt(0)}
		destinationBridge := &bridgeStub{}
		monitor := NewMonitor(
			&bridgeStub{pendingTransactions: []*bridge.DepositTransaction{expect}},
			destinationBridge,
			&timerStub{},
			&topologyProviderStub{peerCount: 2, amITheLeader: false},
			"testMonitor",
		)

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expect, destinationBridge.lastWasProposedTransaction)
	})
	t.Run("it will sign proposed transaction if not leader", func(t *testing.T) {
		expect := &bridge.DepositTransaction{To: "address", DepositNonce: big.NewInt(0)}
		destinationBridge := &bridgeStub{wasProposed: true}
		monitor := NewMonitor(
			&bridgeStub{pendingTransactions: []*bridge.DepositTransaction{expect}},
			destinationBridge,
			&timerStub{},
			&topologyProviderStub{peerCount: 2, amITheLeader: false},
			"testMonitor",
		)

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expect, destinationBridge.lastSignedTransaction)
	})
	t.Run("it will try to propose again if timeout", func(t *testing.T) {
		expect := &bridge.DepositTransaction{To: "address", DepositNonce: big.NewInt(0)}
		destinationBridge := &bridgeStub{wasProposed: false}
		timer := &timerStub{afterDuration: 3 * time.Millisecond}
		provider := &topologyProviderStub{peerCount: 2, amITheLeader: false}
		monitor := NewMonitor(
			&bridgeStub{pendingTransactions: []*bridge.DepositTransaction{expect}},
			destinationBridge,
			timer,
			provider,
			"testMonitor",
		)

		go func() {
			time.Sleep(2 * time.Millisecond)
			provider.amITheLeader = true
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expect, destinationBridge.lastProposedTransaction)
	})
}

func TestWaitForSignatures(t *testing.T) {
	setLoggerLevel()
	t.Run("it will execute when number of signatures is > 67%", func(t *testing.T) {
		expect := &bridge.DepositTransaction{To: "address", DepositNonce: big.NewInt(0)}
		destinationBridge := &bridgeStub{signersCount: 3}
		monitor := NewMonitor(
			&bridgeStub{pendingTransactions: []*bridge.DepositTransaction{expect}},
			destinationBridge,
			&timerStub{},
			&topologyProviderStub{peerCount: 4, amITheLeader: true},
			"testMonitor",
		)

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expect, destinationBridge.lastExecutedTransaction)
	})
	t.Run("it will sleep and try to wait for signatures again", func(t *testing.T) {
		expect := &bridge.DepositTransaction{To: "address", DepositNonce: big.NewInt(0)}
		destinationBridge := &bridgeStub{signersCount: 0}
		monitor := NewMonitor(
			&bridgeStub{pendingTransactions: []*bridge.DepositTransaction{expect}},
			destinationBridge,
			&timerStub{afterDuration: 3 * time.Millisecond},
			&topologyProviderStub{peerCount: 4, amITheLeader: true},
			"testMonitor",
		)

		go func() {
			time.Sleep(2 * time.Millisecond)
			destinationBridge.signersCount = 3
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expect, destinationBridge.lastExecutedTransaction)
	})
}

func TestExecute(t *testing.T) {
	t.Run("it will wait for execution when not leader", func(t *testing.T) {
		expect := &bridge.DepositTransaction{To: "address", DepositNonce: big.NewInt(0)}
		destinationBridge := &bridgeStub{signersCount: 3, wasExecuted: false, wasProposed: true}
		timer := &timerStub{afterDuration: 3 * time.Millisecond}
		provider := &topologyProviderStub{peerCount: 4, amITheLeader: false}
		monitor := NewMonitor(
			&bridgeStub{pendingTransactions: []*bridge.DepositTransaction{expect}},
			destinationBridge,
			timer,
			provider,
			"testMonitor",
		)

		go func() {
			time.Sleep(2 * time.Millisecond)
			provider.amITheLeader = true
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expect, destinationBridge.lastExecutedTransaction)
	})
}

type topologyProviderStub struct {
	amITheLeader bool
	peerCount    int
}

func (s *topologyProviderStub) AmITheLeader() bool {
	return s.amITheLeader
}

func (s *topologyProviderStub) PeerCount() int {
	return s.peerCount
}

type bridgeStub struct {
	pendingTransactionCallIndex int
	pendingTransactions         []*bridge.DepositTransaction
	wasProposed                 bool
	lastProposedTransaction     *bridge.DepositTransaction
	lastWasProposedTransaction  *bridge.DepositTransaction
	lastSignedTransaction       *bridge.DepositTransaction
	signersCount                uint
	lastExecutedTransaction     *bridge.DepositTransaction
	wasExecuted                 bool
}

func (b *bridgeStub) GetPendingDepositTransaction(context.Context) *bridge.DepositTransaction {
	defer func() { b.pendingTransactionCallIndex++ }()

	if b.pendingTransactionCallIndex >= len(b.pendingTransactions) {
		return nil
	} else {
		return b.pendingTransactions[b.pendingTransactionCallIndex]
	}
}

func (b *bridgeStub) Propose(_ context.Context, tx *bridge.DepositTransaction) {
	b.lastProposedTransaction = tx
}

func (b *bridgeStub) WasProposed(_ context.Context, tx *bridge.DepositTransaction) bool {
	b.lastWasProposedTransaction = tx
	return b.wasProposed
}

func (b *bridgeStub) WasExecuted(context.Context, *bridge.DepositTransaction) bool {
	return b.wasExecuted
}

func (b *bridgeStub) Sign(_ context.Context, tx *bridge.DepositTransaction) {
	b.lastSignedTransaction = tx
}

func (b *bridgeStub) Execute(_ context.Context, tx *bridge.DepositTransaction) (string, error) {
	b.lastExecutedTransaction = tx
	return "", nil
}

func (b *bridgeStub) SignersCount(context.Context, *bridge.DepositTransaction) uint {
	return b.signersCount
}
