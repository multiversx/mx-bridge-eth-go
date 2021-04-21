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
	GetPendingTransaction State = 0
	Propose               State = 1
	WaitForSignatures     State = 2
	Execute               State = 3
	WaitForProposal       State = 4
	WaitForExecute        State = 5
	Stop                  State = 6
)

type Monitor struct {
	name             string
	topologyProvider TopologyProvider
	timer            Timer
	log              logger.Logger

	sourceBridge      bridge.Bridge
	destinationBridge bridge.Bridge

	initialState       State
	pendingTransaction *bridge.DepositTransaction
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
			case Propose:
				go m.propose(ctx, ch)
			case WaitForProposal:
				go m.waitForProposal(ctx, ch)
			case WaitForSignatures:
				go m.waitForSignatures(ctx, ch)
			case Execute:
				go m.execute(ctx, ch)
			case WaitForExecute:
				go m.waitForExecute(ctx, ch)
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
		ch <- Propose
	}
}

func (m *Monitor) propose(ctx context.Context, ch chan State) {
	if m.topologyProvider.AmITheLeader() {
		m.log.Info(fmt.Sprintf("Proposing deposit transaction with nonce %d", m.pendingTransaction.DepositNonce))
		m.destinationBridge.Propose(ctx, m.pendingTransaction)
		ch <- WaitForSignatures
	} else {
		ch <- WaitForProposal
	}
}

func (m *Monitor) waitForProposal(ctx context.Context, ch chan State) {
	m.log.Info(fmt.Sprintf("Waiting for proposal on transaction with nonce %d", m.pendingTransaction.DepositNonce))
	select {
	case <-m.timer.after(Timeout):
		if m.destinationBridge.WasProposed(ctx, m.pendingTransaction) {
			m.log.Info(fmt.Sprintf("Signing transaction with nonce %d", m.pendingTransaction.DepositNonce))
			m.destinationBridge.Sign(ctx, m.pendingTransaction)
			ch <- WaitForSignatures
		} else {
			ch <- Propose
		}
	case <-ctx.Done():
		ch <- Stop
	}
}

func (m *Monitor) waitForSignatures(ctx context.Context, ch chan State) {
	m.log.Info(fmt.Sprintf("Waiting for signatures on transaction with nonce %d", m.pendingTransaction.DepositNonce))
	select {
	case <-m.timer.after(Timeout):
		count := m.destinationBridge.SignersCount(ctx, m.pendingTransaction)
		peerCount := m.topologyProvider.PeerCount()
		minCountRequired := math.Ceil(float64(peerCount) * MinSignaturePercent / 100)

		m.log.Info(fmt.Sprintf("Got %d signatures for transaction with nonce %d", count, m.pendingTransaction.DepositNonce))
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
		m.log.Info(fmt.Sprintf("Executing transaction with nonce %d", m.pendingTransaction.DepositNonce))
		hash, err := m.destinationBridge.Execute(ctx, m.pendingTransaction)

		if err != nil {
			m.log.Error(err.Error())
		}

		m.log.Info(fmt.Sprintf("Bridge transaction executed with hash %q", hash))
	}

	ch <- WaitForExecute
}

func (m *Monitor) waitForExecute(ctx context.Context, ch chan State) {
	m.log.Info(fmt.Sprintf("Waiting for execution for transaction with nonce %d", m.pendingTransaction.DepositNonce))
	select {
	case <-m.timer.after(Timeout):
		if m.destinationBridge.WasExecuted(ctx, m.pendingTransaction) {
			m.log.Info(fmt.Sprintf("Transaction with nonce %d was executed", m.pendingTransaction.DepositNonce))
			ch <- GetPendingTransaction
		} else {
			ch <- Execute
		}
	case <-ctx.Done():
		ch <- Stop
	}
}
