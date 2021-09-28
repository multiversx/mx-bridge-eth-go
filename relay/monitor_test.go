package relay

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/testHelpers"
	"github.com/stretchr/testify/assert"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
)

func TestGetPending(t *testing.T) {
	testHelpers.SetTestLogLevel()
	t.Run("it will get the next pending transaction", func(t *testing.T) {
		expected := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		sourceBridge := &bridgeStub{pendingBatches: []*bridge.Batch{expected}}
		provider := &topologyProviderStub{}
		monitor := NewMonitor(sourceBridge, &bridgeStub{}, &testHelpers.TimerStub{}, provider, &quorumProviderStub{}, "testMonitor")

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expected, monitor.pendingBatch)
	})
	t.Run("it will sleep and try again if there is no pending transaction", func(t *testing.T) {
		expected := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		sourceBridge := &bridgeStub{pendingBatches: []*bridge.Batch{nil, expected}}
		monitor := NewMonitor(sourceBridge, &bridgeStub{}, &testHelpers.TimerStub{}, &topologyProviderStub{}, &quorumProviderStub{}, "testMonitor")

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Millisecond)
		defer cancel()
		monitor.Start(ctx)

		assert.Equal(t, expected, monitor.pendingBatch)
		assert.GreaterOrEqual(t, sourceBridge.pendingBatchCallIndex, 1)
	})
}

func TestProposeTransaction(t *testing.T) {
	testHelpers.SetTestLogLevel()
	t.Run("it will proposeTransfer transaction when leader", func(t *testing.T) {
		expected := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		destinationBridge := &bridgeStub{}
		monitor := NewMonitor(
			&bridgeStub{pendingBatches: []*bridge.Batch{expected}},
			destinationBridge,
			&testHelpers.TimerStub{},
			&topologyProviderStub{peerCount: 1, amITheLeader: true},
			&quorumProviderStub{},
			"testMonitor",
		)

		destinationBridge.lock()

		go func() {
			monitor.Start(context.Background())
		}()

		// allow propose transfer
		destinationBridge.proposeTransferMutex.Unlock()

		time.Sleep(5 * time.Millisecond)
		assert.Equal(t, expected, destinationBridge.lastProposedBatch)
	})
	t.Run("it will proposeStatus Rejected when proposeTransfer fails", func(t *testing.T) {
		batch := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		sourceBridge := &bridgeStub{
			pendingBatches: []*bridge.Batch{batch},
		}
		monitor := NewMonitor(
			sourceBridge,
			&bridgeStub{
				proposeTransferError: errors.New("some error"),
			},
			&testHelpers.TimerStub{},
			&topologyProviderStub{peerCount: 1, amITheLeader: true},
			&quorumProviderStub{},
			"testMonitor",
		)

		sourceBridge.lock()

		go func() {
			monitor.Start(context.Background())
		}()

		// allow propose status
		sourceBridge.proposeSetStatusMutex.Unlock()

		time.Sleep(5 * time.Millisecond)
		assert.Equal(t, bridge.Rejected, sourceBridge.proposedStatusBatch.Transactions[0].Status)
	})
	t.Run("it will wait for proposal if not leader", func(t *testing.T) {
		expect := bridge.NewBatchId(0)
		batch := &bridge.Batch{
			Id:           expect,
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		destinationBridge := &bridgeStub{}
		monitor := NewMonitor(
			&bridgeStub{pendingBatches: []*bridge.Batch{batch}},
			destinationBridge,
			&testHelpers.TimerStub{},
			&topologyProviderStub{peerCount: 2, amITheLeader: false},
			&quorumProviderStub{},
			"testMonitor",
		)

		destinationBridge.lock()

		go func() {
			monitor.Start(context.Background())
		}()

		time.Sleep(5 * time.Millisecond)
		assert.Equal(t, expect, destinationBridge.lastWasProposedTransferBatchId)
	})
	t.Run("it will sign proposed transaction if not leader", func(t *testing.T) {
		expect := bridge.NewActionId(42)
		batch := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		destinationBridge := &bridgeStub{wasProposedTransfer: true, proposeTransferActionId: expect}
		monitor := NewMonitor(
			&bridgeStub{pendingBatches: []*bridge.Batch{batch}},
			destinationBridge,
			&testHelpers.TimerStub{},
			&topologyProviderStub{peerCount: 2, amITheLeader: false},
			&quorumProviderStub{quorum: 3},
			"testMonitor",
		)

		destinationBridge.lock()

		go func() {
			monitor.Start(context.Background())
		}()

		// allow propose transfer
		destinationBridge.proposeTransferMutex.Unlock()
		// allow signing
		destinationBridge.signMutex.Unlock()

		time.Sleep(5 * time.Millisecond)
		assert.Equal(t, expect, destinationBridge.lastSignedActionId)
	})
	t.Run("it will try to proposeTransfer again if timeout and it becomes leader", func(t *testing.T) {
		expect := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		destinationBridge := &bridgeStub{wasProposedTransfer: false}
		timer := &testHelpers.TimerStub{}
		provider := &topologyProviderStub{peerCount: 2, amITheLeader: false}
		monitor := NewMonitor(
			&bridgeStub{pendingBatches: []*bridge.Batch{expect}},
			destinationBridge,
			timer,
			provider,
			&quorumProviderStub{},
			"testMonitor",
		)

		destinationBridge.lock()
		go func() {
			monitor.Start(context.Background())
		}()

		// allow propose transfer
		provider.amITheLeader = true
		destinationBridge.proposeTransferMutex.Unlock()

		time.Sleep(5 * time.Millisecond)
		assert.Equal(t, expect, destinationBridge.lastProposedBatch)
	})
}

func TestWaitForSignatures(t *testing.T) {
	testHelpers.SetTestLogLevel()
	t.Run("it will execute transfer when leader and quorum is meet", func(t *testing.T) {
		expect := bridge.NewActionId(42)
		batch := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		destinationBridge := &bridgeStub{signersCount: 4, proposeTransferActionId: expect}
		monitor := NewMonitor(
			&bridgeStub{pendingBatches: []*bridge.Batch{batch}},
			destinationBridge,
			&testHelpers.TimerStub{},
			&topologyProviderStub{peerCount: 10, amITheLeader: true},
			&quorumProviderStub{quorum: 4},
			"testMonitor",
		)

		destinationBridge.lock()

		go func() {
			monitor.Start(context.Background())
		}()

		// allow propose transfer
		destinationBridge.proposeTransferMutex.Unlock()
		// allow signing transfer
		destinationBridge.signMutex.Unlock()
		// allow executing
		destinationBridge.executeMutex.Unlock()

		time.Sleep(5 * time.Millisecond)
		assert.Equal(t, expect, destinationBridge.lastExecutedActionId)
	})
	t.Run("it will clean when signatures after execute", func(t *testing.T) {
		expect := bridge.NewActionId(42)
		batch := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		destinationBridge := &bridgeStub{signersCount: 4, proposeTransferActionId: expect, wasExecuted: true}
		provider := &topologyProviderStub{peerCount: 10, amITheLeader: true}

		monitor := NewMonitor(
			&bridgeStub{pendingBatches: []*bridge.Batch{batch}},
			destinationBridge,
			&testHelpers.TimerStub{},
			provider,
			&quorumProviderStub{quorum: 4},
			"testMonitor",
		)

		destinationBridge.lock()

		go func() {
			monitor.Start(context.Background())
		}()

		// allow propose transfer
		destinationBridge.proposeTransferMutex.Unlock()
		// allow signing transfer
		destinationBridge.signMutex.Unlock()
		// allow executing
		destinationBridge.executeMutex.Unlock()

		time.Sleep(5 * time.Millisecond)
		assert.True(t, provider.cleaned)
	})
	t.Run("it will sleep and try to wait for signatures quorum not achieved", func(t *testing.T) {
		expect := bridge.NewActionId(42)
		batch := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		destinationBridge := &bridgeStub{signersCount: 0, proposeTransferActionId: expect}
		monitor := NewMonitor(
			&bridgeStub{pendingBatches: []*bridge.Batch{batch}},
			destinationBridge,
			&testHelpers.TimerStub{},
			&topologyProviderStub{peerCount: 4, amITheLeader: true},
			&quorumProviderStub{quorum: 3},
			"testMonitor",
		)

		destinationBridge.lock()

		go func() {
			monitor.Start(context.Background())
		}()

		// allow propose transfer
		destinationBridge.proposeTransferMutex.Unlock()
		// allow signing transfer
		destinationBridge.signMutex.Unlock()
		// allow executing
		destinationBridge.signersCount = 3
		destinationBridge.executeMutex.Unlock()

		time.Sleep(5 * time.Millisecond)
		assert.Equal(t, expect, destinationBridge.lastExecutedActionId)
	})
}

func TestExecute(t *testing.T) {
	testHelpers.SetTestLogLevel()
	t.Run("it will not execute if not leader", func(t *testing.T) {
		batch := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		destinationBridge := &bridgeStub{signersCount: 3, wasExecuted: false, wasProposedTransfer: true, proposeTransferActionId: bridge.NewActionId(42)}
		timer := &testHelpers.TimerStub{}
		provider := &topologyProviderStub{peerCount: 4, amITheLeader: false}

		monitor := NewMonitor(
			&bridgeStub{pendingBatches: []*bridge.Batch{batch}},
			destinationBridge,
			timer,
			provider,
			&quorumProviderStub{quorum: 1},
			"testMonitor",
		)

		destinationBridge.lock()

		go func() {
			monitor.Start(context.Background())
		}()

		// allow signing
		destinationBridge.signMutex.Unlock()
		// make executing
		destinationBridge.executeMutex.Unlock()

		time.Sleep(5 * time.Millisecond)
		assert.Nil(t, destinationBridge.lastExecutedActionId)
	})
	t.Run("it will wait for execution when not leader", func(t *testing.T) {
		expect := bridge.NewActionId(42)
		batch := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		destinationBridge := &bridgeStub{signersCount: 3, wasExecuted: false, wasProposedTransfer: true, proposeTransferActionId: expect}
		timer := &testHelpers.TimerStub{}
		provider := &topologyProviderStub{peerCount: 4, amITheLeader: false}

		monitor := NewMonitor(
			&bridgeStub{pendingBatches: []*bridge.Batch{batch}},
			destinationBridge,
			timer,
			provider,
			&quorumProviderStub{quorum: 1},
			"testMonitor",
		)

		destinationBridge.lock()

		go func() {
			monitor.Start(context.Background())
		}()

		// allow signing
		time.Sleep(1 * time.Millisecond)
		destinationBridge.signMutex.Unlock()
		// make leader
		time.Sleep(1 * time.Second)
		provider.amITheLeader = true
		destinationBridge.executeMutex.Unlock()

		time.Sleep(1 * time.Second)
		assert.Equal(t, expect, destinationBridge.lastExecutedActionId)
	})
}

func TestProposeSetStatus(t *testing.T) {
	testHelpers.SetTestLogLevel()
	t.Run("it will propose to set status when leader", func(t *testing.T) {
		destinationBridge := &bridgeStub{
			signersCount:            3,
			wasExecuted:             true,
			wasProposedTransfer:     true,
			proposeTransferActionId: bridge.NewActionId(41),
		}
		provider := &topologyProviderStub{peerCount: 4, amITheLeader: true}
		batch := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		sourceBridge := &bridgeStub{pendingBatches: []*bridge.Batch{batch}}
		monitor := NewMonitor(
			sourceBridge,
			destinationBridge,
			&testHelpers.TimerStub{},
			provider,
			&quorumProviderStub{quorum: 1},
			"testMonitor",
		)

		sourceBridge.lock()

		go func() {
			monitor.Start(context.Background())
		}()

		// allow set status
		sourceBridge.proposeSetStatusMutex.Unlock()

		time.Sleep(5 * time.Millisecond)
		assert.Equal(t, bridge.Executed, sourceBridge.proposedStatusBatch.Transactions[0].Status)
	})
	t.Run("it will sign proposed set status when not leader", func(t *testing.T) {
		expect := bridge.NewActionId(42)
		destinationBridge := &bridgeStub{
			signersCount:            3,
			wasExecuted:             true,
			wasProposedTransfer:     true,
			proposeTransferActionId: bridge.NewActionId(41),
		}
		batch := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		provider := &topologyProviderStub{peerCount: 4, amITheLeader: false}
		sourceBridge := &bridgeStub{
			pendingBatches:           []*bridge.Batch{batch},
			proposedStatusBatch:      batch,
			proposeSetStatusActionId: expect,
		}
		monitor := NewMonitor(
			sourceBridge,
			destinationBridge,
			&testHelpers.TimerStub{},
			provider,
			&quorumProviderStub{quorum: 1},
			"testMonitor",
		)

		sourceBridge.lock()

		go func() {
			monitor.Start(context.Background())
		}()

		// allow set status
		sourceBridge.signMutex.Unlock()

		time.Sleep(5 * time.Millisecond)
		assert.Equal(t, expect, sourceBridge.lastSignedActionId)
	})
	t.Run("it will execute set status when leader and number of signatures > 67%", func(t *testing.T) {
		expect := bridge.NewActionId(42)
		destinationBridge := &bridgeStub{
			signersCount:            3,
			wasExecuted:             true,
			wasProposedTransfer:     true,
			proposeTransferActionId: bridge.NewActionId(41),
		}
		provider := &topologyProviderStub{peerCount: 4, amITheLeader: true}
		batch := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		sourceBridge := &bridgeStub{
			signersCount:             3,
			pendingBatches:           []*bridge.Batch{batch},
			proposedStatusBatch:      batch,
			proposeSetStatusActionId: expect,
		}
		monitor := NewMonitor(
			sourceBridge,
			destinationBridge,
			&testHelpers.TimerStub{},
			provider,
			&quorumProviderStub{quorum: 1},
			"testMonitor",
		)

		sourceBridge.lock()

		go func() {
			monitor.Start(context.Background())
		}()

		// allow set status
		sourceBridge.proposeSetStatusMutex.Unlock()
		// allow signing
		sourceBridge.signMutex.Unlock()
		// allow execute
		sourceBridge.executeMutex.Unlock()

		time.Sleep(5 * time.Millisecond)
		assert.Equal(t, expect, sourceBridge.lastExecutedActionId)
	})
	t.Run("it will execute set status when leader After waiting", func(t *testing.T) {
		expect := bridge.NewActionId(42)
		destinationBridge := &bridgeStub{
			signersCount:            3,
			wasExecuted:             true,
			wasProposedTransfer:     true,
			proposeTransferActionId: bridge.NewActionId(41),
		}
		provider := &topologyProviderStub{peerCount: 4, amITheLeader: false}
		batch := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{{To: "address", DepositNonce: bridge.NewNonce(0)}},
		}
		sourceBridge := &bridgeStub{
			signersCount:             3,
			pendingBatches:           []*bridge.Batch{batch},
			proposedStatusBatch:      batch,
			proposeSetStatusActionId: expect,
		}
		monitor := NewMonitor(
			sourceBridge,
			destinationBridge,
			&testHelpers.TimerStub{},
			provider,
			&quorumProviderStub{quorum: 1},
			"testMonitor",
		)

		sourceBridge.lock()

		go func() {
			monitor.Start(context.Background())
		}()

		// allow set status
		sourceBridge.proposeSetStatusMutex.Unlock()
		// allow signing
		sourceBridge.signMutex.Unlock()
		// allow execute
		provider.amITheLeader = true
		sourceBridge.executeMutex.Unlock()

		time.Sleep(5 * time.Millisecond)
		assert.Equal(t, expect, sourceBridge.lastExecutedActionId)
	})
}

type quorumProviderStub struct {
	quorum uint
}

func (s *quorumProviderStub) GetQuorum(_ context.Context) (uint, error) {
	return s.quorum, nil
}

type topologyProviderStub struct {
	amITheLeader bool
	peerCount    int
	cleaned      bool
}

func (s *topologyProviderStub) AmITheLeader() bool {
	return s.amITheLeader
}

func (s *topologyProviderStub) PeerCount() int {
	return s.peerCount
}

func (s *topologyProviderStub) Clean() {
	s.cleaned = true
}

type bridgeStub struct {
	pendingBatchCallIndex          int
	pendingBatches                 []*bridge.Batch
	wasProposedTransfer            bool
	lastProposedBatch              *bridge.Batch
	lastWasProposedTransferBatchId bridge.BatchId
	lastSignedActionId             bridge.ActionId
	signersCount                   uint
	lastExecutedActionId           bridge.ActionId
	wasExecuted                    bool
	proposeTransferActionId        bridge.ActionId
	proposeTransferError           error
	proposedStatusBatch            *bridge.Batch
	proposeSetStatusActionId       bridge.ActionId

	proposeSetStatusMutex sync.Mutex
	proposeTransferMutex  sync.Mutex
	signMutex             sync.Mutex
	executeMutex          sync.Mutex
}

func (b *bridgeStub) lock() {
	b.proposeSetStatusMutex.Lock()
	b.proposeTransferMutex.Lock()
	b.signMutex.Lock()
	b.executeMutex.Lock()
}

func (b *bridgeStub) GetPending(context.Context) *bridge.Batch {
	defer func() { b.pendingBatchCallIndex++ }()

	if b.pendingBatchCallIndex >= len(b.pendingBatches) {
		return nil
	} else {
		return b.pendingBatches[b.pendingBatchCallIndex]
	}
}

func (b *bridgeStub) ProposeSetStatus(_ context.Context, batch *bridge.Batch) {
	b.proposeSetStatusMutex.Lock()
	b.proposedStatusBatch = batch
}

func (b *bridgeStub) ProposeTransfer(_ context.Context, batch *bridge.Batch) (string, error) {
	b.proposeTransferMutex.Lock()
	b.wasProposedTransfer = true
	b.lastProposedBatch = batch

	return "propose_tx_hash", b.proposeTransferError
}

func (b *bridgeStub) WasProposedTransfer(_ context.Context, batch *bridge.Batch) bool {
	b.lastWasProposedTransferBatchId = batch.Id
	return b.wasProposedTransfer
}

func (b *bridgeStub) GetActionIdForProposeTransfer(context.Context, *bridge.Batch) bridge.ActionId {
	return b.proposeTransferActionId
}

func (b *bridgeStub) WasProposedSetStatus(context.Context, *bridge.Batch) bool {
	return true
}

func (b *bridgeStub) GetActionIdForSetStatusOnPendingTransfer(context.Context, *bridge.Batch) bridge.ActionId {
	return b.proposeSetStatusActionId
}

func (b *bridgeStub) WasExecuted(context.Context, bridge.ActionId, bridge.BatchId) bool {
	return b.wasExecuted
}

func (b *bridgeStub) Sign(_ context.Context, actionId bridge.ActionId) (string, error) {
	b.signMutex.Lock()
	b.lastSignedActionId = actionId

	return "sign_tx_hash", nil
}

func (b *bridgeStub) Execute(_ context.Context, actionId bridge.ActionId, _ *bridge.Batch) (string, error) {
	b.executeMutex.Lock()
	b.lastExecutedActionId = actionId

	return "execution hash", nil
}

func (b *bridgeStub) SignersCount(context.Context, bridge.ActionId) uint {
	return b.signersCount
}
