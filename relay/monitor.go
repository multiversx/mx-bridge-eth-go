package relay

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

type state int

const (
	getPending               state = 0
	proposeTransfer          state = 1
	waitForSignatures        state = 2
	execute                  state = 3
	waitForTransferProposal  state = 4
	waitForExecute           state = 5
	stop                     state = 6
	proposeSetStatus         state = 7
	waitForSetStatusProposal state = 8
)

type Monitor struct {
	name             string
	topologyProvider TopologyProvider
	quorumProvider   bridge.QuorumProvider
	timer            Timer
	log              logger.Logger

	sourceBridge      bridge.Bridge
	destinationBridge bridge.Bridge
	executingBridge   bridge.Bridge

	initialState state
	pendingBatch *bridge.Batch
	actionId     bridge.ActionId
}

func NewMonitor(sourceBridge, destinationBridge bridge.Bridge, timer Timer, topologyProvider TopologyProvider, quorumProvider bridge.QuorumProvider, name string) *Monitor {
	return &Monitor{
		name:             name,
		topologyProvider: topologyProvider,
		quorumProvider:   quorumProvider,
		timer:            timer,
		log:              logger.GetOrCreate(name),

		sourceBridge:      sourceBridge,
		destinationBridge: destinationBridge,
	}
}

func (m *Monitor) Start(ctx context.Context) {
	m.log.Info(fmt.Sprintf("Started monitor %q", m.name))

	ch := make(chan state, 1)
	ch <- m.initialState

	for {
		select {
		case stateValue := <-ch:
			switch stateValue {
			case getPending:
				go m.getPending(ctx, ch)
			case proposeTransfer:
				go m.proposeTransfer(ctx, ch)
			case waitForTransferProposal:
				go m.waitForTransferProposal(ctx, ch)
			case waitForSignatures:
				go m.waitForSignatures(ctx, ch)
			case execute:
				go m.execute(ctx, ch)
			case waitForExecute:
				go m.waitForExecute(ctx, ch)
			case proposeSetStatus:
				go m.proposeSetStatus(ctx, ch)
			case waitForSetStatusProposal:
				go m.waitForSetStatusProposal(ctx, ch)
			case stop:
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// State

func (m *Monitor) getPending(ctx context.Context, ch chan state) {
	m.pendingBatch = m.sourceBridge.GetPending(ctx)

	if m.pendingBatch == nil {
		select {
		case <-m.timer.After(5 * time.Second):
			ch <- getPending
		case <-ctx.Done():
			ch <- stop
		}
	} else {
		ch <- proposeTransfer
	}
}

func (m *Monitor) proposeTransfer(ctx context.Context, ch chan state) {
	if m.topologyProvider.AmITheLeader() {
		_, err := m.destinationBridge.ProposeTransfer(ctx, m.pendingBatch)
		if err != nil {
			m.log.Error(err.Error())
			m.pendingBatch.SetStatusOnAllTransactions(bridge.Rejected, err)
			m.executingBridge = m.sourceBridge
			ch <- proposeSetStatus
		} else {
			ch <- waitForTransferProposal
		}
	} else {
		ch <- waitForTransferProposal
	}
}

func (m *Monitor) waitForTransferProposal(ctx context.Context, ch chan state) {
	m.log.Info(fmt.Sprintf("Waiting for transfer proposal on batch with nonce %v", m.pendingBatch.Id))
	select {
	case <-m.timer.After(timeout):
		if m.destinationBridge.WasProposedTransfer(ctx, m.pendingBatch) {
			m.actionId = m.destinationBridge.GetActionIdForProposeTransfer(ctx, m.pendingBatch)
			_, err := m.destinationBridge.Sign(ctx, m.actionId)
			if err != nil {
				m.log.Error(err.Error())
			}
			m.executingBridge = m.destinationBridge
			ch <- waitForSignatures
		} else {
			ch <- proposeTransfer
		}
	case <-ctx.Done():
		ch <- stop
	}
}

func (m *Monitor) waitForSignatures(ctx context.Context, ch chan state) {
	m.log.Info("Waiting for signatures")
	select {
	case <-m.timer.After(timeout):
		count := big.NewInt(int64(m.executingBridge.SignersCount(ctx, m.actionId)))
		quorum, err := m.quorumProvider.GetQuorum(ctx)
		if err != nil {
			m.log.Error(err.Error())
			ch <- waitForSignatures
		}

		m.log.Info(fmt.Sprintf("Got %d signatures, the quorum is %d", count, quorum))
		if m.wasQuorumReached(quorum, count) {
			ch <- execute
		} else {
			if m.wasExecuted(ctx) {
				m.executed(ctx, ch)
			} else {
				ch <- waitForSignatures
			}
		}
	case <-ctx.Done():
		ch <- stop
	}
}

func (m *Monitor) execute(ctx context.Context, ch chan state) {
	if m.topologyProvider.AmITheLeader() {
		_, err := m.executingBridge.Execute(ctx, m.actionId, m.pendingBatch)

		if err != nil {
			m.log.Error(err.Error())
		}
	}

	ch <- waitForExecute
}

func (m *Monitor) waitForExecute(ctx context.Context, ch chan state) {
	m.log.Info("Waiting for execution")
	select {
	case <-m.timer.After(timeout):
		if m.wasExecuted(ctx) {
			m.executed(ctx, ch)
		} else {
			ch <- execute
		}
	case <-ctx.Done():
		ch <- stop
	}
}

func (m *Monitor) proposeSetStatus(ctx context.Context, ch chan state) {
	if m.topologyProvider.AmITheLeader() {
		m.sourceBridge.ProposeSetStatus(ctx, m.pendingBatch)
	}
	ch <- waitForSetStatusProposal
}

func (m *Monitor) waitForSetStatusProposal(ctx context.Context, ch chan state) {
	m.log.Info(fmt.Sprintf("Waiting for set status proposal on batch with nonce %v", m.pendingBatch.Id))
	select {
	case <-m.timer.After(timeout):
		if m.sourceBridge.WasProposedSetStatus(ctx, m.pendingBatch) {
			m.log.Info(fmt.Sprintf("Signing set status for batch with id %v", m.pendingBatch.Id))
			m.actionId = m.sourceBridge.GetActionIdForSetStatusOnPendingTransfer(ctx, m.pendingBatch)
			_, err := m.sourceBridge.Sign(ctx, m.actionId)
			if err != nil {
				m.log.Error(err.Error())
			}
			m.executingBridge = m.sourceBridge
			ch <- waitForSignatures
		} else {
			ch <- proposeSetStatus
		}
	case <-ctx.Done():
		ch <- stop
	}
}

// helpers

func (m *Monitor) wasQuorumReached(quorum *big.Int, count *big.Int) bool {
	return quorum.Cmp(count) <= 0
}

func (m *Monitor) wasExecuted(ctx context.Context) bool {
	return m.executingBridge.WasExecuted(ctx, m.actionId, m.pendingBatch.Id)
}

func (m *Monitor) executed(_ context.Context, ch chan state) {
	m.topologyProvider.Clean()
	m.pendingBatch.SetStatusOnAllTransactions(bridge.Executed, nil)

	switch m.executingBridge {
	case m.destinationBridge:
		ch <- proposeSetStatus
	case m.sourceBridge:
		ch <- getPending
	}
}
