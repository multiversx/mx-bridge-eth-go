package relay

import (
	"context"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/stretchr/testify/assert"
)

func TestReadPendingTransaction(t *testing.T) {
	setTestLogLevel()
	t.Run("it will read the next pending transaction", func(t *testing.T) {
		expected := &bridge.DepositTransaction{To: "address", DepositNonce: 0}
		sourceBridge := &bridgeStub{pendingTransactions: []*bridge.DepositTransaction{expected}}
		monitor := NewMonitor(sourceBridge, &bridgeStub{}, &timerStub{}, &topologyProviderStub{}, "testMonitor")

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expected, monitor.pendingTransaction)
	})
	t.Run("it will sleep and try again if there is no pending transaction", func(t *testing.T) {
		expected := &bridge.DepositTransaction{To: "address", DepositNonce: 0}
		sourceBridge := &bridgeStub{pendingTransactions: []*bridge.DepositTransaction{nil, expected}}
		monitor := NewMonitor(sourceBridge, &bridgeStub{}, &timerStub{}, &topologyProviderStub{}, "testMonitor")

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expected, monitor.pendingTransaction)
		assert.GreaterOrEqual(t, sourceBridge.pendingTransactionCallIndex, 1)
	})
}

func TestProposeTransaction(t *testing.T) {
	setTestLogLevel()
	t.Run("it will proposeTransfer transaction when leader", func(t *testing.T) {
		expect := &bridge.DepositTransaction{To: "address", DepositNonce: 0}
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
		expect := bridge.Nonce(0)
		destinationBridge := &bridgeStub{}
		monitor := NewMonitor(
			&bridgeStub{pendingTransactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: expect}}},
			destinationBridge,
			&timerStub{},
			&topologyProviderStub{peerCount: 2, amITheLeader: false},
			"testMonitor",
		)

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expect, destinationBridge.lastWasProposedTransferNonce)
	})
	t.Run("it will sign proposed transaction if not leader", func(t *testing.T) {
		expect := bridge.ActionId(42)
		destinationBridge := &bridgeStub{wasProposedTransfer: true, proposeTransferActionId: expect}
		monitor := NewMonitor(
			&bridgeStub{pendingTransactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: 0}}},
			destinationBridge,
			&timerStub{},
			&topologyProviderStub{peerCount: 2, amITheLeader: false},
			"testMonitor",
		)

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expect, destinationBridge.lastSignedActionId)
	})
	t.Run("it will try to proposeTransfer again if timeout", func(t *testing.T) {
		expect := &bridge.DepositTransaction{To: "address", DepositNonce: 0}
		destinationBridge := &bridgeStub{wasProposedTransfer: false}
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
	setTestLogLevel()
	t.Run("it will execute transfer when leader and number of signatures is > 67%", func(t *testing.T) {
		expect := bridge.ActionId(42)
		destinationBridge := &bridgeStub{signersCount: 3, proposeTransferActionId: expect}
		monitor := NewMonitor(
			&bridgeStub{pendingTransactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: 0}}},
			destinationBridge,
			&timerStub{},
			&topologyProviderStub{peerCount: 4, amITheLeader: true},
			"testMonitor",
		)

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expect, destinationBridge.lastExecutedActionId)
	})
	t.Run("it will sleep and try to wait for signatures again when the number of signatures is < 67%", func(t *testing.T) {
		expect := bridge.ActionId(42)
		destinationBridge := &bridgeStub{signersCount: 0, proposeTransferActionId: expect}
		monitor := NewMonitor(
			&bridgeStub{pendingTransactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: 0}}},
			destinationBridge,
			&timerStub{afterDuration: 3 * time.Millisecond},
			&topologyProviderStub{peerCount: 4, amITheLeader: true},
			"testMonitor",
		)

		go func() {
			time.Sleep(8 * time.Millisecond)
			destinationBridge.signersCount = 3
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 11*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expect, destinationBridge.lastExecutedActionId)
	})
}

func TestExecute(t *testing.T) {
	setTestLogLevel()
	t.Run("it will wait for execution when not leader", func(t *testing.T) {
		expect := bridge.ActionId(42)
		destinationBridge := &bridgeStub{signersCount: 3, wasExecuted: false, wasProposedTransfer: true, proposeTransferActionId: expect}
		timer := &timerStub{afterDuration: 3 * time.Millisecond}
		provider := &topologyProviderStub{peerCount: 4, amITheLeader: false}
		monitor := NewMonitor(
			&bridgeStub{pendingTransactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: 0}}},
			destinationBridge,
			timer,
			provider,
			"testMonitor",
		)

		go func() {
			time.Sleep(11 * time.Millisecond)
			provider.amITheLeader = true
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 16*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expect, destinationBridge.lastExecutedActionId)
	})
}

func TestProposeSetStatus(t *testing.T) {
	setTestLogLevel()
	t.Run("it will propose to set status when leader", func(t *testing.T) {
		destinationBridge := &bridgeStub{
			signersCount:            3,
			wasExecuted:             true,
			wasProposedTransfer:     true,
			proposeTransferActionId: bridge.ActionId(41),
		}
		provider := &topologyProviderStub{peerCount: 4, amITheLeader: true}
		sourceBridge := &bridgeStub{pendingTransactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: 0}}}
		monitor := NewMonitor(
			sourceBridge,
			destinationBridge,
			&timerStub{},
			provider,
			"testMonitor",
		)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.True(t, sourceBridge.wasProposedSetStatusSuccessOnPendingTransfer)
	})
	t.Run("it will sign proposed set status when not leader", func(t *testing.T) {
		expect := bridge.ActionId(42)
		destinationBridge := &bridgeStub{
			signersCount:            3,
			wasExecuted:             true,
			wasProposedTransfer:     true,
			proposeTransferActionId: bridge.ActionId(41),
		}
		provider := &topologyProviderStub{peerCount: 4, amITheLeader: false}
		sourceBridge := &bridgeStub{
			pendingTransactions:                          []*bridge.DepositTransaction{{To: "address", DepositNonce: 0}},
			wasProposedSetStatusSuccessOnPendingTransfer: true,
			proposeSetStatusActionId:                     expect,
		}
		monitor := NewMonitor(
			sourceBridge,
			destinationBridge,
			&timerStub{},
			provider,
			"testMonitor",
		)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expect, sourceBridge.lastSignedActionId)
	})
	t.Run("it will execute set status when leader and number of signatures > 67%", func(t *testing.T) {
		expect := bridge.ActionId(42)
		destinationBridge := &bridgeStub{
			signersCount:            3,
			wasExecuted:             true,
			wasProposedTransfer:     true,
			proposeTransferActionId: bridge.ActionId(41),
		}
		provider := &topologyProviderStub{peerCount: 4, amITheLeader: true}
		sourceBridge := &bridgeStub{
			signersCount:        3,
			pendingTransactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: 0}},
			wasProposedSetStatusSuccessOnPendingTransfer: true,
			proposeSetStatusActionId:                     expect,
		}
		monitor := NewMonitor(
			sourceBridge,
			destinationBridge,
			&timerStub{},
			provider,
			"testMonitor",
		)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expect, sourceBridge.lastExecutedActionId)
	})
	t.Run("it will execute set status when leader after waiting", func(t *testing.T) {
		expect := bridge.ActionId(42)
		destinationBridge := &bridgeStub{
			signersCount:            3,
			wasExecuted:             true,
			wasProposedTransfer:     true,
			proposeTransferActionId: bridge.ActionId(41),
		}
		provider := &topologyProviderStub{peerCount: 4, amITheLeader: false}
		sourceBridge := &bridgeStub{
			signersCount:        3,
			pendingTransactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: 0}},
			wasProposedSetStatusSuccessOnPendingTransfer: false,
			proposeSetStatusActionId:                     expect,
		}
		monitor := NewMonitor(
			sourceBridge,
			destinationBridge,
			&timerStub{afterDuration: 3 * time.Millisecond},
			provider,
			"testMonitor",
		)

		go func() {
			time.Sleep(14 * time.Millisecond)
			provider.amITheLeader = true
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expect, sourceBridge.lastExecutedActionId)
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
	pendingTransactionCallIndex                  int
	pendingTransactions                          []*bridge.DepositTransaction
	wasProposedTransfer                          bool
	lastProposedTransaction                      *bridge.DepositTransaction
	lastWasProposedTransferNonce                 bridge.Nonce
	lastSignedActionId                           bridge.ActionId
	signersCount                                 uint
	lastExecutedActionId                         bridge.ActionId
	wasExecuted                                  bool
	proposeTransferActionId                      bridge.ActionId
	wasProposedSetStatusSuccessOnPendingTransfer bool
	proposeSetStatusActionId                     bridge.ActionId
}

func (b *bridgeStub) GetPendingDepositTransaction(context.Context) *bridge.DepositTransaction {
	defer func() { b.pendingTransactionCallIndex++ }()

	if b.pendingTransactionCallIndex >= len(b.pendingTransactions) {
		return nil
	} else {
		return b.pendingTransactions[b.pendingTransactionCallIndex]
	}
}

func (b *bridgeStub) ProposeTransfer(_ context.Context, tx *bridge.DepositTransaction) (string, error) {
	b.wasProposedTransfer = true
	b.lastProposedTransaction = tx

	return "propose_tx_hash", nil
}

func (b *bridgeStub) ProposeSetStatusSuccessOnPendingTransfer(context.Context) {
	b.wasProposedSetStatusSuccessOnPendingTransfer = true
}

func (b *bridgeStub) ProposeSetStatusFailedOnPendingTransfer(context.Context) {}

func (b *bridgeStub) WasProposedTransfer(_ context.Context, nonce bridge.Nonce) bool {
	b.lastWasProposedTransferNonce = nonce
	return b.wasProposedTransfer
}

func (b *bridgeStub) GetActionIdForProposeTransfer(context.Context, bridge.Nonce) bridge.ActionId {
	return b.proposeTransferActionId
}

func (b *bridgeStub) WasProposedSetStatusSuccessOnPendingTransfer(context.Context) bool {
	return b.wasProposedSetStatusSuccessOnPendingTransfer
}

func (b *bridgeStub) WasProposedSetStatusFailedOnPendingTransfer(context.Context) bool {
	return false
}

func (b *bridgeStub) GetActionIdForSetStatusOnPendingTransfer(context.Context) bridge.ActionId {
	return b.proposeSetStatusActionId
}

func (b *bridgeStub) WasExecuted(context.Context, bridge.ActionId) bool {
	return b.wasExecuted
}

func (b *bridgeStub) Sign(_ context.Context, actionId bridge.ActionId) (string, error) {
	b.lastSignedActionId = actionId
	return "sign_tx_hash", nil
}

func (b *bridgeStub) Execute(_ context.Context, actionId bridge.ActionId) (string, error) {
	b.lastExecutedActionId = actionId
	return "execution hash", nil
}

func (b *bridgeStub) SignersCount(context.Context, bridge.ActionId) uint {
	return b.signersCount
}
