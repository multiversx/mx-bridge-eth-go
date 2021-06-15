package relay

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

const (
	MinSignaturePercent = 67
)

type State int

const (
	GetPendingTransaction    State = 0
	ProposeTransfer          State = 1
	WaitForSignatures        State = 2
	Execute                  State = 3
	WaitForTransferProposal  State = 4
	WaitForExecute           State = 5
	Stop                     State = 6
	ProposeSetStatus         State = 7
	WaitForSetStatusProposal State = 8
)

type Monitor struct {
	name             string
	topologyProvider TopologyProvider
	timer            Timer
	log              logger.Logger

	sourceBridge      bridge.Bridge
	destinationBridge bridge.Bridge
	executingBridge   bridge.Bridge

	initialState State
	pendingBatch *bridge.Batch
	actionId     bridge.ActionId
}

func NewMonitor(sourceBridge, destinationBridge bridge.Bridge, timer Timer, topologyProvider TopologyProvider, name string) *Monitor {
	return &Monitor{
		name:             name,
		topologyProvider: topologyProvider,
		timer:            timer,
		log:              logger.GetOrCreate(name),

		sourceBridge:      sourceBridge,
		destinationBridge: destinationBridge,
	}
}

func (m *Monitor) Start(ctx context.Context) {
	m.log.Info(fmt.Sprintf("Started monitor %q", m.name))

	ch := make(chan State, 1)
	ch <- m.initialState

	for {
		select {
		case state := <-ch:
			switch state {
			case GetPendingTransaction:
				go m.getPendingTransaction(ctx, ch)
			case ProposeTransfer:
				go m.proposeTransfer(ctx, ch)
			case WaitForTransferProposal:
				go m.waitForTransferProposal(ctx, ch)
			case WaitForSignatures:
				go m.waitForSignatures(ctx, ch)
			case Execute:
				go m.execute(ctx, ch)
			case WaitForExecute:
				go m.waitForExecute(ctx, ch)
			case ProposeSetStatus:
				go m.proposeSetStatus(ctx, ch)
			case WaitForSetStatusProposal:
				go m.waitForSetStatusProposal(ctx, ch)
			case Stop:
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// State

func (m *Monitor) getPendingTransaction(ctx context.Context, ch chan State) {
	m.log.Info("Getting pending transaction")
	m.pendingBatch = m.sourceBridge.GetPending(ctx)

	if m.pendingBatch == nil {
		select {
		case <-m.timer.After(5 * time.Second):
			ch <- GetPendingTransaction
		case <-ctx.Done():
			ch <- Stop
		}
	} else {
		m.topologyProvider.Clean()
		ch <- ProposeTransfer
	}
}

func (m *Monitor) proposeTransfer(ctx context.Context, ch chan State) {
	if m.topologyProvider.AmITheLeader() {
		m.log.Info(fmt.Sprintf("Proposing deposit transaction for nonce %v", m.pendingBatch.Id))
		hash, err := m.destinationBridge.ProposeTransfer(ctx, m.pendingBatch)
		if err != nil {
			m.log.Error(err.Error())
			// TODO: figure this out
			//m.pendingBatch.Status = bridge.Rejected
			//m.pendingBatch.Error = err
			m.executingBridge = m.sourceBridge
			ch <- ProposeSetStatus
		} else {
			m.log.Info(fmt.Sprintf("Proposed with hash %q", hash))
			ch <- WaitForTransferProposal
		}
	} else {
		ch <- WaitForTransferProposal
	}
}

func (m *Monitor) waitForTransferProposal(ctx context.Context, ch chan State) {
	m.log.Info(fmt.Sprintf("Waiting for proposal on transaction with nonce %v", m.pendingBatch.Id))
	select {
	case <-m.timer.After(Timeout):
		// TODO: deal with this
		if m.destinationBridge.WasProposedTransfer(ctx, nil) {
			m.log.Info(fmt.Sprintf("Signing transaction with nonce %v", m.pendingBatch.Id))
			// TODO: deal with this
			m.actionId = m.destinationBridge.GetActionIdForProposeTransfer(ctx, nil)
			hash, err := m.destinationBridge.Sign(ctx, m.actionId)
			if err != nil {
				m.log.Error(err.Error())
			} else {
				m.log.Info(fmt.Sprintf("Singed with hash %q", hash))
			}
			m.executingBridge = m.destinationBridge
			ch <- WaitForSignatures
		} else {
			ch <- ProposeTransfer
		}
	case <-ctx.Done():
		ch <- Stop
	}
}

func (m *Monitor) waitForSignatures(ctx context.Context, ch chan State) {
	m.log.Info(fmt.Sprintf("Waiting for signatures for actionId %v", m.actionId))
	select {
	case <-m.timer.After(Timeout):
		count := m.executingBridge.SignersCount(ctx, m.actionId)
		peerCount := m.topologyProvider.PeerCount()
		minCountRequired := math.Ceil(float64(peerCount) * MinSignaturePercent / 100)

		m.log.Info(fmt.Sprintf("Got %d signatures for actionId %v", count, m.actionId))
		if count >= uint(minCountRequired) && count > 0 {
			ch <- Execute
		} else {
			ch <- WaitForSignatures
		}
	case <-ctx.Done():
		ch <- Stop
	}
}

func (m *Monitor) execute(ctx context.Context, ch chan State) {
	if m.topologyProvider.AmITheLeader() {
		m.log.Info(fmt.Sprintf("Executing actionId %v", m.actionId))
		// Todo: Revisit this
		hash, err := m.executingBridge.Execute(ctx, m.actionId, nil)

		if err != nil {
			m.log.Error(err.Error())
		}

		m.log.Info(fmt.Sprintf("ActionId %v was executed with hash %q", m.actionId, hash))
	}

	ch <- WaitForExecute
}

func (m *Monitor) waitForExecute(ctx context.Context, ch chan State) {
	m.log.Info(fmt.Sprintf("Waiting for execution for actionID %v", m.actionId))
	select {
	case <-m.timer.After(Timeout):
		if m.executingBridge.WasExecuted(ctx, m.actionId, m.pendingBatch.Id) {
			m.log.Info(fmt.Sprintf("ActionId %v was executed", m.actionId))
			// TODO: figure this out
			//m.pendingBatch.Status = bridge.Executed

			switch m.executingBridge {
			case m.destinationBridge:
				ch <- ProposeSetStatus
			case m.sourceBridge:
				ch <- GetPendingTransaction
			}
		} else {
			ch <- Execute
		}
	case <-ctx.Done():
		ch <- Stop
	}
}

func (m *Monitor) proposeSetStatus(ctx context.Context, ch chan State) {
	if m.topologyProvider.AmITheLeader() {
		m.log.Info(fmt.Sprintf("Proposing set status on transaction with nonce %v", m.pendingBatch.Id))
		// TODO: revisit this
		m.sourceBridge.ProposeSetStatus(ctx, nil)
	}
	ch <- WaitForSetStatusProposal
}

func (m *Monitor) waitForSetStatusProposal(ctx context.Context, ch chan State) {
	m.log.Info(fmt.Sprintf("Waiting for set status proposal on transaction with nonce %v", m.pendingBatch.Id))
	select {
	case <-m.timer.After(Timeout):
		// TODO: figure this out
		if m.sourceBridge.WasProposedSetStatus(ctx, nil) {
			m.log.Info(fmt.Sprintf("Signing set status for transaction with nonce %v", m.pendingBatch.Id))
			m.actionId = m.sourceBridge.GetActionIdForSetStatusOnPendingTransfer(ctx)
			hash, err := m.sourceBridge.Sign(ctx, m.actionId)
			if err != nil {
				m.log.Error(err.Error())
			}
			m.log.Info(fmt.Sprintf("Singed set status for batch with id %v with hash %q", m.pendingBatch.Id, hash))
			m.executingBridge = m.sourceBridge
			ch <- WaitForSignatures
		} else {
			ch <- ProposeSetStatus
		}
	case <-ctx.Done():
		ch <- Stop
	}
}
