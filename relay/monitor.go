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

	initialState       State
	pendingTransaction *bridge.DepositTransaction
	actionId           bridge.ActionId
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
	m.pendingTransaction = m.sourceBridge.GetPendingDepositTransaction(ctx)

	if m.pendingTransaction == nil {
		select {
		case <-m.timer.after((Timeout / 10) * time.Second):
			ch <- GetPendingTransaction
		case <-ctx.Done():
			ch <- Stop
		}
	} else {
		ch <- ProposeTransfer
	}
}

func (m *Monitor) proposeTransfer(ctx context.Context, ch chan State) {
	if m.topologyProvider.AmITheLeader() {
		m.log.Info(fmt.Sprintf("Proposing deposit transaction with nonce %d", m.pendingTransaction.DepositNonce))
		m.destinationBridge.ProposeTransfer(ctx, m.pendingTransaction)
	}
	ch <- WaitForTransferProposal
}

func (m *Monitor) waitForTransferProposal(ctx context.Context, ch chan State) {
	m.log.Info(fmt.Sprintf("Waiting for proposal on transaction with nonce %d", m.pendingTransaction.DepositNonce))
	select {
	case <-m.timer.after(Timeout):
		if m.destinationBridge.WasProposedTransfer(ctx, m.pendingTransaction.DepositNonce) {
			m.log.Info(fmt.Sprintf("Signing transaction with nonce %d", m.pendingTransaction.DepositNonce))
			m.actionId = m.destinationBridge.GetActionIdForProposeTransfer(ctx, m.pendingTransaction.DepositNonce)
			m.destinationBridge.Sign(ctx, m.actionId)
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
	m.log.Info(fmt.Sprintf("Waiting for signatures for actionId %d", m.actionId))
	select {
	case <-m.timer.after(Timeout):
		count := m.executingBridge.SignersCount(ctx, m.actionId)
		peerCount := m.topologyProvider.PeerCount()
		minCountRequired := math.Ceil(float64(peerCount) * MinSignaturePercent / 100)

		m.log.Info(fmt.Sprintf("Got %d signatures for actionId %d", count, m.actionId))
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
		m.log.Info(fmt.Sprintf("Executing actionId %d", m.actionId))
		hash, err := m.executingBridge.Execute(ctx, m.actionId)

		if err != nil {
			m.log.Error(err.Error())
		}

		m.log.Info(fmt.Sprintf("ActionId %d was executed with hash %q", m.actionId, hash))
	}

	ch <- WaitForExecute
}

func (m *Monitor) waitForExecute(ctx context.Context, ch chan State) {
	m.log.Info(fmt.Sprintf("Waiting for execution for actionID %d", m.actionId))
	select {
	case <-m.timer.after(Timeout):
		if m.executingBridge.WasExecuted(ctx, m.actionId) {
			m.log.Info(fmt.Sprintf("ActionId %d was executed", m.actionId))

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
		m.log.Info(fmt.Sprintf("Proposing set status on transaction with nonce %d", m.pendingTransaction.DepositNonce))
		m.sourceBridge.ProposeSetStatusSuccessOnPendingTransfer(ctx)
	}
	ch <- WaitForSetStatusProposal
}

func (m *Monitor) waitForSetStatusProposal(ctx context.Context, ch chan State) {
	m.log.Info(fmt.Sprintf("Waiting for set status proposal on transaction with nonce %d", m.pendingTransaction.DepositNonce))
	select {
	case <-m.timer.after(Timeout):
		if m.sourceBridge.WasProposedSetStatusSuccessOnPendingTransfer(ctx) {
			m.log.Info(fmt.Sprintf("Signing set status for transaction with nonce %d", m.pendingTransaction.DepositNonce))
			m.actionId = m.sourceBridge.GetActionIdForSetStatusOnPendingTransfer(ctx)
			m.sourceBridge.Sign(ctx, m.actionId)
			m.executingBridge = m.sourceBridge
			ch <- WaitForSignatures
		} else {
			ch <- ProposeSetStatus
		}
	case <-ctx.Done():
		ch <- Stop
	}
}
